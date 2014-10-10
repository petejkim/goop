package goop

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"code.google.com/p/go.tools/go/vcs"

	"github.com/nitrous-io/goop/colors"
	"github.com/nitrous-io/goop/parser"
	"github.com/nitrous-io/goop/pkg/env"
)

type UnsupportedVCSError struct {
	VCS string
}

func (e *UnsupportedVCSError) Error() string {
	return fmt.Sprintf("%s is not supported.", e.VCS)
}

type Goop struct {
	dir    string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewGoop(dir string, stdin io.Reader, stdout io.Writer, stderr io.Writer) *Goop {
	return &Goop{dir: dir, stdin: stdin, stdout: stdout, stderr: stderr}
}

func (g *Goop) patchedEnv(replaceGopath bool) env.Env {
	e := env.NewEnv()

	binPath := path.Join(g.vendorDir(), "bin")

	if replaceGopath {
		e["GOPATH"] = g.vendorDir()
	} else {
		e.Prepend("GOPATH", g.vendorDir())
	}
	e["GOBIN"] = binPath
	e.Prepend("PATH", binPath)

	return e
}

func (g *Goop) PrintEnv() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		g.stdout.Write([]byte(fmt.Sprintf("GOPATH=%s\n", g.vendorDir())))
	} else {
		g.stdout.Write([]byte(fmt.Sprintf("GOPATH=%s:%s\n", g.vendorDir(), gopath)))
	}
	g.stdout.Write([]byte(fmt.Sprintf("PATH=%s:%s\n", path.Join(g.vendorDir(), "bin"), os.Getenv("PATH"))))
}

func (g *Goop) Exec(name string, args ...string) error {
	vname := path.Join(g.vendorDir(), "bin", name)
	_, err := os.Stat(vname)
	if err == nil {
		name = vname
	}
	cmd := exec.Command(name, args...)
	cmd.Env = g.patchedEnv(false).Strings()
	cmd.Stdin = g.stdin
	cmd.Stdout = g.stdout
	cmd.Stderr = g.stderr
	return cmd.Run()
}

func (g *Goop) Install() error {
	writeLockFile := false
	f, err := os.Open(path.Join(g.dir, "Goopfile.lock"))
	if err == nil {
		g.stdout.Write([]byte(colors.OK + "Using Goopfile.lock..." + colors.Reset + "\n"))
	} else {
		f, err = os.Open(path.Join(g.dir, "Goopfile"))
		if err != nil {
			return err
		}
		writeLockFile = true
	}
	return g.parseAndInstall(f, writeLockFile)
}

func (g *Goop) Update() error {
	f, err := os.Open(path.Join(g.dir, "Goopfile"))
	if err != nil {
		return err
	}
	return g.parseAndInstall(f, true)
}


func (g *Goop) GenerateGoopFile() error {
	goopFileName := path.Join(g.dir, "Goopfile")
	if exists, err := pathExists(goopFileName); exists || err != nil {
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Goopfile %s already exists", goopFileName)
		}
	}
	g.stdout.Write([]byte(colors.OK + "=> Generating Goopfile..." + colors.Reset + "\n"))
	gf, err := os.Create(goopFileName)
	if err != nil {
		return err
	}
	defer gf.Close()

	imports := make(map[string]struct{})
	packages := make(map[string]struct{})
	// find packages in sub directories and create list of imports
	err = filepath.Walk(g.dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			} else {
				if pkg, err := build.Default.ImportDir(path, 0); err == nil {
					packages[pkg.ImportPath] = struct{}{}
					for _, name := range pkg.Imports {
						if importPkg, err := build.Default.Import(name, g.dir, 0); err == nil {
							if importPkg.Goroot {
								continue
							}
						}
						if _, err := vcs.RepoRootForImportPath(name, false); err == nil {
							imports[name] = struct{}{}
						}
					}
				}
			}
		}
		return err
	})
	if err != nil {
		return err
	}

	// remove any imports that are internal
	for k := range packages {
		if _, ok := imports[k]; ok {
			delete(imports, k)
		}
	}
	importList := make([]string, 0, len(imports))
	for k := range imports {
		importList = append(importList, k)
	}
	sort.Strings(importList)
	for _, importName := range importList {
		gf.WriteString(importName + "\n")
	}
	return nil
}

func (g *Goop) parseAndInstall(goopfile *os.File, writeLockFile bool) error {
	defer goopfile.Close()

	deps, err := parser.Parse(goopfile)
	if err != nil {
		return err
	}

	srcPath := path.Join(g.vendorDir(), "src")
	tmpGoPath := path.Join(g.vendorDir(), "tmp")
	tmpSrcPath := path.Join(tmpGoPath, "src")

	err = os.RemoveAll(tmpGoPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(tmpSrcPath, 0775)
	if err != nil {
		return err
	}

	repos := map[string]*vcs.RepoRoot{}
	lockedDeps := map[string]*parser.Dependency{}

	for _, dep := range deps {
		if dep.URL == "" {
			g.stdout.Write([]byte(colors.OK + "=> Fetching " + dep.Pkg + "..." + colors.Reset + "\n"))
		} else {
			g.stdout.Write([]byte(colors.OK + "=> Fetching " + dep.Pkg + " from " + dep.URL + "..." + colors.Reset + "\n"))
		}

		repo, err := repoForDep(dep)
		if err != nil {
			return err
		}
		repos[dep.Pkg] = repo

		pkgPath := path.Join(srcPath, repo.Root)
		tmpPkgPath := path.Join(tmpSrcPath, repo.Root)

		err = os.MkdirAll(path.Join(tmpPkgPath, ".."), 0775)
		if err != nil {
			return err
		}

		noclone := false

		exists, err := pathExists(pkgPath)
		if err != nil {
			return err
		}
		tmpExists, err := pathExists(tmpPkgPath)
		if err != nil {
			return err
		}
		if exists {
			// if package already exists, just symlink package dir and skip cloning
			g.stderr.Write([]byte(colors.Warn + "Warning: " + pkgPath + " already exists; skipping!" + colors.Reset + "\n"))
			if !tmpExists {
				err = os.Symlink(pkgPath, tmpPkgPath)
				if err != nil {
					return err
				}
			}
			noclone = true
		} else {
			noclone = tmpExists
		}

		if !noclone {
			// clone repo
			err = g.clone(repo.VCS.Cmd, repo.Repo, tmpPkgPath)
			if err != nil {
				return err
			}
		}

		// if rev is not given, record current rev in path
		if dep.Rev == "" {
			rev, err := g.currentRev(repo.VCS.Cmd, tmpPkgPath)
			if err != nil {
				return err
			}
			dep.Rev = rev
		}
		lockedDeps[dep.Pkg] = dep

		// checkout specified rev
		err = g.checkout(repo.VCS.Cmd, tmpPkgPath, dep.Rev)
		if err != nil {
			return err
		}
	}

	for _, dep := range deps {
		g.stdout.Write([]byte(colors.OK + "=> Fetching dependencies for " + dep.Pkg + "..." + colors.Reset + "\n"))

		repo := repos[dep.Pkg]
		tmpPkgPath := path.Join(tmpSrcPath, repo.Root)

		// fetch sub-dependencies
		subdeps, err := g.goGet(tmpPkgPath, tmpGoPath)
		if err != nil {
			return err
		}

		for _, subdep := range subdeps {
			subdepRepo, err := vcs.RepoRootForImportPath(subdep, true)
			if err != nil {
				return err
			}

			subdepPkgPath := path.Join(tmpSrcPath, subdepRepo.Root)

			rev, err := g.currentRev(subdepRepo.VCS.Cmd, subdepPkgPath)
			if err != nil {
				return err
			}

			err = g.checkout(subdepRepo.VCS.Cmd, subdepPkgPath, rev)
			if err != nil {
				return err
			}

			repos[subdep] = subdepRepo
			lockedDeps[subdep] = &parser.Dependency{Pkg: subdep, Rev: rev}
		}
	}

	for _, dep := range lockedDeps {
		g.stdout.Write([]byte(colors.OK + "=> Installing " + dep.Pkg + "..." + colors.Reset + "\n"))

		repo := repos[dep.Pkg]
		pkgPath := path.Join(srcPath, repo.Root)
		tmpPkgPath := path.Join(tmpSrcPath, repo.Root)

		err = os.MkdirAll(path.Join(pkgPath, ".."), 0775)
		if err != nil {
			return err
		}

		lfi, err := os.Lstat(tmpPkgPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			if lfi.Mode()&os.ModeSymlink == 0 {
				// move package to vendor path
				err = os.RemoveAll(pkgPath)
				if err != nil {
					return err
				}
				err = os.Rename(tmpPkgPath, pkgPath)
			} else {
				// package already in vendor path, just remove the symlink
				err = os.Remove(tmpPkgPath)
			}
			if err != nil {
				return err
			}
		}
	}

	for _, dep := range lockedDeps {
		// install
		repo := repos[dep.Pkg]
		pkgPath := path.Join(srcPath, repo.Root)
		cmd := g.command(pkgPath, "go", "install", "-x", dep.Pkg)
		cmd.Env = g.patchedEnv(true).Strings()
		cmd.Run()
	}

	err = os.RemoveAll(tmpGoPath)
	if err != nil {
		return err
	}

	// in order to minimize diffs, we sort lockedDeps first and write the
	// sorted results
	if writeLockFile {
		lf, err := os.Create(path.Join(g.dir, "Goopfile.lock"))
		defer lf.Close()

		var keys []string
		for k := range lockedDeps {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			dep := lockedDeps[k]
			_, err = lf.WriteString(dep.String() + "\n")
			if err != nil {
				return err
			}
		}
	}

	g.stdout.Write([]byte(colors.OK + "=> Done!" + colors.Reset + "\n"))

	return nil
}

func (g *Goop) vendorDir() string {
	return path.Join(g.dir, ".vendor")
}

func (g *Goop) currentRev(vcsCmd string, path string) (string, error) {
	switch vcsCmd {
	case "git":
		cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
		cmd.Dir = path
		cmd.Stderr = g.stderr
		rev, err := cmd.Output()
		if err != nil {
			return "", err
		} else {
			return strings.TrimSpace(string(rev)), err
		}
	case "hg":
		cmd := exec.Command("hg", "log", "-r", ".", "--template", "{node}")
		cmd.Dir = path
		cmd.Stderr = g.stderr
		rev, err := cmd.Output()
		if err != nil {
			return "", err
		} else {
			return strings.TrimSpace(string(rev)), err
		}
	}
	return "", &UnsupportedVCSError{VCS: vcsCmd}
}

func (g *Goop) clone(vcsCmd string, url string, clonePath string) error {
	switch vcsCmd {
	case "git":
		return g.command("", "git", "clone", url, clonePath).Run()
	case "hg":
		return g.command("", "hg", "clone", url, clonePath).Run()
	}
	return &UnsupportedVCSError{VCS: vcsCmd}
}

func (g *Goop) checkout(vcsCmd string, path string, tag string) error {
	g.stdout.Write([]byte("Checking out \"" + tag + "\"\n"))
	switch vcsCmd {
	case "git":
		err := g.command(path, "git", "fetch").Run()
		if err != nil {
			return err
		}
		return g.quietCommand(path, "git", "checkout", tag).Run()
	case "hg":
		err := g.command(path, "hg", "pull").Run()
		if err != nil {
			return err
		}
		return g.quietCommand(path, "hg", "update", tag).Run()
	}
	return &UnsupportedVCSError{VCS: vcsCmd}
}

func (g *Goop) command(path string, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = path
	cmd.Stdin = g.stdin
	cmd.Stdout = g.stdout
	cmd.Stderr = g.stderr
	return cmd
}

func (g *Goop) quietCommand(path string, name string, args ...string) *exec.Cmd {
	cmd := g.command(path, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}

func repoForDep(dep *parser.Dependency) (*vcs.RepoRoot, error) {
	if dep.URL != "" {
		return RepoRootForImportPathWithURLOverride(dep.Pkg, dep.URL)
	}
	return vcs.RepoRootForImportPath(dep.Pkg, true)
}

// pathExists returns:
// * (true, nil) if path exists
// * (false, nil) if path does not exist
// * (false, err) if error happened during stat
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err != nil && !os.IsNotExist(err): // unexpected err
		return false, err
	case err != nil && os.IsNotExist(err):
		return false, nil
	case err == nil:
		return true, nil
	default:
		panic("never reached")
	}
}

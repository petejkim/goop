package goop

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"code.google.com/p/go.tools/go/vcs"

	"github.com/nitrous-io/goop/colors"
	"github.com/nitrous-io/goop/parser"
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

func (g *Goop) patchedEnv(replace bool) []string {
	// Later versions of go require GOPATH to be defined; if it's not defined,
	// default it to the vendor dir
	if os.Getenv("GOPATH") == "" {
		os.Setenv("GOPATH", g.vendorDir())
	}

	sysEnv := os.Environ()
	env := make([]string, len(sysEnv))
	copy(env, sysEnv)
	gopathPatched, pathPatched := false, false

	for i, e := range env {
		if !gopathPatched && strings.HasPrefix(e, "GOPATH=") {
			if replace {
				env[i] = fmt.Sprintf("GOPATH=%s", g.vendorDir())
			} else {
				env[i] = fmt.Sprintf("GOPATH=%s:%s", g.vendorDir(), e[len("GOPATH="):])
			}
			gopathPatched = true
		} else if !pathPatched && strings.HasPrefix(e, "PATH=") {
			env[i] = fmt.Sprintf("PATH=%s:%s", path.Join(g.vendorDir(), "bin"), e[len("PATH="):])
			pathPatched = true
		}
		if gopathPatched && pathPatched {
			break
		}
	}
	return env
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
	cmd.Env = g.patchedEnv(false)
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

func (g *Goop) parseAndInstall(goopfile *os.File, writeLockFile bool) error {
	defer goopfile.Close()

	deps, err := parser.Parse(goopfile)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		g.stdout.Write([]byte(colors.OK + "=> Installing " + dep.Pkg + "..." + colors.Reset + "\n"))
		cmd := exec.Command("go", "get", "-v", dep.Pkg)
		cmd.Env = g.patchedEnv(true)
		cmd.Stdin = g.stdin
		cmd.Stdout = g.stdout
		cmd.Stderr = g.stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

		repo, err := vcs.RepoRootForImportPath(dep.Pkg, true)
		if err != nil {
			return err
		}

		pkgPath := path.Join(g.vendorDir(), "src", repo.Root)

		if dep.Rev == "" {
			rev, err := g.currentRev(repo.VCS.Cmd, pkgPath)
			if err != nil {
				return err
			}
			dep.Rev = rev
			continue
		}

		err = g.checkout(repo.VCS.Cmd, pkgPath, dep.Rev)
		if err != nil {
			return err
		}
	}

	if writeLockFile {
		lf, err := os.Create(path.Join(g.dir, "Goopfile.lock"))
		defer lf.Close()

		for _, dep := range deps {
			_, err = lf.WriteString(fmt.Sprintf("%s #%s\n", dep.Pkg, dep.Rev))
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

func (g *Goop) checkout(vcsCmd string, path string, tag string) error {
	switch vcsCmd {
	case "git":
		err := g.execInPath(path, "git", "fetch")
		if err != nil {
			return err
		}
		return g.execInPath(path, "git", "checkout", tag)
	case "hg":
		err := g.execInPath(path, "hg", "pull")
		if err != nil {
			return err
		}
		return g.execInPath(path, "hg", "update", tag)
	}
	return &UnsupportedVCSError{VCS: vcsCmd}
}

func (g *Goop) execInPath(path string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = path
	cmd.Stdin = g.stdin
	cmd.Stdout = g.stdout
	cmd.Stderr = g.stderr
	return cmd.Run()
}

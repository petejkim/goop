package goop

import (
	"io"
	"os/exec"
	"regexp"
)

var goGetDownloadRe = regexp.MustCompile(`(?m)^(\S+)\s+\(download\)$`)

type DownloadRecorder struct {
	downloads map[string]struct{}
	writer    io.Writer
}

func NewDownloadRecorder(writer io.Writer) *DownloadRecorder {
	return &DownloadRecorder{downloads: map[string]struct{}{}, writer: writer}
}

func (d *DownloadRecorder) Write(p []byte) (n int, err error) {
	s := string(p)
	matches := goGetDownloadRe.FindAllStringSubmatch(s, -1)
	if matches != nil {
		for _, m := range matches {
			d.downloads[m[1]] = struct{}{}
		}
	}
	return d.writer.Write(p)
}

func (d *DownloadRecorder) Downloads() []string {
	s := make([]string, 0, len(d.downloads))
	for k, _ := range d.downloads {
		s = append(s, k)
	}
	return s
}

func (g *Goop) goGet(pkgpath string, gopath string) ([]string, error) {
	cmd := exec.Command("go", "get", "-d", "-v", "./...")
	env := g.patchedEnv(true)
	env["GOPATH"] = gopath
	cmd.Dir = pkgpath
	cmd.Env = env.Strings()
	cmd.Stdin = g.stdin
	cmd.Stdout = g.stdout
	dlRec := NewDownloadRecorder(g.stderr)
	cmd.Stderr = dlRec
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return dlRec.Downloads(), nil
}

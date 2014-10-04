package goop

import (
	"fmt"
	"os/exec"
	"strings"

	"code.google.com/p/go.tools/go/vcs"
)

func GuessVCS(url string) string {
	switch {
	case strings.HasPrefix(url, "https://github.com"):
		return "git"
	case strings.HasPrefix(url, "git://"):
		return "git"
	case strings.HasPrefix(url, "git+ssh://"):
		return "git"
	case strings.HasPrefix(url, "git@"):
		return "git"
	case strings.HasSuffix(url, ".git"):
		return "git"
	case strings.HasPrefix(url, "ssh://hg@"):
		return "hg"
	case strings.HasSuffix(url, ".hg"):
		return "hg"
	default:
		return ""
	}
}

func IdentifyVCS(url string) string {
	v := map[string][]string{
		"git": []string{"git", "ls-remote"},
		"hg":  []string{"hg", "identify"},
	}
	tryVCS := func(vcs string) bool {
		cmd := v[vcs]
		delete(v, vcs)
		return exec.Command(cmd[0], append(cmd[1:], url)...).Run() == nil // use vcs.VCS.Ping?
	}
	guess := GuessVCS(url)
	if guess != "" && v[guess] != nil {
		if tryVCS(guess) {
			return guess
		}
	}
	for k, _ := range v {
		if tryVCS(k) {
			return k
		}
	}
	return ""
}

func RepoRootForImportPathWithURLOverride(importPath string, url string) (*vcs.RepoRoot, error) {
	repo, err := vcs.RepoRootForImportPathStatic(importPath, "ignore")
	if err != nil {
		return nil, err
	}
	repo.Repo = url

	vcs_guess := GuessVCS(url)
	if vcs_guess != "" {
		repo.VCS = vcs.ByCmd(vcs_guess)
		if repo.VCS == nil {
			return nil, fmt.Errorf("Unrecognized VCS %s in %s ", vcs_guess, url)
		}
	}

	return repo, nil
}

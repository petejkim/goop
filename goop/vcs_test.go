package goop_test

import (
	"github.com/nitrous-io/goop/goop"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("vcs", func() {
	tests := []struct {
		importPath string
		url        string
		guess      string
		actual     string
		repoRoot   string
	}{
		{"github.com/dotcloud/docker/pkg/term", "git://github.com/dotcloud/docker.git", "git", "git", "github.com/dotcloud/docker"},
		{"github.com/mattn/go-sqlite3", "git://github.com/mattn/go-sqlite3.git", "git", "git", "github.com/mattn/go-sqlite3"},
		{"github.com/mattn/go-sqlite3", "https://github.com/mattn/go-sqlite3.git", "git", "git", "github.com/mattn/go-sqlite3"},
		{"github.com/mattn/go-sqlite3", "git+ssh://git@github.com/mattn/go-sqlite3.git", "git", "git", "github.com/mattn/go-sqlite3"},
		{"github.com/mattn/go-sqlite3", "git@github.com:mattn/go-sqlite3.git", "git", "git", "github.com/mattn/go-sqlite3"},
		{"github.com/nitrous-io/no-such", "git@github.com/nitrous-io/no-such", "git", "", "github.com/nitrous-io/no-such"},
		{"bitbucket.org/kardianos/osext", "ssh://hg@bitbucket.org/kardianos/osext", "hg", "hg", "bitbucket.org/kardianos/osext"},
		{"bitbucket.org/kardianos/osext", "https://bitbucket.org/kardianos/osext", "", "hg", "bitbucket.org/kardianos/osext"},
		{"bitbucket.org/ymotongpoo/go-bitarray", "git@bitbucket.org:ymotongpoo/go-bitarray.git", "git", "git", "bitbucket.org/ymotongpoo/go-bitarray"},
		{"bitbucket.org/ymotongpoo/go-bitarray", "https://bitbucket.org/ymotongpoo/go-bitarray.git", "git", "git", "bitbucket.org/ymotongpoo/go-bitarray"},
		{"code.google.com/p/go.tools/go/vcs", "https://code.google.com/p/go.tools/", "", "hg", "code.google.com/p/go.tools"},
		// not supported yet - {"example.com/foo/go-sqlite3", "git@github.com:mattn/go-sqlite3.git", "git", "git", "example.com/foo/go-sqlite3"},
	}

	Describe("GuessVCS()", func() {
		for _, test := range tests {
			t := test
			Context(t.url, func() {
				It("returns "+t.guess, func() {
					Expect(goop.GuessVCS(t.url)).To(Equal(t.guess))
				})
			})
		}
	})

	// slow test: this test requires network connection
	XDescribe("IdentifyVCS()", func() {
		for _, test := range tests {
			t := test
			Context(t.url, func() {
				It("returns "+t.actual, func() {
					Expect(goop.IdentifyVCS(t.url)).To(Equal(t.actual))
				})
			})
		}
	})

	Describe("RepoRootForImportPathWithURLOverride()", func() {
		for _, test := range tests {
			t := test
			Context(t.url, func() {
				It("returns "+t.repoRoot, func() {
					repo, err := goop.RepoRootForImportPathWithURLOverride(t.importPath, t.url)
					Expect(err).To(BeNil())
					Expect(repo).NotTo(BeNil())
					Expect(repo.Repo).To(Equal(t.url))
					Expect(repo.Root).To(Equal(t.repoRoot))
				})
			})
		}
	})
})

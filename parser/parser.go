package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Dependency struct {
	Pkg string
	Rev string
}

var reDepDefn = regexp.MustCompile(`\A(\S+)(\s+#(\S+))?\z`)

type ParseError struct {
	LineNum  uint
	LineText string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse failed at line %d: %s", e.LineNum, e.LineText)
}

func Parse(r io.Reader) ([]*Dependency, error) {
	s := bufio.NewScanner(r)
	ln := uint(0)
	deps := []*Dependency{}

	for s.Scan() {
		ln++
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		match := reDepDefn.FindStringSubmatch(line)
		if match == nil {
			return nil, &ParseError{LineNum: ln, LineText: line}
		}
		deps = append(deps, &Dependency{Pkg: match[1], Rev: match[3]})
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	return deps, nil
}

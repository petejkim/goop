package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ParseError struct {
	LineNum  uint
	LineText string
	Message  string
}

const (
	CommentOrPackage = iota
	URLOr
)

const (
	TokenComment = "//"
	TokenRev     = "#"
	TokenURL     = "!"
)

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse failed at line %d - %s\n  %s", e.LineNum, e.LineText, e.Message)
}

func Parse(r io.Reader) ([]*Dependency, error) {
	s := bufio.NewScanner(r)
	ln := uint(0)
	deps := []*Dependency{}

	for s.Scan() {
		ln++
		line := strings.TrimSpace(s.Text())
		tokens := strings.Fields(line)

		if line == "" || strings.HasPrefix(tokens[0], TokenComment) {
			continue
		}

		dep := &Dependency{Pkg: tokens[0]}
		parseErr := &ParseError{LineNum: ln, LineText: line}

		for _, t := range tokens[1:] {
			if strings.HasPrefix(t, TokenComment) {
				break
			}
			switch {
			case strings.HasPrefix(t, TokenRev):
				if dep.Rev != "" {
					parseErr.Message = "Multiple revisions given"
					return nil, parseErr
				}
				dep.Rev = t[1:]
			case strings.HasPrefix(t, TokenURL):
				if dep.URL != "" {
					parseErr.Message = "Multiple URLs given"
					return nil, parseErr
				}
				dep.URL = t[1:]
			default:
				parseErr.Message = "Unrecognized token given"
				return nil, parseErr
			}
		}
		deps = append(deps, dep)
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	return deps, nil
}

package parser

import "strings"

type Dependency struct {
	Pkg string
	Rev string
	URL string
}

func (d *Dependency) String() string {
	s := make([]string, 0, 3)
	s = append(s, d.Pkg)
	if d.Rev != "" {
		s = append(s, "#"+d.Rev)
	}
	if d.URL != "" {
		s = append(s, "!"+d.URL)
	}
	return strings.Join(s, " ")
}

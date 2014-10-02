package parser

import (
	"os/exec"
	"strings"
)

func ParseGoList(input []byte) []string {
	rawstr := string(input)
	for _, strip_char := range []string{"[", "]", "\n"} {
		rawstr = strings.Replace(rawstr, strip_char, "", -1)
	}

	return strings.Split(rawstr, " ")
}

func GetStdPackages() []string {
	cmd := exec.Command("go", "list", "std")
	if std_deps, err := cmd.Output(); err == nil {
		return strings.Split(string(std_deps), "\n")
	}

	return []string{}
}

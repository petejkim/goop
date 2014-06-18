package main

import (
	"os"
	"path"

	"github.com/nitrous-io/goop/colors"
	"github.com/nitrous-io/goop/goop"
)

func main() {
	name := path.Base(os.Args[0])

	pwd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString(colors.Error + name + ": error - failed to determine present working directory!" + colors.Reset + "\n")
	}

	g := goop.NewGoop(path.Join(pwd), os.Stdin, os.Stdout, os.Stderr)

	switch os.Args[1] {
	case "install":
		err = g.Install()
	case "exec":
		err = g.Exec(os.Args[2], os.Args[3:]...)
	case "go":
		err = g.Exec("go", os.Args[2:]...)
	}
	if err != nil {
		os.Stderr.WriteString(colors.Error + name + ": error - " + err.Error() + colors.Reset + "\n")
	}
}

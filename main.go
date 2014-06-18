package main

import (
	"errors"
	"os"
	"path"
	"strings"

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

	ignoreError := false

	if len(os.Args) < 2 {
		printUsage()
	}

	cmd := os.Args[1]
	switch cmd {
	case "help":
		printUsage()
	case "install":
		err = g.Install()
	case "update":
		err = g.Update()
	case "exec":
		if len(os.Args) < 3 {
			printUsage()
		}
		err = g.Exec(os.Args[2], os.Args[3:]...)
		ignoreError = true
	case "go":
		if len(os.Args) < 3 {
			printUsage()
		}
		err = g.Exec("go", os.Args[2:]...)
		ignoreError = true
	case "env":
		g.PrintEnv()
	default:
		err = errors.New(`unrecognized command "` + cmd + `"`)
	}

	if err != nil {
		if ignoreError {
			os.Exit(1)
		}
		os.Stderr.WriteString(colors.Error + name + ": error - " + err.Error() + colors.Reset + "\n")
	}
}

func printUsage() {
	os.Stdout.WriteString(strings.TrimSpace(usage) + "\n\n")
	os.Exit(0)
}

const usage = `
Goop is a tool for managing Go dependencies.

        goop command [arguments]

The commands are:

    install     install the dependencies specified by Goopfile or Goopfile.lock
    update      update dependencies to their latest versions
    env         print GOPATH and PATH environment variables, with the vendor path prepended
    exec        execute a command in the context of the installed dependencies
    go          execute a go command in the context of the installed dependencies
    help        print this message
`

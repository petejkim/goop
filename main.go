package main

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nitrous-io/goop/colors"
	"github.com/nitrous-io/goop/goop"
)

func main() {
	name := path.Base(os.Args[0])

	pwd, err := os.Getwd()
	if err != nil {
		handleNoWorkingDirError(name)
	}

	g := goop.NewGoop(path.Join(pwd), os.Stdin, os.Stdout, os.Stderr)

	if len(os.Args) < 2 {
		printUsage()
	}

	cmd := os.Args[1]
	err = executeCommand(cmd, err)

	if err != nil {
		handleExecutionError(err)
	}
}

func handleNoWorkingDirError(name string) {
	os.Stderr.WriteString(colors.Error + name + ": failed to determine present working directory!" + colors.Reset + "\n")
}

func printUsage() {
	os.Stdout.WriteString(strings.TrimSpace(usage) + "\n\n")
	os.Exit(0)
}

func executeCommand(cmd string, err error) err error {
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
	case "go":
		if len(os.Args) < 3 {
			printUsage()
		}
		err = g.Exec("go", os.Args[2:]...)
	case "env":
		g.PrintEnv()
	default:
		err = errors.New(`unrecognized command "` + cmd + `"`)
	}
	return err
}

func handleExecutionError(err error) {
	errMsg := err.Error()
	code := 1

	// go does not provide a cross-platform way to get exit status, so inspect error message instead
	// https://code.google.com/p/go/source/browse/src/pkg/os/exec_posix.go#119
	if strings.HasPrefix(errMsg, "exit status ") {
		code, err = strconv.Atoi(errMsg[len("exit status "):])
		if err != nil {
			code = 1
		}
		errMsg = "Command failed with " + errMsg
	}

	os.Stderr.WriteString(colors.Error + name + ": " + errMsg + colors.Reset + "\n")
	os.Exit(code)
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

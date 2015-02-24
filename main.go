package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/nitrous-io/goop/colors"
	"github.com/nitrous-io/goop/goop"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	app = kingpin.New("goop", "A a tool for managing Go dependencies.")

	installCmd     = app.Command("install", "Install the dependencies specified by Goopfile or Goopfile.lock")
	installPath    = installCmd.Flag("path", "Install dependencies to this directory").String()
	installVerbose = installCmd.Flag("verbose", "Enable verbose output").Bool()

	updateCmd     = app.Command("update", "Update dependencies to their latest versions")
	updateVerbose = updateCmd.Flag("verbose", "Enable verbose output").Bool()

	execCmd  = app.Command("exec", "Execute a command in the context of the installed dependencies")
	execArgs = StringList(execCmd.Arg("command", "Command and arguments to execute").Required())

	goCmd  = app.Command("go", "Execute a go command in the context of the installed dependencies")
	goArgs = StringList(goCmd.Arg("command", "Command and arguments to execute").Required())

	envCmd = app.Command("env", "Print GOPATH and PATH environment variables, with the vendor path prepended")
)

func main() {
	name := path.Base(os.Args[0])

	pwd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString(colors.Error + name + ": failed to determine present working directory!" + colors.Reset + "\n")
	}

	g, err := goop.NewGoop(path.Join(pwd), os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		os.Stderr.WriteString(colors.Error + name + ": " + err.Error() + colors.Reset + "\n")
		os.Exit(1)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case installCmd.FullCommand():
		if *installVerbose {
			g.Verbose = true
		}
		err = g.Install(*installPath)

	case updateCmd.FullCommand():
		if *updateVerbose {
			g.Verbose = true
		}
		err = g.Update()

	case envCmd.FullCommand():
		g.PrintEnv()

	case execCmd.FullCommand():
		args := *execArgs
		if len(args) > 0 {
			err = g.Exec(args[0], args[1:]...)
		}

	case goCmd.FullCommand():
		args := *goArgs
		if len(args) > 0 {
			err = g.Exec("go", args...)
		}

	default:
		app.Usage(os.Stdout)

	}

	if err != nil {
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
}

type stringList []string

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (l *stringList) String() string {
	return ""
}

func (l *stringList) IsCumulative() bool {
	return true
}

func StringList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*stringList)(target))
	return
}

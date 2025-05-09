package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/awryme/slogf"
)

const helpmsg = `
gb is a simple build runner, taking commands from environment variables and .env / .gb files

available env variables:
	- GB: sets build commands
	- GB_RUN: sets run commands
	- GB_SHELL: overrides shell to use

each env for GB and GB_RUN can contain multiple commands
commands are separated by ';'
each command is passed to the shell separately

env variables can be store in .env files
default file is standard '.env', which is always read, if exists
it does not override env variables passed to gb

if an argument is provided, it will be interpreted as an additional env file
only one argument can be specified, if reading it failes, gb prints the error and exits
variables defined in the file override existing env variables
file passed as <name> will be read as <name>, .<name>.gb or .<name>.env, whichever is available in that order

env sources priority:
1) file passed as argument
2) env variables inherited by command
3) .env file
`

// envs
const (
	envGB      = "GB"
	envGBShell = "GB_SHELL"
	envGBRun   = "GB_RUN"

	envStdShell = "SHELL"
)

const cmdSplitSymbol = ";"

func main() {
	err := run()
	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}
}

func run() error {
	log := slogf.DefaultHandler(os.Stdout)
	logf := slogf.New(log)

	flag.Usage = func() {
		fmt.Println(helpmsg)
		os.Exit(0)
	}

	flag.Parse()
	if flag.NArg() > 1 {
		fmt.Println("error: only one env file can be specified, refer to 'gb -h'")
		os.Exit(1)
	}

	err := DotenvLoad(log, flag.Arg(0))
	if err != nil {
		return err
	}

	buildCmd := os.Getenv(envGB)
	if buildCmd == "" {
		return fmt.Errorf("env %s is not set", envGB)
	}
	runCmd := os.Getenv(envGBRun)
	shell := detectShell()
	logf("using shell", slog.String("shell", shell.Shell), slog.String("from", shell.From))
	logf("using build commands", slog.String("commands", buildCmd))
	if runCmd != "" {
		logf("using run commands", slog.String("commands", runCmd))
	}

	for cmd := range strings.SplitSeq(buildCmd, cmdSplitSymbol) {
		cmd = strings.TrimSpace(cmd)
		logf("building", slog.String("command", cmd))
		err := execCmd(cmd, shell.Shell)
		if err != nil {
			return err
		}
	}

	for cmd := range strings.SplitSeq(runCmd, cmdSplitSymbol) {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		logf("running", slog.String("command", cmd))
		err := execCmd(cmd, shell.Shell)
		if err != nil {
			return err
		}
	}

	return nil
}

func execCmd(cmd string, shell string) error {
	args := []string{
		"-c",
		cmd,
	}
	command := exec.Command(shell, args...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	return command.Run()
}

const defaultShellUnix = "bash"
const defaultShellWindows = "powershell.exe"

type ShellResult struct {
	From  string
	Shell string
}

func detectShell() ShellResult {
	env := os.Getenv(envGBShell)
	if env != "" {
		return ShellResult{
			From:  fmt.Sprintf("env %s", envGBShell),
			Shell: env,
		}
	}

	// check $SHELL
	stdenv := os.Getenv(envStdShell)
	if stdenv != "" {
		return ShellResult{
			From:  fmt.Sprintf("env %s", envStdShell),
			Shell: stdenv,
		}
	}

	// try to use default shell
	switch runtime.GOOS {
	case "windows":
		return ShellResult{
			From:  "default windows shell",
			Shell: defaultShellWindows,
		}
	default:
		return ShellResult{
			From:  "default unix shell",
			Shell: defaultShellUnix,
		}
	}
}

package main

import (
	"fmt"
	"os"
	"strings"
)

func logMessage(message string) {
	if opts.Verbose {
		fmt.Fprintln(os.Stderr, message)
	}
}

func logError(err error) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: %s\n", err))
}

func logFatalErrorAndExit(err error, exitCode int) {
	cmdline := fmt.Sprintf("%s %s", argparser.Command.Name, strings.Join(os.Args[1:], " "))

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: %s", err))
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Command: %s", cmdline))

	os.Exit(exitCode)
}

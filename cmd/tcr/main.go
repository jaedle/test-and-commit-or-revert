package main

import (
	"github.com/jaedle/test-and-commit-or-revert/internal"
	"os"
)

type cmd string

const (
	squash cmd = "squash"
	tcr    cmd = "tcr"
)

func command() cmd {
	if len(os.Args) > 1 && os.Args[1] == "squash" {
		return squash
	} else {
		return tcr
	}
}

func main() {
	switch command() {
	case tcr:
		toExitCode(internal.New().Tcr())
	case squash:
		toExitCode(internal.New().Squash())
	}

}

func toExitCode(result internal.Result) {
	switch result {
	case internal.Success:
		os.Exit(0)
	case internal.Failure:
		os.Exit(1)
	case internal.Error:
		os.Exit(1)
	}
}

package main

import (
	"github.com/jaedle/test-and-commit-or-revert/internal"
	"os"
)

func main() {
	result := internal.New(internal.Config{
		Workdir: ".",
	}).Run()
	switch result {
	case internal.Success:
		os.Exit(0)
	case internal.Failure:
		os.Exit(1)
	case internal.Error:
		os.Exit(1)
	}
}

package main

import (
	"github.com/jaedle/test-and-commit-or-revert/internal"
	"os"
)

func main() {
	switch internal.New().Run() {
	case internal.Success:
		os.Exit(0)
	case internal.Failure:
		os.Exit(1)
	case internal.Error:
		os.Exit(1)
	}
}

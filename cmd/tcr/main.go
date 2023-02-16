package main

import (
	"github.com/jaedle/test-and-commit-or-revert/internal"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "squash" {
			switch internal.New().RunSquash() {
			case internal.Success:
				os.Exit(0)
			case internal.Failure:
				os.Exit(1)
			case internal.Error:
				os.Exit(1)
			}
		}
	}

	switch internal.New().RunTcr() {
	case internal.Success:
		os.Exit(0)
	case internal.Failure:
		os.Exit(1)
	case internal.Error:
		os.Exit(1)
	}
}

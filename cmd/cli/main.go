package main

import (
	"github.com/jaedle/test-and-commit-or-revert/internal"
	"os"
)

type config struct {
	Test string `json:"test"`
}

func main() {
	result, err := internal.New(internal.Config{
		Workdir: ".",
	}).Run()
	if err != nil {
		os.Exit(1)
	} else if *result == internal.Failure {
		os.Exit(1)
	} else {
		os.Exit(0)
	}

}

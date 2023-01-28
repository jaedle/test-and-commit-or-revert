package main

import "os"

func main() {
	_, err := os.Stat(".git")
	if err != nil {
		os.Exit(1)
	}
}

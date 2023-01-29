package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
	"os/exec"
)

type config struct {
	Test string `json:"test"`
}

func main() {
	println("tcr")

	_, err := os.Stat(".git")
	if err != nil {
		os.Exit(1)
	}

	println("reading config")

	_, err = os.Stat("tcr.json")
	if err != nil {
		os.Exit(1)
	}

	file, err := os.Open("tcr.json")
	if err != nil {
		os.Exit(1)
	}
	defer func() { _ = file.Close() }()

	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		os.Exit(1)
	}

	println("opening repository")

	repo, err := git.PlainOpen(".")
	if err != nil {
		os.Exit(1)
	}

	if passed, err := test(c); err != nil {
		os.Exit(1)
	} else if passed {
		commit(repo)
	} else {
		revert(repo)
		os.Exit(1)
	}

}

func test(c config) (bool, error) {
	println("running tests")

	cmd := exec.Command(c.Test)
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		println("exit error", err.Error())
		return false, nil
	} else if err != nil {
		println("general error", err.Error())
		return false, err
	} else {
		return true, nil
	}
}

func commit(repo *git.Repository) {
	worktree, err := repo.Worktree()
	if err != nil {
		os.Exit(1)
	}

	if err := worktree.AddGlob("*"); err != nil {
		os.Exit(1)
	}

	_, err = worktree.Commit("[WIP] refactoring", &git.CommitOptions{})
	if err != nil {
		os.Exit(1)
	}
}

func revert(repo *git.Repository) {
	worktree, err := repo.Worktree()
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("cleaned")

	//err = worktree.Clean(&git.CleanOptions{})
	//if err != nil {
	//	os.Exit(1)
	//}

	err = worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		os.Exit(1)
	}

	status, err := worktree.Status()
	if err != nil {
		os.Exit(1)
	}

	println(status.String())

}

package internal

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
)

type Result int

const (
	Success Result = iota
	Failure Result = iota
)

type config struct {
	Test string `json:"test"`
}

func New(c Config) *Tcr {
	return &Tcr{
		logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

type Tcr struct {
	config config
	repo   *git.Repository
	logger zerolog.Logger
}

func (t *Tcr) Run() (*Result, error) {
	if err := t.readConfig(); err != nil {
		return nil, err
	}

	if err := t.openRepository(); err != nil {
		return nil, err
	}

	if passed, err := t.test(); err != nil {
		return nil, err
	} else if passed {
		return t.commit()
	} else {
		if err := t.revert(); err != nil {
			return nil, err
		}

		var result = Failure
		return &result, nil
	}

}

func (t *Tcr) readConfig() error {
	t.logger.Trace().Msg("reading configuration")

	file, err := os.Open("tcr.json")
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		return err
	}

	t.config = c
	return nil
}

func (t *Tcr) openRepository() error {
	t.logger.Trace().Msg("opening repository")

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	t.repo = repo
	return nil
}

func (t *Tcr) test() (bool, error) {
	t.logger.Trace().Msg("running tests")

	cmd := exec.Command(t.config.Test)
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

func (t *Tcr) commit() (*Result, error) {
	t.logger.Trace().Msg("commiting")

	wt, err := t.repo.Worktree()
	if err != nil {
		return nil, err
	}

	if err := wt.AddGlob("*"); err != nil {
		return nil, err
	}

	if _, err := wt.Commit("[WIP] refactoring", &git.CommitOptions{}); err != nil {
		return nil, err
	}

	var result = Success
	return &result, nil
}

func (t *Tcr) revert() error {
	t.logger.Trace().Msg("reverting commit")

	worktree, err := t.repo.Worktree()
	if err != nil {
		return nil
	}
	fmt.Println("cleaned")

	return worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
}

type Config struct {
	Workdir string
}

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
	Error   Result = iota
)

type config struct {
	Test string `json:"test"`
}

func New(c Config) *Tcr {
	return &Tcr{
		logger: zerolog.New(os.Stdout).
			Output(zerolog.NewConsoleWriter()).
			With().Timestamp().
			Logger(),
	}
}

type Tcr struct {
	config config
	repo   *git.Repository
	logger zerolog.Logger
}

func (t *Tcr) Run() Result {
	if err := t.readConfig(); err != nil {
		t.logger.Err(err).Msg("error on reading configuration")
		return Error
	}

	if err := t.openRepository(); err != nil {
		t.logger.Err(err).Msg("error on opening git repository")
		return Error
	}

	if passed, err := t.test(); err != nil {
		t.logger.Err(err).Msg("error on running tests")
		return Error
	} else if passed {
		t.logger.Info().Msg("tests have passed, committing changes")
		if err := t.commit(); err != nil {
			t.logger.Err(err).Msg("error on commit")
			return Error
		} else {
			return Success
		}
	} else {
		t.logger.Info().Msg("tests have failed, resetting worktree")
		if err := t.revert(); err != nil {
			t.logger.Err(err).Msg("error on reverting commit")
			return Error
		} else {
			return Failure
		}
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
		t.logger.Info().Err(err).Msg("test execution failed")
		return false, nil
	} else if err != nil {
		println("general error", err.Error())
		return false, err
	} else {
		return true, nil
	}
}

func (t *Tcr) commit() error {
	t.logger.Trace().Msg("commiting")

	wt, err := t.repo.Worktree()
	if err != nil {
		return err
	}

	if err := wt.AddGlob("*"); err != nil {
		return err
	}

	_, err = wt.Commit("[WIP] refactoring", &git.CommitOptions{})
	return err
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

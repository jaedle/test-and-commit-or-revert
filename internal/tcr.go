package internal

import (
	"encoding/json"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"strings"
)

type Result int

const (
	Error   Result = iota
	Failure Result = iota
	Success Result = iota
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
	repo        *git.Repository
	logger      zerolog.Logger
	testCommand []string
}

func (t *Tcr) Run() Result {
	if err := t.openRepository(); err != nil {
		t.logger.Err(err).Msg("error on opening git repository")
		return Error
	}

	if err := t.readConfig(); err != nil {
		t.logger.Err(err).Msg("error on reading configuration")
		return Error
	}

	if clean, err := t.cleanWorktree(); err != nil {
		t.logger.Err(err).Msg("error on running tests")
		return Error
	} else if clean {
		t.logger.Info().Msg("worktree is clean, nothing to do")
		return Success
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

func (t *Tcr) openRepository() error {
	t.logger.Trace().Msg("opening repository")

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	t.repo = repo
	return nil
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

	t.testCommand = strings.Split(c.Test, " ")
	return nil
}

func (t *Tcr) cleanWorktree() (bool, error) {
	wt, err := t.repo.Worktree()
	if err != nil {
		return false, err
	}

	if status, err := wt.Status(); err != nil {
		return false, err
	} else {
		return status.IsClean(), nil
	}

}

func (t *Tcr) test() (bool, error) {
	t.logger.Trace().Msg("running tests")

	cmd := exec.Command(t.testCommand[0], t.testCommand[1:]...)
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		t.logger.Info().Err(err).Msg("test execution failed")
		return false, nil
	} else if err != nil {
		t.logger.Info().Err(err).Any("cmd", cmd).Msg("general error on running the tests")
		return false, err
	} else {
		return true, nil
	}
}

func (t *Tcr) commit() error {
	t.logger.Trace().Msg("commit")

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
	t.logger.Trace().Msg("revert")

	worktree, err := t.repo.Worktree()
	if err != nil {
		return nil
	}

	return worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
}

type Config struct {
	Workdir string
}

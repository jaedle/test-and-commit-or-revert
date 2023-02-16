package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func New() *Tcr {
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

func (t *Tcr) Tcr() Result {
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

	var out bytes.Buffer
	cmd := exec.Command(t.testCommand[0], t.testCommand[1:]...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		t.logger.Info().Err(err).Msg("test execution failed")
		fmt.Print(out.String())
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

func (t *Tcr) Squash() Result {
	if err := t.openRepository(); err != nil {
		t.logger.Err(err).Msg("error on opening git repository")
		return Error
	}

	if clean, err := t.cleanWorktree(); err != nil {
		t.logger.Err(err).Msg("error on reading git worktree")
		return Error
	} else if !clean {
		t.logger.Error().Msg("worktree is not clean, no squashing possible")
		return Failure
	}

	history, err := t.getHistory()
	if err != nil {
		t.logger.Err(err).Msg("error on reading git log")
		return Error
	}

	if t.numberOfRefactoringCommits(history) == 0 {
		t.logger.Error().Msg("no refactoring commits, nothing to do")
		return Failure
	} else if t.numberOfRefactoringCommits(history) == 1 {
		t.logger.Info().Msg("only one refactoring commit, nothing to do")
		return Failure
	}

	if err := t.resetToCommit(history[t.numberOfRefactoringCommits(history)]); err != nil {
		t.logger.Error().Err(err).Str("hash", history[t.numberOfRefactoringCommits(history)].Hash.String()).Msg("error on resetting to commit")
		return Error
	}

	if err := t.commit(); err != nil {
		t.logger.Error().Err(err).Msg("error on commit")
		return Failure
	} else {
		return Success
	}

}

func (t *Tcr) numberOfRefactoringCommits(history []*object.Commit) int {
	var result = 0

	for _, s := range history {
		if s.Message == "[WIP] refactoring" {
			result++
		}
	}

	return result
}

func (t *Tcr) getHistory() ([]*object.Commit, error) {
	log, err := t.repo.Log(&git.LogOptions{})
	if err != nil {
		t.logger.Err(err).Msg("error on reading git log")
		return nil, err
	}

	var result []*object.Commit
	err = log.ForEach(func(c *object.Commit) error {
		result = append(result, c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (t *Tcr) resetToCommit(c *object.Commit) error {
	worktree, err := t.repo.Worktree()
	if err != nil {
		return nil
	}

	return worktree.Reset(&git.ResetOptions{
		Commit: c.Hash,
		Mode:   git.SoftReset,
	})
}

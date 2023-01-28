package test

import (
	"github.com/go-git/go-git/v5"
	"os"
	"path"
)

func NewGitHelper(dir string) *GitHelper {
	return &GitHelper{
		dir: dir,
	}
}

type GitHelper struct {
	dir  string
	repo *git.Repository
}

func (h *GitHelper) WithCommits() error {
	repo, err := git.PlainInit(h.dir, false)
	if err != nil {
		return err
	}
	h.repo = repo

	if err := os.WriteFile(path.Join(h.dir, ".test"), nil, os.ModePerm); err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := wt.Add(".test"); err != nil {
		return err
	}

	_, err = wt.Commit("a commit message", &git.CommitOptions{})
	return err

}

package test

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

	if err := os.WriteFile(path.Join(h.dir, "README.md"), []byte("# Dummy"), os.ModePerm); err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := wt.Add("README.md"); err != nil {
		return err
	}

	_, err = wt.Commit("a commit message", &git.CommitOptions{})
	return err

}

func (h *GitHelper) InitRepositoryWithFiles(f Files) error {
	repo, err := git.PlainInit(h.dir, false)
	if err != nil {
		return err
	}
	h.repo = repo

	for _, file := range f {
		if err := os.WriteFile(path.Join(h.dir, file.Name), []byte(file.Content), os.ModePerm); err != nil {
			return err
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	for _, file := range f {
		if _, err := wt.Add(file.Name); err != nil {
			return err
		}
	}

	_, err = wt.Commit("a commit message", &git.CommitOptions{})
	return err
}

func (h *GitHelper) IsWorkingTreeClean() (bool, error) {
	worktree, err := h.repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, err
	}

	return status.IsClean(), nil
}

func (h *GitHelper) Commit() error {
	if wt, err := h.repo.Worktree(); err != nil {
		return err
	} else if err := wt.AddGlob("*"); err != nil {
		return err
	} else {
		_, err := wt.Commit("commit", &git.CommitOptions{
			All:               false,
			AllowEmptyCommits: false,
			Author: &object.Signature{
				Name:  "any",
				Email: "any",
			},
			Committer: nil,
			Parents:   nil,
			SignKey:   nil,
		})
		return err
	}

}

func (h *GitHelper) Init() error {
	repo, err := git.PlainInit(h.dir, false)
	if err != nil {
		return err
	}
	h.repo = repo
	return nil
}

func (h *GitHelper) Head() (string, error) {
	if head, err := h.repo.Head(); err != nil {
		return "", err
	} else {
		return head.Hash().String(), nil
	}
}

type GitHistory []Commit
type Commit struct {
	Hash    string
	Message string
}

func (h *GitHelper) Commits() (GitHistory, error) {
	log, err := h.repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	var result GitHistory
	err = log.ForEach(func(c *object.Commit) error {
		result = append(result, Commit{
			Hash:    c.Hash.String(),
			Message: c.Message,
		})
		return nil
	})
	return result, err
}

func (h *GitHelper) Add(name string) error {
	worktree, err := h.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Add(name)
	return err

}

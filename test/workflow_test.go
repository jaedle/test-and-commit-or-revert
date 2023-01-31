package test_test

import (
	"github.com/jaedle/test-and-commit-or-revert/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
)

const defaultCommitMessage = "[WIP] refactoring"
const aFileName = "new"
const aContent = "some content"
const anUpdatedContent = "updated content"

var _ = Describe("Workflow", Ordered, func() {
	var binary string
	var workdir string
	var gitHelper *test.GitHelper

	BeforeAll(func() {
		var err error
		binary, err = gexec.Build("github.com/jaedle/test-and-commit-or-revert/cmd/tcr")
		Expect(err).ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		d, err := os.MkdirTemp(os.TempDir(), "tcr-workflow-test")
		Expect(err).NotTo(HaveOccurred())
		workdir = d

		gitHelper = test.NewGitHelper(workdir)
	})

	AfterEach(func() {
		_ = os.RemoveAll(workdir)
	})

	AfterAll(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("dirty worktree", func() {
		Context("tests passing", func() {
			It("commits untracked files on succeeded cycle", func() {
				givenAPassingTestSetup(workdir, gitHelper)
				history := givenAGitHistory(gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				result := whenIRunTcr(binary, workdir)

				thenTcrSucceeds(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenANewCommitIsAdded(gitHelper, history, defaultCommitMessage)
			})

			It("commits changes on tracked files on succeeded cycle", func() {
				givenAPassingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				history := givenAGitHistory(gitHelper)

				result := whenIRunTcr(binary, workdir)

				thenTcrSucceeds(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				thenANewCommitIsAdded(gitHelper, history, defaultCommitMessage)
			})
		})
		Context("tests failing", func() {
			It("removes untracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				result := whenIRunTcr(binary, workdir)

				thenTcrFails(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenTheUnstagedChangesAreReset(workdir, test.Files{{Name: aFileName, Content: aContent}})
			})

			It("resets changes to tracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				commits := givenAGitHistory(gitHelper)

				result := whenIRunTcr(binary, workdir)

				thenTcrFails(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})

			It("resets staged changes on tracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenStagedChanges(workdir, gitHelper, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				commits := givenAGitHistory(gitHelper)

				result := whenIRunTcr(binary, workdir)

				thenTcrFails(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})

			It("resets staged changes on untracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenStagedChanges(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				commits := givenAGitHistory(gitHelper)

				result := whenIRunTcr(binary, workdir)

				thenTcrFails(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenTheUnstagedChangesAreReset(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})
		})
		Context("tests can not be executed", func() {
			It("does not revert", func() {
				givenATestSetupWithNonExecutableTests(workdir, gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				result := whenIRunTcr(binary, workdir)

				thenTcrFails(result)
				thenTheWorkingTreeIsNotClean(gitHelper)
			})
		})
	})

	Context("test commands", func() {
		It("supports arguments", func() {
			givenATestCommandThatNeedsArguments(gitHelper, workdir)
			givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

			result := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(result)
			thenTheWorkingTreeIsClean(gitHelper)
		})

		It("outputs test output if tests fail", func() {
			givenAFailingTestSetupWithOutput(workdir, gitHelper, "some random output")
			givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

			result := whenIRunTcr(binary, workdir)

			thenItDisplays(result, "some random output")
		})

		It("swallows test output if tests fail", func() {
			givenAPassingTestSetupWithOutput(workdir, gitHelper, "some random output")
			givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

			result := whenIRunTcr(binary, workdir)

			thenItDoesNotDisplay(result, "some random output")
		})
	})

	Context("clean worktree", func() {
		It("does not create a new commit", func() {
			givenAPassingTestSetup(workdir, gitHelper)
			givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
			commits := givenAGitHistory(gitHelper)
			result := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(result)
			thenTheWorkingTreeIsClean(gitHelper)
			thenTheHistoryIsUnchaged(gitHelper, commits)
		})
	})

	Context("error cases", func() {
		It("fails if not run within a git repository", func() {
			result := whenIRunTcr(binary, workdir)

			thenTcrFails(result)
		})
		It("fails if no config is present", func() {
			helper := test.NewGitHelper(workdir)
			Expect(helper.WithCommits()).NotTo(HaveOccurred())
			result := whenIRunTcr(binary, workdir)

			thenTcrFails(result)
		})
	})

})

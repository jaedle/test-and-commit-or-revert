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
	var tempTestDir string
	var gitHelper *test.GitHelper

	BeforeAll(func() {
		var err error
		binary, err = gexec.Build("github.com/jaedle/test-and-commit-or-revert/cmd/tcr")
		Expect(err).ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		tmp1, err := os.MkdirTemp(os.TempDir(), "tcr-workflow-test-workdir")
		Expect(err).NotTo(HaveOccurred())
		workdir = tmp1
		gitHelper = test.NewGitHelper(workdir)

		tmp2, err := os.MkdirTemp(os.TempDir(), "tcr-workflow-test-tmp-test-dir")
		Expect(err).NotTo(HaveOccurred())
		tempTestDir = tmp2
	})

	AfterEach(func() {
		_ = os.RemoveAll(workdir)
	})

	AfterAll(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("dirty worktree", func() {

		Context("test passes", func() {
			It("commits untracked files", func() {
				givenAPassingTestSetup(workdir, "", gitHelper)
				history := givenAGitHistory(gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				result := whenIRunTcr(binary, workdir)

				thenTcrSucceeds(result)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenANewCommitIsAdded(gitHelper, history, defaultCommitMessage)
			})

			It("commits tracked files", func() {
				givenAPassingTestSetup(workdir, "", gitHelper)
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

		Context("test fails", func() {
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

			It("resets already staged changes on tracked files", func() {
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
	})

	Context("test execution fails", func() {
		It("does not revert", func() {
			givenATestSetupWithNonExecutableTests(workdir, gitHelper)
			givenAnyUnstagedChanges(workdir)

			result := whenIRunTcr(binary, workdir)

			thenTcrFails(result)
			thenTheWorkingTreeIsNotClean(gitHelper)
		})
	})

	Context("test commands", func() {
		It("supports arguments", func() {
			givenATestCommandThatNeedsArguments(gitHelper, workdir)
			givenAnyUnstagedChanges(workdir)

			result := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(result)
		})
	})

	Context("test output", func() {
		It("is swallowed if test passes", func() {
			givenAPassingTestSetupWithOutput(workdir, gitHelper, "some random output")
			givenAnyUnstagedChanges(workdir)

			result := whenIRunTcr(binary, workdir)

			thenItDoesNotDisplay(result, "some random output")
		})

		It("is printed if test fails", func() {
			givenAFailingTestSetupWithOutput(workdir, gitHelper, "some random output")
			givenAnyUnstagedChanges(workdir)

			result := whenIRunTcr(binary, workdir)

			thenItDisplays(result, "some random output")
		})
	})

	Context("clean worktree", func() {
		It("does not create a new commit", func() {
			givenAPassingTestSetup(workdir, "", gitHelper)
			givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
			commits := givenAGitHistory(gitHelper)
			result := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(result)
			thenTheWorkingTreeIsClean(gitHelper)
			thenTheHistoryIsUnchaged(gitHelper, commits)
		})

		It("does not test", func() {
			givenATestThatLogsRun(workdir, tempTestDir, gitHelper)
			givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
			result := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(result)
			thenTheWorkingTreeIsClean(gitHelper)
			thenTestWasNotRun(tempTestDir)
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

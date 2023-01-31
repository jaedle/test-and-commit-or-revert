package test_test

import (
	"github.com/jaedle/test-and-commit-or-revert/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
	"os/exec"
	"path"
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

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrSucceeds(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenANewCommitIsAdded(gitHelper, history, defaultCommitMessage)
			})

			It("commits changes on tracked files on succeeded cycle", func() {
				givenAPassingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				history := givenAGitHistory(gitHelper)

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrSucceeds(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				thenANewCommitIsAdded(gitHelper, history, defaultCommitMessage)
			})
		})
		Context("tests failing", func() {
			It("removes untracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrFails(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenTheUnstagedChangesAreReset(workdir, test.Files{{Name: aFileName, Content: aContent}})
			})

			It("resets changes to tracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				commits := givenAGitHistory(gitHelper)

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrFails(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})

			It("resets staged changes on tracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				givenStagedChanges(workdir, gitHelper, test.Files{{Name: aFileName, Content: anUpdatedContent}})
				commits := givenAGitHistory(gitHelper)

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrFails(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenThoseFilesExist(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})

			It("resets staged changes on untracked files", func() {
				givenAFailingTestSetup(workdir, gitHelper)
				givenStagedChanges(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
				commits := givenAGitHistory(gitHelper)

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrFails(exitCode)
				thenTheWorkingTreeIsClean(gitHelper)
				thenTheUnstagedChangesAreReset(workdir, test.Files{{Name: aFileName, Content: aContent}})
				thenTheHistoryIsUnchaged(gitHelper, commits)
			})
		})
		Context("tests can not be executed", func() {
			It("does not revert", func() {
				givenATestSetupWithNonExecutableTests(workdir, gitHelper)
				givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})

				exitCode := whenIRunTcr(binary, workdir)

				thenTcrFails(exitCode)
				thenTheWorkingTreeIsNotClean(gitHelper)
			})
		})
	})

	Context("clean worktree", func() {
		It("does not create a new commit", func() {
			givenAPassingTestSetup(workdir, gitHelper)
			givenACommit(workdir, gitHelper, test.Files{{Name: aFileName, Content: aContent}})
			commits := givenAGitHistory(gitHelper)
			exitCode := whenIRunTcr(binary, workdir)

			thenTcrSucceeds(exitCode)
			thenTheWorkingTreeIsClean(gitHelper)
			thenTheHistoryIsUnchaged(gitHelper, commits)
		})
	})

	Context("error cases", func() {
		It("fails if not run within a git repository", func() {
			status := whenIRunTcr(binary, workdir)

			Expect(status).To(Equal(1))
		})
		It("fails if no config is present", func() {
			helper := test.NewGitHelper(workdir)
			Expect(helper.WithCommits()).NotTo(HaveOccurred())
			status := whenIRunTcr(binary, workdir)

			Expect(status).To(Equal(1))
		})
	})

})

func thenTheUnstagedChangesAreReset(workdir string, files test.Files) {
	for _, f := range files {
		p := path.Join(workdir, f.Name)
		Expect(p).NotTo(BeAnExistingFile())
	}
}

func thenANewCommitIsAdded(helper *test.GitHelper, previous test.GitHistory, msg string) {
	commits, err := helper.Commits()
	Expect(err).NotTo(HaveOccurred())

	Expect(len(commits)).To(Equal(len(previous)+1), "new commit must be added")
	Expect(commits[1:]).To(Equal(previous))
	Expect(commits[0].Message).To(Equal(msg))
}
func thenTheHistoryIsUnchaged(helper *test.GitHelper, previous test.GitHistory) {
	commits, err := helper.Commits()
	Expect(err).NotTo(HaveOccurred())
	Expect(commits).To(Equal(previous))
}

func givenAGitHistory(helper *test.GitHelper) test.GitHistory {
	commits, err := helper.Commits()
	Expect(err).NotTo(HaveOccurred())
	return commits
}

func thenTcrSucceeds(exitCode int) bool {
	return Expect(exitCode).To(Equal(0))
}

func thenTcrFails(exitCode int) bool {
	return Expect(exitCode).NotTo(Equal(0))
}

func thenTheWorkingTreeIsClean(helper *test.GitHelper) bool {
	return Expect(helper.IsWorkingTreeClean()).To(BeTrue(), "worktree must be clean")
}

func thenTheWorkingTreeIsNotClean(helper *test.GitHelper) bool {
	return Expect(helper.IsWorkingTreeClean()).To(BeFalse(), "worktree must not be clean")
}

func thenThoseFilesExist(workdir string, files test.Files) {
	for _, f := range files {
		p := path.Join(workdir, f.Name)
		Expect(p).To(BeAnExistingFile())

		file, err := os.ReadFile(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(file)).To(Equal(f.Content))
	}
}

func givenUnstangedChanges(workdir string, f test.Files) {
	for _, file := range f {
		Expect(os.WriteFile(path.Join(workdir, file.Name), []byte(file.Content), os.ModePerm)).NotTo(HaveOccurred())
	}
}

func givenStagedChanges(workdir string, helper *test.GitHelper, f test.Files) {
	for _, file := range f {
		Expect(os.WriteFile(path.Join(workdir, file.Name), []byte(file.Content), os.ModePerm)).NotTo(HaveOccurred())
		Expect(helper.Add(file.Name)).NotTo(HaveOccurred())
	}
}

func givenACommit(workdir string, helper *test.GitHelper, f test.Files) {
	for _, file := range f {
		Expect(os.WriteFile(path.Join(workdir, file.Name), []byte(file.Content), os.ModePerm)).NotTo(HaveOccurred())
	}

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

func whenIRunTcr(binary string, workdir string) int {
	cmd := exec.Command(binary)
	cmd.Dir = workdir

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait()

	return session.ExitCode()
}

func givenAPassingTestSetup(workdir string, helper *test.GitHelper) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
		{Name: "test.sh", Content: "#!/usr/bin/env bash\nexit 0"},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

func givenAFailingTestSetup(workdir string, helper *test.GitHelper) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
		{Name: "test.sh", Content: "#!/usr/bin/env bash\nexit 1"},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

func givenATestSetupWithNonExecutableTests(workdir string, helper *test.GitHelper) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

package test_test

import (
	"github.com/jaedle/test-and-commit-or-revert/test"
	. "github.com/onsi/gomega"
	"os"
	"path"
)

func thenItDoesNotDisplay(result tcrOutput, content string) {
	Expect(result.stdOut).NotTo(ContainSubstring(content))

}

func thenItDisplays(result tcrOutput, content string) {
	Expect(result.stdOut).To(ContainSubstring(content))
}

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

func thenTcrSucceeds(o tcrOutput) bool {
	return Expect(o.exitCode).To(Equal(0))
}

func thenTcrFails(o tcrOutput) bool {
	return Expect(o.exitCode).NotTo(Equal(0))
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
func thenTestWasNotRun(dir string) {
	Expect(path.Join(dir, "ran")).NotTo(BeAnExistingFile(), "test must not be run")
}

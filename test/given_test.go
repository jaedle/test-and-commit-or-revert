package test_test

import (
	"github.com/jaedle/test-and-commit-or-revert/test"
	. "github.com/onsi/gomega"
	"os"
	"path"
)

func givenAPassingTestSetup(workdir string, dir string, helper *test.GitHelper) {
	givenAPassingTestSetupWithOutput(workdir, helper, "any")
}

func givenATestThatLogsRun(workdir string, tmpTestDir string, helper *test.GitHelper) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	ran := path.Join(tmpTestDir, "ran")

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
		{Name: "test.sh", Content: "#!/usr/bin/env bash\ntouch '" + (ran) + "'\nexit 1"},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

func givenAFailingTestSetup(workdir string, helper *test.GitHelper) {
	givenAFailingTestSetupWithOutput(workdir, helper, "any")
}

func givenATestSetupWithNonExecutableTests(workdir string, helper *test.GitHelper) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
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

func givenAGitHistory(helper *test.GitHelper) test.GitHistory {
	commits, err := helper.Commits()
	Expect(err).NotTo(HaveOccurred())
	return commits
}

func givenAFailingTestSetupWithOutput(workdir string, helper *test.GitHelper, testOutput string) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
		{Name: "test.sh", Content: "#!/usr/bin/env bash\necho'" + testOutput + "'\nexit 1"},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}
func givenAPassingTestSetupWithOutput(workdir string, helper *test.GitHelper, testOutput string) {
	Expect(helper.Init()).NotTo(HaveOccurred())

	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh"}`},
		{Name: "test.sh", Content: "#!/usr/bin/env bash\necho'" + testOutput + "'\nexit 0"},
	})

	Expect(helper.Commit()).NotTo(HaveOccurred())
}

func givenATestCommandThatNeedsArguments(gitHelper *test.GitHelper, workdir string) {
	Expect(gitHelper.Init()).NotTo(HaveOccurred())
	givenUnstangedChanges(workdir, test.Files{
		{Name: "tcr.json", Content: `{"test": "./test.sh argument1 argument2"}`},
		{Name: "test.sh", Content: `#!/usr/bin/env bash
[[ "$1" == 'argument1' ]]
[[ "$2" == 'argument2' ]]`},
	})
	Expect(gitHelper.Commit()).NotTo(HaveOccurred())
}

func givenAnyUnstagedChanges(workdir string) {
	givenUnstangedChanges(workdir, test.Files{{Name: aFileName, Content: aContent}})
}

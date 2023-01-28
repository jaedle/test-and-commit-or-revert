package test_test

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
	"os/exec"
	"path"
)

var _ = Describe("Workflow", Ordered, func() {
	var binary string
	var testDirectory string

	BeforeAll(func() {
		var err error
		binary, err = gexec.Build("github.com/jaedle/test-and-commit-or-revert/cmd/cli")
		Expect(err).ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		dir, err := os.MkdirTemp(os.TempDir(), "workflow-")
		Expect(err).NotTo(HaveOccurred())
		testDirectory = dir
	})

	AfterEach(func() {
		//_ = os.RemoveAll(testDirectory)
	})

	AfterAll(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("happy path", func() {
		It("succeeds if run within a git repository", func() {
			repo, err := git.PlainInit(testDirectory, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(repo.CreateBranch(&config.Branch{
				Name: "main",
			})).NotTo(HaveOccurred())
			Expect(os.WriteFile(path.Join(testDirectory, ".test"), nil, os.ModePerm)).NotTo(HaveOccurred())

			worktree, err := repo.Worktree()
			Expect(err).NotTo(HaveOccurred())

			_, err = worktree.Add(".test")
			Expect(err).NotTo(HaveOccurred())
			_, err = worktree.Commit("amazing", &git.CommitOptions{
				All:               false,
				AllowEmptyCommits: false,
				Author:            nil,
				Committer:         nil,
				Parents:           nil,
				SignKey:           nil,
			})
			Expect(err).NotTo(HaveOccurred())

			println(testDirectory)

			cmd := exec.Command(binary)
			cmd.Dir = testDirectory

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("error cases", func() {
		It("fails if not run within a git repository", func() {
			cmd := exec.Command(binary)
			cmd.Dir = testDirectory

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Expect(exec.Command("git", "init").Run()).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
		})
	})

	It("exits with zero exit code", func() {
		session, err := gexec.Start(exec.Command(binary), GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
	})
})

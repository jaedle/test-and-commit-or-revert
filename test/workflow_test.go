package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)

var _ = Describe("Workflow", Ordered, func() {
	var binary string

	BeforeAll(func() {
		var err error
		binary, err = gexec.Build("github.com/jaedle/test-and-commit-or-revert/cmd/cli")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterAll(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("exits with zero exit code", func() {

		session, err := gexec.Start(exec.Command(binary), GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
	})
})

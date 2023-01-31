package test_test

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)

type tcrOutput struct {
	exitCode int
	stdOut   string
	stdErr   string
}

func whenIRunTcr(binary string, workdir string) tcrOutput {
	cmd := exec.Command(binary)
	cmd.Dir = workdir

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	session, err := gexec.Start(cmd, &stdOut, &stdErr)
	Expect(err).NotTo(HaveOccurred())
	session.Wait()

	_, _ = GinkgoWriter.Write(stdErr.Bytes())
	_, _ = GinkgoWriter.Write(stdOut.Bytes())

	return tcrOutput{
		exitCode: session.ExitCode(),
		stdOut:   stdOut.String(),
		stdErr:   stdErr.String(),
	}
}

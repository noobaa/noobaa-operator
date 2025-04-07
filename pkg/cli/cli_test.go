package cli

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"
)

const (
	CLIPath = "../../build/_output/bin/noobaa-operator-local"
)

var _ = Describe("CLI tests", func() {

	os.Setenv("TEST_ENV", "true")

	Context("Noobaa CLI functions", func() {
		It("CLI with no arguments", func() {
			cmd := exec.Command(CLIPath)
			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil())
			Expect(ExpectMessage(string(out),
				`.*?`, `Install:`, `\s*\n`,
				`.*?`, `Manage:`, `\s*\n`,
				`.*?`, `Advanced:`, `\s*\n`,
				`.*?`, `Use "noobaa <command> --help" for more information about a given command.`, `\s*\n`,
				`.*?`)).To(BeTrue())
		})

		It("CLI with version", func() {
			cmd := exec.Command(CLIPath, "version")
			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil())
			Expect(ExpectMessage(string(out),
				`.*?`, `CLI version: `, version.Version,
				`.*?`)).To(BeTrue())
		})

		It("CLI with Options", func() {
			cmd := exec.Command(CLIPath, "options")
			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil())
			Expect(ExpectMessage(string(out),
				`.*?`, `The following options can be passed to any command:`,
				`.*?`)).To(BeTrue())
		})
	})

	Context("Noobaa namspacestore CLI", func() {
		It("S3Compatible namespacestore - signature-version validation", func() {
			cmd := exec.Command(CLIPath, "namespacestore", "create", "s3-compatible", "s3compatible-nss", "--endpoint", "http://localhost.com",
				"--access-key", "2EA1ZBLabcd123", "--secret-key", "XL2zyHTYElKuxTiCwTabcd1234",
				"--target-bucket", "test-bucket", "--signature-version", "v4", "-n", "noobaa")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(BeNil())
			Expect(ExpectMessage(string(out),
				`.*?`, `Non-secure endpoint works only with v2 signature-version. Please select signature version v2 for namespacestore`,
				`.*?`)).To(BeTrue())
		})
	})
})

func RunCLI(args ...string) (string, error) {
	cmd := exec.Command(CLIPath, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ExpectMessage(out string, re ...string) bool {
	log := util.Logger()
	fullRE := regexp.MustCompile("(?s)^" + strings.Join(re, "") + "$")
	if !fullRE.MatchString(out) {
		log.Infof("Command output %s and regexp do not match", out)
		return false
	}
	return true
}

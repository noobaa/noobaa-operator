package cli

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/noobaa/noobaa-operator/v5/version"
)

const (
	CLIPath = "../../build/_output/bin/noobaa-operator-local"
)

func TestNoArgs(t *testing.T) {
	out := RunCLI(t)
	Expect(t, out,
		`.*?`, `Install:`, `\s*\n`,
		`.*?`, `Manage:`, `\s*\n`,
		`.*?`, `Advanced:`, `\s*\n`,
		`.*?`, `Use "noobaa <command> --help" for more information about a given command.`, `\s*\n`,
		`.*?`)
}

func TestVersion(t *testing.T) {
	out := RunCLI(t, "version")
	Expect(t, out,
		`.*?`, `CLI version: `, version.Version,
		`.*?`)
}

func TestOptions(t *testing.T) {
	out := RunCLI(t, "options")
	Expect(t, out,
		`.*?`, `The following options can be passed to any command:`,
		`.*?`)
}

func RunCLI(t *testing.T, args ...string) string {
	cmd := exec.Command(CLIPath, args...) // #nosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	return string(out)
}

func Expect(t *testing.T, out string, re ...string) {
	fullRE := regexp.MustCompile("(?s)^" + strings.Join(re, "") + "$")
	if !fullRE.MatchString(out) {
		t.Fatalf("Expected output not matching regexp %q %q", fullRE, out)
	}
}

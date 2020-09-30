package system

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/spf13/cobra"
)

// CmdUI returns a CLI command
func CmdUI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Open the NooBaa UI",
		Run:   RunUI,
	}
	return cmd
}

// RunUI runs a CLI command
func RunUI(cmd *cobra.Command, args []string) {
	log := util.Logger()

	sysClient, err := Connect(true)
	if err != nil {
		log.Fatalf("❌ %s", err)
	}

	mgmtURL := sysClient.MgmtURL.String()

	fmt.Printf("\n")
	fmt.Printf("NooBaa UI (credentials unless using Openshift SSO):\n")
	fmt.Printf("url      : %s\n", mgmtURL)
	fmt.Printf("email    : %s\n", sysClient.SecretAdmin.StringData["email"])
	fmt.Printf("password : %s\n", sysClient.SecretAdmin.StringData["password"])
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("---> NOTE: Keep this process running while using the UI ...")
	fmt.Printf("---> <Ctrl-C> to stop")
	fmt.Printf("\n")
	fmt.Printf("\n")

	err = OpenURLInBrowser(mgmtURL)
	if err != nil {
		log.Fatalf("Encountered error when trying to open management in the browser. %v", err)
	}
	stopChan := make(chan int)
	<-stopChan
}

// OpenURLInBrowser opens the url in a browser according to the current OS
func OpenURLInBrowser(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Start()
	case "linux":
		return exec.Command("xdg-open", u).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

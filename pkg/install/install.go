package install

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v2/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/v2/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v2/pkg/crd"
	"github.com/noobaa/noobaa-operator/v2/pkg/obc"
	"github.com/noobaa/noobaa-operator/v2/pkg/operator"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/pkg/version"
	"github.com/spf13/cobra"
)

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the operator and create the noobaa system",
		Run:   RunInstall,
	}
	cmd.Flags().Bool("use-obc-cleanup-policy", false, "Create NooBaa system with obc cleanup policy")
	return cmd
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the operator and delete the system",
		Run:   RunUninstall,
	}
	cmd.Flags().Bool("cleanup", false, "Enable deletion of Namespace and CRD's")
	cmd.Flags().Bool("cleanup_data", false, "Clean object buckets")
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of the operator and the system",
		Run:   RunStatus,
	}
	return cmd
}

// RunInstall runs a CLI command
func RunInstall(cmd *cobra.Command, args []string) {
	log := util.Logger()
	version.RunVersion(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("CRD Create:")
	crd.RunCreate(cmd, args)
	log.Printf("")
	log.Printf("Operator Install:")
	operator.RunInstall(cmd, args)
	log.Printf("")
	log.Printf("System Create:")
	system.RunCreate(cmd, args)
	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("System Wait Ready:")
	if system.WaitReady() {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	log := util.Logger()
	system.RunSystemVersionsStatus(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("System Delete:")
	system.RunDelete(cmd, args)
	log.Printf("")
	log.Printf("Operator Delete:")
	operator.RunUninstall(cmd, args)
	log.Printf("")
	cleanup, _ := cmd.Flags().GetBool("cleanup")
	if cleanup {
		log.Printf("CRD Delete:")
		crd.RunDelete(cmd, args)
	} else {
		log.Printf("CRD Delete: currently disabled (enable with \"--cleanup\")")
		log.Printf("CRD Status:")
		crd.RunStatus(cmd, args)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	system.RunSystemVersionsStatus(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("CRD Status:")
	crd.RunStatus(cmd, args)
	log.Printf("")
	log.Printf("Operator Status:")
	operator.RunStatus(cmd, args)
	log.Printf("")
	log.Printf("System Wait Ready:")
	if system.WaitReady() {
		log.Printf("")
		log.Printf("")
	}
	log.Printf("System Status:")
	system.RunStatus(cmd, args)

	fmt.Println("#------------------#")
	fmt.Println("#- Backing Stores -#")
	fmt.Println("#------------------#")
	fmt.Println("")
	backingstore.RunList(cmd, args)
	fmt.Println("")

	fmt.Println("#------------------#")
	fmt.Println("#- Bucket Classes -#")
	fmt.Println("#------------------#")
	fmt.Println("")
	bucketclass.RunList(cmd, args)
	fmt.Println("")

	fmt.Println("#-----------------#")
	fmt.Println("#- Bucket Claims -#")
	fmt.Println("#-----------------#")
	fmt.Println("")
	obc.RunList(cmd, args)
	fmt.Println("")
}

package install

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/pkg/crd"
	"github.com/noobaa/noobaa-operator/pkg/obc"
	"github.com/noobaa/noobaa-operator/pkg/operator"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/noobaa/noobaa-operator/pkg/version"
	"github.com/spf13/cobra"
)

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the operator and create the noobaa system",
		Run:   RunInstall,
	}
	return cmd
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the operator and delete the system",
		Run:   RunUninstall,
	}
	cmd.Flags().Bool("crd", false, "Enable deletion of CRD's")
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
	log.Printf("System Wait:")
	if system.WaitReady() {
		RunStatus(cmd, args)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	log := util.Logger()
	version.RunVersion(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("System Delete:")
	system.RunDelete(cmd, args)
	log.Printf("")
	log.Printf("Operator Delete:")
	operator.RunUninstall(cmd, args)
	log.Printf("")
	crdDelete, _ := cmd.Flags().GetBool("crd")
	if crdDelete {
		log.Printf("CRD Delete:")
		crd.RunDelete(cmd, args)
	} else {
		log.Printf("CRD Delete: currently disabled (enable with \"--crd\")")
		log.Printf("CRD Status:")
		crd.RunStatus(cmd, args)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	version.RunVersion(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("CRD Status:")
	crd.RunStatus(cmd, args)
	log.Printf("")
	log.Printf("Operator Status:")
	operator.RunStatus(cmd, args)
	log.Printf("")
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

package install

import (
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/crd"
	"github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/noobaaaccount"
	"github.com/noobaa/noobaa-operator/v5/pkg/obc"
	"github.com/noobaa/noobaa-operator/v5/pkg/operator"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
)

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the operator and create the noobaa system",
		Run:   RunInstall,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().Bool("use-obc-cleanup-policy", false, "Create NooBaa system with obc cleanup policy")
	cmd.AddCommand(
		CmdYaml(),
	)
	return cmd
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show install yaml, expected usage \"noobaa install 1> install.yaml\"",
		Run:   RunYaml,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunYaml dumps a combined yaml of all installation yaml
// including CRD, operator and system
func RunYaml(cmd *cobra.Command, args []string) {
	log := util.Logger()
	log.Println("Dump CRD yamls...")
	crd.RunYaml(cmd, args)
	fmt.Println("---") // yaml resources separator
	log.Println("Dump operator yamls...")
	operator.RunYaml(cmd, args)
	fmt.Println("---") // yaml resources separator
	log.Println("Dump system yamls...")
	system.RunYaml(cmd, args)
	log.Println("✅ Done dumping installation yaml")
}

// CmdUpgrade returns a CLI command
func CmdUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade --noobaa-image <noobaa-image-path-and-tag> --operator-image <operator-image-path-and-tag>",
		Short: "Upgrade the system, its components and CRDS",
		Long:  "The command should be used in conjunction with the global flags --noobaa-image and " + 
		"--operator-image to upgrade the system and its components to the desired versions.",
		Run:   RunUpgrade,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the operator and delete the system",
		Run:   RunUninstall,
		Args:  cobra.NoArgs,
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
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunInstall runs a CLI command
func RunInstall(cmd *cobra.Command, args []string) {
	log := util.Logger()
	system.RunSystemVersionsStatus(cmd, args)
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

// RunUpgrade runs a CLI command
func RunUpgrade(cmd *cobra.Command, args []string) {
	log := util.Logger()
	log.Printf("System versions prior to upgrade:\n")
	system.RunSystemVersionsStatus(cmd, args)
	log.Printf("Namespace: %s\n", options.Namespace)
	log.Printf("CRD upgrade:")
	crd.RunUpgrade(cmd, args)
	log.Printf("\nOperator upgrade:")
	operator.RunUpgrade(cmd, args)
	log.Printf("\nSystem apply:")
	system.RunUpgrade(cmd, args)
	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("\nWaiting for the system to be ready...")
	// Sleep to let the system get out of its old Ready state
	time.Sleep(3 * time.Second)
	if system.WaitReady() {
		log.Printf("\n\n")
		RunStatus(cmd, args)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	log := util.Logger()
	cleanup, _ := cmd.Flags().GetBool("cleanup")

	if cleanup {
		var decision string

		for {
			log.Printf("--cleanup removes the CRDs and affecting all noobaa instances, are you sure? y/n ")
			fmt.Scanln(&decision)

			if decision == "y" {
				log.Printf("Will remove CRD (cluster scope)")
				break
			} else if decision == "n" {
				return
			}
		}
	}

	system.RunSystemVersionsStatus(cmd, args)
	log.Printf("Namespace: %s", options.Namespace)
	log.Printf("")
	log.Printf("System Delete:")
	system.RunDelete(cmd, args)
	log.Printf("")
	log.Printf("Operator Delete:")
	operator.RunUninstall(cmd, args)
	log.Printf("")
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

	fmt.Println("#--------------------#")
	fmt.Println("#- Namespace Stores -#")
	fmt.Println("#--------------------#")
	fmt.Println("")
	namespacestore.RunList(cmd, args)
	fmt.Println("")

	fmt.Println("#------------------#")
	fmt.Println("#- Bucket Classes -#")
	fmt.Println("#------------------#")
	fmt.Println("")
	bucketclass.RunList(cmd, args)
	fmt.Println("")

	fmt.Println("#-------------------#")
	fmt.Println("#- NooBaa Accounts -#")
	fmt.Println("#-------------------#")
	fmt.Println("")
	noobaaaccount.RunList(cmd, args)
	fmt.Println("")

	fmt.Println("#-----------------#")
	fmt.Println("#- Bucket Claims -#")
	fmt.Println("#-----------------#")
	fmt.Println("")
	obc.RunList(cmd, args)
	fmt.Println("")
}

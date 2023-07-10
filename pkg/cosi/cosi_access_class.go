package cosi

import (
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

// CmdCOSIBucketAccessClass returns a CLI command
func CmdCOSIBucketAccessClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accessclass",
		Short: "Manage cosi access class",
	}
	cmd.AddCommand(
		CmdCreateAccessClass(),
		CmdDeleteAccessClass(),
		CmdStatusAccessClass(),
		CmdListAccessClass(),
	)
	return cmd
}

// CmdCreateAccessClass returns a CLI command
func CmdCreateAccessClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <access-class-name>",
		Short: "Create a COSI access class",
		Run:   RunCreateAccessClass,
	}
	// AuthenticationType - valid types are KEY / IAM - currently the only supported type is KEY
	// Parameters - currently no extra parameters are supported
	return cmd
}

// CmdDeleteAccessClass returns a CLI command
func CmdDeleteAccessClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <access-class-name>",
		Short: "Delete a COSI access class",
		Run:   RunDeleteAccessClass,
	}

	return cmd
}

// CmdStatusAccessClass returns a CLI command
func CmdStatusAccessClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <access-class-name>",
		Short: "Status of a COSI access class",
		Run:   RunStatusAccessClass,
	}
	return cmd
}

// CmdListAccessClass returns a CLI command
func CmdListAccessClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List COSI access classes",
		Run:   RunListAccessClass,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreateAccessClass runs a CLI command
func RunCreateAccessClass(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <access-class-name> %s`, cmd.UsageString())
	}
	name := args[0]

	cosiAccessClass := util.KubeObject(bundle.File_deploy_cosi_bucket_access_class_yaml).(*nbv1.COSIBucketAccessClass)
	cosiAccessClass.Name = name
	cosiAccessClass.DriverName = options.COSIDriverName()
	cosiAccessClass.AuthenticationType = nbv1.COSIKEYAuthenticationType

	if !util.KubeCreateFailExisting(cosiAccessClass) {
		log.Fatalf(`❌ Could not create COSI access class %q (conflict)`, cosiAccessClass.Name)
	}

	log.Printf("")
	log.Printf("")
	log.Printf("")
	RunStatusAccessClass(cmd, args)
}

// RunDeleteAccessClass runs a CLI command
func RunDeleteAccessClass(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <access-class-name> %s`, cmd.UsageString())
	}

	cosiAccessClass := util.KubeObject(bundle.File_deploy_cosi_bucket_access_class_yaml).(*nbv1.COSIBucketAccessClass)
	cosiAccessClass.Name = args[0]

	if !util.KubeDelete(cosiAccessClass) {
		log.Fatalf(`❌ Could not delete COSI access class %q `, cosiAccessClass.Name)
	}
}

// RunStatusAccessClass runs a CLI command
func RunStatusAccessClass(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <access-class-name> %s`, cmd.UsageString())
	}

	cosiAccessClass := util.KubeObject(bundle.File_deploy_cosi_bucket_access_class_yaml).(*nbv1.COSIBucketAccessClass)
	cosiAccessClass.Name = args[0]

	if !util.KubeCheck(cosiAccessClass) {
		log.Fatalf(`❌ Could not find COSI access class %q`, cosiAccessClass.Name)
	}

	fmt.Println()
	fmt.Println("# AccessClass spec:")
	fmt.Printf("Name:\n %s\n", cosiAccessClass.Name)
	fmt.Printf("Driver Name:\n %s\n", cosiAccessClass.DriverName)
	fmt.Printf("Authentication Type:\n %+v", cosiAccessClass.AuthenticationType)
	fmt.Println()
}

// RunListAccessClass runs a CLI command
func RunListAccessClass(cmd *cobra.Command, args []string) {
	list := &nbv1.COSIBucketAccessClassList{
		TypeMeta: metav1.TypeMeta{Kind: "AccessClass"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No COSI access classes found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAME",
		"DRIVER-NAME",
		"AUTHENTICATION-TYPE",
		"AGE",
	)
	for i := range list.Items {
		cosiAccessClass := &list.Items[i]
		table.AddRow(
			cosiAccessClass.Name,
			cosiAccessClass.DriverName,
			string(cosiAccessClass.AuthenticationType),
			util.HumanizeDuration(time.Since(cosiAccessClass.CreationTimestamp.Time).Round(time.Second)),
		)
	}
	fmt.Print(table.String())
}

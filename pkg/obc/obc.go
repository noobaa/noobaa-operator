package obc

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/pkg/options"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cmd creates a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obc",
		Short: "Manage bucket claims",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdList(),
	)
	return cmd
}

// CmdCreate creates a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-name>",
		Short: "Create an OBC",
		Run:   RunCreate,
	}
	return cmd
}

// CmdDelete creates a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-name>",
		Short: "Delete an OBC",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList creates a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBC's",
		Run:   RunList,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_obc_cr_yaml)
	obc := o.(*nbv1.ObjectBucketClaim)
	obc.Name = args[0]
	obc.Namespace = options.Namespace
	obc.Spec.BucketName = args[0]
	obc.Spec.StorageClassName = options.Namespace
	obc.Spec.SSL = false
	obc.Spec.Versioned = false

	if !util.KubeCreateSkipExisting(obc) {
		log.Fatalf(`❌ Could not OBC %q in namespace %q (conflict)`, obc.Name, obc.Namespace)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_obc_cr_yaml)
	obc := o.(*nbv1.ObjectBucketClaim)
	obc.Name = args[0]
	obc.Namespace = options.Namespace

	if !util.KubeDelete(obc) {
		log.Fatalf(`❌ Could not delete OBC %q in namespace %q`,
			obc.Name, obc.Namespace)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.ObjectBucketClaimList{
		TypeMeta: metav1.TypeMeta{Kind: "ObjectBucketClaim"},
	}
	if !util.KubeList(list, nil) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No OBC's found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"BUCKET-NAME",
		"STORAGE-CLASS",
		"PHASE",
	)
	for i := range list.Items {
		obc := &list.Items[i]
		table.AddRow(
			obc.Namespace,
			obc.Name,
			obc.Spec.BucketName,
			obc.Spec.StorageClassName,
			string(obc.Status.Phase),
		)
	}
	fmt.Print(table.String())
}

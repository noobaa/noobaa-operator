package bucket

import (
	"fmt"
	"log"

	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

	"github.com/spf13/cobra"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket",
		Short: "Manage noobaa buckets",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdList(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-name>",
		Short: "Create a NooBaa bucket",
		Run:   RunCreate,
	}
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-name>",
		Short: "Delete a NooBaa bucket",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List NooBaa buckets",
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
	bucketName := args[0]
	nbClient := system.GetNBClient()
	err := nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName})
	if err != nil {
		log.Fatal(err)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketName := args[0]
	nbClient := system.GetNBClient()
	err := nbClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: bucketName})
	if err != nil {
		log.Fatal(err)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	nbClient := system.GetNBClient()
	list, err := nbClient.ListBucketsAPI()
	if err != nil {
		log.Fatal(err)
	}
	if len(list.Buckets) == 0 {
		fmt.Printf("No buckets found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow("BUCKET-NAME")
	for i := range list.Buckets {
		b := &list.Buckets[i]
		table.AddRow(b.Name)
	}
	fmt.Print(table.String())
}

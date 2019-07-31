package bucket

import (
	"fmt"
	"strconv"

	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/spf13/cobra"
)

// Cmd creates a CLI command
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

// CmdCreate creates a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-name>",
		Short: "Create a NooBaa bucket",
		Run:   RunCreate,
	}
	return cmd
}

// CmdDelete creates a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-name>",
		Short: "Delete a NooBaa bucket",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList creates a CLI command
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
	_, err := nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName})
	util.Panic(err)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketName := args[0]
	nbClient := system.GetNBClient()
	_, err := nbClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: bucketName})
	util.Panic(err)
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	nbClient := system.GetNBClient()
	list, err := nbClient.ListBucketsAPI()
	util.Panic(err)
	if len(list.Buckets) == 0 {
		fmt.Printf("No buckets found.\n")
	}
	table := (&util.PrintTable{}).AddRow("#", "BUCKET-NAME")
	for i := range list.Buckets {
		b := &list.Buckets[i]
		table.AddRow(strconv.Itoa(i+1), b.Name)
	}
	fmt.Print(table.String())
}

package bucket

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

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
		CmdStatus(),
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

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-name>",
		Short: "Show the status of a NooBaa bucket",
		Run:   RunStatus,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List NooBaa buckets",
		Run:   RunList,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreate runs a CLI command. The default backingstore will be used as the underlying storage for buckets created using the CLI
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketName := args[0]
	nbClient := system.GetNBClient()

	bucketClass, err := bucketclass.GetDefaultBucketClass(options.Namespace)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to get default bucketclass with error: %v", err))
	}

	tierName, err := bucketclass.CreateTieringStructure(*bucketClass, bucketName, nbClient)
	if err != nil {
		log.Fatal(fmt.Errorf("CreateTieringStructure for PlacementPolicy failed to create policy %q with error: %v", tierName, err))
	}

	err = nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName, Tiering: tierName})
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

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketName := args[0]
	nbClient := system.GetNBClient()
	b, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: bucketName})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n")
	fmt.Printf("Bucket status:\n")
	fmt.Printf("  %-22s : %s\n", "Bucket", b.Name)
	if b.BucketClaim != nil {
		fmt.Printf("  %-22s : %s\n", "OBC Namespace", b.BucketClaim.Namespace)
		fmt.Printf("  %-22s : %s\n", "OBC BucketClass", b.BucketClaim.BucketClass)
	}
	fmt.Printf("  %-22s : %s\n", "Type", b.BucketType)
	fmt.Printf("  %-22s : %s\n", "Mode", b.Mode)
	if b.PolicyModes != nil {
		fmt.Printf("  %-22s : %s\n", "ResiliencyStatus", b.PolicyModes.ResiliencyStatus)
		fmt.Printf("  %-22s : %s\n", "QuotaStatus", b.PolicyModes.QuotaStatus)
	}
	if b.Undeletable != "" {
		fmt.Printf("  %-22s : %s\n", "Undeletable", b.Undeletable)
	}
	if b.NumObjects != nil {
		if b.BucketType == "NAMESPACE" {
			fmt.Printf("  %-22s : N/A\n", "Num Objects")
		} else {
			fmt.Printf("  %-22s : %d\n", "Num Objects", b.NumObjects.Value)
		}
	}
	if b.DataCapacity != nil {
		if b.BucketType == "NAMESPACE" {
			fmt.Printf("  %-22s : N/A\n", "Data Size")
			fmt.Printf("  %-22s : N/A\n", "Data Size Reduced")
			fmt.Printf("  %-22s : N/A\n", "Data Space Avail")
		} else {
			fmt.Printf("  %-22s : %s\n", "Data Size", nb.BigIntToHumanBytes(b.DataCapacity.Size))
			fmt.Printf("  %-22s : %s\n", "Data Size Reduced", nb.BigIntToHumanBytes(b.DataCapacity.SizeReduced))
			fmt.Printf("  %-22s : %s\n", "Data Space Avail", nb.BigIntToHumanBytes(b.DataCapacity.AvailableSizeToUpload))
			fmt.Printf("  %-22s : %s\n", "Num Objects Avail", b.DataCapacity.AvailableQuantityToUpload.ToString())
		}
	}
	fmt.Printf("\n")
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	log := util.Logger()
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
	fmt.Printf("\n")
	fmt.Print(table.String())
	fmt.Printf("\n")
}

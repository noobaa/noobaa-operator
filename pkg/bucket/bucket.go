package bucket

import (
	"fmt"
	"strconv"

	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"

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
		CmdUpdate(),
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
	cmd.Flags().Bool("force_md5_etag", false, "This flag enables md5 etag calculation for bucket")
	cmd.Flags().String("max-objects", "",
		"Set quota max objects quantity config to requested bucket")
	cmd.Flags().String("max-size", "",
		"Set quota max size config to requested bucket")
	return cmd
}

// CmdUpdate returns a CLI command
func CmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <bucket-name>",
		Short: "Update a NooBaa bucket",
		Run:   RunUpdate,
	}
	cmd.Flags().Bool("force_md5_etag", false, "This flag enables md5 etag calculation for bucket")
	cmd.Flags().String("max-objects", "",
		"Set quota max objects quantity config to requested bucket")
	cmd.Flags().String("max-size", "",
		"Set quota max size config to requested bucket")
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

	tierName, err := bucketclass.CreateTieringStructure(*bucketClass.Spec.PlacementPolicy, bucketName, nbClient)
	if err != nil {
		log.Fatal(fmt.Errorf("CreateTieringStructure for PlacementPolicy failed to create policy %q with error: %v", tierName, err))
	}

	forceMd5EtagPtr, err := util.GetBoolFlagPtr(cmd, "force_md5_etag")
	if err != nil {
		log.Fatal(err)
	}
	maxSize, _ := cmd.Flags().GetString("max-size")
	maxObjects, _ := cmd.Flags().GetString("max-objects")

	quota, err := prepareQuotaConfig(bucketName, maxSize, maxObjects)
	if err != nil {
		log.Fatalf(`❌ Could not create bucket "%q" quota validation failed %q`, bucketName, err)
	}

	err = nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName, Tiering: tierName, ForceMd5Etag: forceMd5EtagPtr})
	if err != nil {
		log.Fatal(err)
	}
	// calling updateBucketAPI to update quota for bucket
	err = nbClient.UpdateBucketAPI(nb.CreateBucketParams{Name: bucketName, Quota: &quota})
	if err != nil {
		log.Fatal(err)
	}
}

// RunUpdate runs a CLI command
func RunUpdate(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketName := args[0]
	nbClient := system.GetNBClient()
	forceMd5EtagPtr, err := util.GetBoolFlagPtr(cmd, "force_md5_etag")
	if err != nil {
		log.Fatal(err)
	}
	maxSize, _ := cmd.Flags().GetString("max-size")
	maxObjects, _ := cmd.Flags().GetString("max-objects")

	updateParams := nb.CreateBucketParams{
		Name:         bucketName,
		ForceMd5Etag: forceMd5EtagPtr,
	}

	if maxSize != "" || maxObjects != "" {
		quota, err := prepareQuotaConfig(bucketName, maxSize, maxObjects)
		if err != nil {
			log.Fatalf(`❌ Could not update bucket "%q" quota validation failed %q`, bucketName, err)
		}
		updateParams.Quota = &quota
	}

	err = nbClient.UpdateBucketAPI(updateParams)
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
	if b.ForceMd5Etag != nil {
		fmt.Printf("  %-22s : %t\n", "Force Md5 Etag", *b.ForceMd5Etag)
	}
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
	list, err := nbClient.ListBucketsAPI(nb.ListBucketsParams{})
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

func prepareQuotaConfig(bucketName string, maxSize string, maxObjects string) (nb.QuotaConfig, error) {
	var bucketMaxSize, bucketMaxObjects int64
	quota := nb.QuotaConfig{}

	// validate quota config for bucket
	if err := util.ValidateQuotaConfig(bucketName, maxSize, maxObjects); err != nil {
		return quota, err
	}

	if maxSize != "" {
		quantity, _ := resource.ParseQuantity(maxSize)
		bucketMaxSize = quantity.Value()
		if bucketMaxSize > 0 {
			f, u := nb.GetBytesAndUnits(bucketMaxSize, 2)
			quota.Size = &nb.SizeQuotaConfig{Value: f, Unit: u}
		}
	}
	if maxObjects != "" {
		bucketMaxObjects, _ = strconv.ParseInt(maxObjects, 10, 64)
		if bucketMaxObjects > 0 {
			quota.Quantity = &nb.QuantityQuotaConfig{Value: int(bucketMaxObjects)}
		}
	}

	return quota, nil
}

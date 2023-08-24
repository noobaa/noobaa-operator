package cosi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CmdCOSIBucketClaim returns a CLI command
func CmdCOSIBucketClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucketclaim",
		Short: "Manage cosi bucket claims",
	}
	cmd.AddCommand(
		CmdCreateBucketClaim(),
		CmdDeleteBucketClaim(),
		CmdStatusBucketClaim(),
		CmdListBucketClaim(),
	)
	return cmd
}

// CmdCreateBucketClaim returns a CLI command
func CmdCreateBucketClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-claim-name>",
		Short: "Create a COSI bucket claim",
		Run:   RunCreateBucketClaim,
	}

	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket claim should be created")
	cmd.Flags().String("bucketclass", "",
		"Set bucket class to specify the bucket policy")
	cmd.Flags().String("path", "",
		"Set path to specify inner directory in namespace store target path - can be used only while specifing a namespace bucketclass")
	return cmd
}

// CmdDeleteBucketClaim returns a CLI command
func CmdDeleteBucketClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-claim-name>",
		Short: "Delete a COSI bucket claim",
		Run:   RunDeleteBucketClaim,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket claim should be created")
	return cmd
}

// CmdStatusBucketClaim returns a CLI command
func CmdStatusBucketClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-claim-name>",
		Short: "Status of a COSI bucket claim",
		Run:   RunStatusBucketClaim,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket claim should be created")
	return cmd
}

// CmdListBucketClaim returns a CLI command
func CmdListBucketClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List COSI bucket claims",
		Run:   RunListBucketClaim,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreateBucketClaim runs a CLI command
func RunCreateBucketClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}
	name := args[0]

	bucketClassName := util.GetFlagStringOrPrompt(cmd, "bucketclass")
	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	path, _ := cmd.Flags().GetString("path")

	cosiBucketClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_claim_yaml).(*nbv1.COSIBucketClaim)
	cosiBucketClaim.Name = name
	cosiBucketClaim.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketClaim.Namespace = appNamespace
	}
	if path != "" {
		cosiBucketClaim.Labels = map[string]string{"path": path}
	}
	cosiBucketClaim.Spec.Protocols = []nbv1.COSIProtocol{nbv1.COSIS3Protocol}

	bucketClass := util.KubeObject(bundle.File_deploy_cosi_bucket_class_yaml).(*nbv1.COSIBucketClass)
	bucketClass.Name = bucketClassName
	if !util.KubeCheck(bucketClass) {
		log.Fatalf(`❌ Could not get BucketClass %q in namespace %q`,
			bucketClass.Name, bucketClass.Namespace)
	}

	cosiBucketClaim.Spec.BucketClassName = bucketClassName
	bucketClassSpec, errMsg := CreateBucketClassSpecFromParameters(bucketClass.Parameters)
	if errMsg != "" {
		log.Fatalf(`❌ Could not create BucketClass spec out of bucketclass parameters %q %q`, bucketClass.Name, bucketClass.Parameters)
	}
	err := ValidateCOSIBucketClaim(cosiBucketClaim.Name, options.Namespace, *bucketClassSpec, true)
	if err != nil {
		log.Fatalf(`❌ Could not validate COSI bucket claim %q in namespace %q validation failed %q`, cosiBucketClaim.Name, cosiBucketClaim.Namespace, err)
	}

	if !util.KubeCreateFailExisting(cosiBucketClaim) {
		log.Fatalf(`❌ Could not create COSI bucket claim %q in namespace %q (conflict)`, cosiBucketClaim.Name, cosiBucketClaim.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("COSI bucket claim Wait Ready:")
	if WaitReady(cosiBucketClaim) {
		log.Printf("")
		log.Printf("")
		RunStatusBucketClaim(cmd, args)
	}
}

// RunDeleteBucketClaim runs a CLI command
func RunDeleteBucketClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}
	appNamespace, _ := cmd.Flags().GetString("app-namespace")

	cosiBucketClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_claim_yaml).(*nbv1.COSIBucketClaim)
	cosiBucketClaim.Name = args[0]
	cosiBucketClaim.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketClaim.Namespace = appNamespace
	}

	if !util.KubeDelete(cosiBucketClaim) {
		log.Fatalf(`❌ Could not delete COSI bucket claim %q in namespace %q`,
			cosiBucketClaim.Name, cosiBucketClaim.Namespace)
	}
}

// RunStatusBucketClaim runs a CLI command
func RunStatusBucketClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}

	appNamespace, _ := cmd.Flags().GetString("app-namespace")

	cosiBucketClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_claim_yaml).(*nbv1.COSIBucketClaim)
	cosiBucket := util.KubeObject(bundle.File_deploy_cosi_cosi_bucket_yaml).(*nbv1.COSIBucket)

	cosiBucketClaim.Name = args[0]
	cosiBucketClaim.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketClaim.Namespace = appNamespace
	}

	if !util.KubeCheck(cosiBucketClaim) {
		log.Fatalf(`❌ Could not find COSI bucket claim %q in namespace %q`, cosiBucketClaim.Name, cosiBucketClaim.Namespace)
	}

	cosiBucket.Name = cosiBucketClaim.Status.BucketName
	if !util.KubeCheck(cosiBucket) {
		log.Fatalf(`❌ Could not find COSI bucket %q`, cosiBucket.Name)
	}

	bucketClass := &nbv1.COSIBucketClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cosiBucketClaim.Spec.BucketClassName,
			Namespace: options.Namespace,
		},
	}

	if bucketClass.Name != "" {
		if !util.KubeCheck(bucketClass) {
			log.Errorf(`❌ Could not get BucketClass %q in namespace %q`,
				bucketClass.Name, bucketClass.Namespace)
		}
	}

	sysClient, err := system.Connect(true)
	if err != nil {
		util.Logger().Fatalf("❌ %s", err)
	}
	var b *nb.BucketInfo
	if cosiBucketClaim.Status.BucketName != "" {
		nbClient := sysClient.NBClient
		bucket, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiBucketClaim.Status.BucketName})
		if err == nil {
			b = &bucket
		}
	}

	fmt.Printf("\n")
	fmt.Printf("COSI BucketClaim info:\n")
	fmt.Printf("  %-22s : %t\n", "Bucket Ready", cosiBucketClaim.Status.BucketReady)
	fmt.Printf("  %-22s : kubectl get -n %s bucketclaim %s\n", "COSIBucketClaim", cosiBucketClaim.Namespace, cosiBucketClaim.Name)
	fmt.Printf("  %-22s : kubectl get bucket %s\n", "COSIBucket", cosiBucket.Name)
	fmt.Printf("  %-22s : kubectl get bucketclasses.objectstorage.k8s.io %s\n", "BucketClass", bucketClass.Name)
	fmt.Printf("\n")

	fmt.Printf("Shell commands:\n")
	if b != nil {
		fmt.Printf("Bucket status:\n")
		fmt.Printf("  %-22s : %s\n", "Name", b.Name)
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
			fmt.Printf("  %-22s : %d\n", "Num Objects", b.NumObjects.Value)
		}
		if b.DataCapacity != nil {
			fmt.Printf("  %-22s : %s\n", "Data Size", nb.BigIntToHumanBytes(b.DataCapacity.Size))
			fmt.Printf("  %-22s : %s\n", "Data Size Reduced", nb.BigIntToHumanBytes(b.DataCapacity.SizeReduced))
			fmt.Printf("  %-22s : %s\n", "Data Space Avail", nb.BigIntToHumanBytes(b.DataCapacity.AvailableSizeToUpload))
			fmt.Printf("  %-22s : %s\n", "Num Objects Avail", b.DataCapacity.AvailableQuantityToUpload.ToString())
		}
		fmt.Printf("\n")
	}
}

// RunListBucketClaim runs a CLI command
func RunListBucketClaim(cmd *cobra.Command, args []string) {
	list := &nbv1.COSIBucketClaimList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClaim"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No COSI bucket claims found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"BUCKET-NAME",
		"BUCKET-CLASS",
		"BUCKET-READY",
	)
	for i := range list.Items {
		cosiBucketClaim := &list.Items[i]
		table.AddRow(
			cosiBucketClaim.Namespace,
			cosiBucketClaim.Name,
			cosiBucketClaim.Status.BucketName,
			cosiBucketClaim.Spec.BucketClassName,
			fmt.Sprintf("%t", bool(cosiBucketClaim.Status.BucketReady)),
		)
	}
	fmt.Print(table.String())
}

// WaitReady waits until the cosi bucket claim status bucket ready changes to true
func WaitReady(cosiBucketClaim *nbv1.COSIBucketClaim) bool {
	log := util.Logger()
	klient := util.KubeClient()

	interval := time.Duration(3)
	maxRetries := 60
	retries := 0
	err := wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		if retries == maxRetries {
			return false, fmt.Errorf("COSI bucket claim is not ready after max retries - %q", maxRetries)
		}
		retries++
		err := klient.Get(util.Context(), util.ObjectKey(cosiBucketClaim), cosiBucketClaim)
		if err != nil {
			log.Printf("⏳ Failed to get COSI bucket claim: %s", err)
			return false, nil
		}
		CheckPhase(cosiBucketClaim)
		if cosiBucketClaim.Status.BucketReady {
			return true, nil
		}
		return false, nil
	})
	return (err == nil)
}

// CheckPhase prints the phase and reason for it
func CheckPhase(cosiBucketClaim *nbv1.COSIBucketClaim) {
	log := util.Logger()
	if cosiBucketClaim.Status.BucketReady {
		log.Printf("✅ COSI bucket claim %q Bucket is ready", cosiBucketClaim.Name)
	} else {
		log.Printf("⏳ COSI bucket claim %q is not yet ready", cosiBucketClaim.Name)
	}
}

// CreateBucketClassSpecFromParameters converts Cosi bucket class additional Parameters to BucketClassSpec
func CreateBucketClassSpecFromParameters(parameters map[string]string) (*nbv1.BucketClassSpec, string) {
	log := util.Logger()
	spec := &nbv1.BucketClassSpec{}
	if parameters["placementPolicy"] != "" {
		log.Infof("CreateBucketClassSpecFromParameters: placement policy - %+v %+v", parameters["placementPolicy"], &spec.PlacementPolicy)
		if err := json.Unmarshal([]byte(parameters["placementPolicy"]), &spec.PlacementPolicy); err != nil {
			return nil, fmt.Sprintf("failed to parse placement policy in COSI params %s", parameters["placementPolicy"])
		}
	}
	if parameters["namespacePolicy"] != "" {
		log.Infof("CreateBucketClassSpecFromParameters: namespace policy - %+v", parameters["namespacePolicy"])
		if err := json.Unmarshal([]byte(parameters["namespacePolicy"]), &spec.NamespacePolicy); err != nil {
			return nil, fmt.Sprintf("failed to parse namespace policy in COSI params %s", parameters["namespacePolicy"])
		}
	}
	if parameters["replicationPolicy"] != "" {
		log.Infof("CreateBucketClassSpecFromParameters: replication policy - %+v", parameters["replicationPolicy"])
		if err := json.Unmarshal([]byte(parameters["replicationPolicy"]), &spec.ReplicationPolicy); err != nil {
			return nil, fmt.Sprintf("failed to parse replication policy in COSI params %s", parameters["replicationPolicy"])
		}
	}
	if parameters["quota"] != "" {
		log.Infof("CreateBucketClassSpecFromParameters: quota - %+v", parameters["quota"])
		if err := json.Unmarshal([]byte(parameters["quota"]), &spec.Quota); err != nil {
			return nil, fmt.Sprintf("failed to parse quota policy in COSI params %s", parameters["quota"])
		}
	}
	return spec, ""
}

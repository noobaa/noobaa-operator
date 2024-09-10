package cosi

import (
	"encoding/json"
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CmdCOSIBucketClass returns a CLI command
func CmdCOSIBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucketclass",
		Short: "Manage cosi bucket classes",
	}
	cmd.AddCommand(
		CmdCreateBucketClass(),
		CmdDeleteBucketClass(),
		CmdStatusBucketClass(),
		CmdListBucketClass(),
	)
	return cmd
}

// CmdCreateBucketClass returns a CLI command
func CmdCreateBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a COSI bucket class",
	}

	cmd.AddCommand(
		CmdCreateNamespaceBucketclass(),
		CmdCreatePlacementBucketClass(),
	)

	return cmd
}

// CmdCreatePlacementBucketClass returns a CLI command
func CmdCreatePlacementBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "placement-bucketclass <bucket-class-name>",
		Short: "Create placement policy bucket class",
		Run:   RunCreatePlacementBucketClass,
	}

	// placement policy flags
	cmd.Flags().String("placement", "",
		"Set first tier placement policy - Mirror | Spread | \"\" (empty defaults to single backing store)")
	cmd.Flags().StringSlice("backingstores", nil,
		"Set first tier backing stores (use commas or multiple flags)")
	cmd.Flags().String("replication-policy", "",
		"Set the json file name that contains the replication rules")
	cmd.Flags().String("deletion-policy", "retain",
		"Set the deletion policy, specify how COSI will handle deletion of buckets created on top of the created bucketclass "+
			"- delete | retain - the default deletion policy is retain")
	cmd.Flags().String("max-objects", "",
		"Set quota max objects quantity config to requested bucket")
	cmd.Flags().String("max-size", "",
		"Set quota max size config to requested bucket")

	return cmd
}

// CmdCreateNamespaceBucketclass returns a CLI command
func CmdCreateNamespaceBucketclass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace-bucketclass <bucket-class-name>",
		Short: "Create namespace policy bucket class",
	}

	cmd.AddCommand(
		CmdCreateSingleNamespaceBucketclass(),
		CmdCreateMultiNamespaceBucketclass(),
		CmdCreateCacheNamespaceBucketclass(),
	)

	return cmd
}

// CmdCreateSingleNamespaceBucketclass returns a CLI command
func CmdCreateSingleNamespaceBucketclass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "single <bucket-class-name>",
		Short: "Create namespace bucket class of type Single",
		Run:   RunCreateSingleNamespaceBucketClass,
	}

	// single namespace policy
	cmd.Flags().String("resource", "",
		"Set the namespace read and write resource")
	cmd.Flags().String("replication-policy", "",
		"Set the json file name that contains replication rules")
	cmd.Flags().String("deletion-policy", "retain",
		"Set the deletion policy, specify how COSI will handle deletion of buckets created on top of the created bucketclass "+
			"- delete | retain - the default deletion policy is retain")
	return cmd
}

// CmdCreateMultiNamespaceBucketclass returns a CLI command
func CmdCreateMultiNamespaceBucketclass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi <bucket-class-name>",
		Short: "Create namespace bucket class of type Multi",
		Run:   RunCreateMultiNamespaceBucketClass,
	}

	// multi namespace policy
	cmd.Flags().String("write-resource", "",
		"Set the namespace write resource")
	cmd.Flags().StringSlice("read-resources", nil,
		"Set the namespace read resources")
	cmd.Flags().String("replication-policy", "",
		"Set the json file name that contains replication rules")
	cmd.Flags().String("deletion-policy", "retain",
		"Set the deletion policy, specify how COSI will handle deletion of buckets created on top of the created bucketclass "+
			"- delete | retain - the default deletion policy is retain")

	return cmd
}

// CmdCreateCacheNamespaceBucketclass returns a CLI command
func CmdCreateCacheNamespaceBucketclass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache <bucket-class-name>",
		Short: "Create namespace bucket class of type Cache",
		Run:   RunCreateCacheNamespaceBucketClass,
	}

	// cache namespace policy
	cmd.Flags().String("hub-resource", "",
		"Set the namespace read and write resource")
	cmd.Flags().Uint32("ttl", 0,
		"Set the namespace cache ttl")

	// placement policy flags
	cmd.Flags().String("placement", "",
		"Set first tier placement policy - Mirror | Spread | \"\" (empty defaults to single backing store)")
	cmd.Flags().StringSlice("backingstores", nil,
		"Set first tier backing stores (use commas or multiple flags)")
	cmd.Flags().String("replication-policy", "",
		"Set the json file name that contains replication rules")
	cmd.Flags().String("deletion-policy", "retain",
		"Set the deletion policy, specify how COSI will handle deletion of buckets created on top of the created bucketclass "+
			"- delete | retain - the default deletion policy is retain")
	return cmd
}

// CmdDeleteBucketClass returns a CLI command
func CmdDeleteBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-class-name>",
		Short: "Delete a COSI bucket class",
		Run:   RunDeleteBucketClass,
	}

	return cmd
}

// CmdStatusBucketClass returns a CLI command
func CmdStatusBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-class-name>",
		Short: "Status of a COSI bucket class",
		Run:   RunStatusBucketClass,
	}
	return cmd
}

// CmdListBucketClass returns a CLI command
func CmdListBucketClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List COSI bucket classes",
		Run:   RunListBucketClass,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreateSingleNamespaceBucketClass runs a CLI command
func RunCreateSingleNamespaceBucketClass(cmd *cobra.Command, args []string) {
	createCommonCOSIBucketclass(cmd, args, nbv1.NSBucketClassTypeSingle, bucketclass.PopulateSingleNamespaceBucketClass)
}

// RunCreateMultiNamespaceBucketClass runs a CLI command
func RunCreateMultiNamespaceBucketClass(cmd *cobra.Command, args []string) {
	createCommonCOSIBucketclass(cmd, args, nbv1.NSBucketClassTypeMulti, bucketclass.PopulateMultiNamespaceBucketClass)
}

// RunCreateCacheNamespaceBucketClass runs a CLI command
func RunCreateCacheNamespaceBucketClass(cmd *cobra.Command, args []string) {
	createCommonCOSIBucketclass(cmd, args, nbv1.NSBucketClassTypeCache, bucketclass.PopulateCacheNamespaceBucketClass)
}

// RunCreatePlacementBucketClass runs a CLI command
func RunCreatePlacementBucketClass(cmd *cobra.Command, args []string) {
	createCommonCOSIBucketclass(cmd, args, "", bucketclass.PopulatePlacementBucketClass)
}

// createCommonBucketclass runs a CLI command
func createCommonCOSIBucketclass(cmd *cobra.Command, args []string, bucketClassType nbv1.NSBucketClassType, populate func(cmd *cobra.Command, bucketClassSpec *nbv1.BucketClassSpec) ([]string, []string)) {

	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}
	name := args[0]

	// Check and get system
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_cosi_bucket_class_yaml)
	bucketClass := o.(*nbv1.COSIBucketClass)
	bucketClass.Name = name
	bucketClass.DriverName = options.COSIDriverName()

	deletionPolicy, _ := cmd.Flags().GetString("deletion-policy")
	if deletionPolicy == "delete" {
		bucketClass.DeletionPolicy = nbv1.COSIDeletionPolicyDelete
	} else if deletionPolicy == "retain" {
		bucketClass.DeletionPolicy = nbv1.COSIDeletionPolicyRetain
	} else {
		log.Fatalf(`❌ Invalid deletion policy, valid values are delete or retain %s`, deletionPolicy)
	}

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(bucketClass), bucketClass)
	if err == nil {
		log.Fatalf(`❌ BucketClass %q already exists in namespace %q`, bucketClass.Name, bucketClass.Namespace)
	}

	bucketClassSpec := &nbv1.BucketClassSpec{}
	if bucketClassType != "" {
		bucketClassSpec.NamespacePolicy = &nbv1.NamespacePolicy{
			Type: bucketClassType,
		}
	}
	if bucketClassType == "" || bucketClassType == nbv1.NSBucketClassTypeCache {
		bucketClassSpec.PlacementPolicy = &nbv1.PlacementPolicy{
			Tiers: []nbv1.Tier{},
		}
	}
	namespaceStoresArr, backingStoresArr := populate(cmd, bucketClassSpec)

	bucketClassObj := &nbv1.BucketClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bucketClass.Name,
			Namespace: options.Namespace,
		},
		Spec: *bucketClassSpec,
	}
	err = validations.ValidateBucketClass(bucketClassObj)
	if err != nil {
		log.Fatalf(`❌ Bucket class validation failed %q`, err)
	}

	// check that namespace stores exists
	for _, name := range namespaceStoresArr {
		nsStore := &nbv1.NamespaceStore{
			TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: options.Namespace,
			},
		}
		if !util.KubeCheck(nsStore) {
			log.Fatalf(`❌ Could not get NamespaceStore %q in namespace %q`,
				nsStore.Name, nsStore.Namespace)
		}
	}

	// check that backing stores exists (for cache buckets)
	for _, name := range backingStoresArr {
		backStore := &nbv1.BackingStore{
			TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: options.Namespace,
			},
		}
		if !util.KubeCheck(backStore) {
			log.Fatalf(`❌ Could not get BackingStore %q in namespace %q`,
				backStore.Name, backStore.Namespace)
		}
	}

	replicationPolicyJSON, _ := cmd.Flags().GetString("replication-policy")
	if replicationPolicyJSON != "" {
		replication, err := util.LoadConfigurationJSON(replicationPolicyJSON)
		if err != nil {
			log.Fatalf(`❌ %q`, err)
		}
		bucketClassSpec.ReplicationPolicy = replication
	}
	bucketClass.Parameters = map[string]string{}

	if bucketClassSpec.PlacementPolicy != nil {
		pp, err := json.Marshal(bucketClassSpec.PlacementPolicy)
		if err != nil {
			log.Fatalf(`❌ Could not marshal BucketClass spec placement policy %q %+v`, bucketClass.Name, bucketClassSpec)
		}
		bucketClass.Parameters["placementPolicy"] = string(pp)
	}

	if bucketClassSpec.NamespacePolicy != nil {
		np, err := json.Marshal(bucketClassSpec.NamespacePolicy)
		if err != nil {
			log.Fatalf(`❌ Could not marshal BucketClass spec namespace policy %q %+v`, bucketClass.Name, bucketClassSpec)
		}
		bucketClass.Parameters["namespacePolicy"] = string(np)
	}

	if bucketClassSpec.Quota != nil {
		q, err := json.Marshal(bucketClassSpec.Quota)
		if err != nil {
			log.Fatalf(`❌ Could not marshal BucketClass spec quota %q %+v`, bucketClass.Name, bucketClassSpec)
		}
		bucketClass.Parameters["quota"] = string(q)

	}

	if bucketClassSpec.ReplicationPolicy != "" {
		rp, err := json.Marshal(bucketClassSpec.ReplicationPolicy)
		if err != nil {
			log.Fatalf(`❌ Could not marshal BucketClass spec replication policy %q %+v`, bucketClass.Name, bucketClassSpec)
		}
		bucketClass.Parameters["replicationPolicy"] = string(rp)

	}

	if !util.KubeCreateFailExisting(bucketClass) {
		log.Fatalf(`❌ Could not create BucketClass %q in Namespace %q (conflict)`, bucketClass.Name, bucketClass.Namespace)
	}

	log.Printf("")
	log.Printf("")
	RunStatusBucketClass(cmd, args)
}

// RunDeleteBucketClass runs a CLI command
func RunDeleteBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}

	cosiBucketClass := util.KubeObject(bundle.File_deploy_cosi_bucket_class_yaml).(*nbv1.COSIBucketClass)
	cosiBucketClass.Name = args[0]

	if !util.KubeDelete(cosiBucketClass) {
		log.Fatalf(`❌ Could not delete COSI bucket class %q`, cosiBucketClass.Name)
	}
}

// RunStatusBucketClass runs a CLI command
func RunStatusBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}

	cosiBucketClass := util.KubeObject(bundle.File_deploy_cosi_bucket_class_yaml).(*nbv1.COSIBucketClass)
	cosiBucketClass.Name = args[0]

	if !util.KubeCheck(cosiBucketClass) {
		log.Fatalf(`❌ Could not get BucketClass %q`, cosiBucketClass.Name)
	}

	fmt.Println()
	fmt.Println("# BucketClass spec:")
	fmt.Printf("Name:\n %s\n", cosiBucketClass.Name)
	fmt.Printf("DeletionPolicy:\n %s\n", cosiBucketClass.DeletionPolicy)
	fmt.Printf("Spec:\n %+v", cosiBucketClass.Parameters)
	fmt.Println()
}

// RunListBucketClass runs a CLI command
func RunListBucketClass(cmd *cobra.Command, args []string) {
	list := &nbv1.COSIBucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
	}

	if !util.KubeList(list, &client.ListOptions{}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No bucket classes found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAME",
		"PLACEMENT",
		"NAMESPACE-POLICY",
		"QUOTA",
		"AGE",
	)
	for i := range list.Items {
		bc := &list.Items[i]
		pp := bc.Parameters["placementPolicy"]
		np := bc.Parameters["namespacePolicy"]
		quota := bc.Parameters["quota"]
		table.AddRow(
			bc.Name,
			pp,
			np,
			quota,
			util.HumanizeDuration(time.Since(bc.CreationTimestamp.Time).Round(time.Second)),
		)
	}
	fmt.Print(table.String())
}

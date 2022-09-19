package bucketclass

import (
	"encoding/json"
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	sigyaml "sigs.k8s.io/yaml"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucketclass",
		Short: "Manage bucket classes",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdStatus(),
		CmdList(),
		CmdReconcile(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create bucket class",
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

	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-class-name>",
		Short: "Delete bucket class",
		Run:   RunDelete,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-class-name>",
		Short: "Status bucket class",
		Run:   RunStatus,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bucket classes",
		Run:   RunList,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdReconcile returns a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "reconcile",
		Short:  "Runs a reconcile attempt like noobaa-operator",
		Run:    RunReconcile,
	}
	return cmd
}

// RunCreateSingleNamespaceBucketClass runs a CLI command
func RunCreateSingleNamespaceBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommonBucketclass(cmd, args, nbv1.NSBucketClassTypeSingle, func(bucketClass *nbv1.BucketClass) ([]string, []string) {
		resource, _ := cmd.Flags().GetString("resource")
		if resource == "" {
			log.Fatalf(`❌ Must provide one namespace store`)
		}
		bucketClass.Spec.NamespacePolicy.Single = &nbv1.SingleNamespacePolicy{
			Resource: resource,
		}
		var namespaceStoresArr []string
		return append(namespaceStoresArr, resource), []string{}
	})
}

// RunCreateMultiNamespaceBucketClass runs a CLI command
func RunCreateMultiNamespaceBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommonBucketclass(cmd, args, nbv1.NSBucketClassTypeMulti, func(bucketClass *nbv1.BucketClass) ([]string, []string) {
		writeResource, _ := cmd.Flags().GetString("write-resource")
		readResources, _ := cmd.Flags().GetStringSlice("read-resources")
		if len(readResources) == 0 {
			log.Fatalf(`❌ Must provide at least one read resource`)
		}
		bucketClass.Spec.NamespacePolicy.Multi = &nbv1.MultiNamespacePolicy{
			WriteResource: writeResource,
			ReadResources: readResources,
		}
		if writeResource == "" {
			return readResources, []string{}
		}
		return append(readResources, writeResource), []string{}
	})
}

// RunCreateCacheNamespaceBucketClass runs a CLI command
func RunCreateCacheNamespaceBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommonBucketclass(cmd, args, nbv1.NSBucketClassTypeCache, func(bucketClass *nbv1.BucketClass) ([]string, []string) {
		hubResource, _ := cmd.Flags().GetString("hub-resource")
		cacheTTL, _ := cmd.Flags().GetUint32("ttl")
		placement, _ := cmd.Flags().GetString("placement")
		backingStores, _ := cmd.Flags().GetStringSlice("backingstores")
		if hubResource == "" {
			log.Fatalf(`❌ Must provide one namespace store as hub resource`)
		}
		if placement != "" && placement != "Spread" && placement != "Mirror" {
			log.Fatalf(`❌ Must provide valid placement: Mirror | Spread | ""`)
		}
		if len(backingStores) == 0 {
			log.Fatalf(`❌ Must provide at least one backing store`)
		}
		bucketClass.Spec.NamespacePolicy.Cache = &nbv1.CacheNamespacePolicy{
			HubResource: hubResource,
			Caching: &nbv1.CacheSpec{
				TTL: int(cacheTTL),
				// bucketClass.Spec.NamespacePolicy.Cache.Prefix = cachePrefix
			},
		}
		bucketClass.Spec.PlacementPolicy.Tiers = append(bucketClass.Spec.PlacementPolicy.Tiers,
			nbv1.Tier{Placement: nbv1.TierPlacement(placement), BackingStores: backingStores})

		var namespaceStoresArr []string
		return append(namespaceStoresArr, hubResource), backingStores
	})
}

// RunCreatePlacementBucketClass runs a CLI command
func RunCreatePlacementBucketClass(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommonBucketclass(cmd, args, "", func(bucketClass *nbv1.BucketClass) ([]string, []string) {
		placement, _ := cmd.Flags().GetString("placement")
		if placement != "" && placement != "Spread" && placement != "Mirror" {
			log.Fatalf(`❌ Must provide valid placement: Mirror | Spread | ""`)
		}
		backingStores, _ := cmd.Flags().GetStringSlice("backingstores")
		if len(backingStores) == 0 {
			log.Fatalf(`❌ Must provide at least one backing store`)
		}
		bucketClass.Spec.PlacementPolicy.Tiers = append(bucketClass.Spec.PlacementPolicy.Tiers,
			nbv1.Tier{Placement: nbv1.TierPlacement(placement), BackingStores: backingStores})

		maxSize, _ := cmd.Flags().GetString("max-size")
		maxObjects, _ := cmd.Flags().GetString("max-objects")
		if maxSize != "" || maxObjects != "" {
			bucketClass.Spec.Quota = &nbv1.Quota{
				MaxSize:    maxSize,
				MaxObjects: maxObjects,
			}
		}

		return []string{}, backingStores
	})
}

// createCommonBucketclass runs a CLI command
func createCommonBucketclass(cmd *cobra.Command, args []string, bucketClassType nbv1.NSBucketClassType, populate func(bucketClass *nbv1.BucketClass) ([]string, []string)) {

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

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml)
	bucketClass := o.(*nbv1.BucketClass)
	bucketClass.Name = name
	bucketClass.Namespace = options.Namespace

	if bucketClassType != "" {
		bucketClass.Spec.NamespacePolicy = &nbv1.NamespacePolicy{
			Type: bucketClassType,
		}
	}
	if bucketClassType == "" || bucketClassType == nbv1.NSBucketClassTypeCache {
		bucketClass.Spec.PlacementPolicy = &nbv1.PlacementPolicy{
			Tiers: []nbv1.Tier{},
		}
	}

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(bucketClass), bucketClass)
	if err == nil {
		log.Fatalf(`❌ BucketClass %q already exists in namespace %q`, bucketClass.Name, bucketClass.Namespace)
	}

	namespaceStoresArr, backingStoresArr := populate(bucketClass)

	err = validations.ValidateBucketClass(bucketClass)
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
		replication, err := util.LoadBucketReplicationJSON(replicationPolicyJSON)
		if err != nil {
			log.Fatalf(`❌ %q`, err)
		}
		bucketClass.Spec.ReplicationPolicy = replication
	}
	// Create bucket class CR
	util.Panic(controllerutil.SetControllerReference(sys, bucketClass, scheme.Scheme))
	if !util.KubeCreateFailExisting(bucketClass) {
		log.Fatalf(`❌ Could not create BucketClass %q in Namespace %q (conflict)`, bucketClass.Name, bucketClass.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("BucketClass Wait Ready:")
	if WaitReady(bucketClass) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml)
	bucketClass := o.(*nbv1.BucketClass)
	bucketClass.Name = args[0]
	bucketClass.Namespace = options.Namespace

	if !util.KubeDelete(bucketClass) {
		log.Fatalf(`❌ Could not delete BucketClass %q in namespace %q`,
			bucketClass.Name, bucketClass.Namespace)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml)
	bucketClass := o.(*nbv1.BucketClass)

	bucketClass.Name = args[0]
	bucketClass.Namespace = options.Namespace

	if !util.KubeCheck(bucketClass) {
		log.Fatalf(`❌ Could not get BucketClass %q in namespace %q`,
			bucketClass.Name, bucketClass.Namespace)
	}

	CheckPhase(bucketClass)

	fmt.Println()
	fmt.Println("# BucketClass spec:")
	output, err := sigyaml.Marshal(bucketClass.Spec)
	util.Panic(err)
	fmt.Print(string(output))
	fmt.Println()
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady(bucketClass *nbv1.BucketClass) bool {
	log := util.Logger()
	klient := util.KubeClient()

	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(bucketClass), bucketClass)
		if err != nil {
			log.Printf("⏳ Failed to get BucketClass: %s", err)
			return false, nil
		}
		CheckPhase(bucketClass)
		if bucketClass.Status.Phase == nbv1.BucketClassPhaseRejected {
			return false, fmt.Errorf("BucketClassPhaseRejected")
		}
		if bucketClass.Status.Phase != nbv1.BucketClassPhaseReady {
			return false, nil
		}
		return true, nil
	})
	return (err == nil)
}

// CheckPhase prints the phase and reason for it
func CheckPhase(bucketClass *nbv1.BucketClass) {
	log := util.Logger()

	reason := "waiting..."
	for _, c := range bucketClass.Status.Conditions {
		if c.Type == "Available" {
			reason = fmt.Sprintf("%s %s", c.Reason, c.Message)
		}
	}

	switch bucketClass.Status.Phase {

	case nbv1.BucketClassPhaseReady:
		log.Printf("✅ BucketClass %q Phase is Ready", bucketClass.Name)

	case nbv1.BucketClassPhaseRejected:
		log.Errorf("❌ BucketClass %q Phase is %q: %s", bucketClass.Name, bucketClass.Status.Phase, reason)

	case nbv1.BucketClassPhaseVerifying:
		fallthrough
	case nbv1.BucketClassPhaseDeleting:
		fallthrough
	default:
		log.Printf("⏳ BucketClass %q Phase is %q: %s", bucketClass.Name, bucketClass.Status.Phase, reason)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
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
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		bc := &list.Items[i]
		pp, _ := json.Marshal(bc.Spec.PlacementPolicy)
		np, _ := json.Marshal(bc.Spec.NamespacePolicy)
		quota, _ := json.Marshal(bc.Spec.Quota)
		table.AddRow(
			bc.Name,
			fmt.Sprintf("%+v", string(pp)),
			fmt.Sprintf("%+v", string(np)),
			fmt.Sprintf("%+v", string(quota)),
			string(bc.Status.Phase),
			util.HumanizeDuration(time.Since(bc.CreationTimestamp.Time).Round(time.Second)),
		)
	}
	fmt.Print(table.String())
}

// RunReconcile runs a CLI command
func RunReconcile(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	bucketClassName := args[0]
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      bucketClassName,
			},
		}
		res, err := NewReconciler(req.NamespacedName, klient, scheme.Scheme, nil).Reconcile()
		if err != nil {
			return false, err
		}
		if res.Requeue || res.RequeueAfter != 0 {
			log.Printf("\nRetrying in %d seconds\n", intervalSec)
			return false, nil
		}
		return true, nil
	}))
}

// ParseBucketClassType parses the --type flag to a StoreType enum
func ParseBucketClassType(cmd *cobra.Command) nbv1.StoreType {
	log := util.Logger()
	s, _ := cmd.Flags().GetString("type")
	if s == "" {
		fmt.Printf("Enter BucketClass Type - 'aws-s3' or 's3-compatible': ")
		_, err := fmt.Scan(&s)
		util.Panic(err)
	}
	switch s {
	case string(nbv1.StoreTypeAWSS3):
		return nbv1.StoreTypeAWSS3
	case "":
		log.Fatalf(`❌ Missing type value %s`, cmd.UsageString())
	default:
		log.Fatalf(`❌ Unsupported type value %q %s`, s, cmd.UsageString())
	}
	return ""
}

// MapBackingstoreToBucketclasses returns a list of bucketclasses that uses the backingstore in their tiering policy
// used by bucketclass_contorller to watch backingstore changes
func MapBackingstoreToBucketclasses(backingstore types.NamespacedName) []reconcile.Request {
	logrus.Infof("checking which bucketclasses to reconcile. mapping backingstore %v to bucketclasses", backingstore)
	bucketclassList := &nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	}
	if !util.KubeList(bucketclassList, &client.ListOptions{Namespace: backingstore.Namespace}) {
		logrus.Infof("Could not found bucketClasses in namespace %q", backingstore.Namespace)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, bc := range bucketclassList.Items {
		if bc.Spec.PlacementPolicy == nil {
			continue
		}
		for _, tier := range bc.Spec.PlacementPolicy.Tiers {
			for _, bs := range tier.BackingStores {
				if bs == backingstore.Name {
					reqs = append(reqs, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      bc.Name,
							Namespace: bc.Namespace,
						},
					})
				}
			}
		}
	}
	logrus.Infof("will reconcile these bucketclasses: %v", reqs)

	return reqs
}

// MapNamespacestoreToBucketclasses returns a list of bucketclasses that uses the namespacestore in their namespace policy
// used by bucketclass_contorller to watch namespacestores changes
func MapNamespacestoreToBucketclasses(namespacestore types.NamespacedName) []reconcile.Request {
	logrus.Infof("checking which bucketclasses to reconcile. mapping namespacestore %v to bucketclasses", namespacestore)
	bucketclassList := &nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	}
	if !util.KubeList(bucketclassList, &client.ListOptions{Namespace: namespacestore.Namespace}) {
		logrus.Infof("did not find namespace stores in namespace %q", namespacestore.Namespace)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, bc := range bucketclassList.Items {
		if bc.Spec.NamespacePolicy == nil {
			continue
		}
		policyType := bc.Spec.NamespacePolicy.Type
		if policyType == nbv1.NSBucketClassTypeSingle {
			nsr := bc.Spec.NamespacePolicy.Single.Resource
			if nsr == namespacestore.Name {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      bc.Name,
						Namespace: bc.Namespace,
					},
				})
			}
		} else if policyType == nbv1.NSBucketClassTypeCache {
			nsr := bc.Spec.NamespacePolicy.Cache.HubResource
			if nsr == namespacestore.Name {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      bc.Name,
						Namespace: bc.Namespace,
					},
				})
			}
		} else if policyType == nbv1.NSBucketClassTypeMulti {
			nsr := bc.Spec.NamespacePolicy.Multi.WriteResource
			if nsr == namespacestore.Name {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      bc.Name,
						Namespace: bc.Namespace,
					},
				})
			}
			for _, nsr := range bc.Spec.NamespacePolicy.Multi.ReadResources {
				if nsr == namespacestore.Name {
					reqs = append(reqs, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      bc.Name,
							Namespace: bc.Namespace,
						},
					})
				}
			}
		}
	}
	logrus.Infof("will reconcile these bucketclasses: %v", reqs)

	return reqs
}

// CreateTieringStructure creates a tering policy for a bucket
func CreateTieringStructure(BucketClass nbv1.BucketClass, BucketName string, nbClient nb.Client) (string, error) {
	tierName := fmt.Sprintf("%s.%x", BucketName, time.Now().Unix())
	tiers := []nb.TierItem{}

	for i := range BucketClass.Spec.PlacementPolicy.Tiers {
		tier := BucketClass.Spec.PlacementPolicy.Tiers[i]
		name := fmt.Sprintf("%s.%d", tierName, i)
		tiers = append(tiers, nb.TierItem{Order: int64(i), Tier: name})
		// we assume either mirror or spread but no mix and the bucket class controller rejects mixed classes.
		placement := "SPREAD"
		if tier.Placement == nbv1.TierPlacementMirror {
			placement = "MIRROR"
		}

		err := nbClient.CreateTierAPI(nb.CreateTierParams{
			Name:          name,
			AttachedPools: tier.BackingStores,
			DataPlacement: placement,
		})
		if err != nil {
			return tierName, fmt.Errorf("Failed to create tier %q with error: %v", name, err)
		}
	}

	err := nbClient.CreateTieringPolicyAPI(nb.TieringPolicyInfo{
		Name:  tierName,
		Tiers: tiers,
	})
	if err != nil {
		return tierName, fmt.Errorf("Failed to create tier %q with error: %v", tierName, err)
	}
	return tierName, nil
}

// GetDefaultBucketClass will get the default bucket class
func GetDefaultBucketClass(Namespace string) (*nbv1.BucketClass, error) {
	bucketClassName := options.SystemName + "-default-bucket-class"

	bucketClass := &nbv1.BucketClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bucketClassName,
			Namespace: Namespace,
		},
	}

	if !util.KubeCheck(bucketClass) {
		msg := fmt.Sprintf("GetDefaultBucketClass BucketClass %q not found in provisioner namespace %q", bucketClassName, Namespace)
		return nil, fmt.Errorf(msg)
	}

	if bucketClass.Status.Phase != nbv1.BucketClassPhaseReady {
		msg := fmt.Sprintf("GetDefaultBucketClass BucketClass %q is %v", bucketClassName, bucketClass.Status.Phase)
		return nil, fmt.Errorf(msg)
	}

	return bucketClass, nil
}

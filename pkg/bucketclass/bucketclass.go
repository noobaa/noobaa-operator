package bucketclass

import (
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

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
		Use:   "create <bucket-class-name>",
		Short: "Create bucket class",
		Run:   RunCreate,
	}
	cmd.Flags().String("placement", "",
		"Set first tier placement policy - Mirror | Spread | \"\" (empty defaults to single backing store)")
	cmd.Flags().StringSlice("backingstores", nil,
		"Set first tier backing stores (use commas or multiple flags)")
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
	}
	return cmd
}

// CmdReconcile returns a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Runs a reconcile attempt like noobaa-operator",
		Run:   RunReconcile,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}
	name := args[0]

	placement, _ := cmd.Flags().GetString("placement")
	if placement != "" && placement != "Spread" && placement != "Mirror" {
		log.Fatalf(`❌ Must provide valid placement: Mirror | Spread | ""`)
	}
	backingStores, _ := cmd.Flags().GetStringSlice("backingstores")
	if len(backingStores) == 0 {
		log.Fatalf(`❌ Must provide at least one backing store`)
	}

	// Check and get system
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml)
	bucketClass := o.(*nbv1.BucketClass)
	bucketClass.Name = name
	bucketClass.Namespace = options.Namespace
	bucketClass.Spec.PlacementPolicy.Tiers[0].Placement = nbv1.TierPlacement(placement)
	bucketClass.Spec.PlacementPolicy.Tiers[0].BackingStores = backingStores

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(bucketClass), bucketClass)
	if err == nil {
		log.Fatalf(`❌ BucketClass %q already exists in namespace %q`, bucketClass.Name, bucketClass.Namespace)
	}

	// check that backing stores exists
	for _, name := range backingStores {
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

	// Create bucket class CR
	util.Panic(controllerutil.SetControllerReference(sys, bucketClass, scheme.Scheme))
	if !util.KubeCreateSkipExisting(bucketClass) {
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

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml)
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

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml)
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
	if err != nil {
		return false
	}
	return true
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
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
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
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		bc := &list.Items[i]
		table.AddRow(
			bc.Name,
			fmt.Sprintf("%+v", bc.Spec.PlacementPolicy),
			string(bc.Status.Phase),
			time.Since(bc.CreationTimestamp.Time).Round(time.Second).String(),
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

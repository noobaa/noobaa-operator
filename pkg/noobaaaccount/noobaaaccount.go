package noobaaaccount

import (
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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
		Use:   "account",
		Short: "Manage noobaa accounts",
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
		Use:   "create <noobaa-account-name>",
		Short: "Create noobaa account",
		Run:   RunCreate,
	}
	cmd.Flags().Bool("allow_bucket_create", true,
		"Should this account be allowed to create new buckets")
	cmd.Flags().Bool("full_permission", false,
		"Should this account be allowed to access all the buckets (including future ones)")
	cmd.Flags().StringSlice("allowed_buckets", nil,
		"Set the user allowed buckets list (use commas or multiple flags)")
	cmd.Flags().String("default_resource", "", "Set the default resource, on which new buckets will be created")
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <noobaa-account-name>",
		Short: "Delete noobaa account",
		Run:   RunDelete,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <noobaa-account-name>",
		Short: "Status noobaa account",
		Run:   RunStatus,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List noobaa accounts",
		Run:   RunList,
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

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}
	name := args[0]

	allowedBuckets := []string{}
	fullPermission, _ := cmd.Flags().GetBool("full_permission")
	bucketList, _ := cmd.Flags().GetStringSlice("allowed_buckets")
	allowedBuckets = append(allowedBuckets, bucketList...)
	if !fullPermission && len(allowedBuckets) == 0 {
		log.Fatalf(`❌ Must provide at least one allowed buckets, or full_permission`)
	}
	if len(allowedBuckets) > 0 &&  fullPermission {
		log.Fatalf(`❌ Can't provide both full_permission and an allowed buckets list`)
	}
	allowBucketCreate, _ := cmd.Flags().GetBool("allow_bucket_create")
	defaultResource, _ := cmd.Flags().GetString("default_resource")

	// Check and get system
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	noobaaAccount.Name = name
	noobaaAccount.Namespace = options.Namespace
	noobaaAccount.Spec.AllowBucketCreate = allowBucketCreate
	noobaaAccount.Spec.AllowedBuckets.FullPermission = fullPermission
	noobaaAccount.Spec.AllowedBuckets.PermissionList = allowedBuckets

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	if defaultResource == "" { // if user doesn't provide default resource we will use the default backingstore
		defaultResource = sys.Name + "-default-backing-store"
	} 
	// check that default backing store exists
	defaultRes := &nbv1.BackingStore{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultResource,
			Namespace: options.Namespace,
		},
	}
	if !util.KubeCheck(defaultRes) {
		log.Fatalf(`❌ Could not get BackingStore %q in namespace %q`,
			defaultResource, noobaaAccount.Namespace)
	}
	noobaaAccount.Spec.DefaultResource = defaultResource

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(noobaaAccount), noobaaAccount)
	if err == nil {
		log.Fatalf(`❌ noobaaAccount %q already exists in namespace %q`, noobaaAccount.Name, noobaaAccount.Namespace)
	}

	// Create noobaa account CR
	util.Panic(controllerutil.SetControllerReference(sys, noobaaAccount, scheme.Scheme))
	if !util.KubeCreateSkipExisting(noobaaAccount) {
		log.Fatalf(`❌ Could not create noobaaAccount %q in Namespace %q (conflict)`, noobaaAccount.Name, noobaaAccount.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("NooBaaAccount Wait Ready:")
	if WaitReady(noobaaAccount) {
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

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	noobaaAccount.Name = args[0]
	noobaaAccount.Namespace = options.Namespace

	if !util.KubeDelete(noobaaAccount) {
		log.Fatalf(`❌ Could not delete NooBaaAccount %q in namespace %q`,
		noobaaAccount.Name, noobaaAccount.Namespace)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)

	noobaaAccount.Name = args[0]
	secret.Name = fmt.Sprintf("noobaa-account-%s", args[0])
	noobaaAccount.Namespace = options.Namespace
	secret.Namespace = options.Namespace

	if !util.KubeCheck(noobaaAccount) {
		log.Fatalf(`❌ Could not get NooBaaAccount %q in namespace %q`,
		noobaaAccount.Name, noobaaAccount.Namespace)
	}
	util.KubeCheck(secret)

	CheckPhase(noobaaAccount)

	fmt.Println()
	fmt.Println("# NooBaaAccount spec:")
	output, err := sigyaml.Marshal(noobaaAccount.Spec)
	util.Panic(err)
	fmt.Print(string(output))
	fmt.Println()
	fmt.Printf("Connection info:\n")
	credsEnv := ""
	for k, v := range secret.StringData {
		if v != "" {
			fmt.Printf("  %-22s : %s\n", k, v)
			credsEnv += k + "=" + v + " "
		}
	}
	fmt.Println()
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady(noobaaAccount *nbv1.NooBaaAccount) bool {
	log := util.Logger()
	klient := util.KubeClient()

	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(noobaaAccount), noobaaAccount)
		if err != nil {
			log.Printf("⏳ Failed to get NooBaaAccount: %s", err)
			return false, nil
		}
		CheckPhase(noobaaAccount)
		if noobaaAccount.Status.Phase == nbv1.NooBaaAccountPhaseRejected {
			return false, fmt.Errorf("NooBaaAccountPhaseRejected")
		}
		if noobaaAccount.Status.Phase != nbv1.NooBaaAccountPhaseReady {
			return false, nil
		}
		return true, nil
	})
	return err == nil
}

// CheckPhase prints the phase and reason for it
func CheckPhase(noobaaAccount *nbv1.NooBaaAccount) {
	log := util.Logger()

	reason := "waiting..."
	for _, c := range noobaaAccount.Status.Conditions {
		if c.Type == "Available" {
			reason = fmt.Sprintf("%s %s", c.Reason, c.Message)
		}
	}

	switch noobaaAccount.Status.Phase {

	case nbv1.NooBaaAccountPhaseReady:
		log.Printf("✅ NooBaaAccount %q Phase is Ready", noobaaAccount.Name)

	case nbv1.NooBaaAccountPhaseRejected:
		log.Errorf("❌ NooBaaAccount %q Phase is %q: %s", noobaaAccount.Name, noobaaAccount.Status.Phase, reason)

	case nbv1.NooBaaAccountPhaseVerifying:
		fallthrough
	case nbv1.NooBaaAccountPhaseDeleting:
		fallthrough
	default:
		log.Printf("⏳ NooBaaAccount %q Phase is %q: %s", noobaaAccount.Name, noobaaAccount.Status.Phase, reason)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.NooBaaAccountList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaAccountList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No noobaa accounts found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAME",
		"ALLOWED_BUCKETS",
		"DEFAULT_RESOURCE",
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		na := &list.Items[i]
		var allowedBuckets []string
		if na.Spec.AllowedBuckets.FullPermission {
			allowedBuckets = append(allowedBuckets, "*")
		} else {
			allowedBuckets = na.Spec.AllowedBuckets.PermissionList
		}
		defaultResource := na.Spec.DefaultResource
		if !na.Spec.AllowBucketCreate {
			defaultResource = "-NO-BUCKET-CREATION-"
		}
		table.AddRow(
			na.Name,
			fmt.Sprintf("%+v", allowedBuckets),
			defaultResource,
			string(na.Status.Phase),
			time.Since(na.CreationTimestamp.Time).Round(time.Second).String(),
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
	noobaaAccountName := args[0]
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      noobaaAccountName,
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


package backingstore

import (
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backingstore",
		Short: "Manage noobaa backing stores",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdList(),
		CmdReconcile(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <backing-store-name>",
		Short: "Create backing store",
		Run:   RunCreate,
	}
	cmd.Flags().String(
		"type", "",
		`Backing store type: 'aws-s3' or 'google-cloud-store' or 'azure-blob' or 's3-compatible'`,
	)
	cmd.Flags().String(
		"bucket-name", "",
		"The target bucket name",
	)
	cmd.Flags().String(
		"access-key", "",
		`Access key for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"secret-key", "",
		`Secret key for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"endpoint", "",
		"The target endpoint",
	)
	cmd.Flags().String(
		"aws-region", "",
		"The AWS bucket region",
	)
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <backing-store-name>",
		Short: "Delete backing store",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backing stores",
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
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}
	name := args[0]

	// Check and get system
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	o = util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)

	sys.Name = options.SystemName
	secret.Name = "backing-store-secret-" + name
	backStore.Name = name

	sys.Namespace = options.Namespace
	secret.Namespace = options.Namespace
	backStore.Namespace = options.Namespace

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(backStore), backStore)
	if err == nil {
		log.Fatalf(`❌ BackingStore %q already exists in namespace %q`, backStore.Name, backStore.Namespace)
	}

	typeVal := ParseBackingStoreType(cmd)
	endpoint, _ := cmd.Flags().GetString("endpoint")
	bucketName, _ := cmd.Flags().GetString("bucket-name")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")

	if bucketName == "" {
		fmt.Printf("Enter target bucket name: ")
		_, err := fmt.Scan(&bucketName)
		util.Panic(err)
		if bucketName == "" {
			log.Fatalf(`❌ Missing bucket name %s`, cmd.UsageString())
		}
	}

	secret.StringData = map[string]string{}
	secret.Data = nil

	backStore.Spec.Type = typeVal
	backStore.Spec.BucketName = bucketName
	backStore.Spec.S3Options = &nbv1.S3Options{Endpoint: endpoint}
	backStore.Spec.Secret = corev1.SecretReference{Name: secret.Name, Namespace: secret.Namespace}

	if accessKey == "" {
		fmt.Printf("Enter Access Key: ")
		accessKeyBytes, err := terminal.ReadPassword(0)
		util.Panic(err)
		accessKey = string(accessKeyBytes)
		fmt.Printf("[got %d characters]\n", len(accessKey))
	}
	if secretKey == "" {
		fmt.Printf("Enter Secret Key: ")
		secretBytes, err := terminal.ReadPassword(0)
		util.Panic(err)
		secretKey = string(secretBytes)
		fmt.Printf("[got %d characters]\n", len(secretKey))
	}
	secret.StringData["AWS_ACCESS_KEY_ID"] = accessKey
	secret.StringData["AWS_SECRET_ACCESS_KEY"] = secretKey

	// Create backing store CR
	util.Panic(controllerutil.SetControllerReference(sys, backStore, scheme.Scheme))
	if !util.KubeCreateSkipExisting(backStore) {
		log.Fatalf(`❌ Could not create backing-store %q in namespace %q (conflict)`, backStore.Name, backStore.Namespace)
	}

	// Create secret
	util.Panic(controllerutil.SetControllerReference(backStore, secret, scheme.Scheme))
	if !util.KubeCreateSkipExisting(secret) {
		log.Fatalf(`❌ Could not create secret %q in namespace %q (conflict)`, secret.Name, secret.Namespace)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = args[0]
	backStore.Namespace = options.Namespace

	if !util.KubeDelete(backStore) {
		log.Fatalf(`❌ Could not delete backing-store %q in namespace %q`,
			backStore.Name, backStore.Namespace)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No backing stores found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow("NAME", "TYPE", "BUCKET-NAME", "PHASE")
	for i := range list.Items {
		bs := &list.Items[i]
		table.AddRow(bs.Name, string(bs.Spec.Type), bs.Spec.BucketName, string(bs.Status.Phase))
	}
	fmt.Print(table.String())
}

// RunReconcile runs a CLI command
func RunReconcile(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	backingStoreName := args[0]
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      backingStoreName,
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

// ParseBackingStoreType parses the --type flag to a StoreType enum
func ParseBackingStoreType(cmd *cobra.Command) nbv1.StoreType {
	log := util.Logger()
	s, _ := cmd.Flags().GetString("type")
	if s == "" {
		fmt.Printf("Enter BackingStore Type - 'aws-s3' or 'google-cloud-store' or 'azure-blob' or 's3-compatible': ")
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

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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Cmd creates a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backing-store",
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

// CmdCreate creates a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <backing-store-name>",
		Short: "Create backing store",
		Run:   RunCreate,
	}
	cmd.Flags().String(
		"type", "",
		`**REQUIRED** Backing store type: 'aws-s3' or 'google-cloud-store' or 'azure-blob' or 's3-compatible'`,
	)
	cmd.Flags().String(
		"bucket-name", "",
		"**REQUIRED** The target bucket name",
	)
	cmd.Flags().String(
		"access-key", "",
		"**REQUIRED** Access key for authentication",
	)
	cmd.Flags().String(
		"secret-key", "",
		`[Optional] Secret key for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to read the secret from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"endpoint", "",
		"[Optional] The target endpoint",
	)
	cmd.Flags().String(
		"aws-region", "",
		"[Optional] The AWS bucket region",
	)
	return cmd
}

// CmdDelete creates a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <backing-store-name>",
		Short: "Delete backing store",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList creates a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backing stores",
		Run:   RunList,
	}
	return cmd
}

// CmdReconcile creates a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile backing stores",
		Run:   RunReconcile,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	name := args[0]
	typeVal := ParseBackingStoreType(cmd)
	endpoint, _ := cmd.Flags().GetString("endpoint")
	bucketName, _ := cmd.Flags().GetString("bucket-name")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")

	if bucketName == "" {
		log.Fatalf(`Flag --bucket-name is required %s`, cmd.UsageString())
	}
	if accessKey == "" {
		log.Fatalf(`Flag --access-key is required %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	o = util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)

	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	secret.Name = "backing-store-secret-" + name
	secret.Namespace = options.Namespace
	secret.StringData = map[string]string{}
	secret.Data = nil

	backStore.Name = name
	backStore.Namespace = options.Namespace
	backStore.Spec.Type = typeVal
	backStore.Spec.BucketName = bucketName
	backStore.Spec.S3Options = &nbv1.S3Options{Endpoint: endpoint}
	backStore.Spec.Secret = corev1.SecretReference{Name: secret.Name, Namespace: secret.Namespace}

	// Check and get system
	if !util.KubeCheck(sys) {
		log.Fatalf(`System "%s" not found in namespace "%s"`, sys.Name, sys.Namespace)
	}

	if secretKey == "" {
		fmt.Printf("Enter Secret Key: ")
		secretBytes, err := terminal.ReadPassword(0)
		util.Panic(err)
		secretKey = string(secretBytes)
		fmt.Println()
	}
	secret.StringData["AWS_ACCESS_KEY_ID"] = accessKey
	secret.StringData["AWS_SECRET_ACCESS_KEY"] = secretKey

	// Create backing store CR
	util.Panic(controllerutil.SetControllerReference(sys, backStore, scheme.Scheme))
	if !util.KubeCreateSkipExisting(backStore) {
		log.Fatalf(`Backing store name "%s" conflict in namespace "%s"`, backStore.Name, backStore.Namespace)
	}

	// Create secret
	util.Panic(controllerutil.SetControllerReference(backStore, secret, scheme.Scheme))
	if !util.KubeCreateSkipExisting(secret) {
		log.Fatalf(`Backing store secret name "%s" conflict in namespace "%s"`, secret.Name, secret.Namespace)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = args[0]
	backStore.Namespace = options.Namespace

	if !util.KubeDelete(backStore) {
		log.Fatalf(`Backing store name "%s" not found in namespace "%s"`, backStore.Name, backStore.Namespace)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.BackingStoreList{}
	util.KubeClient().List(
		util.Context(),
		&client.ListOptions{Namespace: options.Namespace},
		list,
	)
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
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      options.SystemName,
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
	switch s {
	case string(nbv1.StoreTypeAWSS3):
		return nbv1.StoreTypeAWSS3
	default:
		if s == "" {
			log.Fatalf(`Flag --type is required %s`, cmd.UsageString())
		} else {
			log.Fatalf(`Flag "--type %s" unsupported value %s`, s, cmd.UsageString())
		}
	}
	return ""
}

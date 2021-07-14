package namespacestore

import (
	"fmt"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"

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
		Use:   "namespacestore",
		Short: "Manage namespace stores",
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
		Short: "Create namespace store",
	}
	cmd.AddCommand(
		CmdCreateAWSS3(),
		CmdCreateS3Compatible(),
		CmdCreateIBMCos(),
		CmdCreateAzureBlob(),
		CmdCreateNSFS(),
	)
	return cmd
}

// CmdCreateAWSS3 returns a CLI command
func CmdCreateAWSS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-s3 <namespace-store-name>",
		Short: "Create aws-s3 namespace store",
		Run:   RunCreateAWSS3,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on the cloud",
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
		"secret-name", "",
		`The name of a secret for authentication - should have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY properties`,
	)
	cmd.Flags().String(
		"region", "",
		"The AWS bucket region",
	)
	return cmd
}

// CmdCreateS3Compatible returns a CLI command
func CmdCreateS3Compatible() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "s3-compatible <namespace-store-name>",
		Short: "Create s3-compatible namespace store",
		Run:   RunCreateS3Compatible,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on the cloud",
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
		"secret-name", "",
		`The name of a secret for authentication - should have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY properties`,
	)
	cmd.Flags().String(
		"endpoint", "",
		"The target S3 endpoint",
	)
	cmd.Flags().String(
		"signature-version", "v4",
		"The S3 signature version v4|v2",
	)
	return cmd
}

// CmdCreateIBMCos returns a CLI command
func CmdCreateIBMCos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibm-cos <namespace-store-name>",
		Short: "Create ibm-cos namespace store",
		Run:   RunCreateIBMCos,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on the cloud",
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
		"secret-name", "",
		`The name of a secret for authentication - should have IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY properties`,
	)
	cmd.Flags().String(
		"endpoint", "",
		"The target IBM Cos endpoint",
	)
	return cmd
}

// CmdCreateAzureBlob returns a CLI command
func CmdCreateAzureBlob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-blob <namespace-store-name>",
		Short: "Create azure-blob namespace store",
		Run:   RunCreateAzureBlob,
	}
	cmd.Flags().String(
		"target-blob-container", "",
		"The target container name on Azure storage account",
	)
	cmd.Flags().String(
		"account-name", "",
		`Account name for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"account-key", "",
		`Account key for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"secret-name", "",
		`The name of a secret for authentication - should have AccountName and AccountKey properties`,
	)
	return cmd
}

// CmdCreateNSFS returns a CLI command
func CmdCreateNSFS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nsfs <namespace-store-name>",
		Short: "Create nsfs namespace store",
		Run:   RunCreateNSFS,
	}
	cmd.Flags().String(
		"fs-backend", "",
		"The file system backend type - CEPH_FS | GPFS | NFSv4",
	)
	cmd.Flags().String(
		"sub-path", "",
		"The path to a sub directory inside the pvc file system",
	)
	cmd.Flags().String(
		"pvc-name", "",
		"The pvc name in which the file system resides",
	)
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <namespace-store-name>",
		Short: "Delete namespace store",
		Run:   RunDelete,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <namespace-store-name>",
		Short: "Status namespace store",
		Run:   RunStatus,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List namespace stores",
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

func createCommon(cmd *cobra.Command, args []string, storeType nbv1.NSType, populate func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret)) {

	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <namespace-store-name> %s`, cmd.UsageString())
	}
	name := args[0]
	secretName, _ := cmd.Flags().GetString("secret-name")

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml)
	namespaceStore := o.(*nbv1.NamespaceStore)
	namespaceStore.Name = name
	namespaceStore.Namespace = options.Namespace
	namespaceStore.Spec = nbv1.NamespaceStoreSpec{Type: storeType}

	o = util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)
	secret.Name = fmt.Sprintf("namespace-store-%s-%s", storeType, name)
	secret.Namespace = options.Namespace
	secret.StringData = map[string]string{}
	secret.Data = nil

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(namespaceStore), namespaceStore)
	if err == nil {
		log.Fatalf(`❌ NamespaceStore %q already exists in namespace %q`, namespaceStore.Name, namespaceStore.Namespace)
	}

	populate(namespaceStore, secret)

	suggestedSecret := util.CheckForIdenticalSecretsCreds(secret, util.MapStorTypeToMandatoryProperties[nbv1.StoreType(namespaceStore.Spec.Type)])
	if suggestedSecret != nil {
		var decision string
		log.Printf("Found a Secret in the system with the same credentials (%s)", suggestedSecret.Name)
		log.Printf("Note that using more then one secret with the same credentials is not supported")
		log.Printf("do you want to use it for this Namespacestore? y/n")
		fmt.Scanln(&decision)
		if strings.ToLower(decision) == "y" {
			log.Printf("Will use %s as the Namespacestore Secret", suggestedSecret.Name)
			err := util.SetNamespaceStoreSecretRef(namespaceStore, &corev1.SecretReference{
				Name:      suggestedSecret.Name,
				Namespace: suggestedSecret.Namespace,
			})
			if err != nil {
				log.Fatalf(`❌ %s`, err)
			}
		} else if strings.ToLower(decision) == "n" {
			log.Fatalf("Not creating Namespacestore")
		} else {
			log.Fatalf(`❌ Invalid input, please select y/n`)
		}
	}

	validationErr := validations.ValidateNamespaceStore(namespaceStore)
	if validationErr != nil {
		log.Fatalf(`❌ %s %s`, validationErr, cmd.UsageString())
	}

	// Create namespace store CR
	util.Panic(controllerutil.SetControllerReference(sys, namespaceStore, scheme.Scheme))
	if !util.KubeCreateFailExisting(namespaceStore) {
		log.Fatalf(`❌ Could not create NamespaceStore %q in Namespace %q (conflict)`, namespaceStore.Name, namespaceStore.Namespace)
	}

	secretRef, _ := util.GetNamespaceStoreSecret(namespaceStore)
	if secretRef != nil && secretName == "" && suggestedSecret == nil {
		// Create secret
		util.Panic(controllerutil.SetControllerReference(namespaceStore, secret, scheme.Scheme))
		if !util.KubeCreateFailExisting(secret) {
			log.Fatalf(`❌ Could not create Secret %q in Namespace %q (conflict)`, secret.Name, secret.Namespace)
		}
	} else if secretRef != nil && secretName != "" {
		_, err := util.GetSecretFromSecretReference(secretRef)
		if err != nil {
			util.Logger().Fatalf(`❌ Could not found Secret %q from SecretReference`, secret.Name)
		}
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("NamespaceStore Wait Ready:")
	if WaitReady(namespaceStore) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunCreateAWSS3 runs a CLI command
func RunCreateAWSS3(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.NSStoreTypeAWSS3, func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret) {
		targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
		region, _ := cmd.Flags().GetString("region")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}

		if secretName == "" {
			accessKey := util.GetFlagStringOrPromptPassword(cmd, "access-key")
			secretKey := util.GetFlagStringOrPromptPassword(cmd, "secret-key")
			secret.StringData["AWS_ACCESS_KEY_ID"] = accessKey
			secret.StringData["AWS_SECRET_ACCESS_KEY"] = secretKey
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		namespaceStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
			TargetBucket: targetBucket,
			Region:       region,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateS3Compatible runs a CLI command
func RunCreateS3Compatible(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.NSStoreTypeS3Compatible, func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret) {
		endpoint := util.GetFlagStringOrPrompt(cmd, "endpoint")
		targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
		sigVer, _ := cmd.Flags().GetString("signature-version")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}

		if secretName == "" {
			accessKey := util.GetFlagStringOrPromptPassword(cmd, "access-key")
			secretKey := util.GetFlagStringOrPromptPassword(cmd, "secret-key")
			secret.StringData["AWS_ACCESS_KEY_ID"] = accessKey
			secret.StringData["AWS_SECRET_ACCESS_KEY"] = secretKey
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		namespaceStore.Spec.S3Compatible = &nbv1.S3CompatibleSpec{
			TargetBucket:     targetBucket,
			Endpoint:         endpoint,
			SignatureVersion: nbv1.S3SignatureVersion(sigVer),
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateIBMCos runs a CLI command
func RunCreateIBMCos(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.NSStoreTypeIBMCos, func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret) {
		endpoint := util.GetFlagStringOrPrompt(cmd, "endpoint")
		targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
		// sigVer, _ := cmd.Flags().GetString("signature-version")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{"IBM_COS_ACCESS_KEY_ID", "IBM_COS_SECRET_ACCESS_KEY"}

		if secretName == "" {
			accessKey := util.GetFlagStringOrPromptPassword(cmd, "access-key")
			secretKey := util.GetFlagStringOrPromptPassword(cmd, "secret-key")
			secret.StringData["IBM_COS_ACCESS_KEY_ID"] = accessKey
			secret.StringData["IBM_COS_SECRET_ACCESS_KEY"] = secretKey
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		namespaceStore.Spec.IBMCos = &nbv1.IBMCosSpec{
			TargetBucket:     targetBucket,
			Endpoint:         endpoint,
			SignatureVersion: nbv1.S3SignatureVersion("v2"),
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateAzureBlob runs a CLI command
func RunCreateAzureBlob(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.NSStoreTypeAzureBlob, func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret) {
		targetBlobContainer := util.GetFlagStringOrPrompt(cmd, "target-blob-container")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{"AccountName", "AccountKey"}

		if secretName == "" {
			accountName := util.GetFlagStringOrPrompt(cmd, "account-name")
			accountKey := util.GetFlagStringOrPrompt(cmd, "account-key")
			secret.StringData["AccountName"] = accountName
			secret.StringData["AccountKey"] = accountKey
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		namespaceStore.Spec.AzureBlob = &nbv1.AzureBlobSpec{
			TargetBlobContainer: targetBlobContainer,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateNSFS runs a CLI command
func RunCreateNSFS(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.NSStoreTypeNSFS, func(namespaceStore *nbv1.NamespaceStore, secret *corev1.Secret) {
		pvcName := util.GetFlagStringOrPrompt(cmd, "pvc-name")
		fsBackend, _ := cmd.Flags().GetString("fs-backend")
		subPath, _ := cmd.Flags().GetString("sub-path")

		namespaceStore.Spec.NSFS = &nbv1.NSFSSpec{
			PvcName:   pvcName,
			SubPath:   subPath,
			FsBackend: fsBackend,
		}
	})
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <namespace-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml)
	namespaceStore := o.(*nbv1.NamespaceStore)
	namespaceStore.Name = args[0]
	namespaceStore.Namespace = options.Namespace
	namespaceStore.Spec = nbv1.NamespaceStoreSpec{}

	nbClient := system.GetNBClient()

	namespaceResourceinfo, err := nbClient.ReadNamespaceResourceAPI(nb.ReadNamespaceResourceParams{Name: namespaceStore.Name})
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if !isRPCErr || rpcErr.RPCCode != "NO_SUCH_NAMESPACE_RESOURCE" {
			log.Fatalf(`❌ Failed to read NamespaceStore info: %s`, err)
		}
	} else if namespaceResourceinfo.Undeletable != "" && namespaceResourceinfo.Undeletable != "IS_NAMESPACESTORE" {
		switch namespaceResourceinfo.Undeletable {
		case "CONNECTED_BUCKET_DELETING":
			fallthrough
		case "IN_USE":
			log.Fatalf(`❌ Could not delete NamespaceStore %q in namespace %q as it is being used by one or more buckets`,
				namespaceStore.Name, namespaceStore.Namespace)
		default:
			log.Fatalf(`❌ Could not delete NamespaceStore %q in namespace %q, undeletable due to %q`,
				namespaceStore.Name, namespaceStore.Namespace, namespaceResourceinfo.Undeletable)
		}
	}
	if !util.KubeDelete(namespaceStore) {
		log.Fatalf(`❌ Could not delete NamespaceStore %q in namespace %q`,
			namespaceStore.Name, namespaceStore.Namespace)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <namespace-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml)
	namespaceStore := o.(*nbv1.NamespaceStore)
	namespaceStore.Name = args[0]
	namespaceStore.Namespace = options.Namespace
	namespaceStore.Spec = nbv1.NamespaceStoreSpec{}

	if !util.KubeCheck(namespaceStore) {
		log.Fatalf(`❌ Could not get NamespaceStore %q in namespace %q`,
			namespaceStore.Name, namespaceStore.Namespace)
	}

	secretRef, _ := util.GetNamespaceStoreSecret(namespaceStore)
	if secretRef != nil {
		secret.Name = secretRef.Name
		secret.Namespace = secretRef.Namespace
		if secret.Namespace == "" {
			secret.Namespace = namespaceStore.Namespace
		}
		if !util.KubeCheck(secret) {
			log.Errorf(`❌ Could not get Secret %q in namespace %q`,
				secret.Name, secret.Namespace)
		}
	}

	CheckPhase(namespaceStore)

	fmt.Println()
	fmt.Println("# NamespaceStore spec:")
	output, err := sigyaml.Marshal(namespaceStore.Spec)
	util.Panic(err)
	fmt.Print(string(output))
	fmt.Println()
	if secretRef != nil {
		_, err = sigyaml.Marshal(secret.StringData)
		util.Panic(err)
		fmt.Println()
	}
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady(namespaceStore *nbv1.NamespaceStore) bool {
	log := util.Logger()
	klient := util.KubeClient()

	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(namespaceStore), namespaceStore)
		if err != nil {
			log.Printf("⏳ Failed to get NamespaceStore: %s", err)
			return false, nil
		}
		CheckPhase(namespaceStore)
		if namespaceStore.Status.Phase == nbv1.NamespaceStorePhaseRejected {
			return false, fmt.Errorf("NamespaceStorePhaseRejected")
		}
		if namespaceStore.Status.Phase != nbv1.NamespaceStorePhaseReady {
			return false, nil
		}
		return true, nil
	})
	return (err == nil)
}

// CheckPhase prints the phase and reason for it
func CheckPhase(namespaceStore *nbv1.NamespaceStore) {
	log := util.Logger()

	reason := "waiting..."
	for _, c := range namespaceStore.Status.Conditions {
		if c.Type == "Available" {
			reason = fmt.Sprintf("%s %s", c.Reason, c.Message)
		}
	}

	switch namespaceStore.Status.Phase {

	case nbv1.NamespaceStorePhaseReady:
		log.Printf("✅ NamespaceStore %q Phase is Ready", namespaceStore.Name)

	case nbv1.NamespaceStorePhaseRejected:
		log.Errorf("❌ NamespaceStore %q Phase is %q: %s", namespaceStore.Name, namespaceStore.Status.Phase, reason)

	case nbv1.NamespaceStorePhaseVerifying:
		log.Printf("NamespaceStorePhaseVerifying")
		fallthrough
	case nbv1.NamespaceStorePhaseConnecting:
		log.Printf("NamespaceStorePhaseVerifying")
		fallthrough
	case nbv1.NamespaceStorePhaseCreating:
		fallthrough
	case nbv1.NamespaceStorePhaseDeleting:
		fallthrough
	default:
		log.Printf("⏳ NamespaceStore %q Phase is %q: %s", namespaceStore.Name, namespaceStore.Status.Phase, reason)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No namespace stores found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAME",
		"TYPE",
		"TARGET-BUCKET",
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		bs := &list.Items[i]
		tb, err := util.GetNamespaceStoreTargetBucket(bs)
		if err == nil {
			table.AddRow(
				bs.Name,
				string(bs.Spec.Type),
				tb,
				string(bs.Status.Phase),
				util.HumanizeDuration(time.Since(bs.CreationTimestamp.Time).Round(time.Second)),
			)
		}
	}
	fmt.Print(table.String())
}

// RunReconcile runs a CLI command
func RunReconcile(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <namespace-store-name> %s`, cmd.UsageString())
	}
	namespaceStoreName := args[0]
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      namespaceStoreName,
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

// MapSecretToNamespaceStores returns a list of namespacestores that uses the secret in their secretRefernce
// used by namespacestore_contorller to watch secrets changes
func MapSecretToNamespaceStores(secret types.NamespacedName) []reconcile.Request {
	log := util.Logger()
	log.Infof("checking which namespaceStores to reconcile. mapping secret %v to namespaceStores", secret)
	nsList := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	if !util.KubeList(nsList, &client.ListOptions{Namespace: secret.Namespace}) {
		log.Infof("Cloud not found namespaceStores in namespace %q", secret.Namespace)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, ns := range nsList.Items {
		nsSecret, err := util.GetNamespaceStoreSecret(&ns)
		if nsSecret.Name == secret.Name && err != nil {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      ns.Name,
					Namespace: ns.Namespace,
				},
			})
		}
	}
	log.Infof("will reconcile these namespaceStores: %v", reqs)

	return reqs
}

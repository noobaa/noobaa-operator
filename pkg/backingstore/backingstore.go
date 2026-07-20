package backingstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	sigyaml "sigs.k8s.io/yaml"
)

const (
	S3Compatible          = "s3-compatible"
	IBMCOS                = "ibm-cos"
	AWSS3                 = "aws-s3"
	AWSSTSS3              = "aws-sts-s3"
	GoogleCloudStorage    = "google-cloud-storage"
	GoogleCloudStorageSTS = "google-cloud-storage-sts"
	AzureBlob             = "azure-blob"
	AzureSTSBlob          = "azure-sts-blob"
	PVPool                = "pv-pool"
)

var validBackingStores = []string{S3Compatible, IBMCOS, AWSS3, AWSSTSS3, GoogleCloudStorage,
	GoogleCloudStorageSTS, AzureBlob, AzureSTSBlob, PVPool}

var ctx = context.TODO()

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backingstore",
		Short: "Manage backing stores",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdUpdate(),
		CmdDelete(),
		CmdStatus(),
		CmdList(),
		CmdReconcile(),
		CmdRunRemovePendingPods(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <backing-store-type> <backing-store-name>",
		Short: "Create backing store",
		Run:   RunCreate,
	}
	cmd.AddCommand(
		CmdCreateAWSS3(),
		CmdCreateAWSSTSS3(),
		CmdCreateS3Compatible(),
		CmdCreateIBMCos(),
		CmdCreateAzureBlob(),
		CmdCreateGoogleCloudStorage(),
		CmdCreateGoogleCloudStorageSTS(),
		CmdCreatePVPool(),
		CmdCreateAzureSTS(),
	)
	return cmd
}

func CmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <backing-store-name>",
		Short: "Update a backing store",
		Run:   RunUpdate,
	}
	cmd.Flags().String(
		"endpoint", "",
		"New endpoint URL",
	)

	return cmd
}

// CmdCreateAWSSTSS3 returns a cli command
func CmdCreateAWSSTSS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-sts-s3 <backing-store-name>",
		Short: "Create aws-s3 backing store (using STS, short-lived credentials)",
		Run:   RunCreateAWSSTSS3,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on the cloud",
	)
	cmd.Flags().String(
		"aws-sts-arn", "",
		"The AWS STS Role ARN which will assume role",
	)
	cmd.Flags().String(
		"region", "",
		"The AWS bucket region",
	)
	return cmd
}

// CmdCreateAzureSTS returns a CLI command
func CmdCreateAzureSTS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-sts-blob <backing-store-name>",
		Short: "Create azure-blob backing store (using STS, short-lived credentials)",
		Run:   RunCreateAzureSTS,
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
		"tenant-id", "",
		"The Azure Tenant ID which will assume role",
	)
	cmd.Flags().String(
		"client-id", "",
		"The Azure Client ID which will assume role",
	)
	cmd.Flags().String(
		"secret-name", "",
		"Optional name of an existing secret containing azure_tenant_id (and optionally AccountName); if omitted, a secret is created from flags",
	)
	return cmd
}

// CmdCreateAWSS3 returns a CLI command
func CmdCreateAWSS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-s3 <backing-store-name>",
		Short: "Create aws-s3 backing store (using long-lived credentials)",
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
		Use:   "s3-compatible <backing-store-name>",
		Short: "Create s3-compatible backing store",
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
		Use:   "ibm-cos <backing-store-name>",
		Short: "Create ibm-cos backing store",
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
		Use:   "azure-blob <backing-store-name>",
		Short: "Create azure-blob backing store",
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

// CmdCreateGoogleCloudStorage returns a CLI command
func CmdCreateGoogleCloudStorage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "google-cloud-storage <backing-store-name>",
		Short: "Create google-cloud-storage backing store (using long-lived credentials)",
		Run:   RunCreateGoogleCloudStorage,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on Google cloud storage",
	)
	cmd.Flags().String(
		"private-key-json-file", "",
		`private-key-json-file is the path to the json file provided by google for service account authentication`,
	)
	cmd.Flags().String(
		"secret-name", "",
		`The name of a secret for authentication - should have `+util.GoogleServiceAccountPrivateKeyJson+` property`,
	)
	return cmd
}

// CmdCreateGoogleCloudStorageSTS returns a CLI command
func CmdCreateGoogleCloudStorageSTS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "google-cloud-storage-sts <backing-store-name>",
		Short: "Create google-cloud-storage backing store using GCP WIF (STS, short-lived credentials)",
		Run:   RunCreateGoogleCloudStorageSTS,
	}
	cmd.Flags().String(
		"target-bucket", "",
		"The target bucket name on Google cloud storage",
	)
	cmd.Flags().String(
		"service-account-email", "",
		"The GCP service account email to impersonate",
	)
	cmd.Flags().String(
		"project-number", "",
		"The GCP project number (numeric string, e.g. 123456789; not the project ID)",
	)
	cmd.Flags().String(
		"pool-id", "",
		"The GCP workload identity pool ID",
	)
	cmd.Flags().String(
		"provider-id", "",
		"The GCP workload identity provider ID",
	)
	cmd.Flags().String(
		"secret-name", "",
		`The name of a secret for authentication - should have `+util.GoogleCredentialsJson+` property (external_account JSON)`,
	)
	return cmd
}

const (
	// Default volume size for pv-pool backing store
	defaultVolumeSize = int64(20 * 1024 * 1024 * 1024) // 20Gi=20*1024^3
)

// CmdCreatePVPool returns a CLI command
func CmdCreatePVPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pv-pool <backing-store-name>",
		Short: "Create pv-pool backing store",
		Run:   RunCreatePVPool,
	}
	cmd.Flags().Uint32(
		"num-volumes", 0,
		`Number of volumes in the store`,
	)
	cmd.Flags().Uint32(
		"pv-size-gb", 0,
		`PV size of each volume in the store`,
	)
	cmd.Flags().String(
		"request-cpu", "",
		"Request cpu for an agent pod (default: determined by performance profile)",
	)
	cmd.Flags().String(
		"request-memory", "",
		"Request memory for an agent pod (default: determined by performance profile)",
	)
	cmd.Flags().String(
		"limit-cpu", "",
		"Limit cpu for an agent pod (default: determined by performance profile)",
	)
	cmd.Flags().String(
		"limit-memory", "",
		"Limit memory for an agent pod (default: determined by performance profile)",
	)
	cmd.Flags().String(
		"storage-class", "",
		"The storage class to use for PV provisioning",
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

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <backing-store-name>",
		Short: "Status backing store",
		Run:   RunStatus,
	}
	return cmd
}

// CmdRunRemovePendingPods returns a CLI command
func CmdRunRemovePendingPods() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove_pending <backing-store-name>",
		Short: "Deletes all the pending pods that failed to connect to server",
		Run:   RunRemovePendingPods,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backing stores",
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

func createCommon(cmd *cobra.Command, args []string, storeType nbv1.StoreType, populate func(backStore *nbv1.BackingStore, secret *corev1.Secret)) {

	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}
	name := args[0]
	secretName, _ := cmd.Flags().GetString("secret-name")

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = name
	backStore.Namespace = options.Namespace
	backStore.Spec = nbv1.BackingStoreSpec{Type: storeType}

	o = util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)
	secret.Namespace = options.Namespace
	secret.Name = fmt.Sprintf("backing-store-%s-%s", storeType, name)
	secret.StringData = map[string]string{}
	secret.Data = nil

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(backStore), backStore)
	if err == nil {
		log.Fatalf(`❌ BackingStore %q already exists in namespace %q`, backStore.Name, backStore.Namespace)
	}

	populate(backStore, secret)

	if storeType == nbv1.StoreTypePVPool && backStore.Spec.PVPool != nil {
		profileDefaults := system.GetPVPoolResources(sys)
		effective := corev1.ResourceRequirements{
			Requests: profileDefaults.Requests.DeepCopy(),
			Limits:   profileDefaults.Limits.DeepCopy(),
		}
		if vr := backStore.Spec.PVPool.VolumeResources; vr != nil {
			applyResourceOverrides(&effective, vr)
		}
		if err := validateResourceRequestsVsLimits(effective); err != nil {
			log.Fatalf(`❌ PV pool resource conflict (against profile defaults): %v`, err)
		}
	}

	if secretName != "" {
		if !util.KubeCheck(secret) {
			log.Fatalf(`❌ Could not find the suggested secret: name %q namespace %q`, secret.Name, secret.Namespace)
			return
		}
	}

	validationErr := validations.ValidateBackingStore(*backStore)
	if validationErr != nil {
		log.Fatalf(`❌ %s %s`, validationErr, cmd.UsageString())
	}

	suggestedSecret := util.CheckForIdenticalSecretsCreds(secret, string(storeType))
	if suggestedSecret != nil {
		var decision string
		log.Printf("Found a Secret in the system with the same credentials (%s)", suggestedSecret.Name)
		log.Printf("Note that using more then one secret with the same credentials is not supported")
		log.Printf("do you want to use it for this Backingstore? y/n")
		if _, err := fmt.Scanln(&decision); err != nil {
			log.Fatalf(`❌ Invalid input, please select y/n`)
		}
		if strings.ToLower(decision) == "y" {
			log.Printf("Will use %s as the Backingstore %s Secret", suggestedSecret.Name, backStore.Name)
			err := util.SetBackingStoreSecretRef(backStore, &corev1.SecretReference{
				Name:      suggestedSecret.Name,
				Namespace: suggestedSecret.Namespace,
			})
			if err != nil {
				log.Fatalf(`❌ %s`, err)
			}
		} else if strings.ToLower(decision) == "n" {
			log.Fatalf("Not creating Backingstore")
		}
	}

	// Create backing store CR
	util.Panic(controllerutil.SetControllerReference(sys, backStore, scheme.Scheme))
	if !util.KubeCreateFailExisting(backStore) {
		log.Fatalf(`❌ Could not create BackingStore %q in Namespace %q (conflict)`, backStore.Name, backStore.Namespace)
	}

	secretRef, _ := util.GetBackingStoreSecret(backStore)
	if secretRef != nil && secretName == "" && suggestedSecret == nil {
		// Create secret
		util.Panic(controllerutil.SetControllerReference(backStore, secret, scheme.Scheme))
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
	log.Printf("BackingStore Wait Ready:")
	if WaitReady(backStore) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunCreate runs a cli command
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-type> %s`, cmd.UsageString())
	}
	if !slices.Contains(validBackingStores, args[0]) {
		log.Fatalf(`❌ Unsupported <backing-store-type> -> %s %s`, args[0], cmd.UsageString())
	}
}

// RunCreateAWSSTSS3 runs a cli command
func RunCreateAWSSTSS3(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}
	name := args[0]
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = name
	backStore.Namespace = options.Namespace
	backStore.Spec = nbv1.BackingStoreSpec{Type: nbv1.StoreTypeAWSS3}

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(backStore), backStore)
	if err == nil {
		log.Fatalf(`❌ BackingStore %q already exists in namespace %q`, backStore.Name, backStore.Namespace)
	}
	awsSTSARN := util.GetFlagStringOrPrompt(cmd, "aws-sts-arn")
	if !arn.IsARN(awsSTSARN) {
		log.Fatalf(`❌ aws-sts-arn %q is invalid`, awsSTSARN)
	}
	targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
	region, _ := cmd.Flags().GetString("region")
	backStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
		TargetBucket:  targetBucket,
		Region:        region,
		AWSSTSRoleARN: &awsSTSARN,
	}
	// Create backing store CR
	util.Panic(controllerutil.SetControllerReference(sys, backStore, scheme.Scheme))
	if !util.KubeCreateFailExisting(backStore) {
		log.Fatalf(`❌ Could not create BackingStore %q in Namespace %q (conflict)`, backStore.Name, backStore.Namespace)
	}
	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("BackingStore Wait Ready:")
	if WaitReady(backStore) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunCreateAWSS3 runs a CLI command
func RunCreateAWSS3(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.StoreTypeAWSS3, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
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

		backStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
			TargetBucket: targetBucket,
			Region:       region,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateAzureSTS runs a CLI command
func RunCreateAzureSTS(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}
	createCommon(cmd, args, nbv1.StoreTypeAzureBlob, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {

		secretName, _ := cmd.Flags().GetString("secret-name")
		targetBlobContainer := util.GetFlagStringOrPrompt(cmd, "target-blob-container")
		azureSTSClientID := ""
		if secretName == "" {
			azureSTSAccountName := util.GetFlagStringOrPromptPassword(cmd, "account-name")
			azureSTSTenantID := util.GetFlagStringOrPrompt(cmd, "tenant-id")
			azureSTSClientID = util.GetFlagStringOrPrompt(cmd, "client-id")
			if err := validations.ValidateAzureSTSCredsPresent(&targetBlobContainer, &azureSTSAccountName, &azureSTSTenantID, &azureSTSClientID); err != nil {
				log.Fatalf(`❌ %v %s`, err, cmd.UsageString())
			}
			secret.StringData["AccountName"] = azureSTSAccountName
			secret.StringData["azure_tenant_id"] = azureSTSTenantID
			secret.StringData["azure_client_id"] = azureSTSClientID
		} else {
			mandatoryProperties := []string{"AccountName", "azure_tenant_id", "azure_client_id"}
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
			util.KubeCheck(secret)
			azureSTSClientID = secret.StringData["azure_client_id"]
		}

		backStore.Spec.AzureBlob = &nbv1.AzureBlobSpec{
			TargetBlobContainer: targetBlobContainer,
			ClientId:            &azureSTSClientID,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateS3Compatible runs a CLI command
func RunCreateS3Compatible(cmd *cobra.Command, args []string) {
	createCommon(cmd, args, nbv1.StoreTypeS3Compatible, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
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

		backStore.Spec.S3Compatible = &nbv1.S3CompatibleSpec{
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
	createCommon(cmd, args, nbv1.StoreTypeIBMCos, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
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

		backStore.Spec.IBMCos = &nbv1.IBMCosSpec{
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
	createCommon(cmd, args, nbv1.StoreTypeAzureBlob, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
		targetBlobContainer := util.GetFlagStringOrPrompt(cmd, "target-blob-container")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{"AccountName", "AccountKey"}

		if secretName == "" {
			accountName := util.GetFlagStringOrPromptPassword(cmd, "account-name")
			accountKey := util.GetFlagStringOrPromptPassword(cmd, "account-key")
			secret.StringData["AccountName"] = accountName
			secret.StringData["AccountKey"] = accountKey
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		backStore.Spec.AzureBlob = &nbv1.AzureBlobSpec{
			TargetBlobContainer: targetBlobContainer,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateGoogleCloudStorage runs a CLI command
func RunCreateGoogleCloudStorage(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommon(cmd, args, nbv1.StoreTypeGoogleCloudStorage, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
		targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{util.GoogleServiceAccountPrivateKeyJson}

		if secretName == "" {
			privateKeyJSONFile := util.GetFlagStringOrPrompt(cmd, "private-key-json-file")
			bytes, err := os.ReadFile(privateKeyJSONFile)
			if err != nil {
				log.Fatalf("Failed to read file %q: %v", privateKeyJSONFile, err)
			}
			var privateKeyJSON map[string]interface{}
			err = json.Unmarshal(bytes, &privateKeyJSON)
			if err != nil {
				log.Fatalf("Failed to parse json file %q: %v", privateKeyJSONFile, err)
			}
			credentialsType, ok := privateKeyJSON["type"].(string)
			if !ok || credentialsType != "service_account" {
				log.Fatalf("GCP credentials JSON 'type' field must be a string with value %q", "service_account")
			}
			secret.StringData[util.GoogleServiceAccountPrivateKeyJson] = string(bytes)
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			util.VerifyGoogleCredentialsJSONTypeInSecret(secretName, options.Namespace, false)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		backStore.Spec.GoogleCloudStorage = &nbv1.GoogleCloudStorageSpec{
			TargetBucket: targetBucket,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreateGoogleCloudStorageSTS runs a CLI command
func RunCreateGoogleCloudStorageSTS(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommon(cmd, args, nbv1.StoreTypeGoogleCloudStorage, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
		targetBucket := util.GetFlagStringOrPrompt(cmd, "target-bucket")
		secretName, _ := cmd.Flags().GetString("secret-name")
		mandatoryProperties := []string{util.GoogleCredentialsJson}

		if secretName == "" {
			projectNumber := util.GetFlagStringOrPrompt(cmd, "project-number")
			poolID := util.GetFlagStringOrPrompt(cmd, "pool-id")
			providerID := util.GetFlagStringOrPrompt(cmd, "provider-id")
			serviceAccountEmail := util.GetFlagStringOrPrompt(cmd, "service-account-email")
			credentialsJSON, err := util.BuildGoogleWIFCredentialsJSON(projectNumber,
				poolID, providerID, serviceAccountEmail)
			if err != nil {
				log.Fatalf("Failed to build GCP WIF credentials: %v", err)
			}
			secret.StringData[util.GoogleCredentialsJson] = credentialsJSON
		} else {
			util.VerifyCredsInSecret(secretName, options.Namespace, mandatoryProperties)
			util.VerifyGoogleCredentialsJSONTypeInSecret(secretName, options.Namespace, true)
			secret.Name = secretName
			secret.Namespace = options.Namespace
		}

		backStore.Spec.GoogleCloudStorage = &nbv1.GoogleCloudStorageSpec{
			TargetBucket: targetBucket,
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

// RunCreatePVPool runs a CLI command
func RunCreatePVPool(cmd *cobra.Command, args []string) {
	log := util.Logger()
	createCommon(cmd, args, nbv1.StoreTypePVPool, func(backStore *nbv1.BackingStore, secret *corev1.Secret) {
		numVolumes, _ := cmd.Flags().GetUint32("num-volumes")
		pvSizeGB, _ := cmd.Flags().GetUint32("pv-size-gb")
		storageClass, _ := cmd.Flags().GetString("storage-class")
		requestCPU, _ := cmd.Flags().GetString("request-cpu")
		requestMemory, _ := cmd.Flags().GetString("request-memory")
		limitCPU, _ := cmd.Flags().GetString("limit-cpu")
		limitMemory, _ := cmd.Flags().GetString("limit-memory")
		pvPoolName := args[0]
		if len(pvPoolName) > 43 {
			log.Fatalf(`❌ Number of characters in <backing-store-name> should not exceed 63 `)
		}
		if numVolumes == 0 {
			fmt.Printf("Enter number of volumes: ")
			_, err := fmt.Scan(&numVolumes)
			if err != nil {
				log.Fatalf(`❌ Number of volumes must be a positive number %s`, cmd.UsageString())
			}
			if numVolumes == 0 {
				log.Fatalf(`❌ Missing number of volumes %s`, cmd.UsageString())
			}
		}
		if numVolumes > 20 {
			log.Fatalf(`❌ Number of volumes seems to be too large %d, maximal size of volumes is 20 %s`, numVolumes, cmd.UsageString())
		}

		if pvSizeGB == 0 {
			fmt.Printf("Enter PV size (GB): ")
			_, err := fmt.Scan(&pvSizeGB)
			if err != nil {
				log.Fatalf(`❌ PV size (GB) must be a positive number %s`, cmd.UsageString())
			}
			if pvSizeGB == 0 {
				log.Fatalf(`❌ Missing PV size (GB) %s`, cmd.UsageString())
			}
		}
		if pvSizeGB > 1024 {
			log.Fatalf(`❌ PV size seems to be too large %d %s`, pvSizeGB, cmd.UsageString())
		}
		if pvSizeGB < 16 {
			log.Fatalf(`❌ PV size seems to be too small (%dGB), minimal size for a pv is 16GB %s`, pvSizeGB, cmd.UsageString())
		}

		if storageClass != "" {
			sc := &storagev1.StorageClass{
				TypeMeta:   metav1.TypeMeta{Kind: "StorageClass"},
				ObjectMeta: metav1.ObjectMeta{Name: storageClass},
			}
			if !util.KubeCheck(sc) {
				log.Fatalf(`❌ Could not get StorageClass %q for system in namespace %q`,
					sc.Name, options.Namespace)
			}
			if strings.HasSuffix(sc.Provisioner, "/obc") || strings.HasSuffix(sc.Provisioner, "/bucket") {
				log.Fatalf(`❌ Could not set StorageClass %q for system in namespace %q - as this class reserved for obc only`,
					sc.Name, options.Namespace)
			}
		}

		volumeRequests := corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(int64(pvSizeGB)*1024*1024*1024, resource.BinarySI),
		}
		volumeLimits := corev1.ResourceList{}

		if requestCPU != "" {
			qty, err := resource.ParseQuantity(requestCPU)
			if err != nil {
				log.Fatalf(`❌ Could not parse request cpu %q`, requestCPU)
			}
			volumeRequests[corev1.ResourceCPU] = qty
		}
		if requestMemory != "" {
			qty, err := resource.ParseQuantity(requestMemory)
			if err != nil {
				log.Fatalf(`❌ Could not parse request memory %q`, requestMemory)
			}
			volumeRequests[corev1.ResourceMemory] = qty
		}
		if limitCPU != "" {
			qty, err := resource.ParseQuantity(limitCPU)
			if err != nil {
				log.Fatalf(`❌ Could not parse limit cpu %q`, limitCPU)
			}
			volumeLimits[corev1.ResourceCPU] = qty
		}
		if limitMemory != "" {
			qty, err := resource.ParseQuantity(limitMemory)
			if err != nil {
				log.Fatalf(`❌ Could not parse limit memory %q`, limitMemory)
			}
			volumeLimits[corev1.ResourceMemory] = qty
		}

		if reqCPU, hasReqCPU := volumeRequests[corev1.ResourceCPU]; hasReqCPU {
			if limCPU, hasLimCPU := volumeLimits[corev1.ResourceCPU]; hasLimCPU {
				if reqCPU.Cmp(limCPU) > 0 {
					log.Fatalf(`❌ Request CPU %v is larger than limit CPU %v`,
						reqCPU.String(), limCPU.String())
				}
			}
		}
		if reqMem, hasReqMem := volumeRequests[corev1.ResourceMemory]; hasReqMem {
			if limMem, hasLimMem := volumeLimits[corev1.ResourceMemory]; hasLimMem {
				if reqMem.Cmp(limMem) > 0 {
					log.Fatalf(`❌ Request memory %v is larger than limit memory %v`,
						reqMem.String(), limMem.String())
				}
			}
		}

		backStore.Spec.PVPool = &nbv1.PVPoolSpec{
			StorageClass: storageClass,
			NumVolumes:   int(numVolumes),
			VolumeResources: &corev1.VolumeResourceRequirements{
				Requests: volumeRequests,
				Limits:   volumeLimits,
			},
			Secret: corev1.SecretReference{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
		}
	})
}

func RunUpdate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	endpoint, err := cmd.Flags().GetString("endpoint")
	endpoint = strings.TrimSpace(endpoint)
	if err != nil || len(endpoint) == 0 {
		log.Fatalf("❌ Endpoint is required for update command")
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml)
	bs := o.(*nbv1.BackingStore)
	bs.Name = args[0]
	bs.Namespace = options.Namespace

	err = util.KubeClient().Get(util.Context(), util.ObjectKey(bs), bs)
	if err != nil {
		log.Fatalf("❌ Could not fetch the backing store: %v", err)
	}

	storeType := bs.Spec.Type
	if !slices.Contains(validBackingStores, string(storeType)) {
		log.Fatalf(`❌ Unsupported <backing-store-type> %s`, storeType)
	} else if storeType != IBMCOS && storeType != S3Compatible {
		log.Fatalf("❌ Target endpoint for backing store type %s cannot be updated", storeType)
	}

	switch storeType {
	case S3Compatible:
		if bs.Spec.S3Compatible == nil {
			log.Fatalf(`❌ Invalid backing store spec`)
		}
		bs.Spec.S3Compatible.Endpoint = endpoint
	case IBMCOS:
		if bs.Spec.IBMCos == nil {
			log.Fatalf(`❌ Invalid backing store spec`)
		}
		bs.Spec.IBMCos.Endpoint = endpoint
	}

	success := util.KubeUpdate(bs)
	if !success {
		log.Fatalf("❌ Error updating backing store %s", bs.Name)
	}

	log.Printf("✅ BackingStore %s Spec updated. Changes will be applied", bs.Name)
	log.Printf("")
	log.Printf("BackingStore Wait Ready:")

	if WaitReady(bs) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = args[0]
	backStore.Namespace = options.Namespace
	backStore.Spec = nbv1.BackingStoreSpec{}

	nbClient := system.GetNBClient()

	poolinfo, err := nbClient.ReadPoolAPI(nb.ReadPoolParams{Name: backStore.Name})
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if !isRPCErr || rpcErr.RPCCode != "NO_SUCH_POOL" {
			log.Fatalf(`❌ Failed to read BackingStore info: %s`, err)
		}
	} else if poolinfo.Undeletable != "" && poolinfo.Undeletable != "IS_BACKINGSTORE" {
		switch poolinfo.Undeletable {
		case "CONNECTED_BUCKET_DELETING":
			fallthrough
		case "IN_USE":
			log.Fatalf(`❌ Could not delete BackingStore %q in namespace %q as it is being used by one or more buckets`,
				backStore.Name, backStore.Namespace)

		case "DEFAULT_RESOURCE":
			log.Fatalf(`❌ Could not delete BackingStore %q in namespace %q as it is the default resource of one or more accounts`,
				backStore.Name, backStore.Namespace)
		default:
			log.Fatalf(`❌ Could not delete BackingStore %q in namespace %q, undeletable due to %q`,
				backStore.Name, backStore.Namespace, poolinfo.Undeletable)
		}
	}
	if !util.KubeDelete(backStore) {
		log.Fatalf(`❌ Could not delete BackingStore %q in namespace %q`,
			backStore.Name, backStore.Namespace)
	}
}

// RunRemovePendingPods runs a CLI command
func RunRemovePendingPods(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml)
	backStore := o.(*nbv1.BackingStore)
	backStore.Name = args[0]
	backStore.Namespace = options.Namespace
	backStore.Spec = nbv1.BackingStoreSpec{}
	if !util.KubeCheck(backStore) {
		log.Fatalf(`❌ Could not get BackingStore %q in namespace %q`,
			backStore.Name, backStore.Namespace)
	}
	if backStore.Spec.Type != nbv1.StoreTypePVPool {
		log.Fatalf(`❌ Could not get Run this Command on None PV-Pool backingstore`)
	}

	nbClient := system.GetNBClient()
	hostsInfo, err := nbClient.ListHostsAPI(nb.ListHostsParams{Query: nb.ListHostsQuery{Pools: []string{backStore.Name}}})
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if !isRPCErr || rpcErr.RPCCode != "NO_SUCH_POOL" {
			log.Fatalf(`❌ Failed to read BackingStore host info: %s`, err)
		}
	}
	podsList := &corev1.PodList{}
	util.KubeList(podsList, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": backStore.Name})
	for _, pod := range podsList.Items {
		if !isPodinNoobaa(&pod, &hostsInfo.Hosts) {
			util.RemoveFinalizer(&pod, nbv1.Finalizer)
			if !util.KubeUpdate(&pod) {
				log.Errorf("Pod %q failed to remove finalizer %q",
					pod.Name, nbv1.Finalizer)
			}
			pvc := &corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pod.Spec.Volumes[1].PersistentVolumeClaim.ClaimName,
					Namespace: options.Namespace,
				},
			}
			util.KubeDelete(&pod)
			util.KubeDelete(pvc)
		}
	}
}

func isPodinNoobaa(pod *corev1.Pod, hostsInfo *[]nb.HostInfo) bool {
	for _, host := range *hostsInfo {
		if strings.HasPrefix(host.Name, pod.Name) {
			return true
		}
	}
	return false
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	backStore := GetBackingStoreFromArgs(cmd, args)
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secretRef, _ := util.GetBackingStoreSecret(backStore)
	if !util.IsAWSSTSClusterBS(backStore) {
		if secretRef != nil {
			secret.Name = secretRef.Name
			secret.Namespace = secretRef.Namespace
			if secret.Namespace == "" {
				secret.Namespace = backStore.Namespace
			}
			if backStore.Spec.Type != nbv1.StoreTypePVPool && !util.KubeCheck(secret) {
				log.Errorf(`❌ Could not get Secret %q in namespace %q`,
					secret.Name, secret.Namespace)
			}
		}
	}

	CheckPhase(backStore)

	fmt.Println()
	fmt.Println("# BackingStore spec:")
	output, err := sigyaml.Marshal(backStore.Spec)
	util.Panic(err)
	fmt.Print(string(output))
	fmt.Println()
	if secretRef != nil && secret.Name != "" {
		_, err = sigyaml.Marshal(secret.StringData)
		util.Panic(err)
		fmt.Println()
	}
}

// GetBackingStoreFromArgs returns the backingstore from CLI arg
func GetBackingStoreFromArgs(cmd *cobra.Command, args []string) *nbv1.BackingStore {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}

	backStore := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
	backStore.Name = args[0]
	backStore.Namespace = options.Namespace
	backStore.Spec = nbv1.BackingStoreSpec{}
	if !util.KubeCheck(backStore) {
		log.Fatalf(`❌ Could not get BackingStore %q in namespace %q`,
			backStore.Name, backStore.Namespace)
	}
	return backStore
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady(backStore *nbv1.BackingStore) bool {
	log := util.Logger()
	klient := util.KubeClient()

	interval := time.Duration(30)

	err := wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(backStore), backStore)
		if err != nil {
			log.Printf("⏳ Failed to get BackingStore: %s", err)
			return false, nil
		}
		CheckPhase(backStore)
		if backStore.Status.Phase == nbv1.BackingStorePhaseRejected {
			return false, fmt.Errorf("BackingStorePhaseRejected")
		}
		if backStore.Status.Phase != nbv1.BackingStorePhaseReady {
			return false, nil
		}
		return true, nil
	})
	return (err == nil)
}

// CheckPhase prints the phase and reason for it
func CheckPhase(backStore *nbv1.BackingStore) {
	log := util.Logger()

	reason := "waiting..."
	for _, c := range backStore.Status.Conditions {
		if c.Type == "Available" {
			reason = fmt.Sprintf("%s %s", c.Reason, c.Message)
		}
	}

	switch backStore.Status.Phase {

	case nbv1.BackingStorePhaseReady:
		log.Printf("✅ BackingStore %q Phase is Ready", backStore.Name)

	case nbv1.BackingStorePhaseRejected:
		log.Errorf("❌ BackingStore %q Phase is %q: %s", backStore.Name, backStore.Status.Phase, reason)

	case nbv1.BackingStorePhaseVerifying:
		fallthrough
	case nbv1.BackingStorePhaseConnecting:
		fallthrough
	case nbv1.BackingStorePhaseCreating:
		fallthrough
	case nbv1.BackingStorePhaseDeleting:
		fallthrough
	default:
		log.Printf("⏳ BackingStore %q Phase is %q: %s", backStore.Name, backStore.Status.Phase, reason)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No backing stores found.\n")
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
		tb, err := util.GetBackingStoreTargetBucket(bs)
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
		log.Fatalf(`❌ Missing expected arguments: <backing-store-name> %s`, cmd.UsageString())
	}
	backingStoreName := args[0]
	klient := util.KubeClient()
	interval := time.Duration(3)
	util.Panic(wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
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
		if res.RequeueAfter != 0 {
			log.Printf("\nRetrying in %d seconds\n", interval)
			return false, nil
		}
		return true, nil
	}))
}

// MapSecretToBackingStores returns a list of backingstores that uses the secret in their secretReference
// used by backingstore_controller to watch secrets changes
func MapSecretToBackingStores(secret types.NamespacedName) []reconcile.Request {
	log := util.Logger()
	bsList := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(bsList, &client.ListOptions{Namespace: secret.Namespace}) {
		log.Infof("Could not found backingStores in namespace %q, while trying to find Backingstore that uses %s secret", secret.Namespace, secret.Name)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, bs := range bsList.Items {
		bsSecret, err := util.GetBackingStoreSecret(&bs)
		if err != nil {
			log.Error(err)
		}
		if bsSecret != nil && bsSecret.Name == secret.Name {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      bs.Name,
					Namespace: bs.Namespace,
				},
			})
		}
	}

	return reqs
}

// MapNoobaaToBackingStores returns a list of backingstores that are inside Noobaa system
// used by backingstore_controller to watch Noobaa CR changes
func MapNoobaaToBackingStores(noobaa types.NamespacedName) []reconcile.Request {
	log := util.Logger()
	bsList := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(bsList, &client.ListOptions{Namespace: noobaa.Namespace}) {
		log.Infof("Could not found backingStores in namespace %q, while trying to find Backingstore that inside %s Noobaa system", noobaa.Namespace, noobaa.Name)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, bs := range bsList.Items {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      bs.Name,
				Namespace: bs.Namespace,
			},
		})
	}

	return reqs
}

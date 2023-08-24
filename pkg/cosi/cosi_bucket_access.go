package cosi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

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

var ctx = context.TODO()

// CmdCOSIBucketAccessClaim returns a CLI command
func CmdCOSIBucketAccessClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accessclaim",
		Short: "Manage cosi access claims",
	}
	cmd.AddCommand(
		CmdCreateBucketAccessClaim(),
		CmdDeleteBucketAccessClaim(),
		CmdStatusBucketAccessClaim(),
		CmdListBucketAccessClaim(),
	)
	return cmd
}

// CmdCreateBucketAccessClaim returns a CLI command
func CmdCreateBucketAccessClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-access-claim-name>",
		Short: "Create a COSI bucket access claim",
		Run:   RunCreateBucketAccessClaim,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket access claim should be created")
	cmd.Flags().String("bucket-claim", "",
		"Set the bucket claim name to which the user require access credentials")
	cmd.Flags().String("bucket-access-class", "",
		"Set bucket access class name to specify the bucket access policy")
	cmd.Flags().String("creds-secret-name", "",
		"Set the secret name in which COSI will set the access credentials to the bucket")
	return cmd
}

// CmdDeleteBucketAccessClaim returns a CLI command
func CmdDeleteBucketAccessClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-access-claim-name>",
		Short: "Delete a COSI bucket access claim",
		Run:   RunDeleteBucketAccessClaim,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket access claim exists")
	return cmd
}

// CmdStatusBucketAccessClaim returns a CLI command
func CmdStatusBucketAccessClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-access-claim-name>",
		Short: "Status of a COSI bucket access claim",
		Run:   RunStatusBucketAccessClaim,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the COSI bucket access claim exists")
	return cmd
}

// CmdListBucketAccessClaim returns a CLI command
func CmdListBucketAccessClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List COSI bucket access claims",
		Run:   RunListBucketAccessClaim,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreateBucketAccessClaim runs a CLI command
func RunCreateBucketAccessClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}
	name := args[0]

	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	bucketClaimName := util.GetFlagStringOrPrompt(cmd, "bucket-claim")
	bucketAccessClassName := util.GetFlagStringOrPrompt(cmd, "bucket-access-class")
	credsSecretName := util.GetFlagStringOrPrompt(cmd, "creds-secret-name")

	cosiBucketAccessClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_access_claim_yaml).(*nbv1.COSIBucketAccessClaim)
	cosiBucketAccessClaim.Name = name
	cosiBucketAccessClaim.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketAccessClaim.Namespace = appNamespace
	}

	cosiBucketAccessClaim.Spec.BucketClaimName = bucketClaimName
	cosiBucketAccessClaim.Spec.CredentialsSecretName = credsSecretName
	cosiBucketAccessClaim.Spec.BucketAccessClassName = bucketAccessClassName

	bucketAccessClass := util.KubeObject(bundle.File_deploy_cosi_bucket_access_class_yaml).(*nbv1.COSIBucketAccessClass)
	bucketAccessClass.Name = bucketAccessClassName
	if !util.KubeCheck(bucketAccessClass) {
		log.Fatalf(`❌ Could not get bucketAccessClass %q`, bucketAccessClass.Name)
	}

	bucketClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_claim_yaml).(*nbv1.COSIBucketClaim)
	bucketClaim.Name = bucketClaimName
	bucketClaim.Namespace = cosiBucketAccessClaim.Namespace
	if !util.KubeCheck(bucketClaim) {
		log.Fatalf(`❌ Could not get BucketClaim %q`, bucketClaim.Name)
	}

	// NOTE - when/if extra parameters are supported in bucketAccessClass we will need to validate them here

	if !util.KubeCreateFailExisting(cosiBucketAccessClaim) {
		log.Fatalf(`❌ Could not create COSI bucket access claim %q in namespace %q (conflict)`, cosiBucketAccessClaim.Name, cosiBucketAccessClaim.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("COSI bucket access claim Wait Ready:")
	if WaitBucketAccessClaimReady(cosiBucketAccessClaim) {
		log.Printf("")
		log.Printf("")
		RunStatusBucketAccessClaim(cmd, args)
	}
}

// RunDeleteBucketAccessClaim runs a CLI command
func RunDeleteBucketAccessClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-access-claim-name> %s`, cmd.UsageString())
	}
	appNamespace, _ := cmd.Flags().GetString("app-namespace")

	cosiBucketAccessClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_access_claim_yaml).(*nbv1.COSIBucketAccessClaim)
	cosiBucketAccessClaim.Name = args[0]
	cosiBucketAccessClaim.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketAccessClaim.Namespace = appNamespace
	}

	if !util.KubeDelete(cosiBucketAccessClaim) {
		log.Fatalf(`❌ Could not delete COSI bucket access claim %q in namespace %q`,
			cosiBucketAccessClaim.Name, cosiBucketAccessClaim.Namespace)
	}
}

// RunStatusBucketAccessClaim runs a CLI command
func RunStatusBucketAccessClaim(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-access-claim-name> %s`, cmd.UsageString())
	}

	appNamespace, _ := cmd.Flags().GetString("app-namespace")

	cosiBucketAccessClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_access_claim_yaml).(*nbv1.COSIBucketAccessClaim)
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)

	cosiBucketAccessClaim.Name = args[0]
	cosiBucketAccessClaim.Namespace = options.Namespace
	secret.Namespace = options.Namespace
	if appNamespace != "" {
		cosiBucketAccessClaim.Namespace = appNamespace
		secret.Namespace = appNamespace

	}

	if !util.KubeCheck(cosiBucketAccessClaim) {
		log.Fatalf(`❌ Could not find COSI bucket access claim %q in namespace %q`, cosiBucketAccessClaim.Name, cosiBucketAccessClaim.Namespace)
	}

	bucketAccessClass := &nbv1.COSIBucketAccessClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketAccessClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name: cosiBucketAccessClaim.Spec.BucketAccessClassName,
		},
	}

	if !util.KubeCheck(bucketAccessClass) {
		log.Errorf(`❌ Could not get BucketAccessClass %s`, bucketAccessClass.Name)
	}

	secret.Name = cosiBucketAccessClaim.Spec.CredentialsSecretName
	if !util.KubeCheck(secret) {
		log.Fatalf(`❌ Could not find COSI bucket access claim secret %q in namespace %q`, secret.Name, cosiBucketAccessClaim.Namespace)
	}

	sysClient, err := system.Connect(true)
	if err != nil {
		util.Logger().Fatalf("❌ %s", err)
	}
	var a *nb.AccountInfo
	if cosiBucketAccessClaim.Status.AccountID != "" {
		nbClient := sysClient.NBClient
		account, err := nbClient.ReadAccountAPI(nb.ReadAccountParams{Email: cosiBucketAccessClaim.Status.AccountID})
		if err == nil {
			a = &account
		}
	}

	fmt.Printf("\n")
	fmt.Printf("COSI BucketAccessClaim info:\n")
	fmt.Printf("  %-22s : %t\n", "Bucket Access Granted", cosiBucketAccessClaim.Status.AccessGranted)
	fmt.Printf("  %-22s : kubectl get -n %s bucketaccessclaim %s\n", "COSIBucketAccessClaim", cosiBucketAccessClaim.Namespace, cosiBucketAccessClaim.Name)
	fmt.Printf("  %-22s : kubectl get bucketaccessclasses.objectstorage.k8s.io %s\n", "BucketAccessClass", bucketAccessClass.Name)
	fmt.Printf("\n")

	credsEnv := ""
	bucketInfo := secret.StringData["BucketInfo"]
	if bucketInfo != "" {
		var bucketInfoObj nbv1.COSIBucketInfo
		if err := json.Unmarshal([]byte(bucketInfo), &bucketInfoObj); err != nil {
			log.Fatal("error deserializing secret BucketInfo")
		}
		accessKey := bucketInfoObj.Spec.S3.AccessKeyID
		secretKey := bucketInfoObj.Spec.S3.AccessSecretKey
		accessKeyProperty := "AWS_ACCESS_KEY_ID"
		secretKeyProperty := "AWS_SECRET_ACCESS_KEY"
		if options.ShowSecrets {
			fmt.Printf("  %-22s : %s\n", accessKeyProperty, accessKey)
			fmt.Printf("  %-22s : %s\n", accessKeyProperty, secretKey)
			credsEnv += accessKeyProperty + "=" + accessKey + " "
			credsEnv += secretKeyProperty + "=" + secretKey + " "
		} else {
			fmt.Printf("  %-22s : %s\n", accessKeyProperty, nb.MaskedString(accessKey))
			fmt.Printf("  %-22s : %s\n", accessKeyProperty, nb.MaskedString(secretKey))
			credsEnv += accessKeyProperty + "=" + string(nb.MaskedString(accessKey)) + " "
			credsEnv += secretKeyProperty + "=" + string(nb.MaskedString(secretKey)) + " "
		}

	}

	fmt.Printf("Shell commands:\n")
	fmt.Printf("  %-22s : alias s3='%saws s3 --no-verify-ssl --endpoint-url %s'\n", "AWS S3 Alias", credsEnv, sysClient.S3URL.String())
	fmt.Printf("\n")
	if a != nil {
		fmt.Printf("Account status:\n")
		fmt.Printf("  %-22s : %s\n", "Name", a.Name)
		fmt.Printf("  %-22s : %s\n", "Email", a.Email)
		fmt.Printf("  %-22s : %s\n", "DefaultResource", a.DefaultResource)
		fmt.Printf("  %-22s : %t\n", "S3Access", a.HasS3Access)
		fmt.Printf("  %-22s : %t\n", "AllowBucketCreate", a.CanCreateBuckets)
		fmt.Printf("\n")
	}

}

// RunListBucketAccessClaim runs a CLI command
func RunListBucketAccessClaim(cmd *cobra.Command, args []string) {
	list := &nbv1.COSIBucketAccessClaimList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketAccessClaim"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No COSI bucket access claims found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"ACCOUNT-NAME",
		"BUCKET-CLAIM",
		"BUCKET-ACCESS-CLASS",
		"ACCESS-GRANTED",
	)
	for i := range list.Items {
		cosiBucketAccessClaim := &list.Items[i]
		table.AddRow(
			cosiBucketAccessClaim.Namespace,
			cosiBucketAccessClaim.Name,
			cosiBucketAccessClaim.Status.AccountID,
			cosiBucketAccessClaim.Spec.BucketClaimName,
			cosiBucketAccessClaim.Spec.BucketAccessClassName,
			fmt.Sprintf("%t", bool(cosiBucketAccessClaim.Status.AccessGranted)),
		)
	}
	fmt.Print(table.String())
}

// WaitBucketAccessClaimReady waits until the cosi bucket claim status bucket ready changes to true
func WaitBucketAccessClaimReady(cosiBucketAccessClaim *nbv1.COSIBucketAccessClaim) bool {
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
		err := klient.Get(util.Context(), util.ObjectKey(cosiBucketAccessClaim), cosiBucketAccessClaim)
		if err != nil {
			log.Printf("⏳ Failed to get COSI bucket access claim: %s", err)
			return false, nil
		}
		CheckBucketAccessClaimPhase(cosiBucketAccessClaim)
		if cosiBucketAccessClaim.Status.AccessGranted {
			return true, nil
		}
		return false, nil
	})
	return (err == nil)
}

// CheckBucketAccessClaimPhase prints the phase and reason for it
func CheckBucketAccessClaimPhase(cosiBucketAccessClaim *nbv1.COSIBucketAccessClaim) {
	log := util.Logger()
	if cosiBucketAccessClaim.Status.AccessGranted {
		log.Printf("✅ COSI bucket access claim %q granted", cosiBucketAccessClaim.Name)
	} else {
		log.Printf("⏳ COSI bucket access claim %q is not yet granted", cosiBucketAccessClaim.Name)
	}
}

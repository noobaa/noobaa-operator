package obc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	obv1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var ctx = context.TODO()

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obc",
		Short: "Manage object bucket claims",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdRegenerate(),
		CmdDelete(),
		CmdStatus(),
		CmdList(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <bucket-claim-name>",
		Short: "Create an OBC",
		Run:   RunCreate,
	}
	cmd.Flags().Bool("exact", false,
		"Request an exact bucketName instead of the default generateBucketName")
	cmd.Flags().String("bucketclass", "",
		"Set bucket class to specify the bucket policy")
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the OBC should be created")
	cmd.Flags().Int("gid", -1,
		"Set the GID for the NSFS account config")
	cmd.Flags().Int("uid", -1,
		"Set the UID for the NSFS account config")
	cmd.Flags().String("distinguished-name", "",
		"Set the distinguished name for the NSFS account config")
	cmd.Flags().String("path", "",
		"Set path to specify inner directory in namespace store target path, or in the case of NSFS - filesystem mount point (can be used only when specifying a namespace bucketclass)")
	cmd.Flags().String("replication-policy", "",
		"Set the json file path that contains replication rules")
	cmd.Flags().String("max-objects", "",
		"Set quota max objects quantity config to requested bucket")
	cmd.Flags().String("max-size", "",
		"Set quota max size config to requested bucket")
	return cmd
}

// CmdRegenerate returns a CLI command
func CmdRegenerate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate <bucket-claim-name>",
		Short: "Regenerate OBC S3 Credentials",
		Run:   RunRegenerate,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the OBC should be regenerated")
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bucket-claim-name>",
		Short: "Delete an OBC",
		Run:   RunDelete,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the OBC should be deleted")
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <bucket-claim-name>",
		Short: "Status of an OBC",
		Run:   RunStatus,
	}
	cmd.Flags().String("app-namespace", "",
		"Set the namespace of the application where the OBC should be created")
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBC's",
		Run:   RunList,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}
	name := args[0]

	exact, _ := cmd.Flags().GetBool("exact")
	bucketClassName, _ := cmd.Flags().GetString("bucketclass")
	path, _ := cmd.Flags().GetString("path")
	replicationPolicy, _ := cmd.Flags().GetString("replication-policy")
	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	if appNamespace == "" {
		appNamespace = options.Namespace
	}
	maxSize, _ := cmd.Flags().GetString("max-size")
	maxObjects, _ := cmd.Flags().GetString("max-objects")
	gid, _ := cmd.Flags().GetInt("gid")
	uid, _ := cmd.Flags().GetInt("uid")
	distinguishedName, _ := cmd.Flags().GetString("distinguished-name")

	if distinguishedName != "" && (gid > -1 || uid > -1) {
		log.Fatalf(`❌ NSFS account config cannot include both distinguished name and UID/GID`)
	}

	if (gid > -1 && uid == -1) || (gid == -1 && uid > -1) {
		log.Fatalf(`❌ NSFS account config must include both UID and GID as positive integers`)
	}

	if bucketClassName == "" && (gid > -1 || uid > -1 || distinguishedName != "") {
		log.Fatalf(`❌ NSFS account config cannot be set without an NSFS bucketclass`)
	}

	o := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml)
	obc := o.(*nbv1.ObjectBucketClaim)
	obc.Name = name
	obc.Namespace = appNamespace
	if exact {
		obc.Spec.BucketName = name
		obc.Spec.GenerateBucketName = ""
	} else {
		obc.Spec.BucketName = ""
		obc.Spec.GenerateBucketName = name
	}

	sc := &storagev1.StorageClass{
		TypeMeta:   metav1.TypeMeta{Kind: "StorageClass"},
		ObjectMeta: metav1.ObjectMeta{Name: options.SubDomainNS()},
	}
	if !util.KubeCheck(sc) {
		log.Fatalf(`❌ Could not get StorageClass %q for system in namespace %q`,
			sc.Name, options.Namespace)
	}
	obc.Spec.StorageClassName = sc.Name
	obc.Spec.AdditionalConfig = map[string]string{}

	if bucketClassName != "" {
		bucketClass := &nbv1.BucketClass{
			TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      bucketClassName,
				Namespace: options.Namespace,
			},
		}
		if !util.KubeCheck(bucketClass) {
			log.Fatalf(`❌ Could not get BucketClass %q in namespace %q`,
				bucketClass.Name, bucketClass.Namespace)
		}
		if bucketClass.Spec.NamespacePolicy == nil && path != "" {
			log.Fatalf(`❌ Could not create OBC %q with inner path while missing namespace bucketclass`, obc.Name)
		}
		obc.Spec.AdditionalConfig["bucketclass"] = bucketClassName
		obc.Spec.AdditionalConfig["path"] = path
		obc.Spec.AdditionalConfig["replicationPolicy"] = bucketClass.Spec.ReplicationPolicy

	} else if path != "" {
		log.Fatalf(`❌ Could not create OBC %q with inner path while missing namespace bucketclass`, obc.Name)
	}

	if replicationPolicy != "" {
		replication, err := util.LoadConfigurationJSON(replicationPolicy)
		if err != nil {
			log.Fatalf(`❌ %q`, err)
		}
		obc.Spec.AdditionalConfig["replicationPolicy"] = replication
	}

	if gid > -1 {
		var nsfsAccountConfig nbv1.AccountNsfsConfig
		nsfsAccountConfig.GID = &gid
		nsfsAccountConfig.UID = &uid
		nsfsAccountConfig.NsfsOnly = true
		marshalledCfg, _ := json.Marshal(nsfsAccountConfig)
		obc.Spec.AdditionalConfig["nsfsAccountConfig"] = string(marshalledCfg)
	}

	if distinguishedName != "" {
		var nsfsAccountConfig nbv1.AccountNsfsConfig
		nsfsAccountConfig.DistinguishedName = distinguishedName
		nsfsAccountConfig.GID = nil
		nsfsAccountConfig.UID = nil
		nsfsAccountConfig.NsfsOnly = true
		marshalledCfg, _ := json.Marshal(nsfsAccountConfig)
		obc.Spec.AdditionalConfig["nsfsAccountConfig"] = string(marshalledCfg)
	}

	if maxSize != "" {
		obc.Spec.AdditionalConfig["maxSize"] = maxSize
	}
	if maxObjects != "" {
		obc.Spec.AdditionalConfig["maxObjects"] = maxObjects
	}

	err := ValidateOBC(obc, true)
	if err != nil {
		log.Fatalf(`❌ Could not create OBC %q in namespace %q validation failed %q`, obc.Name, obc.Namespace, err)
	}

	if !util.KubeCreateFailExisting(obc) {
		log.Fatalf(`❌ Could not create OBC %q in namespace %q (conflict)`, obc.Name, obc.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("OBC Wait Ready:")
	if WaitReady(obc) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunRegenerate runs a CLI command
func RunRegenerate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}

	var decision string
	log.Printf("You are about to regenerate an OBC's security credentials.")
	log.Printf("This will invalidate all connections between S3 clients and NooBaa which are connected using the current credentials.")
	log.Printf("are you sure? y/n")

	for {
		if _, err := fmt.Scanln(&decision); err != nil {
			log.Printf(`are you sure? y/n`)
		}
		if decision == "y" {
			break
		} else if decision == "n" {
			return
		}
	}

	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	if appNamespace == "" {
		appNamespace = options.Namespace
	}

	name := args[0]

	obc := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml).(*nbv1.ObjectBucketClaim)
	ob := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucket_cr_yaml).(*nbv1.ObjectBucket)
	obc.Name = name
	obc.Namespace = appNamespace

	if !util.KubeCheck(obc) {
		log.Fatalf(`❌ Could not find OBC %q in namespace %q`,
			obc.Name, obc.Namespace)
	}

	if obc.Spec.ObjectBucketName != "" {
		ob.Name = obc.Spec.ObjectBucketName
	} else {
		ob.Name = fmt.Sprintf("ob-%s-%s", appNamespace, name)
	}

	if !util.KubeCheck(ob) {
		log.Fatalf(`❌ Could not find OB %q`, ob.Name)
	}

	accountName := ob.Spec.AdditionalState["account"]

	err := GenerateAccountKeys(name, accountName, appNamespace)
	if err != nil {
		log.Fatalf(`❌ Could not regenerate credentials for %q: %v`, name, err)
	}

	RunStatus(cmd, args)

}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}

	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	if appNamespace == "" {
		appNamespace = options.Namespace
	}

	o := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml)
	obc := o.(*nbv1.ObjectBucketClaim)
	obc.Name = args[0]
	obc.Namespace = appNamespace

	if !util.KubeCheck(obc) {
		log.Fatalf(`❌ Could not delete. OBC %q in namespace %q does not exist`, obc.Name, obc.Namespace)
	}

	if !util.KubeDelete(obc) {
		log.Fatalf(`❌ Could not delete OBC %q in namespace %q`,
			obc.Name, obc.Namespace)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-claim-name> %s`, cmd.UsageString())
	}

	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	if appNamespace == "" {
		appNamespace = options.Namespace
	}

	obc := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml).(*nbv1.ObjectBucketClaim)
	ob := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucket_cr_yaml).(*nbv1.ObjectBucket)
	sc := util.KubeObject(bundle.File_deploy_obc_storage_class_yaml).(*storagev1.StorageClass)
	cm := util.KubeObject(bundle.File_deploy_internal_configmap_empty_yaml).(*corev1.ConfigMap)
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)

	obc.Namespace = appNamespace
	cm.Namespace = appNamespace
	secret.Namespace = appNamespace

	obc.Name = args[0]
	cm.Name = args[0]
	secret.Name = args[0]

	if !util.KubeCheck(obc) {
		log.Fatalf(`❌ Could not find OBC %q in namespace %q`, obc.Name, obc.Namespace)
	}

	if obc.Spec.ObjectBucketName != "" {
		ob.Name = obc.Spec.ObjectBucketName
	} else {
		ob.Name = fmt.Sprintf("obc-%s-%s", appNamespace, args[0])
	}

	util.KubeCheck(ob)
	util.KubeCheck(cm)
	util.KubeCheck(secret)

	bucketClass := &nbv1.BucketClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      obc.Spec.AdditionalConfig["bucketclass"],
			Namespace: options.Namespace,
		},
	}
	sc.Name = obc.Spec.StorageClassName
	if sc.Name != "" {
		util.KubeCheck(sc)
		if bucketClass.Name == "" {
			bucketClass.Name = sc.Parameters["bucketclass"]
		}
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
	if obc.Spec.BucketName != "" {
		nbClient := sysClient.NBClient
		bucket, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: obc.Spec.BucketName})
		if err == nil {
			b = &bucket
		}
	}

	fmt.Printf("\n")
	fmt.Printf("ObjectBucketClaim info:\n")
	fmt.Printf("  %-22s : %s\n", "Phase", obc.Status.Phase)
	fmt.Printf("  %-22s : kubectl get -n %s objectbucketclaim %s\n", "ObjectBucketClaim", obc.Namespace, obc.Name)
	fmt.Printf("  %-22s : kubectl get -n %s configmap %s\n", "ConfigMap", cm.Namespace, cm.Name)
	fmt.Printf("  %-22s : kubectl get -n %s secret %s\n", "Secret", secret.Namespace, secret.Name)
	fmt.Printf("  %-22s : kubectl get objectbucket %s\n", "ObjectBucket", ob.Name)
	fmt.Printf("  %-22s : kubectl get storageclass %s\n", "StorageClass", sc.Name)
	fmt.Printf("  %-22s : kubectl get -n %s bucketclass %s\n", "BucketClass", bucketClass.Namespace, bucketClass.Name)
	fmt.Printf("\n")
	fmt.Printf("Connection info:\n")
	for k, v := range cm.Data {
		if v != "" {
			fmt.Printf("  %-22s : %s\n", k, v)
		}
	}
	credsEnv := ""
	for k, v := range secret.StringData {
		if v != "" {
			if options.ShowSecrets {
				fmt.Printf("  %-22s : %s\n", k, v)
				credsEnv += k + "=" + v + " "
			} else {
				fmt.Printf("  %-22s : %s\n", k, nb.MaskedString(v))
				credsEnv += k + "=" + string(nb.MaskedString(v)) + " "
			}
		}
	}
	fmt.Printf("\n")
	fmt.Printf("Shell commands:\n")
	fmt.Printf("  %-22s : alias s3='%saws s3 --no-verify-ssl --endpoint-url %s'\n", "AWS S3 Alias", credsEnv, sysClient.S3URL.String())
	fmt.Printf("\n")
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

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.ObjectBucketClaimList{
		TypeMeta: metav1.TypeMeta{Kind: "ObjectBucketClaim"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No OBCs found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"BUCKET-NAME",
		"STORAGE-CLASS",
		"BUCKET-CLASS",
		"PHASE",
	)
	scMap := map[string]*storagev1.StorageClass{}
	for i := range list.Items {
		obc := &list.Items[i]
		bucketClass := obc.Spec.AdditionalConfig["bucketclass"]
		if bucketClass == "" && obc.Spec.StorageClassName != "" {
			sc := scMap[obc.Spec.StorageClassName]
			if sc == nil {
				sc = util.KubeObject(bundle.File_deploy_obc_storage_class_yaml).(*storagev1.StorageClass)
				sc.Name = obc.Spec.StorageClassName
				if util.KubeClient().Get(util.Context(), util.ObjectKey(sc), sc) != nil {
					scMap[obc.Spec.StorageClassName] = sc
				}
			}
			bucketClass = sc.Parameters["bucketclass"]
		}
		table.AddRow(
			obc.Namespace,
			obc.Name,
			obc.Spec.BucketName,
			obc.Spec.StorageClassName,
			bucketClass,
			string(obc.Status.Phase),
		)
	}
	fmt.Print(table.String())
}

// WaitReady waits until the obc phase changes to bound by the operator
func WaitReady(obc *nbv1.ObjectBucketClaim) bool {
	log := util.Logger()
	klient := util.KubeClient()

	interval := time.Duration(3)

	err := wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(obc), obc)
		if err != nil {
			log.Printf("⏳ Failed to get OBC: %s", err)
			return false, nil
		}
		CheckPhase(obc)
		if obc.Status.Phase == obv1.ObjectBucketClaimStatusPhaseFailed {
			return false, fmt.Errorf("ObjectBucketClaimStatusPhaseFailed")
		}
		if obc.Status.Phase != obv1.ObjectBucketClaimStatusPhaseBound {
			return false, nil
		}
		return true, nil
	})
	return (err == nil)
}

// CheckPhase prints the phase and reason for it
func CheckPhase(obc *nbv1.ObjectBucketClaim) {
	log := util.Logger()

	switch obc.Status.Phase {

	case obv1.ObjectBucketClaimStatusPhaseBound:
		log.Printf("✅ OBC %q Phase is Bound", obc.Name)

	case obv1.ObjectBucketClaimStatusPhaseFailed:
		log.Errorf("❌ OBC %q Phase is %q", obc.Name, obc.Status.Phase)

	case obv1.ObjectBucketClaimStatusPhasePending:
		fallthrough
	default:
		log.Printf("⏳ OBC %q Phase is %q", obc.Name, obc.Status.Phase)
	}
}

// GenerateAccountKeys regenerate noobaa OBC account S3 keys
func GenerateAccountKeys(name, accountName, appNamespace string) error {
	log := util.Logger()

	if accountName == "" {
		log.Fatalf(`❌  account name cannot be empty.\n`)
	}

	var accessKeys nb.S3AccessKeys

	sysClient, err := system.Connect(true)
	if err != nil {
		return err
	}

	// Checking that we can find the secret before we are calling the RPC to change the credentials.
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = appNamespace
	secret.Name = name
	if !util.KubeCheckQuiet(secret) {
		log.Fatalf(`❌  Could not find secret: %s, will not regenerate keys.`, secret.Name)
	}

	err = sysClient.NBClient.GenerateAccountKeysAPI(nb.GenerateAccountKeysParams{
		Email: accountName,
	})
	if err != nil {
		return err
	}

	// GenerateAccountKeysAPI have no replay so we need to read the account in order to get the new credentials
	accountInfo, err := sysClient.NBClient.ReadAccountAPI(nb.ReadAccountParams{
		Email: accountName,
	})
	if err != nil {
		return err
	}

	accessKeys = accountInfo.AccessKeys[0]

	secret.StringData = map[string]string{}
	secret.StringData["AWS_ACCESS_KEY_ID"] = string(accessKeys.AccessKey)
	secret.StringData["AWS_SECRET_ACCESS_KEY"] = string(accessKeys.SecretKey)

	//If we will not be able to update the secret we will print the credentials as they already been changed by the RPC
	if !util.KubeUpdate(secret) {
		log.Printf(`❌  Please write the new credentials for OBC %s:`, name)
		fmt.Printf("\nAWS_ACCESS_KEY_ID     : %s\n", string(accessKeys.AccessKey))
		fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n\n", string(accessKeys.SecretKey))
		log.Fatalf(`❌  Failed to update the secret %s with the new accessKeys`, secret.Name)
	}

	log.Printf("✅ Successfully regenerate s3 credentials for the OBC %q", name)
	return nil
}

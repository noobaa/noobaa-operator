package obc

import (
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

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obc",
		Short: "Manage object bucket claims",
	}
	cmd.AddCommand(
		CmdCreate(),
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
	cmd.Flags().String("path", "",
		"Set path to specify inner directory in namespace store target path - can be used only while specifing a namespace bucketclass")
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
		"Set the namespace of the application where the OBC should be created")
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
	appNamespace, _ := cmd.Flags().GetString("app-namespace")
	if appNamespace == "" {
		appNamespace = options.Namespace
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
	} else if path != "" {
		log.Fatalf(`❌ Could not create OBC %q with inner path while missing namespace bucketclass`, obc.Name)
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
			fmt.Printf("  %-22s : %s\n", k, v)
			credsEnv += k + "=" + v + " "
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
			fmt.Printf("  %-22s : %s\n", "Data Space Avail", nb.BigIntToHumanBytes(b.DataCapacity.AvailableToUpload))
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

	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
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

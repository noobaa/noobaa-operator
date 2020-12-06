package system

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/version"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage noobaa systems",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdStatus(),
		CmdList(),
		CmdReconcile(),
		CmdYaml(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a noobaa system",
		Run:   RunCreate,
	}
	cmd.Flags().String("core-resources", "", "Core resources JSON")
	cmd.Flags().String("db-resources", "", "DB resources JSON")
	cmd.Flags().String("endpoint-resources", "", "Endpoint resources JSON")
	cmd.Flags().Bool("use-obc-cleanup-policy", false, "Create NooBaa system with obc cleanup policy")
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a noobaa system",
		Run:   RunDelete,
	}
	cmd.Flags().Bool("cleanup_data", false, "clean object buckets")
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List noobaa systems",
		Run:   RunList,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa system",
		Run:   RunStatus,
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

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled noobaa yaml",
		Run:   RunYaml,
	}
	return cmd
}

// LoadSystemDefaults loads a noobaa system CR from bundled yamls
// and apply's changes from CLI flags to the defaults.
func LoadSystemDefaults() *nbv1.NooBaa {
	sys := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	sys.Namespace = options.Namespace
	sys.Name = options.SystemName
	sys.Finalizers = []string{nbv1.GracefulFinalizer}
	dbType := options.DBType
	sys.Spec.DBType = nbv1.DBTypes(dbType)

	if options.NooBaaImage != "" {
		image := options.NooBaaImage
		sys.Spec.Image = &image
	}
	if options.ImagePullSecret != "" {
		sys.Spec.ImagePullSecret = &corev1.LocalObjectReference{Name: options.ImagePullSecret}
	}
	if options.DBImage != "" {
		dbImage := options.DBImage
		sys.Spec.DBImage = &dbImage
	}
	//  naively changing the db postgres image to the hardcoded one if the db type is postgres
	if options.DBType == "postgres" {
		dbPostgresImage := options.DBPostgresImage
		sys.Spec.DBImage = &dbPostgresImage
	}
	if options.DBVolumeSizeGB != 0 {
		sys.Spec.DBVolumeResources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: *resource.NewScaledQuantity(int64(options.DBVolumeSizeGB), resource.Giga),
			},
		}
	}
	if options.DBStorageClass != "" {
		sc := options.DBStorageClass
		sys.Spec.DBStorageClass = &sc
	}
	if options.PVPoolDefaultStorageClass != "" {
		sc := options.PVPoolDefaultStorageClass
		sys.Spec.PVPoolDefaultStorageClass = &sc
	}
	if options.MiniEnv {
		coreResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(1), resource.Giga),
		}
		dbResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(500), resource.Mega),
		}
		endpointResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(500), resource.Mega),
		}
		sys.Spec.CoreResources = &corev1.ResourceRequirements{
			Requests: coreResourceList,
			Limits:   coreResourceList,
		}
		sys.Spec.DBResources = &corev1.ResourceRequirements{
			Requests: dbResourceList,
			Limits:   dbResourceList,
		}
		sys.Spec.Endpoints = &nbv1.EndpointsSpec{
			MinCount: 1,
			MaxCount: 1,
			Resources: &corev1.ResourceRequirements{
				Requests: endpointResourceList,
				Limits:   endpointResourceList,
			}}
	}
	return sys
}

// RunOperatorCreate creates the default system when the operator starts.
func RunOperatorCreate(cmd *cobra.Command, args []string) {
	// The concern with this behavior was that it might be unexpected to the user that this happens.
	// For now we disable it but might reconsider later.
	//
	// sys := LoadSystemDefaults()
	// util.KubeCreateSkipExisting(sys)
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	sys := LoadSystemDefaults()
	ns := util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	ns.Name = sys.Namespace

	coreResourcesJSON, _ := cmd.Flags().GetString("core-resources")
	dbResourcesJSON, _ := cmd.Flags().GetString("db-resources")
	endpointResourcesJSON, _ := cmd.Flags().GetString("endpoint-resources")
	useOBCCleanupPolicy, _ := cmd.Flags().GetBool("use-obc-cleanup-policy")
	if useOBCCleanupPolicy {
		sys.Spec.CleanupPolicy.Confirmation = nbv1.DeleteOBCConfirmation
	}
	if coreResourcesJSON != "" {
		util.Panic(json.Unmarshal([]byte(coreResourcesJSON), &sys.Spec.CoreResources))
	}
	if dbResourcesJSON != "" {
		util.Panic(json.Unmarshal([]byte(dbResourcesJSON), &sys.Spec.DBResources))
	}
	if endpointResourcesJSON != "" {
		if sys.Spec.Endpoints == nil {
			sys.Spec.Endpoints = &nbv1.EndpointsSpec{
				MinCount: 1,
				MaxCount: 1,
			}
		}
		util.Panic(json.Unmarshal([]byte(endpointResourcesJSON), &sys.Spec.Endpoints.Resources))
	}

	// TODO check PVC if exist and the system does not exist -
	// fail and suggest to delete them first with cli system delete.
	util.KubeCreateSkipExisting(ns)
	util.KubeCreateSkipExisting(sys)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()

	sys := &nbv1.NooBaa{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		},
	}

	util.KubeCheck(sys)

	cleanupData, _ := cmd.Flags().GetBool("cleanup_data")
	objectBuckets := &nbv1.ObjectBucketList{}
	obcSelector, _ := labels.Parse("noobaa-domain=" + options.SubDomainNS())
	util.KubeList(objectBuckets, &client.ListOptions{LabelSelector: obcSelector})
	finalizersArray := sys.GetFinalizers()

	if util.Contains(nbv1.GracefulFinalizer, finalizersArray) {

		confirm := sys.Spec.CleanupPolicy.Confirmation
		if !cleanupData && len(objectBuckets.Items) != 0 && confirm != nbv1.DeleteOBCConfirmation {
			log.Fatalf(`❌ %s`, fmt.Sprintf("Failed to delete NooBaa. object buckets in namespace %q are not cleaned up.", options.Namespace))

		} else {
			log.Infof("Deleting All object buckets in namespace %q", options.Namespace)

			util.RemoveFinalizer(sys, nbv1.GracefulFinalizer)
			if !util.KubeUpdate(sys) {
				log.Errorf("NooBaa %q failed to remove finalizer %q", options.SystemName, nbv1.GracefulFinalizer)
			}
		}
	}

	if err := util.VerifyExternalSecretsDeletion(sys.Spec.Security.KeyManagementService, sys.Namespace); err != nil {
		log.Warnf("could not delete external secrets: %s", err)
	}

	util.KubeDelete(sys)

	// NoobaaDB
	noobaaDB := util.KubeObject(bundle.File_deploy_internal_statefulset_db_yaml).(*appsv1.StatefulSet)
	for i := range noobaaDB.Spec.VolumeClaimTemplates {
		t := &noobaaDB.Spec.VolumeClaimTemplates[i]
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.Name + "-" + options.SystemName + "-db-0",
				Namespace: options.Namespace,
			},
		}
		util.KubeDelete(pvc)
	}
	backingStores := &nbv1.BackingStoreList{}
	util.KubeList(backingStores, &client.ListOptions{Namespace: options.Namespace})
	for i := range backingStores.Items {
		bstore := &backingStores.Items[i]
		if bstore.Spec.Type == nbv1.StoreTypePVPool {
			pods := &corev1.PodList{}
			util.KubeList(pods, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name})
			for i := range pods.Items {
				pod := &pods.Items[i]
				util.RemoveFinalizer(pod, nbv1.Finalizer)
				if !util.KubeUpdate(pod) {
					log.Errorf("Pod %q failed to remove finalizer %q",
						pod.Name, nbv1.Finalizer)
				}
			}
			util.KubeDeleteAllOf(&corev1.Pod{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name},
				client.GracePeriodSeconds(0))
			util.KubeDeleteAllOf(&corev1.PersistentVolumeClaim{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name},
				client.GracePeriodSeconds(0))
		}
		util.RemoveFinalizer(bstore, nbv1.Finalizer)
		if !util.KubeUpdate(bstore) {
			log.Errorf("BackingStore %q failed to remove finalizer %q",
				bstore.Name, nbv1.Finalizer)
		}
		util.KubeDelete(bstore, client.GracePeriodSeconds(0))
	}
	objectBucketClaims := &nbv1.ObjectBucketClaimList{}
	configMaps := &corev1.ConfigMapList{}
	secrets := &corev1.SecretList{}

	util.KubeList(objectBucketClaims, &client.ListOptions{LabelSelector: obcSelector})
	util.KubeList(configMaps, &client.ListOptions{LabelSelector: obcSelector})
	util.KubeList(secrets, &client.ListOptions{LabelSelector: obcSelector})

	for i := range objectBucketClaims.Items {
		obj := &objectBucketClaims.Items[i]
		log.Warnf("ObjectBucketClaim %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucketClaim %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range objectBuckets.Items {
		obj := &objectBuckets.Items[i]
		log.Warnf("ObjectBucket %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range configMaps.Items {
		obj := &configMaps.Items[i]
		log.Warnf("ConfigMap %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range secrets.Items {
		obj := &secrets.Items[i]
		log.Warnf("Secret %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No systems found.\n")
		return
	}

	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"MGMT-ENDPOINTS",
		"S3-ENDPOINTS",
		"IMAGE",
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		s := &list.Items[i]
		table.AddRow(
			s.Namespace,
			s.Name,
			fmt.Sprint(s.Status.Services.ServiceMgmt.NodePorts),
			fmt.Sprint(s.Status.Services.ServiceS3.NodePorts),
			s.Status.ActualImage,
			string(s.Status.Phase),
			time.Since(s.CreationTimestamp.Time).Round(time.Second).String(),
		)
	}
	fmt.Print(table.String())
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	sys := LoadSystemDefaults()
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(sys, os.Stdout))
}

// CheckNooBaaImages runs a CLI command
func CheckNooBaaImages(cmd *cobra.Command, sys *nbv1.NooBaa, args []string) string {
	log := util.Logger()

	desiredImage := ""
	runningImage := ""
	if sys.Status.ActualImage != "" {
		desiredImage = sys.Status.ActualImage
	} else if sys.Spec.Image != nil {
		desiredImage = *sys.Spec.DBImage
	}
	sts := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	sts.Namespace = options.Namespace
	if util.KubeCheckQuiet(sts) {
		runningImage = sts.Spec.Template.Spec.Containers[0].Image
	}
	if desiredImage != "" && runningImage != "" && desiredImage != runningImage {
		log.Warnf("⚠️  The desired noobaa image and the running image are not the same")
	}
	return runningImage
}

// CheckNooBaaDBImages runs a CLI command
func CheckNooBaaDBImages(cmd *cobra.Command, sys *nbv1.NooBaa, args []string) string {
	log := util.Logger()

	desiredImage := ""
	runningImage := ""
	if sys.Spec.DBImage != nil {
		desiredImage = *sys.Spec.DBImage
	}
	sts := util.KubeObject(bundle.File_deploy_internal_statefulset_db_yaml).(*appsv1.StatefulSet)
	if (sys.Spec.DBType == "postgres") {
		sts.Name = "noobaa-db-pg";
	}
	sts.Namespace = options.Namespace
	if util.KubeCheckQuiet(sts) {
		runningImage = sts.Spec.Template.Spec.Containers[0].Image
	}
	if desiredImage != "" && runningImage != "" && desiredImage != runningImage {
		log.Warnf("⚠️  The desired db image and the running db image are not the same")
	}
	return runningImage
}

// CheckOperatorImage runs a CLI command
func CheckOperatorImage(cmd *cobra.Command, args []string) string {
	runningImage := ""
	deployment := util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)
	deployment.Namespace = options.Namespace
	if util.KubeCheckQuiet(deployment) {
		runningImage = deployment.Spec.Template.Spec.Containers[0].Image
	}
	return runningImage
}

// RunSystemVersionsStatus runs a CLI command
func RunSystemVersionsStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace
	isSystemExists := util.KubeCheckQuiet(sys)

	noobaaImage := ""
	noobaaDbImage := ""
	noobaaOperatorImage := ""

	if isSystemExists {
		noobaaImage = CheckNooBaaImages(cmd, sys, args)
		noobaaDbImage = CheckNooBaaDBImages(cmd, sys, args)
		noobaaOperatorImage = CheckOperatorImage(cmd, args)
	} else {
		noobaaImage = options.NooBaaImage
		noobaaDbImage = options.DBImage
		noobaaOperatorImage = options.OperatorImage
	}

	log.Printf("CLI version: %s\n", version.Version)
	log.Printf("noobaa-image: %s\n", noobaaImage)
	log.Printf("operator-image: %s\n", noobaaOperatorImage)
	log.Printf("noobaa-db-image: %s\n", noobaaDbImage)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysKey, klient, scheme.Scheme, nil)
	r.CheckAll()
	var NooBaaDB *appsv1.StatefulSet = nil
	if r.NooBaa.Spec.DBType == "postgres" {
		NooBaaDB = r.NooBaaPostgresDB
	} else {
		NooBaaDB = r.NooBaaMongoDB
	}
	// NobbaaDB
	for i := range NooBaaDB.Spec.VolumeClaimTemplates {
		t := &NooBaaDB.Spec.VolumeClaimTemplates[i]
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.Name + "-" + options.SystemName + "-db-0",
				Namespace: options.Namespace,
			},
		}
		util.KubeCheck(pvc)
	}

	// sys := cli.LoadSystemDefaults()
	// util.KubeCheck(cli.Client, sys)

	if r.NooBaa.Status.Phase != nbv1.SystemPhaseReady {
		log.Printf("❌ System Phase is %q\n", r.NooBaa.Status.Phase)
		util.IgnoreError(CheckWaitingFor(r.NooBaa))
		return
	}

	log.Printf("✅ System Phase is %q\n", r.NooBaa.Status.Phase)
	secret := r.SecretAdmin.DeepCopy()
	util.KubeCheck(secret)

	fmt.Println("")
	fmt.Println("#------------------#")
	fmt.Println("#- Mgmt Addresses -#")
	fmt.Println("#------------------#")
	fmt.Println("")

	fmt.Println("ExternalDNS :", r.NooBaa.Status.Services.ServiceMgmt.ExternalDNS)
	fmt.Println("ExternalIP  :", r.NooBaa.Status.Services.ServiceMgmt.ExternalIP)
	fmt.Println("NodePorts   :", r.NooBaa.Status.Services.ServiceMgmt.NodePorts)
	fmt.Println("InternalDNS :", r.NooBaa.Status.Services.ServiceMgmt.InternalDNS)
	fmt.Println("InternalIP  :", r.NooBaa.Status.Services.ServiceMgmt.InternalIP)
	fmt.Println("PodPorts    :", r.NooBaa.Status.Services.ServiceMgmt.PodPorts)

	fmt.Println("")
	fmt.Println("#--------------------#")
	fmt.Println("#- Mgmt Credentials -#")
	fmt.Println("#--------------------#")
	fmt.Println("")
	fmt.Printf("email    : %s\n", secret.StringData["email"])
	fmt.Printf("password : %s\n", secret.StringData["password"])

	fmt.Println("")
	fmt.Println("#----------------#")
	fmt.Println("#- S3 Addresses -#")
	fmt.Println("#----------------#")
	fmt.Println("")

	fmt.Println("ExternalDNS :", r.NooBaa.Status.Services.ServiceS3.ExternalDNS)
	fmt.Println("ExternalIP  :", r.NooBaa.Status.Services.ServiceS3.ExternalIP)
	fmt.Println("NodePorts   :", r.NooBaa.Status.Services.ServiceS3.NodePorts)
	fmt.Println("InternalDNS :", r.NooBaa.Status.Services.ServiceS3.InternalDNS)
	fmt.Println("InternalIP  :", r.NooBaa.Status.Services.ServiceS3.InternalIP)
	fmt.Println("PodPorts    :", r.NooBaa.Status.Services.ServiceS3.PodPorts)

	fmt.Println("")
	fmt.Println("#------------------#")
	fmt.Println("#- S3 Credentials -#")
	fmt.Println("#------------------#")
	fmt.Println("")
	fmt.Printf("AWS_ACCESS_KEY_ID     : %s\n", secret.StringData["AWS_ACCESS_KEY_ID"])
	fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n", secret.StringData["AWS_SECRET_ACCESS_KEY"])
	fmt.Println("")
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

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady() bool {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		sys := &nbv1.NooBaa{}
		err := klient.Get(util.Context(), sysKey, sys)
		if err != nil {
			log.Printf("⏳ Failed to get system: %s", err)
			return false, nil
		}
		if sys.Status.Phase == nbv1.SystemPhaseRejected {
			log.Errorf("❌ System Phase is %q. describe noobaa for more information", sys.Status.Phase)
			return false, fmt.Errorf("SystemPhaseRejected")
		}
		if sys.Status.Phase != nbv1.SystemPhaseReady {
			return false, CheckWaitingFor(sys)
		}
		log.Printf("✅ System Phase is %q.\n", sys.Status.Phase)
		return true, nil
	})
	return (err == nil)
}

// CheckWaitingFor checks what the system deployment is waiting for in order to become ready
// in order to help the user troubleshoot common deployment issues.
func CheckWaitingFor(sys *nbv1.NooBaa) error {
	log := util.Logger()
	klient := util.KubeClient()

	operApp := &appsv1.Deployment{}
	operAppName := "noobaa-operator"
	operAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: operAppName},
		operApp)

	if errors.IsNotFound(operAppErr) {
		log.Printf(`❌ Deployment %q is missing.`, operAppName)
		return operAppErr
	}
	if operAppErr != nil {
		log.Printf(`❌ Deployment %q unknown error in Get(): %s`, operAppName, operAppErr)
		return operAppErr
	}
	desiredReplicas := int32(1)
	if operApp.Spec.Replicas != nil {
		desiredReplicas = *operApp.Spec.Replicas
	}
	if operApp.Status.Replicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. Deployment %q is not ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			operAppName,
			operApp.Status.ReadyReplicas,
			desiredReplicas)
		return nil
	}

	operPodList := &corev1.PodList{}
	operPodSelector, _ := labels.Parse("noobaa-operator=deployment")
	operPodErr := klient.List(util.Context(),
		operPodList,
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: operPodSelector})

	if operPodErr != nil {
		return operPodErr
	}
	if len(operPodList.Items) != int(desiredReplicas) {
		return fmt.Errorf("Can't find the operator pods")
	}
	operPod := &operPodList.Items[0]
	if operPod.Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not yet ready: %s`,
			sys.Status.Phase, operPod.Name, util.GetPodStatusLine(operPod))
		return nil
	}
	for i := range operPod.Status.ContainerStatuses {
		c := &operPod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q is not yet ready: %s`,
				sys.Status.Phase, c.Name, util.GetContainerStatusLine(c))
			return nil
		}
	}

	coreApp := &appsv1.StatefulSet{}
	coreAppName := sys.Name + "-core"
	coreAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: coreAppName},
		coreApp)

	if errors.IsNotFound(coreAppErr) {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not found yet`,
			sys.Status.Phase, coreAppName)
		return nil
	}
	if coreAppErr != nil {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not found yet (error): %s`,
			sys.Status.Phase, coreAppName, coreAppErr)
		return nil
	}
	desiredReplicas = int32(1)
	if coreApp.Spec.Replicas != nil {
		desiredReplicas = *coreApp.Spec.Replicas
	}
	if coreApp.Status.Replicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not yet ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			coreAppName,
			coreApp.Status.ReadyReplicas,
			desiredReplicas)
		return nil
	}

	corePodList := &corev1.PodList{}
	corePodSelector, _ := labels.Parse("noobaa-core=" + sys.Name)
	corePodErr := klient.List(util.Context(),
		corePodList,
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: corePodSelector})

	if corePodErr != nil {
		return corePodErr
	}
	if len(corePodList.Items) != int(desiredReplicas) {
		return fmt.Errorf("Can't find the core pods")
	}
	corePod := &corePodList.Items[0]
	if corePod.Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not yet ready: %s`,
			sys.Status.Phase, corePod.Name, util.GetPodStatusLine(corePod))
		return nil
	}
	for i := range corePod.Status.ContainerStatuses {
		c := &corePod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q is not yet ready: %s`,
				sys.Status.Phase, c.Name, util.GetContainerStatusLine(c))
			return nil
		}
	}

	log.Printf(`⏳ System Phase is %q. Waiting for phase ready ...`, sys.Status.Phase)
	return nil
}

// Client is the system client for making mgmt or s3 calls (with operator/admin credentials)
type Client struct {
	NooBaa      *nbv1.NooBaa
	ServiceMgmt *corev1.Service
	SecretOp    *corev1.Secret
	SecretAdmin *corev1.Secret
	NBClient    nb.Client
	MgmtURL     *url.URL
	S3URL       *url.URL
}

// GetNBClient returns an api client
func GetNBClient() nb.Client {
	c, err := Connect(true)
	if err != nil {
		util.Logger().Fatalf("❌ %s", err)
	}
	return c.NBClient
}

// Connect loads the mgmt and S3 api details from the system.
// It gets the endpoint address and token from the system status and secret that the
// operator creates for the system.
// When isExternal is true we return  : s3 => external DNS, mgmt => port-forwarding (router)
// When isExternal is false we return : s3 => internal DNS, mgmt => node-port
func Connect(isExternal bool) (*Client, error) {

	klient := util.KubeClient()
	sysObjKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysObjKey, klient, scheme.Scheme, nil)

	if !CheckSystem(r.NooBaa) {
		return nil, fmt.Errorf("Connect(): System not found")
	}
	if !util.KubeCheck(r.ServiceMgmt) {
		return nil, fmt.Errorf("Connect(): ServiceMgmt not found")
	}
	if !util.KubeCheck(r.SecretOp) {
		return nil, fmt.Errorf("Connect(): SecretOp not found")
	}
	if !util.KubeCheck(r.SecretAdmin) {
		return nil, fmt.Errorf("Connect(): SecretAdmin not found")
	}

	authToken := r.SecretOp.StringData["auth_token"]
	mgmtStatus := &r.NooBaa.Status.Services.ServiceMgmt
	s3Status := &r.NooBaa.Status.Services.ServiceS3

	if authToken == "" {
		return nil, fmt.Errorf("Connect(): Operator secret with auth token is not ready")
	}
	if len(mgmtStatus.NodePorts) == 0 {
		return nil, fmt.Errorf("Connect(): System mgmt service (NodePorts) is not ready")
	}
	if len(s3Status.InternalDNS) == 0 {
		return nil, fmt.Errorf("Connect(): System s3 service (InternalDNS) is not ready")
	}

	mgmtEndpoint := mgmtStatus.NodePorts[0]
	s3Endpoint := s3Status.InternalDNS[0]
	var nbClient nb.Client

	if isExternal {

		// setup port forwarding
		router := &nb.APIRouterPortForward{
			ServiceMgmt:  r.ServiceMgmt,
			PodNamespace: r.NooBaa.Namespace,
			PodName:      r.NooBaa.Name + "-core-0",
		}
		err := router.Start()
		if err != nil {
			return nil, err
		}
		nbClient = nb.NewClient(router)
		mgmtEndpoint = strings.TrimSuffix(router.GetAddress(""), "/rpc/")

		// set s3 to external if possible, otherwise fallback to node-port which is like external for minikube
		if len(s3Status.ExternalIP) > 0 {
			s3Endpoint = s3Status.ExternalIP[0]
		} else if len(s3Status.NodePorts) > 0 {
			s3Endpoint = s3Status.NodePorts[0]
		}

	} else {
		nbClient = nb.NewClient(&nb.APIRouterServicePort{
			ServiceMgmt: r.ServiceMgmt,
		})
	}

	nbClient.SetAuthToken(authToken)

	mgmtURL, err := url.Parse(mgmtEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Connect(): Failed to parse mgmt url %q. got error: %v", mgmtEndpoint, err)
	}
	s3URL, err := url.Parse(s3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Connect(): Failed to parse s3 url %q. got error: %v", s3Endpoint, err)
	}

	return &Client{
		NooBaa:      r.NooBaa,
		ServiceMgmt: r.ServiceMgmt,
		SecretOp:    r.SecretOp,
		SecretAdmin: r.SecretAdmin,
		NBClient:    nbClient,
		MgmtURL:     mgmtURL,
		S3URL:       s3URL,
	}, nil
}

// CheckSystem checks the state of the system and initializes its status fields
func CheckSystem(sys *nbv1.NooBaa) bool {
	found := util.KubeCheck(sys)
	if sys.Status.Accounts == nil {
		sys.Status.Accounts = &nbv1.AccountsStatus{}
	}
	if sys.Status.Services == nil {
		sys.Status.Services = &nbv1.ServicesStatus{}
	}
	return found && sys.UID != ""
}

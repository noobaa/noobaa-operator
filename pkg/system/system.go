package system

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a noobaa system",
		Run:   RunDelete,
	}
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
		Use:   "reconcile",
		Short: "Runs a reconcile attempt like noobaa-operator",
		Run:   RunReconcile,
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
	sys := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	sys.Namespace = options.Namespace
	sys.Name = options.SystemName
	if options.NooBaaImage != "" {
		image := options.NooBaaImage
		sys.Spec.Image = &image
	}
	if options.DBImage != "" {
		dbImage := options.DBImage
		sys.Spec.DBImage = &dbImage
	}
	if options.ImagePullSecret != "" {
		sys.Spec.ImagePullSecret = &corev1.LocalObjectReference{Name: options.ImagePullSecret}
	}
	if options.StorageClassName != "" {
		sc := options.StorageClassName
		sys.Spec.StorageClassName = &sc
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

	util.KubeDelete(sys)

	// TEMPORARY ? delete PVCs here because we couldn't own them in openshift
	// See https://github.com/noobaa/noobaa-operator/issues/12
	// So we delete the PVC here on system delete.
	coreApp := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	for i := range coreApp.Spec.VolumeClaimTemplates {
		t := &coreApp.Spec.VolumeClaimTemplates[i]
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.Name + "-" + options.SystemName + "-core-0",
				Namespace: options.Namespace,
			},
		}
		util.KubeDelete(pvc)
	}
	// NoobaaDB
	noobaaDB := util.KubeObject(bundle.File_deploy_internal_statefulset_mongo_yaml).(*appsv1.StatefulSet)
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
		obj := &backingStores.Items[i]
		util.RemoveFinalizer(obj, nbv1.Finalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("BackingStore %q failed to remove finalizer %q",
				obj.Name, nbv1.Finalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}

	objectBucketClaims := &nbv1.ObjectBucketClaimList{}
	objectBuckets := &nbv1.ObjectBucketList{}
	configMaps := &corev1.ConfigMapList{}
	secrets := &corev1.SecretList{}
	obcSelector, _ := labels.Parse("noobaa-domain=" + options.SubDomainNS())

	util.KubeList(objectBucketClaims, &client.ListOptions{LabelSelector: obcSelector})
	util.KubeList(objectBuckets, &client.ListOptions{LabelSelector: obcSelector})
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
	if !util.KubeList(list, nil) {
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

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysKey, klient, scheme.Scheme, nil)
	r.CheckAll()

	// TEMPORARY ? check PVCs here because we couldn't own them in openshift
	// See https://github.com/noobaa/noobaa-operator/issues/12
	for i := range r.CoreApp.Spec.VolumeClaimTemplates {
		t := &r.CoreApp.Spec.VolumeClaimTemplates[i]
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.Name + "-" + options.SystemName + "-core-0",
				Namespace: options.Namespace,
			},
		}
		util.KubeCheck(pvc)
	}

	// NobbaaDB
	for i := range r.NoobaaDB.Spec.VolumeClaimTemplates {
		t := &r.NoobaaDB.Spec.VolumeClaimTemplates[i]
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
	secretRef := r.NooBaa.Status.Accounts.Admin.SecretRef
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretRef.Name,
			Namespace: secretRef.Namespace,
		},
	}
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
	for key, value := range secret.StringData {
		if !strings.HasPrefix(key, "AWS") {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

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
	for key, value := range secret.StringData {
		if strings.HasPrefix(key, "AWS") {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
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
	if err != nil {
		return false
	}
	return true
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
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: operPodSelector},
		operPodList)

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

	// NoobaaDB
	noobaaDB := &appsv1.StatefulSet{}
	noobaaDBName := sys.Name + "-db"
	noobaaDBErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: noobaaDBName},
		noobaaDB)

	if errors.IsNotFound(noobaaDBErr) {
		log.Printf(`❌ StatefulSet %q is missing.`, noobaaDBName)
		return noobaaDBErr
	}
	if noobaaDBErr != nil {
		log.Printf(`❌ StatefulSet %q unknown error in Get(): %s`, noobaaDBName, noobaaDBErr)
		return noobaaDBErr
	}
	desiredReplicas = int32(1)
	if noobaaDB.Spec.Replicas != nil {
		desiredReplicas = *noobaaDB.Spec.Replicas
	}
	if noobaaDB.Status.Replicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not yet ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			noobaaDBName,
			noobaaDB.Status.ReadyReplicas,
			desiredReplicas)
		return nil
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

	//
	corePodList := &corev1.PodList{}
	corePodSelector, _ := labels.Parse("noobaa-core=" + sys.Name)
	corePodErr := klient.List(util.Context(),
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: corePodSelector},
		corePodList)

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
	NooBaa   *nbv1.NooBaa
	NBClient nb.Client
	S3Client *s3.S3
}

// GetNBClient returns an api client
func GetNBClient() nb.Client {
	c, err := Connect()
	if err != nil {
		util.Logger().Fatalf("❌ %s", err)
	}
	return c.NBClient
}

// Connect loads the mgmt and S3 api details from the system.
// It gets the endpoint address and token from the system status and secret that the
// operator creates for the system.
func Connect() (*Client, error) {

	klient := util.KubeClient()
	sysObjKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysObjKey, klient, scheme.Scheme, nil)
	util.KubeCheck(r.NooBaa)
	util.KubeCheck(r.ServiceMgmt)
	util.KubeCheck(r.SecretOp)

	authToken := r.SecretOp.StringData["auth_token"]
	accessKey := r.SecretOp.StringData["AWS_ACCESS_KEY_ID"]
	secretKey := r.SecretOp.StringData["AWS_SECRET_ACCESS_KEY"]
	mgmtStatus := &r.NooBaa.Status.Services.ServiceMgmt
	s3Status := &r.NooBaa.Status.Services.ServiceS3

	if authToken == "" {
		return nil, fmt.Errorf("Operator secret with auth token is not ready")
	}
	if len(mgmtStatus.NodePorts) == 0 {
		return nil, fmt.Errorf("System mgmt service (nodeport) is not ready")
	}
	if len(s3Status.NodePorts) == 0 {
		return nil, fmt.Errorf("System s3 service (nodeport) is not ready")
	}

	mgmtNodePort := mgmtStatus.NodePorts[0]
	mgmtURL, err := url.Parse(mgmtNodePort)
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 endpoint %q. got error: %v", mgmtNodePort, err)
	}

	s3NodePort := s3Status.NodePorts[0]
	// s3URL, err := url.Parse(s3NodePort)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse s3 endpoint %q. got error: %v", s3NodePort, err)
	// }

	nbClient := nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: r.ServiceMgmt,
		NodeIP:      mgmtURL.Hostname(),
	})
	nbClient.SetAuthToken(authToken)

	s3Config := &aws.Config{
		Endpoint: &s3NodePort,
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
	}

	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	return &Client{
		NooBaa:   r.NooBaa,
		NBClient: nbClient,
		S3Client: s3.New(s3Session),
	}, nil
}

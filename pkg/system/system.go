package system

import (
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Cmd creates a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage noobaa systems (create delete etc.)",
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

// CmdCreate creates a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a noobaa system",
		Run:   RunCreate,
	}
	return cmd
}

// CmdDelete creates a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a noobaa system",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList creates a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List noobaa systems",
		Run:   RunList,
	}
	return cmd
}

// CmdStatus creates a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa system",
		Run:   RunStatus,
	}
	return cmd
}

// CmdReconcile creates a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile-local",
		Short: "Runs a reconcile attempt like noobaa-operator",
		Run:   RunReconcile,
	}
	return cmd
}

// CmdYaml creates a CLI command
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
func LoadSystemDefaults(cmd *cobra.Command) *nbv1.NooBaa {
	sys := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	sys.Namespace = options.Namespace
	sys.Name = options.SystemName
	if options.NooBaaImage != "" {
		image := options.NooBaaImage
		sys.Spec.Image = &image
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

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	sys := LoadSystemDefaults(cmd)
	ns := util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	ns.Name = sys.Namespace
	// TODO check PVC if exist and the system does not exist -
	// fail and suggest to delete them first with cli system delete.
	util.KubeCreateSkipExisting(ns)
	util.KubeCreateSkipExisting(sys)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
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
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	log := util.Logger()
	list := &nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
	}
	if !util.KubeList(list, nil) {
		return
	}
	if len(list.Items) == 0 {
		log.Printf("No systems found.\n")
		return
	}

	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"PHASE",
		"MGMT-ENDPOINTS",
		"S3-ENDPOINTS",
		"IMAGE",
		"AGE",
	)
	for i := range list.Items {
		s := &list.Items[i]
		table.AddRow(
			s.Namespace,
			s.Name,
			string(s.Status.Phase),
			fmt.Sprint(s.Status.Services.ServiceMgmt.NodePorts),
			fmt.Sprint(s.Status.Services.ServiceS3.NodePorts),
			s.Status.ActualImage,
			since(s.ObjectMeta.CreationTimestamp.Time),
		)
	}
	fmt.Print(table.String())
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	sys := LoadSystemDefaults(cmd)
	p := printers.YAMLPrinter{}
	p.PrintObj(sys, os.Stdout)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	s := NewReconciler(sysKey, klient, scheme.Scheme, nil)
	s.Load()

	// TEMPORARY ? check PVCs here because we couldn't own them in openshift
	// See https://github.com/noobaa/noobaa-operator/issues/12
	for i := range s.CoreApp.Spec.VolumeClaimTemplates {
		t := &s.CoreApp.Spec.VolumeClaimTemplates[i]
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.Name + "-" + options.SystemName + "-core-0",
				Namespace: options.Namespace,
			},
		}
		util.KubeCheck(pvc)
	}

	// sys := cli.LoadSystemDefaults()
	// util.KubeCheck(cli.Client, sys)
	if s.NooBaa.Status.Phase == nbv1.SystemPhaseReady {
		log.Printf("✅ System Phase is %q\n", s.NooBaa.Status.Phase)
		secretRef := s.NooBaa.Status.Accounts.Admin.SecretRef
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretRef.Name,
				Namespace: secretRef.Namespace,
			},
		}
		util.KubeCheck(secret)

		log.Println("")
		log.Println("#------------------#")
		log.Println("#- Mgmt Addresses -#")
		log.Println("#------------------#")
		log.Println("")

		log.Println("ExternalDNS :", s.NooBaa.Status.Services.ServiceMgmt.ExternalDNS)
		log.Println("ExternalIP  :", s.NooBaa.Status.Services.ServiceMgmt.ExternalIP)
		log.Println("NodePorts   :", s.NooBaa.Status.Services.ServiceMgmt.NodePorts)
		log.Println("InternalDNS :", s.NooBaa.Status.Services.ServiceMgmt.InternalDNS)
		log.Println("InternalIP  :", s.NooBaa.Status.Services.ServiceMgmt.InternalIP)
		log.Println("PodPorts    :", s.NooBaa.Status.Services.ServiceMgmt.PodPorts)

		log.Println("")
		log.Println("#--------------------#")
		log.Println("#- Mgmt Credentials -#")
		log.Println("#--------------------#")
		log.Println("")
		for key, value := range secret.Data {
			if !strings.HasPrefix(key, "AWS") {
				log.Printf("%s: %s\n", key, string(value))
			}
		}

		log.Println("")
		log.Println("#----------------#")
		log.Println("#- S3 Addresses -#")
		log.Println("#----------------#")
		log.Println("")

		log.Println("ExternalDNS :", s.NooBaa.Status.Services.ServiceS3.ExternalDNS)
		log.Println("ExternalIP  :", s.NooBaa.Status.Services.ServiceS3.ExternalIP)
		log.Println("NodePorts   :", s.NooBaa.Status.Services.ServiceS3.NodePorts)
		log.Println("InternalDNS :", s.NooBaa.Status.Services.ServiceS3.InternalDNS)
		log.Println("InternalIP  :", s.NooBaa.Status.Services.ServiceS3.InternalIP)
		log.Println("PodPorts    :", s.NooBaa.Status.Services.ServiceS3.PodPorts)

		log.Println("")
		log.Println("#------------------#")
		log.Println("#- S3 Credentials -#")
		log.Println("#------------------#")
		log.Println("")
		for key, value := range secret.Data {
			if strings.HasPrefix(key, "AWS") {
				log.Printf("%s: %s\n", key, string(value))
			}
		}
		log.Println("")
	} else {
		log.Printf("❌ System Phase is %q\n", s.NooBaa.Status.Phase)
	}
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
			log.Errorf("❌ Failed to get system: %s", err)
			return false, err
		}
		if sys.Status.Phase == nbv1.SystemPhaseRejected {
			log.Errorf("❌ System Phase is %q. describe noobaa for more information", sys.Status.Phase)
			return false, fmt.Errorf("SystemPhaseRejected")
		}
		if sys.Status.Phase != nbv1.SystemPhaseReady {
			return CheckWaitingFor(sys)
		}
		log.Printf("✅ System Phase is %q.\n", sys.Status.Phase)
		return true, nil
	})
	if err != nil {
		return false
	}
	return true
}

func CheckWaitingFor(sys *nbv1.NooBaa) (bool, error) {
	log := util.Logger()
	klient := util.KubeClient()

	operApp := &appsv1.Deployment{}
	operAppName := "noobaa-operator"
	operAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: operAppName},
		operApp)

	if errors.IsNotFound(operAppErr) {
		log.Printf(`❌ Deployment %q is missing.`, operAppName)
		return false, operAppErr
	}
	if operAppErr != nil {
		log.Printf(`❌ Deployment %q unknown error in Get(): %s`, operAppName, operAppErr)
		return false, operAppErr
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
		return false, nil
	}

	operPodList := &corev1.PodList{}
	operPodSelector, _ := labels.Parse("noobaa-operator=deployment")
	operPodErr := klient.List(util.Context(),
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: operPodSelector},
		operPodList)

	if operPodErr != nil {
		return false, operPodErr
	}
	if len(operPodList.Items) != int(desiredReplicas) {
		return false, fmt.Errorf("Can't find the operator pods")
	}
	operPod := &operPodList.Items[0]
	if operPodList.Items[0].Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not running:`+
			` Phase=%q Reason=%q Message=%q`,
			sys.Status.Phase,
			operPod.Name,
			operPod.Status.Phase,
			operPod.Status.Reason,
			operPod.Status.Message)
		return false, nil
	}
	for i := range operPod.Status.ContainerStatuses {
		c := &operPod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q is not ready:`+
				` RestartCount=%d`,
				sys.Status.Phase,
				c.Name,
				c.RestartCount)
			return false, nil
		}
	}

	coreApp := &appsv1.StatefulSet{}
	coreAppName := sys.Name + "-core"
	coreAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: coreAppName},
		coreApp)

	if errors.IsNotFound(coreAppErr) {
		log.Printf(`❌ StatefulSet %q is missing.`, coreAppName)
		return false, coreAppErr
	}
	if coreAppErr != nil {
		log.Printf(`❌ StatefulSet %q unknown error in Get(): %s`, coreAppName, coreAppErr)
		return false, coreAppErr
	}
	desiredReplicas = int32(1)
	if coreApp.Spec.Replicas != nil {
		desiredReplicas = *coreApp.Spec.Replicas
	}
	if coreApp.Status.Replicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			coreAppName,
			coreApp.Status.ReadyReplicas,
			desiredReplicas)
		return false, nil
	}

	corePodList := &corev1.PodList{}
	corePodSelector, _ := labels.Parse("noobaa-core=" + sys.Name)
	corePodErr := klient.List(util.Context(),
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: corePodSelector},
		corePodList)

	if corePodErr != nil {
		return false, corePodErr
	}
	if len(corePodList.Items) != int(desiredReplicas) {
		return false, fmt.Errorf("Can't find the core pods")
	}
	corePod := &corePodList.Items[0]
	if corePod.Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not running:`+
			` Phase=%q Reason=%q Message=%q`,
			sys.Status.Phase,
			corePod.Name,
			corePod.Status.Phase,
			corePod.Status.Reason,
			corePod.Status.Message)
		return false, nil
	}
	for i := range corePod.Status.ContainerStatuses {
		c := &corePod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q is not ready:`+
				` RestartCount=%d`,
				sys.Status.Phase,
				c.Name,
				c.RestartCount)
			return false, nil
		}
	}

	log.Printf(`⏳ System Phase is %q. Waiting for phase ready ...`, sys.Status.Phase)
	return false, nil
}

// GetNBClient is a CLI common tool that loads the mgmt api details from the system.
// It gets the endpoint address and token from the system status and secret that the
// operator creates for the system.
func GetNBClient() nb.Client {
	log := util.Logger()
	klient := util.KubeClient()
	sysObjKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	s := NewReconciler(sysObjKey, klient, scheme.Scheme, nil)
	s.Load()

	mgmtStatus := s.NooBaa.Status.Services.ServiceMgmt
	if len(mgmtStatus.NodePorts) == 0 {
		log.Fatalf("❌ System mgmt service (nodeport) is not ready")
	}
	if s.SecretOp.StringData["auth_token"] == "" {
		log.Fatalf("❌ Operator secret with auth token is not ready")
	}

	nodePort := mgmtStatus.NodePorts[0]
	nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]
	nbClient := nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: s.ServiceMgmt,
		NodeIP:      nodeIP,
	})
	nbClient.SetAuthToken(s.SecretOp.StringData["auth_token"])
	return nbClient
}

func since(t time.Time) string {
	return time.Since(t).Round(time.Second).String()
}

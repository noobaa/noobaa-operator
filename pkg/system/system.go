package system

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
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
	klient := util.KubeClient()

	list := nbv1.NooBaaList{}
	err := klient.List(util.Context(), nil, &list)
	if meta.IsNoMatchError(err) {
		log.Warningf("CRD not installed.\n")
		return
	}
	util.Panic(err)
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
		log.Printf("✅ System Phase is \"%s\"\n", s.NooBaa.Status.Phase)
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
		log.Printf("❌ System Phase is \"%s\"\n", s.NooBaa.Status.Phase)
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
func WaitReady() {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	intervalSec := time.Duration(3)

	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		sys := &nbv1.NooBaa{}
		err := klient.Get(util.Context(), sysKey, sys)
		if err != nil {
			log.Errorf("❌ Failed to get system: %s", err)
			return false, err
		}
		if sys.Status.Phase == nbv1.SystemPhaseRejected {
			log.Errorf("❌ System Phase is \"%s\". describe noobaa for more information", sys.Status.Phase)
			return false, fmt.Errorf("SystemPhaseRejected")
		}
		if sys.Status.Phase != nbv1.SystemPhaseReady {
			log.Printf("⏳ System Phase is \"%s\". Waiting for phase ready ...\n", sys.Status.Phase)
			return false, nil
		}
		log.Printf("✅ System Phase is \"%s\".\n", sys.Status.Phase)
		return true, nil
	}))
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

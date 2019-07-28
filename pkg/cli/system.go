package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LoadSystemDefaults loads a noobaa system CR from bundled yamls
// and apply's changes from CLI flags to the defaults.
func (cli *CLI) LoadSystemDefaults() *nbv1.NooBaa {
	sys := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	sys.Namespace = cli.Namespace
	sys.Name = cli.SystemName
	if cli.NooBaaImage != "" {
		image := cli.NooBaaImage
		sys.Spec.Image = &image
	}
	if cli.ImagePullSecret != "" {
		sys.Spec.ImagePullSecret = &corev1.LocalObjectReference{Name: cli.ImagePullSecret}
	}
	if cli.StorageClassName != "" {
		sc := cli.StorageClassName
		sys.Spec.StorageClassName = &sc
	}
	return sys
}

// SystemCreate runs a CLI command
func (cli *CLI) SystemCreate() {
	sys := cli.LoadSystemDefaults()
	util.KubeCreateSkipExisting(cli.Client, sys)
}

// SystemDelete runs a CLI command
func (cli *CLI) SystemDelete() {
	sys := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cli.SystemName,
			Namespace: cli.Namespace,
		},
	}
	util.KubeDelete(cli.Client, sys)
}

// SystemList runs a CLI command
func (cli *CLI) SystemList() {
	list := nbv1.NooBaaList{}
	err := cli.Client.List(cli.Ctx, nil, &list)
	_, noKind := err.(*meta.NoKindMatchError)
	if noKind {
		cli.Log.Warningf("CRD not installed.\n")
		return
	}
	util.Panic(err)
	if len(list.Items) == 0 {
		cli.Log.Printf("No systems found.\n")
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

// SystemYaml runs a CLI command
func (cli *CLI) SystemYaml() {
	sys := cli.LoadSystemDefaults()
	p := printers.YAMLPrinter{}
	p.PrintObj(sys, os.Stdout)
}

// SystemStatus runs a CLI command
func (cli *CLI) SystemStatus() {
	s := system.New(types.NamespacedName{Namespace: cli.Namespace, Name: cli.SystemName}, cli.Client, scheme.Scheme, nil)
	s.Load()

	// sys := cli.LoadSystemDefaults()
	// util.KubeCheck(cli.Client, sys)
	if s.NooBaa.Status.Phase == nbv1.SystemPhaseReady {
		cli.Log.Printf("✅ System Phase is \"%s\"\n", s.NooBaa.Status.Phase)
		secretRef := s.NooBaa.Status.Accounts.Admin.SecretRef
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretRef.Name,
				Namespace: secretRef.Namespace,
			},
		}
		util.KubeCheck(cli.Client, secret)

		cli.Log.Println("")
		cli.Log.Println("#------------------#")
		cli.Log.Println("#- Mgmt Addresses -#")
		cli.Log.Println("#------------------#")
		cli.Log.Println("")

		cli.Log.Println("ExternalDNS :", s.NooBaa.Status.Services.ServiceMgmt.ExternalDNS)
		cli.Log.Println("ExternalIP  :", s.NooBaa.Status.Services.ServiceMgmt.ExternalIP)
		cli.Log.Println("NodePorts   :", s.NooBaa.Status.Services.ServiceMgmt.NodePorts)
		cli.Log.Println("InternalDNS :", s.NooBaa.Status.Services.ServiceMgmt.InternalDNS)
		cli.Log.Println("InternalIP  :", s.NooBaa.Status.Services.ServiceMgmt.InternalIP)
		cli.Log.Println("PodPorts    :", s.NooBaa.Status.Services.ServiceMgmt.PodPorts)

		cli.Log.Println("")
		cli.Log.Println("#----------------#")
		cli.Log.Println("#- S3 Addresses -#")
		cli.Log.Println("#----------------#")
		cli.Log.Println("")

		cli.Log.Println("ExternalDNS :", s.NooBaa.Status.Services.ServiceS3.ExternalDNS)
		cli.Log.Println("ExternalIP  :", s.NooBaa.Status.Services.ServiceS3.ExternalIP)
		cli.Log.Println("NodePorts   :", s.NooBaa.Status.Services.ServiceS3.NodePorts)
		cli.Log.Println("InternalDNS :", s.NooBaa.Status.Services.ServiceS3.InternalDNS)
		cli.Log.Println("InternalIP  :", s.NooBaa.Status.Services.ServiceS3.InternalIP)
		cli.Log.Println("PodPorts    :", s.NooBaa.Status.Services.ServiceS3.PodPorts)

		cli.Log.Println("")
		cli.Log.Println("#---------------#")
		cli.Log.Println("#- Credentials -#")
		cli.Log.Println("#---------------#")
		cli.Log.Println("")
		for key, value := range secret.Data {
			cli.Log.Printf("%s: %s\n", key, string(value))
		}
		cli.Log.Println("")
	} else {
		cli.Log.Printf("❌ System Phase is \"%s\"\n", s.NooBaa.Status.Phase)

	}
}

// SystemWaitReady waits until the system phase changes to ready by the operator
func (cli *CLI) SystemWaitReady() {
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		sys := &nbv1.NooBaa{}
		err := cli.Client.Get(cli.Ctx, client.ObjectKey{Namespace: cli.Namespace, Name: cli.SystemName}, sys)
		if err != nil {
			return false, err
		}
		if sys.Status.Phase == nbv1.SystemPhaseReady {
			cli.Log.Printf("✅ System Phase is \"%s\".\n", sys.Status.Phase)
			return true, nil
		}
		if sys.Status.Phase == nbv1.SystemPhaseRejected {
			return false, fmt.Errorf("❌ System Phase is \"%s\". describe noobaa for more information", sys.Status.Phase)
		}
		cli.Log.Printf("⏳ System Phase is \"%s\". Waiting for it to be ready ...\n", sys.Status.Phase)
		return false, nil
	}))
}

// GetNBClient is a CLI common tool that loads the mgmt api details from the system.
// It gets the endpoint address and token from the system status and secret that the
// operator creates for the system.
func (cli *CLI) GetNBClient() nb.Client {
	s := system.New(types.NamespacedName{Namespace: cli.Namespace, Name: cli.SystemName}, cli.Client, scheme.Scheme, nil)
	s.Load()

	mgmtStatus := s.NooBaa.Status.Services.ServiceMgmt
	if len(mgmtStatus.NodePorts) == 0 {
		fmt.Println("❌ System mgmt service (nodeport) is not ready")
		os.Exit(1)
	}
	if s.SecretOp.StringData["auth_token"] == "" {
		fmt.Println("❌ Operator secret with auth token is not ready")
		os.Exit(1)
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

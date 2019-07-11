package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cli *CLI) SystemCreate() {
	sys := cli.loadSystemDefaults()
	util.KubeCreateSkipExisting(cli.Client, sys)
}

func (cli *CLI) SystemDelete() {
	cli.Log.Println("Delete: Starting")
	sys := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cli.SystemName,
			Namespace: cli.Namespace,
		},
	}
	util.KubeDelete(cli.Client, sys)
	cli.Log.Println("Delete: Done")
}

func (cli *CLI) SystemList() {
	list := nbv1.NooBaaList{}
	util.Fatal(cli.Client.List(cli.Ctx, nil, &list))
	cli.Log.Println("Namespace, Name, Image")
	for _, n := range list.Items {
		cli.Log.Printf("%s, %s, %s\n", n.Namespace, n.Name, n.Status.Phase)
	}
}

func (cli *CLI) SystemStatus() {
	s := system.New(types.NamespacedName{Namespace: cli.Namespace, Name: cli.SystemName}, cli.Client, scheme.Scheme, nil)
	s.Load()

	// sys := cli.loadSystemDefaults()
	// util.KubeCheck(cli.Client, sys)
	cli.Log.Printf("System Phase = \"%s\"\n", s.NooBaa.Status.Phase)
	if s.NooBaa.Status.Phase == nbv1.SystemPhaseReady {
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
	}
}

func (cli *CLI) SystemYaml() {
	sys := cli.loadSystemDefaults()
	p := printers.YAMLPrinter{}
	p.PrintObj(sys, os.Stdout)
}

func (cli *CLI) loadSystemDefaults() *nbv1.NooBaa {
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

func (cli *CLI) SystemWaitReady() {
	intervalSec := time.Duration(3)
	util.Fatal(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
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

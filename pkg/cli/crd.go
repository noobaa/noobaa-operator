package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
)

func (cli *CLI) CrdsCreate() {
	crds := cli.loadCrds()
	util.KubeCreateSkipExisting(cli.Client, crds.NooBaa)
	util.KubeCreateSkipExisting(cli.Client, crds.BackingStore)
	util.KubeCreateSkipExisting(cli.Client, crds.BucketClass)
}

func (cli *CLI) CrdsDelete() {
	crds := cli.loadCrds()
	util.KubeDelete(cli.Client, crds.NooBaa)
	util.KubeDelete(cli.Client, crds.BackingStore)
	util.KubeDelete(cli.Client, crds.BucketClass)
}

func (cli *CLI) CrdsStatus() {
	crds := cli.loadCrds()
	util.KubeCheck(cli.Client, crds.NooBaa)
	util.KubeCheck(cli.Client, crds.BackingStore)
	util.KubeCheck(cli.Client, crds.BucketClass)
}

func (cli *CLI) CrdsWaitReady() {
	crds := cli.loadCrds()
	intervalSec := time.Duration(3)
	util.Fatal(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		ready, err := cli.CrdWaitReady(crds.NooBaa)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
		ready, err = cli.CrdWaitReady(crds.BackingStore)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
		ready, err = cli.CrdWaitReady(crds.BucketClass)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
		return true, nil
	}))
}

func (cli *CLI) CrdWaitReady(crd *apiextv1beta1.CustomResourceDefinition) (bool, error) {
	err := cli.Client.Get(cli.Ctx, client.ObjectKey{Name: crd.Name}, crd)
	if err != nil {
		return false, err
	}
	for _, cond := range crd.Status.Conditions {
		switch cond.Type {
		case apiextv1beta1.Established:
			if cond.Status == apiextv1beta1.ConditionTrue {
				return true, nil
			}
		case apiextv1beta1.NamesAccepted:
			if cond.Status == apiextv1beta1.ConditionFalse {
				return false, fmt.Errorf("Name conflict: %v\n", cond.Reason)
			}
		}
	}
	cli.Log.Printf("CRD not ready")
	return false, nil
}

func (cli *CLI) CrdsYaml() {
	crds := cli.loadCrds()
	p := printers.YAMLPrinter{}
	p.PrintObj(crds.NooBaa, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.BackingStore, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.BucketClass, os.Stdout)
}

type Crds struct {
	NooBaa       *apiextv1beta1.CustomResourceDefinition
	BackingStore *apiextv1beta1.CustomResourceDefinition
	BucketClass  *apiextv1beta1.CustomResourceDefinition
}

func (cli *CLI) loadCrds() *Crds {
	crds := &Crds{}
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml)
	crds.NooBaa = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml)
	crds.BackingStore = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml)
	crds.BucketClass = o.(*apiextv1beta1.CustomResourceDefinition)
	return crds
}

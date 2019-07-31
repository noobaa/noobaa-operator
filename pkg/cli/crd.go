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

// Crds is the
type Crds struct {
	NooBaa            *apiextv1beta1.CustomResourceDefinition
	BackingStore      *apiextv1beta1.CustomResourceDefinition
	BucketClass       *apiextv1beta1.CustomResourceDefinition
	ObjectBucket      *apiextv1beta1.CustomResourceDefinition
	ObjectBucketClaim *apiextv1beta1.CustomResourceDefinition
}

// CrdsCreate runs a CLI command
func (cli *CLI) CrdsCreate() {
	crds := cli.LoadCrds()
	util.KubeCreateSkipExisting(cli.Client, crds.NooBaa)
	util.KubeCreateSkipExisting(cli.Client, crds.BackingStore)
	util.KubeCreateSkipExisting(cli.Client, crds.BucketClass)
	util.KubeCreateSkipExisting(cli.Client, crds.ObjectBucket)
	util.KubeCreateSkipExisting(cli.Client, crds.ObjectBucketClaim)
}

// CrdsDelete runs a CLI command
func (cli *CLI) CrdsDelete() {
	crds := cli.LoadCrds()
	util.KubeDelete(cli.Client, crds.NooBaa)
	util.KubeDelete(cli.Client, crds.BackingStore)
	util.KubeDelete(cli.Client, crds.BucketClass)
	util.KubeDelete(cli.Client, crds.ObjectBucket)
	util.KubeDelete(cli.Client, crds.ObjectBucketClaim)
}

// CrdsStatus runs a CLI command
func (cli *CLI) CrdsStatus() {
	crds := cli.LoadCrds()
	util.KubeCheck(cli.Client, crds.NooBaa)
	util.KubeCheck(cli.Client, crds.BackingStore)
	util.KubeCheck(cli.Client, crds.BucketClass)
	util.KubeCheck(cli.Client, crds.ObjectBucket)
	util.KubeCheck(cli.Client, crds.ObjectBucketClaim)
}

// CrdsYaml dumps a combined yaml of all the CRDs from the bundled yamls
func (cli *CLI) CrdsYaml() {
	crds := cli.LoadCrds()
	p := printers.YAMLPrinter{}
	p.PrintObj(crds.NooBaa, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.BackingStore, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.BucketClass, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.ObjectBucket, os.Stdout)
	fmt.Println("---")
	p.PrintObj(crds.ObjectBucketClaim, os.Stdout)
}

// LoadCrds loads the CRDs structures from the bundled yamls
func (cli *CLI) LoadCrds() *Crds {
	crds := &Crds{}
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml)
	crds.NooBaa = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml)
	crds.BackingStore = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml)
	crds.BucketClass = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_manual_crds_ob_v1alpha1_crd_yaml)
	crds.ObjectBucket = o.(*apiextv1beta1.CustomResourceDefinition)
	o = util.KubeObject(bundle.File_deploy_manual_crds_obc_v1alpha1_crd_yaml)
	crds.ObjectBucketClaim = o.(*apiextv1beta1.CustomResourceDefinition)
	return crds
}

// CrdsWaitReady waits for all CRDs to be ready
func (cli *CLI) CrdsWaitReady() {
	crds := cli.LoadCrds()
	intervalSec := time.Duration(3)
	list := []*apiextv1beta1.CustomResourceDefinition{
		crds.NooBaa, crds.BackingStore, crds.BucketClass,
	}
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		for _, crd := range list {
			ready, err := cli.CrdIsReady(crd)
			if err != nil {
				return false, err
			}
			if !ready {
				return false, nil
			}
		}
		return true, nil
	}))
}

// CrdIsReady checks the status of a CRD
func (cli *CLI) CrdIsReady(crd *apiextv1beta1.CustomResourceDefinition) (bool, error) {
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
				return false, fmt.Errorf("Name conflict: %v", cond.Reason)
			}
		}
	}
	cli.Log.Printf("CRD not ready")
	return false, nil
}

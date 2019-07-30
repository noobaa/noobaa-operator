package crd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
)

type CrdType = apiextv1beta1.CustomResourceDefinition

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crd",
		Short: "Deployment of CRDs",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdStatus(),
		CmdWait(),
		CmdYaml(),
	)
	return cmd
}

func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create noobaa CRDs",
		Run:   RunCreate,
	}
	return cmd
}

func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete noobaa CRDs",
		Run:   RunDelete,
	}
	return cmd
}

func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa CRDs",
		Run:   RunStatus,
	}
	return cmd
}

func CmdWait() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for CRD to be ready",
		Run:   RunWait,
	}
	return cmd
}

func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled CRDs",
		Run:   RunYaml,
	}
	return cmd
}

// Crds is the
type Crds struct {
	NooBaa            *CrdType
	BackingStore      *CrdType
	BucketClass       *CrdType
	ObjectBucket      *CrdType
	ObjectBucketClaim *CrdType
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	CrdCreate(crds.NooBaa)
	CrdCreate(crds.BackingStore)
	CrdCreate(crds.BucketClass)
	CrdCreate(crds.ObjectBucket)
	CrdCreate(crds.ObjectBucketClaim)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	CrdDelete(crds.NooBaa)
	CrdDelete(crds.BackingStore)
	CrdDelete(crds.BucketClass)
	CrdDelete(crds.ObjectBucket)
	CrdDelete(crds.ObjectBucketClaim)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	CrdCheck(crds.NooBaa)
	CrdCheck(crds.BackingStore)
	CrdCheck(crds.BucketClass)
	CrdCheck(crds.ObjectBucket)
	CrdCheck(crds.ObjectBucketClaim)
}

// RunWait runs a CLI command
func RunWait(cmd *cobra.Command, args []string) {
	RunStatus(cmd, args)
	WaitAllReady()
}

// RunYaml dumps a combined yaml of all the CRDs from the bundled yamls
func RunYaml(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
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
func LoadCrds() *Crds {
	o1 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml)
	o2 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml)
	o3 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml)
	o4 := util.KubeObject(bundle.File_deploy_manual_crds_ob_v1alpha1_crd_yaml)
	o5 := util.KubeObject(bundle.File_deploy_manual_crds_obc_v1alpha1_crd_yaml)
	return &Crds{
		NooBaa:            o1.(*CrdType),
		BackingStore:      o2.(*CrdType),
		BucketClass:       o3.(*CrdType),
		ObjectBucket:      o4.(*CrdType),
		ObjectBucketClaim: o5.(*CrdType),
	}
}

// CrdCreate creates a CRD
func CrdCreate(crd *CrdType) {
	util.KubeCreateSkipExisting(crd)
}

// CrdDelete deletes a CRD
func CrdDelete(crd *CrdType) {
	util.KubeDelete(crd)
}

// CrdCheck checks a CRD
func CrdCheck(crd *CrdType) {
	log := util.Logger()
	desired := crd.DeepCopyObject().(*CrdType)
	util.KubeCheck(crd)
	if crd.Spec.Version != desired.Spec.Version {
		log.Printf("❌ CRD Version Mismatch: found %s desired %s",
			crd.Spec.Version, desired.Spec.Version)
	}
}

// WaitAllReady waits for all CRDs to be ready
func WaitAllReady() {
	log := util.Logger()
	klient := util.KubeClient()
	crds := LoadCrds()
	intervalSec := time.Duration(3)
	list := []*CrdType{
		crds.NooBaa, crds.BackingStore, crds.BucketClass,
	}
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		allReady := true
		for _, crd := range list {
			err := klient.Get(util.Context(), client.ObjectKey{Name: crd.Name}, crd)
			util.Panic(err)
			ready, err := IsReady(crd)
			if err != nil {
				log.Printf("❌ %s", err)
				allReady = false
				continue
			}
			if !ready {
				log.Printf("❌ CRD is not ready. Need to wait ...")
				allReady = false
				continue
			}
		}
		return allReady, nil
	}))
}

// IsReady checks the status of a CRD
func IsReady(crd *CrdType) (bool, error) {
	for _, cond := range crd.Status.Conditions {
		switch cond.Type {
		case apiextv1beta1.NamesAccepted:
			if cond.Status == apiextv1beta1.ConditionFalse {
				return false, fmt.Errorf("CRD Name conflict: %v", cond.Reason)
			}
			if cond.Status != apiextv1beta1.ConditionTrue {
				return false, nil
			}
		case apiextv1beta1.Established:
			if cond.Status != apiextv1beta1.ConditionTrue {
				return false, nil
			}
		}
	}
	return true, nil
}

func IsSame(a, b runtime.Object) bool {
	return true
}

// func (crd *CrdType) DeepCopyObject() runtime.Object {
// 	return (*crd).DeepCopyObject()
// }

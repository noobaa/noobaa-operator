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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
)

// CRD is just an alias for a long name
type CRD = apiextv1beta1.CustomResourceDefinition

// Cmd returns a CLI command
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

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create noobaa CRDs",
		Run:   RunCreate,
	}
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete noobaa CRDs",
		Run:   RunDelete,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa CRDs",
		Run:   RunStatus,
	}
	return cmd
}

// CmdWait returns a CLI command
func CmdWait() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for CRD to be ready",
		Run:   RunWait,
	}
	return cmd
}

// CmdYaml returns a CLI command
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
	NooBaa            *CRD
	BackingStore      *CRD
	BucketClass       *CRD
	ObjectBucket      *CRD
	ObjectBucketClaim *CRD
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	CreateCRD(crds.NooBaa)
	CreateCRD(crds.BackingStore)
	CreateCRD(crds.BucketClass)
	CreateCRD(crds.ObjectBucket)
	CreateCRD(crds.ObjectBucketClaim)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	DeleteCRD(crds.NooBaa)
	DeleteCRD(crds.BackingStore)
	DeleteCRD(crds.BucketClass)
	DeleteCRD(crds.ObjectBucket)
	DeleteCRD(crds.ObjectBucketClaim)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	crds := LoadCrds()
	CheckCRD(crds.NooBaa)
	CheckCRD(crds.BackingStore)
	CheckCRD(crds.BucketClass)
	CheckCRD(crds.ObjectBucket)
	CheckCRD(crds.ObjectBucketClaim)
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
	p.PrintObj(crds.BackingStore, os.Stdout)
	p.PrintObj(crds.BucketClass, os.Stdout)
	p.PrintObj(crds.ObjectBucket, os.Stdout)
	p.PrintObj(crds.ObjectBucketClaim, os.Stdout)
}

// LoadCrds loads the CRDs structures from the bundled yamls
func LoadCrds() *Crds {
	o1 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml)
	o2 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml)
	o3 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml)
	o4 := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_ob_crd_yaml)
	o5 := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_obc_crd_yaml)
	return &Crds{
		NooBaa:            o1.(*CRD),
		BackingStore:      o2.(*CRD),
		BucketClass:       o3.(*CRD),
		ObjectBucket:      o4.(*CRD),
		ObjectBucketClaim: o5.(*CRD),
	}
}

// CreateCRD creates a CRD
func CreateCRD(crd *CRD) {
	util.KubeCreateSkipExisting(crd)
}

// DeleteCRD deletes a CRD
func DeleteCRD(crd *CRD) {
	util.KubeDelete(crd)
}

// CheckCRD checks a CRD
func CheckCRD(crd *CRD) {
	log := util.Logger()
	desired := crd.DeepCopyObject().(*CRD)
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
	list := []*CRD{
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
func IsReady(crd *CRD) (bool, error) {
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

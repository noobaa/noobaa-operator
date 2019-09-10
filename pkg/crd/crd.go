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
	All               []*CRD
	NooBaa            *CRD
	BackingStore      *CRD
	BucketClass       *CRD
	ObjectBucket      *CRD
	ObjectBucketClaim *CRD
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	ForEachCRD(CreateCRD)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	ForEachCRD(DeleteCRD)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	ForEachCRD(CheckCRD)
}

// RunWait runs a CLI command
func RunWait(cmd *cobra.Command, args []string) {
	RunStatus(cmd, args)
	WaitAllReady()
}

// RunYaml dumps a combined yaml of all the CRDs from the bundled yamls
func RunYaml(cmd *cobra.Command, args []string) {
	p := printers.YAMLPrinter{}
	ForEachCRD(func(c *CRD) { util.Panic(p.PrintObj(c, os.Stdout)) })
}

// LoadCrds loads the CRDs structures from the bundled yamls
func LoadCrds() *Crds {
	o1 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml)
	o2 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml)
	o3 := util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml)
	o4 := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_crd_yaml)
	o5 := util.KubeObject(bundle.File_deploy_obc_objectbucket_v1alpha1_objectbucket_crd_yaml)
	crds := &Crds{
		NooBaa:            o1.(*CRD),
		BackingStore:      o2.(*CRD),
		BucketClass:       o3.(*CRD),
		ObjectBucketClaim: o4.(*CRD),
		ObjectBucket:      o5.(*CRD),
	}
	crds.All = []*CRD{
		crds.NooBaa,
		crds.BackingStore,
		crds.BucketClass,
		crds.ObjectBucketClaim,
		crds.ObjectBucket,
	}
	return crds
}

// ForEachCRD iterates and calls fn for every CRD
func ForEachCRD(fn func(*CRD)) {
	crds := LoadCrds()
	for _, c := range crds.All {
		fn(c)
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
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		allReady := true
		for _, crd := range crds.All {
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

package crd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
)

// CRD is just an alias for a long name
type CRD = apiextv1.CustomResourceDefinition

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
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete noobaa CRDs",
		Run:   RunDelete,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa CRDs",
		Run:   RunStatus,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdWait returns a CLI command
func CmdWait() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for CRD to be ready",
		Run:   RunWait,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled CRDs",
		Run:   RunYaml,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// Crds is the
type Crds struct {
	All               []*CRD
	NooBaa            *CRD
	BackingStore      *CRD
	NamespaceStore    *CRD
	BucketClass       *CRD
	NooBaaAccount     *CRD
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
	o1 := util.KubeObject(bundle.File_deploy_crds_noobaa_io_noobaas_crd_yaml)
	o2 := util.KubeObject(bundle.File_deploy_crds_noobaa_io_backingstores_crd_yaml)
	o3 := util.KubeObject(bundle.File_deploy_crds_noobaa_io_namespacestores_crd_yaml)
	o4 := util.KubeObject(bundle.File_deploy_crds_noobaa_io_bucketclasses_crd_yaml)
	o5 := util.KubeObject(bundle.File_deploy_crds_noobaa_io_noobaaaccounts_crd_yaml)
	o6 := util.KubeObject(bundle.File_deploy_obc_objectbucket_io_objectbucketclaims_crd_yaml)
	o7 := util.KubeObject(bundle.File_deploy_obc_objectbucket_io_objectbuckets_crd_yaml)
	crds := &Crds{
		NooBaa:            o1.(*CRD),
		BackingStore:      o2.(*CRD),
		NamespaceStore:    o3.(*CRD),
		BucketClass:       o4.(*CRD),
		NooBaaAccount:     o5.(*CRD),
		ObjectBucketClaim: o6.(*CRD),
		ObjectBucket:      o7.(*CRD),
	}
	crds.All = []*CRD{
		crds.NooBaa,
		crds.BackingStore,
		crds.NamespaceStore,
		crds.BucketClass,
		crds.NooBaaAccount,
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
	if crd.Spec.Versions[0].Name != desired.Spec.Versions[0].Name {
		log.Printf("❌ CRD Version Mismatch: found %s desired %s",
			crd.Spec.Versions[0].Name, desired.Spec.Versions[0].Name)
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
		case apiextv1.NamesAccepted:
			if cond.Status == apiextv1.ConditionFalse {
				return false, fmt.Errorf("CRD Name conflict: %v", cond.Reason)
			}
			if cond.Status != apiextv1.ConditionTrue {
				return false, nil
			}
		case apiextv1.Established:
			if cond.Status != apiextv1.ConditionTrue {
				return false, nil
			}
		}
	}
	return true, nil
}

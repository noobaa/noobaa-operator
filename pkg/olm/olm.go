package olm

import (
	"fmt"
	"io"
	"net/http"
	"os"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/crd"
	"github.com/noobaa/noobaa-operator/v2/pkg/operator"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/version"

	"github.com/blang/semver"
	operv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/printers"
	sigyaml "sigs.k8s.io/yaml"
)

type unObj = map[string]interface{}
type unArr = []interface{}

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "olm",
		Short: "OLM related commands",
	}
	cmd.AddCommand(
		CmdCSV(),
		CmdPackage(),
		CmdHubInstall(),
		CmdHubUninstall(),
		CmdHubStatus(),
		// CmdLocalInstall(),
		// CmdLocalUninstall(),
	)
	return cmd
}

// CmdHubInstall returns a CLI command
func CmdHubInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator from operatorhub.io",
		Run:   RunHubInstall,
	}
	return cmd
}

// CmdHubUninstall returns a CLI command
func CmdHubUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator from operatorhub.io",
		Run:   RunHubUninstall,
	}
	return cmd
}

// CmdHubStatus returns a CLI command
func CmdHubStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa-operator from operatorhub.io",
		Run:   RunHubStatus,
	}
	return cmd
}

// CmdLocalInstall returns a CLI command
func CmdLocalInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-install",
		Short: "Install noobaa-operator from local OLM catalog",
		Run:   RunLocalInstall,
	}
	return cmd
}

// CmdLocalUninstall returns a CLI command
func CmdLocalUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-uninstall",
		Short: "Uninstall noobaa-operator from local OLM catalog",
		Run:   RunLocalUninstall,
	}
	return cmd
}

// CmdCSV returns a CLI command
func CmdCSV() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "Print CSV yaml",
		Run:   RunCSV,
	}
	return cmd
}

// CmdPackage returns a CLI command
func CmdPackage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Print OLM package yaml",
		Run:   RunPackage,
	}
	return cmd
}

// RunHubInstall runs a CLI command
func RunHubInstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCreateSkipExisting(obj)
	}
}

// RunHubUninstall runs a CLI command
func RunHubUninstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for i := len(hub.Objects) - 1; i >= 0; i-- {
		obj := hub.Objects[i]
		util.KubeDelete(obj)
	}
}

// RunHubStatus runs a CLI command
func RunHubStatus(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCheck(obj)
	}
}

// RunLocalInstall runs a CLI command
func RunLocalInstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalInstall()")
}

// RunLocalUninstall runs a CLI command
func RunLocalUninstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalUninstall()")
}

// RunCSV runs a CLI command
func RunCSV(cmd *cobra.Command, args []string) {

	opConf := operator.LoadOperatorConf(cmd)
	install := &unstructured.Unstructured{Object: unObj{
		"deployments":        unArr{unObj{"name": opConf.Deployment.Name, "spec": opConf.Deployment.Spec}},
		"permissions":        unArr{unObj{"serviceAccountName": opConf.SA.Name, "rules": opConf.Role.Rules}},
		"clusterPermissions": unArr{unObj{"serviceAccountName": opConf.SA.Name, "rules": opConf.ClusterRole.Rules}},
	}}
	installRaw, err := install.MarshalJSON()
	util.Panic(err)

	o := util.KubeObject(bundle.File_deploy_olm_catalog_noobaa_operator_clusterserviceversion_yaml)
	csv := o.(*operv1.ClusterServiceVersion)
	csv.Name = "noobaa-operator.v" + version.Version
	csv.Namespace = options.Namespace
	csv.Annotations["containerImage"] = options.OperatorImage
	// csv.Annotations["createdAt"] = ???
	// csv.Annotations["alm-examples"] = ""
	csv.Spec.Version.Version = semver.MustParse(version.Version)
	csv.Spec.Description = bundle.File_deploy_olm_catalog_description_md
	csv.Spec.Icon[0].Data = bundle.File_deploy_olm_catalog_noobaa_icon_base64
	csv.Spec.InstallStrategy.StrategySpecRaw = installRaw
	csv.Spec.CustomResourceDefinitions.Owned = []operv1.CRDDescription{}
	csv.Spec.CustomResourceDefinitions.Required = []operv1.CRDDescription{}
	crdDescriptions := map[string]string{
		"NooBaa": `A NooBaa system - Create this to start`,
		"BackingStore": `Storage target spec such as aws-s3, s3-compatible, PV's and more. ` +
			`Used in BacketClass to construct data placement policies.`,
		"BucketClass": `Storage policy spec  tiering, mirroring, spreading. ` +
			`Combines BackingStores. Referenced by ObjectBucketClaims.`,
		"ObjectBucketClaim": `Claim a bucket just like claiming a PV. ` +
			`Automate you app bucket provisioning by creating OBC with your app deployment. ` +
			`A secret and configmap (name=claim) will be created with access details for the app pods.`,
		"ObjectBucket": `Used under-the-hood. Created per ObjectBucketClaim and keeps provisioning information.`,
	}
	crd.ForEachCRD(func(c *crd.CRD) {
		crdDesc := operv1.CRDDescription{
			Name:        c.Name,
			Kind:        c.Spec.Names.Kind,
			Version:     c.Spec.Version,
			DisplayName: c.Spec.Names.Kind,
			Description: crdDescriptions[c.Spec.Names.Kind],
			Resources: []operv1.APIResourceReference{
				operv1.APIResourceReference{Name: "services", Kind: "Service", Version: "v1"},
				operv1.APIResourceReference{Name: "secrets", Kind: "Secret", Version: "v1"},
				operv1.APIResourceReference{Name: "configmaps", Kind: "ConfigMap", Version: "v1"},
				operv1.APIResourceReference{Name: "statefulsets.apps", Kind: "StatefulSet", Version: "v1"},
			},
		}
		if c.Spec.Group == nbv1.SchemeGroupVersion.Group {
			csv.Spec.CustomResourceDefinitions.Owned = append(csv.Spec.CustomResourceDefinitions.Owned, crdDesc)
		} else {
			csv.Spec.CustomResourceDefinitions.Required = append(csv.Spec.CustomResourceDefinitions.Required, crdDesc)
		}
	})

	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(csv, os.Stdout))
}

// RunPackage runs a CLI command
func RunPackage(cmd *cobra.Command, args []string) {
	obj := &unObj{
		"packageName":    "noobaa-operator",
		"defaultChannel": "alpha",
		"channels": unArr{
			unObj{"name": "alpha", "currentCSV": "noobaa-operator.v" + version.Version},
		},
	}
	output, err := sigyaml.Marshal(obj)
	util.Panic(err)
	fmt.Print(string(output))
}

// HubConf keeps the operatorhub yaml objects
type HubConf struct {
	Objects []*unstructured.Unstructured
}

// LoadHubConf loads the operatorhub yaml objects
func LoadHubConf() *HubConf {
	hub := &HubConf{}
	req, err := http.Get("https://operatorhub.io/install/noobaa-operator.yaml")
	util.Panic(err)
	decoder := yaml.NewYAMLOrJSONDecoder(req.Body, 16*1024)
	for {
		obj := &unstructured.Unstructured{}
		err = decoder.Decode(obj)
		if err == io.EOF {
			break
		}
		util.Panic(err)
		hub.Objects = append(hub.Objects, obj)
		// p := printers.YAMLPrinter{}
		// p.PrintObj(obj, os.Stdout)
	}
	return hub
}

package cnpg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"

	cnpgReleases "github.com/cloudnative-pg/cloudnative-pg/releases"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

type CnpgResources struct {
	ClusterRoles                    []*rbacv1.ClusterRole
	ClusterRoleBindings             []*rbacv1.ClusterRoleBinding
	ConfigMaps                      []*corev1.ConfigMap
	CustomResourceDefinitions       []*apiextv1.CustomResourceDefinition
	Deployments                     []*appsv1.Deployment
	MutatingWebhookConfigurations   []*admissionv1.MutatingWebhookConfiguration
	Services                        []*corev1.Service
	ValidatingWebhookConfigurations []*admissionv1.ValidatingWebhookConfiguration
	ServiceAccounts                 []*corev1.ServiceAccount
}

// CmdCNPG returns a CLI command
func CmdCNPG() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cnpg",
		Short: "CloudNativePG operator commands",
		Long: `CloudNativePG operator commands allow managing the PostgreSQL 
operator used by NooBaa for high availability database deployments`,
	}

	cmd.AddCommand(
		CmdInstall(),
		CmdUninstall(),
		CmdYaml(),
		CmdStatus(),
	)
	return cmd
}

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install CloudNativePG operator",
		Run:   RunInstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunInstall runs the CloudNativePG operator installation
func RunInstall(cmd *cobra.Command, args []string) {

	cnpgRes, err := LoadCnpgResources()
	if err != nil {
		util.Panic(err)
	}

	for _, cr := range cnpgRes.ClusterRoles {
		util.KubeCreateSkipExisting(cr)
	}

	for _, sa := range cnpgRes.ServiceAccounts {
		util.KubeCreateSkipExisting(sa)
	}

	for _, crb := range cnpgRes.ClusterRoleBindings {
		util.KubeCreateSkipExisting(crb)
	}

	for _, cm := range cnpgRes.ConfigMaps {
		util.KubeCreateSkipExisting(cm)
	}

	for _, crd := range cnpgRes.CustomResourceDefinitions {
		util.KubeCreateSkipExisting(crd)
	}

	for _, mwh := range cnpgRes.MutatingWebhookConfigurations {
		util.KubeCreateSkipExisting(mwh)
	}

	for _, svc := range cnpgRes.Services {
		util.KubeCreateSkipExisting(svc)
	}

	for _, vwh := range cnpgRes.ValidatingWebhookConfigurations {
		util.KubeCreateSkipExisting(vwh)
	}

	for _, depl := range cnpgRes.Deployments {
		util.KubeCreateSkipExisting(depl)
	}
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall CloudNativePG operator",
		Run:   RunUninstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunUninstall runs a CLI command to uninstall CloudNativePG operator
func RunUninstall(cmd *cobra.Command, args []string) {
	resources, err := LoadCnpgResources()
	if err != nil {
		util.Panic(err)
	}

	// Delete resources in reverse order of creation
	for _, depl := range resources.Deployments {
		util.KubeDelete(depl)
	}

	for _, sa := range resources.ServiceAccounts {
		util.KubeDelete(sa)
	}

	for _, vwh := range resources.ValidatingWebhookConfigurations {
		util.KubeDelete(vwh)
	}

	for _, svc := range resources.Services {
		util.KubeDelete(svc)
	}

	for _, mwh := range resources.MutatingWebhookConfigurations {
		util.KubeDelete(mwh)
	}

	for _, crd := range resources.CustomResourceDefinitions {
		util.KubeDelete(crd)
	}

	for _, cm := range resources.ConfigMaps {
		util.KubeDelete(cm)
	}

	for _, crb := range resources.ClusterRoleBindings {
		util.KubeDelete(crb)
	}

	for _, cr := range resources.ClusterRoles {
		util.KubeDelete(cr)
	}
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled CloudNativePG yaml",
		Run:   RunYaml,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunYaml runs a CLI command to print all CloudNativePG resources
func RunYaml(cmd *cobra.Command, args []string) {
	resources, err := LoadCnpgResources()
	if err != nil {
		util.Panic(err)
	}

	p := printers.YAMLPrinter{}

	// Print all resources by type
	for _, cr := range resources.ClusterRoles {
		util.Panic(p.PrintObj(cr, os.Stdout))
	}

	for _, crb := range resources.ClusterRoleBindings {
		util.Panic(p.PrintObj(crb, os.Stdout))
	}

	for _, cm := range resources.ConfigMaps {
		util.Panic(p.PrintObj(cm, os.Stdout))
	}

	for _, crd := range resources.CustomResourceDefinitions {
		util.Panic(p.PrintObj(crd, os.Stdout))
	}

	for _, deploy := range resources.Deployments {
		util.Panic(p.PrintObj(deploy, os.Stdout))
	}

	for _, mwh := range resources.MutatingWebhookConfigurations {
		util.Panic(p.PrintObj(mwh, os.Stdout))
	}

	for _, svc := range resources.Services {
		util.Panic(p.PrintObj(svc, os.Stdout))
	}

	for _, vwh := range resources.ValidatingWebhookConfigurations {
		util.Panic(p.PrintObj(vwh, os.Stdout))
	}

	for _, sa := range resources.ServiceAccounts {
		util.Panic(p.PrintObj(sa, os.Stdout))
	}
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of CloudNativePG operator deployment",
		Run:   RunStatus,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	log.Printf("CloudNativePG Operator Deployment Status:")

	resources, err := LoadCnpgResources()
	if err != nil {
		util.Panic(err)
	}

	// Check only the deployment
	for _, deploy := range resources.Deployments {
		util.KubeCheck(deploy)
	}

}

// LoadCnpgResources loads all CloudNativePG resources from the embedded manifests
func LoadCnpgResources() (*CnpgResources, error) {
	// Parse the YAML file into resources
	cnpgRes, err := getResourcesFromYaml()
	if err != nil {
		return nil, err
	}

	// Validate that we have all expected resources
	if err := validateResources(cnpgRes); err != nil {
		return nil, err
	}

	//modify resources according to options
	modifyResources(cnpgRes)

	return cnpgRes, nil
}

// modifyResources modifies the resources according to options
func modifyResources(cnpgRes *CnpgResources) {

	// update the deployment namespace, image
	for _, depl := range cnpgRes.Deployments {
		depl.Namespace = options.Namespace
		depl.Spec.Template.Spec.Containers[0].Image = options.CnpgImage
		// add app:noobaa label to the deployments pod
		depl.Spec.Template.Labels["app"] = "noobaa"
		// remove RunAsUser and RunAsGroup from the deployment.
		// This is done for openshift compatibility, otherwise it fails to match an SCC
		depl.Spec.Template.Spec.SecurityContext.RunAsUser = nil
		depl.Spec.Template.Spec.SecurityContext.RunAsGroup = nil
		depl.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = nil
		depl.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup = nil
	}

	// update the cluster role bindings namespace
	for _, crb := range cnpgRes.ClusterRoleBindings {
		for i := range crb.Subjects {
			crb.Subjects[i].Namespace = options.Namespace
		}
	}

	// update the service account namespace
	for _, sa := range cnpgRes.ServiceAccounts {
		sa.Namespace = options.Namespace
	}

	// update the configmap namespace
	for _, cm := range cnpgRes.ConfigMaps {
		cm.Namespace = options.Namespace
	}

	// update the namespace in the  mutating webhooks
	for _, mwh := range cnpgRes.MutatingWebhookConfigurations {
		for i := range mwh.Webhooks {
			mwh.Webhooks[i].ClientConfig.Service.Namespace = options.Namespace
		}
	}

	// update the namespace in the validating webhooks
	for _, vwh := range cnpgRes.ValidatingWebhookConfigurations {
		for i := range vwh.Webhooks {
			vwh.Webhooks[i].ClientConfig.Service.Namespace = options.Namespace
		}
	}

	// update the namespace in the service
	for _, svc := range cnpgRes.Services {
		svc.Namespace = options.Namespace
	}

}

// modifyYamlBytes modifies the YAML bytes to replace the API group and webhooks path
// according to options.UseCnpgApiGroup
func modifyYamlBytes(manifestBytes []byte) []byte {
	if options.UseCnpgApiGroup {
		return manifestBytes
	}

	// Replace API group if different from the original
	originalGroup := "cnpg.io"
	noobaaGroupDomain := "cnpg.noobaa.io"
	mofifiedManifestBytes := bytes.ReplaceAll(manifestBytes,
		[]byte(originalGroup),
		[]byte(noobaaGroupDomain))

	// Replace the webhooks path in the mutating webhook configuration
	originalPath := "cnpg-io"
	newPath := "cnpg-noobaa-io"
	mofifiedManifestBytes = bytes.ReplaceAll(mofifiedManifestBytes,
		[]byte(originalPath),
		[]byte(newPath))

	return mofifiedManifestBytes
}

// getResourcesFromYaml reads and decodes kubernetes resources from a YAML file
// returns the resources with the API group and webhooks path modified according to options.CnpgApiGroup
func getResourcesFromYaml() (*CnpgResources, error) {

	// Get the manifests content from embedded FS
	manifestPath := fmt.Sprintf("cnpg-%s.yaml", options.CnpgVersion)
	manifestBytes, err := cnpgReleases.OperatorManifests.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read operator manifest for version %s: %v", options.CnpgVersion, err)
	}

	manifestBytes = modifyYamlBytes(manifestBytes)

	// Create a new YAML reader
	reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(manifestBytes)))

	// Create decoder for kubernetes resources
	decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	// Initialize resources struct
	cnpgRes := &CnpgResources{}

	// Read documents one at a time
	for {
		// Read single YAML document
		docBytes, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading YAML document: %v", err)
		}

		// Skip empty documents
		if len(bytes.TrimSpace(docBytes)) == 0 {
			continue
		}

		obj, gvk, err := decoder.Decode(docBytes, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode document: %v", err)
		}

		// Store the resource in the appropriate slice based on its type
		switch gvk.Kind {
		case "ClusterRole":
			cnpgRes.ClusterRoles = append(cnpgRes.ClusterRoles, obj.(*rbacv1.ClusterRole))

		case "ClusterRoleBinding":
			cnpgRes.ClusterRoleBindings = append(cnpgRes.ClusterRoleBindings, obj.(*rbacv1.ClusterRoleBinding))

		case "ConfigMap":
			cnpgRes.ConfigMaps = append(cnpgRes.ConfigMaps, obj.(*corev1.ConfigMap))

		case "CustomResourceDefinition":
			cnpgRes.CustomResourceDefinitions = append(cnpgRes.CustomResourceDefinitions, obj.(*apiextv1.CustomResourceDefinition))

		case "Deployment":
			cnpgRes.Deployments = append(cnpgRes.Deployments, obj.(*appsv1.Deployment))

		case "MutatingWebhookConfiguration":
			cnpgRes.MutatingWebhookConfigurations = append(cnpgRes.MutatingWebhookConfigurations, obj.(*admissionv1.MutatingWebhookConfiguration))

		case "Service":
			cnpgRes.Services = append(cnpgRes.Services, obj.(*corev1.Service))

		case "ValidatingWebhookConfiguration":
			cnpgRes.ValidatingWebhookConfigurations = append(cnpgRes.ValidatingWebhookConfigurations, obj.(*admissionv1.ValidatingWebhookConfiguration))

		case "ServiceAccount":
			cnpgRes.ServiceAccounts = append(cnpgRes.ServiceAccounts, obj.(*corev1.ServiceAccount))

		case "Namespace":
			// Skip namespace as it's handled separately

		default:
			return nil, fmt.Errorf("unhandled resource kind: %s, name: %s", gvk.Kind, obj.(metav1.Object).GetName())
		}
	}

	return cnpgRes, nil
}

// validateResources checks that all required resources are present
func validateResources(res *CnpgResources) error {
	var errors []string

	// Expected counts for each resource type
	expectations := map[string]struct {
		actualCount   int
		expactedCount int
	}{
		"ClusterRole":                    {actualCount: len(res.ClusterRoles), expactedCount: 7},
		"ClusterRoleBinding":             {actualCount: len(res.ClusterRoleBindings), expactedCount: 1},
		"ConfigMap":                      {actualCount: len(res.ConfigMaps), expactedCount: 1},
		"CustomResourceDefinition":       {actualCount: len(res.CustomResourceDefinitions), expactedCount: 9},
		"Deployment":                     {actualCount: len(res.Deployments), expactedCount: 1},
		"MutatingWebhookConfiguration":   {actualCount: len(res.MutatingWebhookConfigurations), expactedCount: 1},
		"Service":                        {actualCount: len(res.Services), expactedCount: 1},
		"ValidatingWebhookConfiguration": {actualCount: len(res.ValidatingWebhookConfigurations), expactedCount: 1},
		"ServiceAccount":                 {actualCount: len(res.ServiceAccounts), expactedCount: 1},
	}

	for resourceType, exp := range expectations {
		if exp.actualCount != exp.expactedCount {
			errors = append(errors, fmt.Sprintf("expected %d %s(s), got %d", exp.expactedCount, resourceType, exp.actualCount))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("resource validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

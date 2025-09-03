package cnpg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
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
	CnpgOperatorDeployment         *appsv1.Deployment
	CnpgManagerClusterRole         *rbacv1.ClusterRole
	CnpgManagerClusterRoleBinding  *rbacv1.ClusterRoleBinding
	CnpgManagerRole                *rbacv1.Role
	CnpgManagerRoleBinding         *rbacv1.RoleBinding
	ConfigMap                      *corev1.ConfigMap
	MutatingWebhookConfiguration   *admissionv1.MutatingWebhookConfiguration
	WebhooksService                *corev1.Service
	ValidatingWebhookConfiguration *admissionv1.ValidatingWebhookConfiguration
	ServiceAccount                 *corev1.ServiceAccount
	CRDs                           []*apiextv1.CustomResourceDefinition

	// cluster role and binding for the webhooks permissions
	CnpgWebhooksClusterRole        *rbacv1.ClusterRole
	CnpgWebhooksClusterRoleBinding *rbacv1.ClusterRoleBinding
}

var (
	CnpgAPIGroup   = getCnpgAPIGroup()
	CnpgAPIVersion = CnpgAPIGroup + "/v1"
)

const (
	CnpgDeploymentName = "cnpg-controller-manager"
)

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

	for _, crd := range cnpgRes.CRDs {
		util.KubeCreateSkipExisting(crd)
	}
	util.KubeCreateSkipExisting(cnpgRes.ServiceAccount)
	util.KubeCreateSkipExisting(cnpgRes.CnpgManagerRoleBinding)
	util.KubeCreateSkipExisting(cnpgRes.CnpgManagerClusterRoleBinding)
	util.KubeCreateSkipExisting(cnpgRes.CnpgWebhooksClusterRoleBinding)
	util.KubeCreateSkipExisting(cnpgRes.CnpgWebhooksClusterRole)
	util.KubeCreateSkipExisting(cnpgRes.ConfigMap)
	util.KubeCreateSkipExisting(cnpgRes.MutatingWebhookConfiguration)
	util.KubeCreateSkipExisting(cnpgRes.ValidatingWebhookConfiguration)
	util.KubeCreateSkipExisting(cnpgRes.CnpgOperatorDeployment)
	util.KubeCreateSkipExisting(cnpgRes.CnpgManagerClusterRole)
	util.KubeCreateSkipExisting(cnpgRes.CnpgManagerRole)
	util.KubeCreateSkipExisting(cnpgRes.WebhooksService)
}

// RunUpgrade runs the CloudNativePG operator installation
func RunUpgrade(cmd *cobra.Command, args []string) {

	cnpgRes, err := LoadCnpgResources()
	if err != nil {
		util.Panic(err)
	}

	for _, crd := range cnpgRes.CRDs {
		util.KubeApply(crd)
	}
	util.KubeApply(cnpgRes.ServiceAccount)
	util.KubeApply(cnpgRes.CnpgManagerRoleBinding)
	util.KubeApply(cnpgRes.CnpgManagerClusterRoleBinding)
	util.KubeApply(cnpgRes.CnpgWebhooksClusterRoleBinding)
	util.KubeApply(cnpgRes.CnpgWebhooksClusterRole)
	util.KubeApply(cnpgRes.ConfigMap)
	util.KubeApply(cnpgRes.MutatingWebhookConfiguration)
	util.KubeApply(cnpgRes.ValidatingWebhookConfiguration)
	util.KubeApply(cnpgRes.CnpgOperatorDeployment)
	util.KubeApply(cnpgRes.CnpgManagerClusterRole)
	util.KubeApply(cnpgRes.CnpgManagerRole)
	util.KubeApply(cnpgRes.WebhooksService)
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

	// Delete the resources
	util.KubeDelete(resources.CnpgOperatorDeployment)
	util.KubeDelete(resources.WebhooksService)
	util.KubeDelete(resources.MutatingWebhookConfiguration)
	util.KubeDelete(resources.ValidatingWebhookConfiguration)
	util.KubeDelete(resources.ConfigMap)
	util.KubeDelete(resources.CnpgManagerRoleBinding)
	util.KubeDelete(resources.CnpgManagerRole)
	util.KubeDelete(resources.CnpgManagerClusterRoleBinding)
	util.KubeDelete(resources.CnpgManagerClusterRole)
	util.KubeDelete(resources.CnpgWebhooksClusterRoleBinding)
	util.KubeDelete(resources.CnpgWebhooksClusterRole)
	util.KubeDelete(resources.ServiceAccount)

	// if using the cleanup flag, delete the CRDs
	cleanup, _ := cmd.Flags().GetBool("cleanup")
	if cleanup {
		for _, crd := range resources.CRDs {
			util.KubeDelete(crd)
		}
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
	util.Panic(p.PrintObj(resources.ServiceAccount, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgManagerClusterRole, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgManagerClusterRoleBinding, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgManagerRole, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgManagerRoleBinding, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgWebhooksClusterRole, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgWebhooksClusterRoleBinding, os.Stdout))
	util.Panic(p.PrintObj(resources.ConfigMap, os.Stdout))
	util.Panic(p.PrintObj(resources.MutatingWebhookConfiguration, os.Stdout))
	util.Panic(p.PrintObj(resources.ValidatingWebhookConfiguration, os.Stdout))
	util.Panic(p.PrintObj(resources.CnpgOperatorDeployment, os.Stdout))
	util.Panic(p.PrintObj(resources.WebhooksService, os.Stdout))
	for _, crd := range resources.CRDs {
		util.Panic(p.PrintObj(crd, os.Stdout))
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
	util.KubeCheck(resources.CnpgOperatorDeployment)
}

// LoadCnpgResources loads all CloudNativePG resources from the embedded manifests
func LoadCnpgResources() (*CnpgResources, error) {
	// Parse the YAML file into resources
	cnpgRes, err := getResourcesFromYaml()
	if err != nil {
		return nil, err
	}

	//modify resources according to options
	modifyResources(cnpgRes)

	return cnpgRes, nil
}

// modifyCnpgRbac modifies the RBAC resources for CloudNativePG operator
func modifyCnpgRbac(cnpgRes *CnpgResources) {

	// most of the rules in the cnpg-manager role can be namespace scoped
	// only rules for "nodes" and "clusterimagecatalogs" should be cluster scoped
	const cnpgManagerRoleName = "cnpg-manager"
	cnpgRes.CnpgManagerRole = &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cnpgManagerRoleName,
			Namespace: options.Namespace,
		},
		// we copy the rules from the cluster role
		Rules: cnpgRes.CnpgManagerClusterRole.Rules,
	}

	// add role binding for the role
	cnpgRes.CnpgManagerRoleBinding = &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cnpg-manager-rolebinding",
			Namespace: options.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     cnpgManagerRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      cnpgManagerRoleName,
				Namespace: options.Namespace,
			},
		},
	}

	// remove all unnecessary rules from the cluster role
	// find the rules for "nodes" and "clusterimagecatalogs" and append to new rules
	newRules := []rbacv1.PolicyRule{}
	for _, rule := range cnpgRes.CnpgManagerClusterRole.Rules {
		if slices.Contains(rule.Resources, "nodes") {
			newRules = append(newRules, rbacv1.PolicyRule{
				Verbs:     rule.Verbs,
				APIGroups: rule.APIGroups,
				Resources: []string{"nodes"},
			})
		}
		if slices.Contains(rule.Resources, "clusterimagecatalogs") {
			newRules = append(newRules, rbacv1.PolicyRule{
				Verbs:     rule.Verbs,
				APIGroups: rule.APIGroups,
				Resources: []string{"clusterimagecatalogs"},
			})
		}
	}
	cnpgRes.CnpgManagerClusterRole.Rules = newRules

	//update namespace in the cluster role binding
	for i := range cnpgRes.CnpgManagerClusterRoleBinding.Subjects {
		cnpgRes.CnpgManagerClusterRoleBinding.Subjects[i].Namespace = options.Namespace
	}

	// for non-OLM deployments we need to add cluster-wide permissions for the webhooks
	// create a new cluster role and binding for the webhooks. These are not used in the CSV
	// see here: https://github.com/cloudnative-pg/cloudnative-pg/blob/bad5a251642655399eca392abf5d981668fbd8cc/internal/cmd/manager/controller/controller.go#L362-L390
	cnpgRes.CnpgWebhooksClusterRole = &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cnpg-webhooks-cluster-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
				Verbs:     []string{"get", "patch"},
			},
		},
	}
	cnpgRes.CnpgWebhooksClusterRoleBinding = &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cnpg-webhooks-cluster-rolebinding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cnpg-webhooks-cluster-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "cnpg-manager",
				Namespace: options.Namespace,
			},
		},
	}
}

// modifyResources modifies the resources according to options
func modifyResources(cnpgRes *CnpgResources) {

	// update the deployment namespace, image
	depl := cnpgRes.CnpgOperatorDeployment
	depl.Name = CnpgDeploymentName
	depl.Namespace = options.Namespace
	depl.Spec.Template.Spec.Containers[0].Image = options.CnpgImage
	// add app:noobaa label to the deployments pod
	depl.Spec.Template.Labels["app"] = "noobaa"
	// add WATCH_NAMESPACE env variable to the deployment to restrict the operator to current namespace
	depl.Spec.Template.Spec.Containers[0].Env = append(depl.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name: "WATCH_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	})
	if options.UseCnpgApiGroup {
		depl.Spec.Template.Spec.Containers[0].Env = append(depl.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "USE_CNPG_API_GROUP",
			Value: "true",
		})
	}
	// modify the env variable OPERATOR_IMAGE_NAME according to options.CnpgImage
	for i := range depl.Spec.Template.Spec.Containers[0].Env {
		if depl.Spec.Template.Spec.Containers[0].Env[i].Name == "OPERATOR_IMAGE_NAME" {
			depl.Spec.Template.Spec.Containers[0].Env[i].Value = options.CnpgImage
		}
	}

	modifyCnpgRbac(cnpgRes)

	// update the service account namespace
	cnpgRes.ServiceAccount.Namespace = options.Namespace

	// update the configmap namespace
	cnpgRes.ConfigMap.Namespace = options.Namespace

	// update the namespace in the  mutating webhooks
	for i := range cnpgRes.MutatingWebhookConfiguration.Webhooks {
		cnpgRes.MutatingWebhookConfiguration.Webhooks[i].ClientConfig.Service.Namespace = options.Namespace
	}
	for i := range cnpgRes.ValidatingWebhookConfiguration.Webhooks {
		cnpgRes.ValidatingWebhookConfiguration.Webhooks[i].ClientConfig.Service.Namespace = options.Namespace
	}

	// update the namespace in the service
	cnpgRes.WebhooksService.Namespace = options.Namespace
}

// modifyYamlBytes modifies the YAML bytes to replace the API group and webhooks path
// according to options.UseCnpgApiGroup
func modifyYamlBytes(manifestBytes []byte) []byte {
	if options.UseCnpgApiGroup {
		return manifestBytes
	}

	// Replace API group if different from the original
	originalGroup := "postgresql.cnpg.io"
	noobaaGroupDomain := "postgresql.cnpg.noobaa.io"
	mofifiedManifestBytes := bytes.ReplaceAll(manifestBytes,
		[]byte(originalGroup),
		[]byte(noobaaGroupDomain))

	// Replace the webhooks path in the mutating webhook configuration
	originalPath := "postgresql-cnpg-io"
	newPath := "postgresql-cnpg-noobaa-io"
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

		// Store the resource in the appropriate object in the cnpgRes struct
		switch gvk.Kind {
		case "ClusterRole":
			if obj.(*rbacv1.ClusterRole).Name == "cnpg-manager" {
				cnpgRes.CnpgManagerClusterRole = obj.(*rbacv1.ClusterRole)
			}

		case "ClusterRoleBinding":
			cnpgRes.CnpgManagerClusterRoleBinding = obj.(*rbacv1.ClusterRoleBinding)

		case "ConfigMap":
			cnpgRes.ConfigMap = obj.(*corev1.ConfigMap)

		case "CustomResourceDefinition":
			cnpgRes.CRDs = append(cnpgRes.CRDs, obj.(*apiextv1.CustomResourceDefinition))

		case "Deployment":
			cnpgRes.CnpgOperatorDeployment = obj.(*appsv1.Deployment)

		case "MutatingWebhookConfiguration":
			cnpgRes.MutatingWebhookConfiguration = obj.(*admissionv1.MutatingWebhookConfiguration)

		case "Service":
			cnpgRes.WebhooksService = obj.(*corev1.Service)

		case "ValidatingWebhookConfiguration":
			cnpgRes.ValidatingWebhookConfiguration = obj.(*admissionv1.ValidatingWebhookConfiguration)

		case "ServiceAccount":
			cnpgRes.ServiceAccount = obj.(*corev1.ServiceAccount)

		case "Namespace":
			// Skip namespace as it's handled separately

		default:
			return nil, fmt.Errorf("unhandled resource kind: %s, name: %s", gvk.Kind, obj.(metav1.Object).GetName())
		}
	}

	// After reading all documents, validate that all required resources were found
	if cnpgRes.CnpgManagerClusterRole == nil {
		return nil, fmt.Errorf("required ClusterRole 'cnpg-manager' not found in manifest")
	}
	if cnpgRes.CnpgManagerClusterRoleBinding == nil {
		return nil, fmt.Errorf("required ClusterRoleBinding not found in manifest")
	}
	if cnpgRes.ConfigMap == nil {
		return nil, fmt.Errorf("required ConfigMap not found in manifest")
	}
	if cnpgRes.CnpgOperatorDeployment == nil {
		return nil, fmt.Errorf("required Deployment not found in manifest")
	}
	if cnpgRes.MutatingWebhookConfiguration == nil {
		return nil, fmt.Errorf("required MutatingWebhookConfiguration not found in manifest")
	}
	if cnpgRes.ValidatingWebhookConfiguration == nil {
		return nil, fmt.Errorf("required ValidatingWebhookConfiguration not found in manifest")
	}
	if cnpgRes.ServiceAccount == nil {
		return nil, fmt.Errorf("required ServiceAccount not found in manifest")
	}
	if cnpgRes.WebhooksService == nil {
		return nil, fmt.Errorf("required Service not found in manifest")
	}
	expectedNumberOfCrds := 10
	if len(cnpgRes.CRDs) != expectedNumberOfCrds {
		return nil, fmt.Errorf("expected %d CustomResourceDefinitions, got %d", expectedNumberOfCrds, len(cnpgRes.CRDs))
	}

	return cnpgRes, nil
}

// getCnpgAPIGroup returns the API group to use for CNPG resources
// by default it's "postgresql.cnpg.noobaa.io"
func getCnpgAPIGroup() string {
	useCnpgApiGroup := os.Getenv("USE_CNPG_API_GROUP")
	if useCnpgApiGroup == "true" {
		return "postgresql.cnpg.io"
	}
	return "postgresql.cnpg.noobaa.io"
}

func GetCnpgImageCatalogObj(namespace string, name string) *cnpgv1.ImageCatalog {

	cnpgImageCatalog := &cnpgv1.ImageCatalog{
		TypeMeta: metav1.TypeMeta{
			APIVersion: CnpgAPIVersion,
			Kind:       "ImageCatalog",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return cnpgImageCatalog
}

// GetCnpgCluster returns a new CNPG cluster resource
func GetCnpgClusterObj(namespace string, name string) *cnpgv1.Cluster {
	// cnpgCluster := util.KubeObject(bundle.File_deploy_internal_cnpg_cluster_yaml).(*cnpgv1.Cluster)
	cnpgCluster := &cnpgv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: CnpgAPIVersion,
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{},
		},
	}
	return cnpgCluster
}

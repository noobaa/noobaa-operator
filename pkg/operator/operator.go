package operator

import (
	"os"
	"strings"

	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Deployment using operator",
	}
	cmd.AddCommand(
		CmdInstall(),
		CmdUninstall(),
		CmdStatus(),
		CmdYaml(),
		CmdRun(),
	)
	return cmd
}

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator",
		Run:   RunInstall,
	}
	cmd.Flags().Bool("no-deploy", false, "Install only the needed resources but do not create the operator deployment")
	return cmd
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator",
		Run:   RunUninstall,
	}
	cmd.Flags().Bool("cleanup", false, "Enable deletion of the Namespace")
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa-operator",
		Run:   RunStatus,
	}
	return cmd
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled noobaa-operator yaml",
		Run:   RunYaml,
	}
	return cmd
}

// CmdRun returns a CLI command
func CmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the noobaa-operator",
		Run:   RunOperator,
	}
	return cmd
}

// RunInstall runs a CLI command
func RunInstall(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	util.KubeCreateSkipExisting(c.NS)
	util.KubeCreateSkipExisting(c.SA)
	util.KubeCreateSkipExisting(c.Role)
	util.KubeCreateSkipExisting(c.RoleBinding)
	util.KubeCreateSkipExisting(c.ClusterRole)
	util.KubeCreateSkipExisting(c.ClusterRoleBinding)
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.KubeCreateSkipExisting(c.Deployment)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.KubeDelete(c.Deployment)
	}
	util.KubeDelete(c.ClusterRoleBinding)
	util.KubeDelete(c.ClusterRole)
	util.KubeDelete(c.RoleBinding)
	util.KubeDelete(c.Role)
	util.KubeDelete(c.SA)

	cleanup, _ := cmd.Flags().GetBool("cleanup")
	reservedNS := c.NS.Name == "default" ||
		strings.HasPrefix(c.NS.Name, "openshift-") ||
		strings.HasPrefix(c.NS.Name, "kubernetes-") ||
		strings.HasPrefix(c.NS.Name, "kube-")
	if reservedNS {
		log.Printf("Namespace Delete: disabled for reserved namespace")
		log.Printf("Namespace Status:")
		util.KubeCheck(c.NS)
	} else if !cleanup {
		log.Printf("Namespace Delete: currently disabled (enable with \"--cleanup\")")
		log.Printf("Namespace Status:")
		util.KubeCheck(c.NS)
	} else {
		log.Printf("Namespace Delete:")
		util.KubeDelete(c.NS)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	util.KubeCheck(c.NS)
	if util.KubeCheck(c.SA) {
		// in OLM deployment the roles and bindings have generated names
		// so we list and lookup bindings to our service account to discover the actual names
		DetectRole(c)
		DetectClusterRole(c)
	}
	util.KubeCheck(c.Role)
	util.KubeCheck(c.RoleBinding)
	util.KubeCheck(c.ClusterRole)
	util.KubeCheck(c.ClusterRoleBinding)
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.KubeCheck(c.Deployment)
	}
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(c.NS, os.Stdout))
	util.Panic(p.PrintObj(c.SA, os.Stdout))
	util.Panic(p.PrintObj(c.Role, os.Stdout))
	util.Panic(p.PrintObj(c.RoleBinding, os.Stdout))
	util.Panic(p.PrintObj(c.ClusterRole, os.Stdout))
	util.Panic(p.PrintObj(c.ClusterRoleBinding, os.Stdout))
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.Panic(p.PrintObj(c.Deployment, os.Stdout))
	}
}

// Conf struct holds all the objects needed to install the operator
type Conf struct {
	NS                 *corev1.Namespace
	SA                 *corev1.ServiceAccount
	Role               *rbacv1.Role
	RoleBinding        *rbacv1.RoleBinding
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	Deployment         *appsv1.Deployment
}

// LoadOperatorConf loads and initializes all the objects needed to install the operator
func LoadOperatorConf(cmd *cobra.Command) *Conf {
	c := &Conf{}

	c.NS = util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	c.SA = util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount)
	c.Role = util.KubeObject(bundle.File_deploy_role_yaml).(*rbacv1.Role)
	c.RoleBinding = util.KubeObject(bundle.File_deploy_role_binding_yaml).(*rbacv1.RoleBinding)
	c.ClusterRole = util.KubeObject(bundle.File_deploy_cluster_role_yaml).(*rbacv1.ClusterRole)
	c.ClusterRoleBinding = util.KubeObject(bundle.File_deploy_cluster_role_binding_yaml).(*rbacv1.ClusterRoleBinding)
	c.Deployment = util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)

	c.NS.Name = options.Namespace
	c.SA.Namespace = options.Namespace
	c.Role.Namespace = options.Namespace
	c.RoleBinding.Namespace = options.Namespace
	c.ClusterRole.Namespace = options.Namespace
	c.Deployment.Namespace = options.Namespace

	c.ClusterRole.Name = options.SubDomainNS()
	c.ClusterRoleBinding.Name = c.ClusterRole.Name
	c.ClusterRoleBinding.RoleRef.Name = c.ClusterRole.Name
	for i := range c.ClusterRoleBinding.Subjects {
		c.ClusterRoleBinding.Subjects[i].Namespace = options.Namespace
	}

	c.Deployment.Spec.Template.Spec.Containers[0].Image = options.OperatorImage
	if options.ImagePullSecret != "" {
		c.Deployment.Spec.Template.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{{Name: options.ImagePullSecret}}
	}
	return c
}

// DetectRole looks up a role binding referencing our service account
func DetectRole(c *Conf) {
	roleBindings := &rbacv1.RoleBindingList{}
	selector := labels.SelectorFromSet(labels.Set{
		"olm.owner.kind":      "ClusterServiceVersion",
		"olm.owner.namespace": c.SA.Namespace,
	})
	util.KubeList(roleBindings, &client.ListOptions{
		Namespace:     c.SA.Namespace,
		LabelSelector: selector,
	})
	for i := range roleBindings.Items {
		b := &roleBindings.Items[i]
		for j := range b.Subjects {
			s := b.Subjects[j]
			if s.Kind == "ServiceAccount" &&
				s.Name == c.SA.Name &&
				s.Namespace == c.SA.Namespace {
				c.Role.Name = b.RoleRef.Name
				c.RoleBinding.Name = b.Name
				return
			}
		}
	}
}

// DetectClusterRole looks up a cluster role binding referencing our service account
func DetectClusterRole(c *Conf) {
	clusterRoleBindings := &rbacv1.ClusterRoleBindingList{}
	selector := labels.SelectorFromSet(labels.Set{
		"olm.owner.kind":      "ClusterServiceVersion",
		"olm.owner.namespace": c.SA.Namespace,
	})
	util.KubeList(clusterRoleBindings, &client.ListOptions{
		LabelSelector: selector,
	})
	for i := range clusterRoleBindings.Items {
		b := &clusterRoleBindings.Items[i]
		for j := range b.Subjects {
			s := b.Subjects[j]
			if s.Kind == "ServiceAccount" &&
				s.Name == c.SA.Name &&
				s.Namespace == c.SA.Namespace {
				c.ClusterRole.Name = b.RoleRef.Name
				c.ClusterRoleBinding.Name = b.Name
				return
			}
		}
	}
}

package operator

import (
	"os"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	"github.com/noobaa/noobaa-operator/pkg/controller"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

// Cmd creates a CLI command
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

// CmdInstall creates a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator",
		Run:   RunInstall,
	}
	cmd.Flags().Bool("rbac-only", false, "Install only the RBAC needed for local operator")
	return cmd
}

// CmdUninstall creates a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator",
		Run:   RunUninstall,
	}
	return cmd
}

// CmdStatus creates a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa-operator",
		Run:   RunStatus,
	}
	return cmd
}

// CmdYaml creates a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled noobaa-operator yaml",
		Run:   RunYaml,
	}
	return cmd
}

// CmdRun creates a CLI command
func CmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the noobaa-operator",
		Run: func(cmd *cobra.Command, args []string) {
			controller.OperatorMain()
		},
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
	util.KubeCreateSkipExisting(c.StorageClass)
	rbacOnly, _ := cmd.Flags().GetBool("rbac-only")
	if !rbacOnly {
		util.KubeCreateSkipExisting(c.Deployment)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	util.KubeDelete(c.NS)
	util.KubeDelete(c.SA)
	util.KubeDelete(c.Role)
	util.KubeDelete(c.RoleBinding)
	util.KubeDelete(c.ClusterRole)
	util.KubeDelete(c.ClusterRoleBinding)
	util.KubeDelete(c.StorageClass)
	util.KubeDelete(c.Deployment)
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	util.KubeCheck(c.NS)
	util.KubeCheck(c.SA)
	util.KubeCheck(c.Role)
	util.KubeCheck(c.RoleBinding)
	util.KubeCheck(c.ClusterRole)
	util.KubeCheck(c.ClusterRoleBinding)
	util.KubeCheck(c.StorageClass)
	util.KubeCheck(c.Deployment)
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	p := printers.YAMLPrinter{}
	p.PrintObj(c.NS, os.Stdout)
	p.PrintObj(c.SA, os.Stdout)
	p.PrintObj(c.Role, os.Stdout)
	p.PrintObj(c.RoleBinding, os.Stdout)
	p.PrintObj(c.ClusterRole, os.Stdout)
	p.PrintObj(c.ClusterRoleBinding, os.Stdout)
	p.PrintObj(c.StorageClass, os.Stdout)
	p.PrintObj(c.Deployment, os.Stdout)
}

// Conf struct holds all the objects needed to install the operator
type Conf struct {
	NS                 *corev1.Namespace
	SA                 *corev1.ServiceAccount
	Role               *rbacv1.Role
	RoleBinding        *rbacv1.RoleBinding
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	StorageClass       *storagev1.StorageClass
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
	c.StorageClass = util.KubeObject(bundle.File_deploy_obc_storage_class_yaml).(*storagev1.StorageClass)
	c.Deployment = util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)

	c.StorageClass.Provisioner = "noobaa.io/" + options.Namespace + ".bucket"
	c.StorageClass.Name = options.Namespace + "-storage-class"
	c.NS.Name = options.Namespace
	c.SA.Namespace = options.Namespace
	c.Role.Namespace = options.Namespace
	c.RoleBinding.Namespace = options.Namespace
	c.ClusterRole.Namespace = options.Namespace
	c.ClusterRoleBinding.Name = c.ClusterRoleBinding.Name + "-" + options.Namespace
	for i := range c.ClusterRoleBinding.Subjects {
		c.ClusterRoleBinding.Subjects[i].Namespace = options.Namespace
	}
	c.Deployment.Namespace = options.Namespace
	c.Deployment.Spec.Template.Spec.Containers[0].Image = options.OperatorImage
	if options.ImagePullSecret != "" {
		c.Deployment.Spec.Template.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{{Name: options.ImagePullSecret}}
	}
	return c
}

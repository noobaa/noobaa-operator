package cli

import (
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (cli *CLI) OperatorInstall() {
	c := cli.loadOperatorConf()
	util.KubeCreateSkipExisting(cli.Client, c.NS)
	util.KubeCreateSkipExisting(cli.Client, c.SA)
	util.KubeCreateSkipExisting(cli.Client, c.Role)
	util.KubeCreateSkipExisting(cli.Client, c.RoleBinding)
	util.KubeCreateSkipExisting(cli.Client, c.ClusterRole)
	util.KubeCreateSkipExisting(cli.Client, c.ClusterRoleBinding)
	util.KubeCreateSkipExisting(cli.Client, c.Deployment)
}

func (cli *CLI) OperatorLocalInstall() {
	c := cli.loadOperatorConf()
	util.KubeCreateSkipExisting(cli.Client, c.NS)
	util.KubeCreateSkipExisting(cli.Client, c.SA)
	util.KubeCreateSkipExisting(cli.Client, c.Role)
	util.KubeCreateSkipExisting(cli.Client, c.RoleBinding)
	util.KubeCreateSkipExisting(cli.Client, c.ClusterRole)
	util.KubeCreateSkipExisting(cli.Client, c.ClusterRoleBinding)
}

func (cli *CLI) OperatorUninstall() {
	c := cli.loadOperatorConf()
	util.KubeDelete(cli.Client, c.NS)
	util.KubeDelete(cli.Client, c.SA)
	util.KubeDelete(cli.Client, c.Role)
	util.KubeDelete(cli.Client, c.RoleBinding)
	util.KubeDelete(cli.Client, c.ClusterRole)
	util.KubeDelete(cli.Client, c.ClusterRoleBinding)
	util.KubeDelete(cli.Client, c.Deployment)
}

func (cli *CLI) OperatorLocalUninstall() {
	c := cli.loadOperatorConf()
	util.KubeDelete(cli.Client, c.NS)
	util.KubeDelete(cli.Client, c.SA)
	util.KubeDelete(cli.Client, c.Role)
	util.KubeDelete(cli.Client, c.RoleBinding)
	util.KubeDelete(cli.Client, c.ClusterRole)
	util.KubeDelete(cli.Client, c.ClusterRoleBinding)
}

func (cli *CLI) OperatorStatus() {
	c := cli.loadOperatorConf()
	util.KubeCheck(cli.Client, c.NS)
	util.KubeCheck(cli.Client, c.SA)
	util.KubeCheck(cli.Client, c.Role)
	util.KubeCheck(cli.Client, c.RoleBinding)
	util.KubeCheck(cli.Client, c.ClusterRole)
	util.KubeCheck(cli.Client, c.ClusterRoleBinding)
	util.KubeCheck(cli.Client, c.Deployment)
}

func (cli *CLI) OperatorLocalReconcile() {
	intervalSec := time.Duration(3)
	util.Fatal(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: cli.Namespace,
				Name:      cli.SystemName,
			},
		}
		res, err := system.New(req.NamespacedName, cli.Client, scheme.Scheme, nil).Reconcile()
		if err != nil {
			return false, err
		}
		if res.Requeue || res.RequeueAfter != 0 {
			cli.Log.Printf("\nRetrying in %d seconds\n", intervalSec)
			return false, nil
		}
		return true, nil
	}))
}

func (cli *CLI) OperatorYamls() {
	c := cli.loadOperatorConf()
	p := printers.YAMLPrinter{}
	p.PrintObj(c.NS, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.SA, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.Role, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.RoleBinding, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.ClusterRole, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.ClusterRoleBinding, os.Stdout)
	fmt.Println("---")
	p.PrintObj(c.Deployment, os.Stdout)
}

type OperatorConf struct {
	NS                 *corev1.Namespace
	SA                 *corev1.ServiceAccount
	Role               *rbacv1.Role
	RoleBinding        *rbacv1.RoleBinding
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	Deployment         *appsv1.Deployment
}

func (cli *CLI) loadOperatorConf() *OperatorConf {
	c := &OperatorConf{}
	c.NS = util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	c.SA = util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount)
	c.Role = util.KubeObject(bundle.File_deploy_role_yaml).(*rbacv1.Role)
	c.RoleBinding = util.KubeObject(bundle.File_deploy_role_binding_yaml).(*rbacv1.RoleBinding)
	c.ClusterRole = util.KubeObject(bundle.File_deploy_cluster_role_yaml).(*rbacv1.ClusterRole)
	c.ClusterRoleBinding = util.KubeObject(bundle.File_deploy_cluster_role_binding_yaml).(*rbacv1.ClusterRoleBinding)
	c.Deployment = util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)

	c.NS.Name = cli.Namespace
	c.SA.Namespace = cli.Namespace
	c.Role.Namespace = cli.Namespace
	c.RoleBinding.Namespace = cli.Namespace
	c.ClusterRole.Namespace = cli.Namespace
	c.ClusterRoleBinding.Name = c.ClusterRoleBinding.Name + "-" + cli.Namespace
	for i := range c.ClusterRoleBinding.Subjects {
		c.ClusterRoleBinding.Subjects[i].Namespace = cli.Namespace
	}
	c.Deployment.Namespace = cli.Namespace
	c.Deployment.Spec.Template.Spec.Containers[0].Image = cli.OperatorImage
	if cli.ImagePullSecret != "" {
		c.Deployment.Spec.Template.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{{Name: cli.ImagePullSecret}}
	}
	return c
}

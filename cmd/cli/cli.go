package main

import (
	"context"
	"fmt"

	"github.com/noobaa/noobaa-operator/cmd/cli/bundle"
	"github.com/noobaa/noobaa-operator/pkg/apis"

	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//go:generate go run gen/bundle.go

// ASCIILogo is noobaa's logo ascii art
const ASCIILogo = `
 /~~\\__~__//~~\
|               |
 \~\\_     _//~/
     \\   //
      |   |
      \~~~/
`

var namespace = "noobaa"
var ctx = context.TODO()

// CLI is the top command for noobaa CLI
var CLI = &cobra.Command{
	Use:  "noobaa",
	Long: "\n   NooBaa CLI \n" + ASCIILogo,
}

// InstallCommand installs to kubernetes
var InstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install to kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Installing namespace:%s\n", namespace)
		c, err := client.New(config.GetConfigOrDie(), client.Options{})
		fatal(err)
		createNamespace(c)
		createRBAC(c)
		createCRDs(c)
		createOperator(c)
	},
}

// UninstallCommand uninstalls from kubernetes
var UninstallCommand = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall from kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Uninstalling ...")
	},
}

// CreateCommand uninstalls from kubernetes
var CreateCommand = &cobra.Command{
	Use:   "create",
	Short: "Create a NooBaa system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating ...")
		c, err := client.New(config.GetConfigOrDie(), client.Options{})
		fatal(err)
		createSystem(c)
	},
}

// DeleteCommand uninstalls from kubernetes
var DeleteCommand = &cobra.Command{
	Use:   "delete",
	Short: "Delete a NooBaa system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Deleting ...")
	},
}

func main() {
	apiextv1beta1.AddToScheme(scheme.Scheme)
	apis.AddToScheme(scheme.Scheme)
	CLI.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Target namespace")
	CLI.AddCommand(InstallCommand)
	CLI.AddCommand(UninstallCommand)
	CLI.AddCommand(CreateCommand)
	CLI.AddCommand(DeleteCommand)
	fatal(CLI.Execute())
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func kubeObject(ns string, text string) runtime.Object {
	// Decode text (yaml/json) to kube api object
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	obj, group, err := deserializer.Decode([]byte(text), nil, nil)
	// obj, group, err := scheme.Codecs.UniversalDecoder().Decode([]byte(text), nil, nil)
	fatal(err)
	// not sure if really needed, but set it anyway
	obj.GetObjectKind().SetGroupVersionKind(*group)
	metaObj := obj.(metav1.Object)
	if ns != "" {
		metaObj.SetNamespace(ns)
	}
	return obj
}

func kubeApply(c client.Client, obj runtime.Object) {
	err := c.Create(ctx, obj)
	if errors.IsAlreadyExists(err) {
		err = c.Update(ctx, obj)
	}
	fatal(err)
}

func kubeCreateSkipExisting(c client.Client, obj runtime.Object) {
	err := c.Create(ctx, obj)
	if errors.IsAlreadyExists(err) {
		return
	}
	fatal(err)
}

func createNamespace(c client.Client) {
	kubeApply(c, kubeObject("", fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)))
}

func createRBAC(c client.Client) {
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_service_account_yaml))
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_role_yaml))
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_role_binding_yaml))
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_cluster_role_yaml))

	crb := kubeObject(namespace, bundle.File_deploy_cluster_role_binding_yaml).(*rbacv1.ClusterRoleBinding)
	crb.Name = crb.Name + "-" + namespace
	for i := range crb.Subjects {
		crb.Subjects[i].Namespace = namespace
	}
	kubeApply(c, crb)
}

func createCRDs(c client.Client) {
	kubeCreateSkipExisting(c, kubeObject("", bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml))
}

func createOperator(c client.Client) {
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_operator_yaml))
}

func createSystem(c client.Client) {
	kubeApply(c, kubeObject(namespace, bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml))
}

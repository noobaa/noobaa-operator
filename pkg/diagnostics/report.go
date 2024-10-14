package diagnostics

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	appNoobaaCore     = "NOOBAA-CORE"
	appNoobaaEndpoint = "NOOBAA-ENDPOINT"
)

// RunReport runs a CLI command
func RunReport(cmd *cobra.Command, args []string) {
	log := util.Logger()

	// Fetching coreApp configurations
	coreApp := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	coreApp.Namespace = options.Namespace
	if !util.KubeCheck(coreApp) {
		log.Fatalf(`❌ Could not get core StatefulSet %q in Namespace %q`,
			coreApp.Name, coreApp.Namespace)
	}

	// Fetching endpoint configurations
	endpointApp := util.KubeObject(bundle.File_deploy_internal_deployment_endpoint_yaml).(*appsv1.Deployment)
	endpointApp.Namespace = options.Namespace
	if !util.KubeCheck(endpointApp) {
		log.Fatalf(`❌ Could not get endpoint Deployment %q in Namespace %q`,
			endpointApp.Name, endpointApp.Namespace)
	}

	// Fetching all Backingstores
	bsList := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(bsList, &client.ListOptions{Namespace: options.Namespace}) {
		log.Fatalf(`❌ No backingstores were found in the %q namespace`, options.Namespace)
	}

	// Fetching all Namespacestores
	nsList := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	if !util.KubeList(nsList, &client.ListOptions{Namespace: options.Namespace}) {
		log.Fatalf(`❌ No namespacestores were found in the %q namespace`, options.Namespace)
	}
	fmt.Println("")

	// retrieving the status of proxy environment variables
	proxyStatus(coreApp, endpointApp)

	// retrieving the overridden env variables using `CONFIG_JS_` prefix
	overriddenEnvVar(coreApp, endpointApp)

	// validating ARNs for backingstores and namespacestores
	arnValidationCheck(bsList, nsList)

	// TODO: Add support for additional features
}

// proxyStatus returns the status of the environment variables: HTTP_PROXY, HTTPS_PROXY, and NO_PROXY
func proxyStatus(coreApp *appsv1.StatefulSet, endpointApp *appsv1.Deployment) {
	log := util.Logger()

	log.Print("⏳ Retrieving proxy environment variable details...\n")

	printProxyStatus(appNoobaaCore, coreApp.Spec.Template.Spec.Containers[0].Env)

	printProxyStatus(appNoobaaEndpoint, endpointApp.Spec.Template.Spec.Containers[0].Env)

	fmt.Println("")
}

// overriddenEnvVar retrieves and displays overridden environment variables with the prefix `CONFIG_JS_` from the noobaa-core-0 pod
func overriddenEnvVar(coreApp *appsv1.StatefulSet, endpointApp *appsv1.Deployment) {
	log := util.Logger()

	log.Print("⏳ Retrieving overridden environment variable details...\n")

	printOverriddenEnvVar(appNoobaaCore, coreApp.Spec.Template.Spec.Containers[0].Env)

	printOverriddenEnvVar(appNoobaaEndpoint, endpointApp.Spec.Template.Spec.Containers[0].Env)

	fmt.Println("")
}

// arnValidationCheck validates the ARNs for backingstores and namespacestores
func arnValidationCheck(bsList *nbv1.BackingStoreList, nsList *nbv1.NamespaceStoreList) {
	log := util.Logger()

	log.Print("⏳ Validating store ARNs...\n")

	// Validate ARNs for backingstores
	bsArnList := make(map[string]string)
	for _, bs := range bsList.Items {
		if bs.Spec.AWSS3 != nil && bs.Spec.AWSS3.AWSSTSRoleARN != nil {
			bsArnList[bs.Name] = *bs.Spec.AWSS3.AWSSTSRoleARN
		}
	}
	printARNStatus("BACKINGSTORE", bsArnList)

	// Validate ARNs for namespacestores
	nsArnList := make(map[string]string)
	for _, ns := range nsList.Items {
		if ns.Spec.AWSS3 != nil && ns.Spec.AWSS3.AWSSTSRoleARN != nil {
			nsArnList[ns.Name] = *ns.Spec.AWSS3.AWSSTSRoleARN
		}
	}
	printARNStatus("NAMESPACESTORE", nsArnList)

	fmt.Println("")
}

// printProxyStatus prints the proxy status
func printProxyStatus(appName string, envVars []corev1.EnvVar) {
	fmt.Printf("Proxy Environment Variables Check (%s):\n----------------------------------\n", appName)
	for _, proxyName := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		envVar := util.GetEnvVariable(&envVars, proxyName)
		if envVar != nil && envVar.Value != "" {
			fmt.Printf("	✅ %-12s : %s\n", envVar.Name, envVar.Value)
		} else {
			fmt.Printf("	❌ %-12s : not set or empty.\n", proxyName)
		}
	}
	fmt.Println("")
}

// printOverriddenEnvVar prints the overridden envVars
func printOverriddenEnvVar(appName string, envVars []corev1.EnvVar) {
	fmt.Printf("Overridden Environment Variables Check (%s):\n----------------------------------\n", appName)
	foundOverriddenEnv := false
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar.Name, "CONFIG_JS_") {
			fmt.Printf("    	✔ %s : %s\n", envVar.Name, envVar.Value)
			foundOverriddenEnv = true
		}
	}
	if !foundOverriddenEnv {
		fmt.Print("	❌ No overridden environment variables found.\n")
	}
	fmt.Println("")
}

// isValidSTSArn is a function to validate the STS ARN format
func isValidSTSArn(arnStr *string) bool {
	if arnStr == nil {
		return false
	}

	parsedArn, err := arn.Parse(*arnStr)
	if err != nil {
		return false
	}

	if parsedArn.Service == "sts" {
		return true
	}
	return false
}

// printARNStatus is a function to print ARN validation status
func printARNStatus(listType string, arnList map[string]string) {
	foundARNString := false
	fmt.Printf("%s ARNs:\n----------------------------------\n", listType)
	for name, arn := range arnList {
		fmt.Printf("\t%s \"%s\":\n\t	ARN: %s\n\t", listType, name, arn)
		// currently validating only for AWS STS ARN, can be changed accordingly for other formats and validation
		if isValidSTSArn(&arn) {
			fmt.Printf("	Status: ✅ Valid STS ARN\n")
		} else {
			fmt.Printf("	Status: ⚠️ Invalid (Not an STS ARN)\n")
		}
		foundARNString = true
		fmt.Println("")
	}

	if !foundARNString {
		fmt.Print("	❌ No AWS ARN string found.\n")
	}
	fmt.Println("")
}

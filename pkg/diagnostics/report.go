package diagnostics

import (
	"fmt"
	"strings"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	fmt.Println("")

	// retrieving the status of proxy environment variables
	proxyStatus(coreApp, endpointApp)

	// retrieving the overridden env variables using `CONFIG_JS_` prefix
	overriddenEnvVar(coreApp, endpointApp)

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

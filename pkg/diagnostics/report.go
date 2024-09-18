package diagnostics

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
)

// RunReport runs a CLI command
func RunReport(cmd *cobra.Command, args []string) {

	// retrieving the status of proxy environment variables
	proxyStatus()

	// TODO: Add support for additional features
}

// proxyStatus returns the status of the environment variables: HTTP_PROXY, HTTPS_PROXY, and NO_PROXY
func proxyStatus() {
	log := util.Logger()

	log.Print("⏳ Retrieving proxy environment variable details...\n")
	coreApp := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	coreApp.Namespace = options.Namespace
	if !util.KubeCheck(coreApp) {
		log.Fatalf(`❌ Could not get core StatefulSet %q in Namespace %q`,
			coreApp.Name, coreApp.Namespace)
	}

	fmt.Print("\nProxy Environment Variables Check:\n----------------------------------\n")
	for _, proxyName := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		envVar := util.GetEnvVariable(&coreApp.Spec.Template.Spec.Containers[0].Env, proxyName)
		if envVar != nil && envVar.Value != "" {
			fmt.Printf("	✅ %-12s : %s\n", envVar.Name, envVar.Value)
		} else {
			fmt.Printf("	❌ %-12s : not set or empty.\n", proxyName)
		}
	}
}

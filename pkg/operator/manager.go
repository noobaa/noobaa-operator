package operator

import (
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/noobaa/noobaa-operator/v2/pkg/version"

	"github.com/spf13/cobra"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/noobaa/noobaa-operator/v2/pkg/apis"
	"github.com/noobaa/noobaa-operator/v2/pkg/controller"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

	"github.com/operator-framework/operator-sdk/pkg/leader"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
	log               = util.Logger()
)

// RunOperator is the main function of the operator but it is called from a cobra.Command
func RunOperator(cmd *cobra.Command, args []string) {

	version.RunVersion(cmd, args)

	config := util.KubeConfig()

	// Become the leader before proceeding
	err := leader.Become(util.Context(), "noobaa-operator-lock")
	if err != nil {
		log.Fatalf("Failed to become leader: %s", err)
	}

	recocileDelta := time.Duration(5 * 60) // 5 minutes

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(config, manager.Options{
		Namespace:          options.Namespace,
		MapperProvider:     util.MapperProvider, // restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		SyncPeriod:         &recocileDelta,
	})
	if err != nil {
		log.Fatalf("Failed to create manager: %s", err)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatalf("Failed AddToScheme: %s", err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatalf("Failed AddToManager: %s", err)
	}

	util.Panic(mgr.Add(manager.RunnableFunc(func(stopChan <-chan struct{}) error {
		system.RunOperatorCreate(cmd, args)
		<-stopChan
		return nil
	})))

	// // Create Service object to expose the metrics port.
	// _, err = metrics.CreateMetricsService(util.Context(), config, metricsPort)
	// if err != nil {
	// 	log.Warnf("Failed ExposeMetricsPort: %s", err)
	// }

	// Start the manager
	log.Info("Starting the Operator ...")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatalf("Manager exited non-zero: %s", err)
	}
}

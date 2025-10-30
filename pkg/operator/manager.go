package operator

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/noobaa/noobaa-operator/v5/pkg/admission"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/version"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/noobaa/noobaa-operator/v5/pkg/apis"
	"github.com/noobaa/noobaa-operator/v5/pkg/controller"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/operator-framework/operator-lib/leader"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsServer "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
	log               = util.Logger()
)

// RunOperator is the main function of the operator but it is called from a cobra.Command
func RunOperator(cmd *cobra.Command, args []string) {
	if options.DebugLevel == "warn" {
		util.InitLogger(logrus.WarnLevel)
	} else {
		util.InitLogger(logrus.DebugLevel)
	}
	version.RunVersion(cmd, args)
	// Probe address from CLI flag (defaults to :8081)
	probeAddr := os.Getenv("HEALTH_PROBE_BIND_ADDRESS")
	if probeAddr == "" {
		probeAddr = ":8081"
	}

	config := util.KubeConfig()

	// Become the leader before proceeding
	err := leader.Become(util.Context(), "noobaa-operator-lock")
	if err != nil {
		log.Fatalf("Failed to become leader: %s", err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	// mgr => namespace scoped manager
	mgr, err := manager.New(config, manager.Options{
		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			opts.DefaultNamespaces = map[string]cache.Config{
				options.Namespace: {},
			}
			return cache.New(config, opts)
		},
		MapperProvider: util.MapperProvider, // restmapper.NewDynamicRESTMapper,
		Metrics: metricsServer.Options{
			BindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		},
		HealthProbeBindAddress: probeAddr, // Serve /healthz and /readyz here
	})
	if err != nil {
		log.Fatalf("Failed to create manager: %s", err)
	}

	// cmgr => cluster scoped manager
	cmgr, err := manager.New(config, manager.Options{
		MapperProvider: util.MapperProvider, // restmapper.NewDynamicRESTMapper,
		Metrics: metricsServer.Options{
			BindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort+1),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create cluster scoped manager: %s", err)
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
	if err := controller.AddToClusterScopedManager(cmgr); err != nil {
		log.Fatalf("Failed AddToClusterScopedManager: %s", err)
	}

	// Register health and readiness endpoints on mgr
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("Failed to add health check: %s", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("Failed to add readiness check: %s", err)
	}

	util.Panic(mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		system.RunOperatorCreate(cmd, args)
		return nil
	})))

	// // Create Service object to expose the metrics port.
	// _, err = metrics.CreateMetricsService(util.Context(), config, metricsPort)
	// if err != nil {
	// 	log.Warnf("Failed ExposeMetricsPort: %s", err)
	// }

	enableAdmission, ok := os.LookupEnv("ENABLE_NOOBAA_ADMISSION")
	if ok && enableAdmission == "true" {
		// start webhook server in new routine
		go func() {
			admission.RunAdmissionServer()
		}()
	}

	// Start the manager
	log.Info("Starting the Operator ...")
	mgrs := []manager.Manager{mgr, cmgr}

	ctx := signals.SetupSignalHandler()
	var wg sync.WaitGroup

	for _, mgr := range mgrs {
		wg.Add(1)

		go func(mgr manager.Manager) {
			defer wg.Done()
			if err := mgr.Start(ctx); err != nil {
				log.Errorf("Manager exited non-zero: %s", err)
			}
		}(mgr)
	}

	wg.Wait()
}

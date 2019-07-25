package controller

import (
	"context"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/noobaa/noobaa-operator/pkg/apis"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("Version of operator-sdk: %v", sdkVersion.Version)
}

func OperatorMain() {

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.WithError(err).Errorln("Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logrus.WithError(err).Errorln("Failed to get config")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "noobaa-operator-lock")
	if err != nil {
		logrus.WithError(err).Errorln("Failed to become leader")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		logrus.WithError(err).Errorln("Failed to create manager")
		os.Exit(1)
	}

	logrus.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logrus.WithError(err).Errorln("Failed AddToScheme")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := AddToManager(mgr); err != nil {
		logrus.WithError(err).Errorln("Failed AddToManager")
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	_, err = metrics.ExposeMetricsPort(ctx, metricsPort)
	if err != nil {
		logrus.WithError(err).Warningln("Failed ExposeMetricsPort")
	}

	logrus.Info("Starting the Operator.")
	// Start the manager
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.WithError(err).Errorln("Manager exited non-zero")
		os.Exit(1)
	}
}

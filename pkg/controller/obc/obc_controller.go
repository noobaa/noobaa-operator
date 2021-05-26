package obc

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/obc"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add starts running the noobaa bucket provisioner
func Add(mgr manager.Manager) error {
	return obc.RunProvisioner(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("noobaa-operator"),
	)
}

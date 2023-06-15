package cosi

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/cosi"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add starts running the noobaa cosi driver
func Add(mgr manager.Manager) error {
	return cosi.RunProvisioner(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("noobaa-operator"),
	)
}

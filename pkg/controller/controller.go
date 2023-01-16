package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToClusterScopedManagerFuncs is a list of functions to add all Controllers which
// need access to a cluster scoped manager
var AddToClusterScopedManagerFuncs []func(manager.Manager) error

// addToManager takes a manager and a list of functions and adds them to the given manager
func addToManager(m manager.Manager, funcs []func(manager.Manager) error) error {
	for _, f := range funcs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	return addToManager(m, AddToManagerFuncs)
}

// AddToClusterScopedManager adds all Controllers which need access to a cluster scoped manager
func AddToClusterScopedManager(m manager.Manager) error {
	return addToManager(m, AddToClusterScopedManagerFuncs)
}

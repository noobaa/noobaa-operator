package controller

import (
	"github.com/noobaa/noobaa-operator/pkg/controller/backingstore"
	"github.com/noobaa/noobaa-operator/pkg/controller/bucketclass"
	"github.com/noobaa/noobaa-operator/pkg/controller/system"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type addFuncType func(manager.Manager) error

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs = []addFuncType{
	system.Add,
	backingstore.Add,
	bucketclass.Add,
}

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

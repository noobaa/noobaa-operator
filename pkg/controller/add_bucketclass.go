package controller

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/controller/bucketclass"
)

func init() {
	// AddToClusterScopedManagerFuncs is a list of functions to create controllers and add them to
	// cluster scoped manager.
	AddToClusterScopedManagerFuncs = append(AddToClusterScopedManagerFuncs, bucketclass.Add)
}

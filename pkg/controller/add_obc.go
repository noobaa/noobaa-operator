package controller

import (
	"github.com/noobaa/noobaa-operator/v2/pkg/controller/obc"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, obc.Add)
}

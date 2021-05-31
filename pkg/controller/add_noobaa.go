package controller

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/controller/noobaa"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, noobaa.Add)
}

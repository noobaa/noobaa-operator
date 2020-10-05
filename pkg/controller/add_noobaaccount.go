package controller

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/controller/noobaaaccount"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, noobaaaccount.Add)
}

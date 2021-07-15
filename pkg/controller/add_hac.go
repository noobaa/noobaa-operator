package controller

import (
	hac "github.com/noobaa/noobaa-operator/v5/pkg/controller/ha"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, hac.Add)
}

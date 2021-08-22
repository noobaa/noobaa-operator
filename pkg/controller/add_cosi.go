package controller

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/controller/cosi"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	if options.EnableCosi {
		AddToManagerFuncs = append(AddToManagerFuncs, cosi.Add)
	}
}

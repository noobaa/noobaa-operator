package controller

import (
	"os"

	hac "github.com/noobaa/noobaa-operator/v5/pkg/controller/ha"
)

func init() {
	hacEnabled := os.Getenv("HAC_ENABLED")
	if hacEnabled != "false" {
		// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
		AddToManagerFuncs = append(AddToManagerFuncs, hac.Add)
	}
}

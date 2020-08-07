package controller

import (
	"github.com/tmax-cloud/l2c-operator/pkg/controller/l2c"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, l2c.Add)
}

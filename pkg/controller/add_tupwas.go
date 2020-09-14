package controller

import (
	"github.com/tmax-cloud/l2c-operator/pkg/controller/tupwas"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, tupwas.Add)
}

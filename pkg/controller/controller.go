package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, *sonarqube.SonarQube) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, s *sonarqube.SonarQube) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, s); err != nil {
			return err
		}
	}
	return nil
}

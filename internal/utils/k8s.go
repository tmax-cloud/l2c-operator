package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/tmax-cloud/l2c-operator/internal"
	_ "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func Namespace() (string, error) {
	nsPath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if FileExists(nsPath) {
		// Running in k8s cluster
		nsBytes, err := ioutil.ReadFile(nsPath)
		if err != nil {
			return "", fmt.Errorf("could not read file %s", nsPath)
		}
		return string(nsBytes), nil
	} else {
		// Not running in k8s cluster (may be running locally)
		ns := os.Getenv("NAMESPACE")
		if ns == "" {
			ns = "default"
		}
		return ns, nil
	}
}

func ApiServiceName() string {
	svcName := os.Getenv("API_SERVICE_NAME")
	if svcName == "" {
		svcName = internal.ServiceName
	}
	return svcName
}

func CheckAndCreateObject(obj interface{}, parent metav1.Object, c client.Client, scheme *runtime.Scheme) error {
	metaObj, isMetaObj := obj.(metav1.Object)
	if !isMetaObj {
		return fmt.Errorf("given object is not a meta object")
	}

	// Get the object first to check if the object exists
	runtimeObj, isRuntimeObj := metaObj.(runtime.Object)
	if !isRuntimeObj {
		return fmt.Errorf("given object is not a runtime object")
	}
	err := c.Get(context.TODO(), types.NamespacedName{Name: metaObj.GetName(), Namespace: metaObj.GetNamespace()}, runtimeObj)
	if err != nil && errors.IsNotFound(err) {
		// Not found! create one!
		// First set ownerReference
		if err := controllerutil.SetControllerReference(parent, metaObj, scheme); err != nil {
			return fmt.Errorf("ownerRef: %s", err.Error())
		}

		// Cast to runtime object
		runtimeObj, isRuntimeObj := metaObj.(runtime.Object)
		if !isRuntimeObj {
			return fmt.Errorf("given object is not a runtime object")
		}

		// Now create
		if err := c.Create(context.TODO(), runtimeObj); err != nil {
			return fmt.Errorf("create: %s", err.Error())
		}
	} else if err != nil {
		return fmt.Errorf("get: %s", err.Error())
	}

	return nil
}

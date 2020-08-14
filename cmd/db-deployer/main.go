package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	l2cv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func main() {
	// Env. Var.s check
	doMigrate := os.Getenv("DO_MIGRATE_DB")
	if strings.ToLower(doMigrate) != "true" {
		log.Println("Skip deploying db")
		os.Exit(0)
	}

	configMapName := os.Getenv("CONFIGMAP_NAME")
	if configMapName == "" {
		log.Fatal("CONFIGMAP_NAME should be set")
	}
	log.Printf("Configmap: %s\n", configMapName)

	// Get namespace
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatal(err)
	}
	ns := string(nsBytes)
	log.Printf("Namespace: %s\n", ns)

	// Get client
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// Get Config map
	cm := &corev1.ConfigMap{}
	if err = c.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: ns}, cm); err != nil {
		log.Fatal(err)
	}

	// Get YAMLs from Config Map
	deserializer := scheme.Codecs.UniversalDeserializer()

	// Get & Create PVC
	pvcString, ok := cm.Data[l2cv1.ConfigMapKeyPvc]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.ConfigMapKeyPvc, configMapName)
	}
	obj, _, err := deserializer.Decode([]byte(pvcString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	switch pvc := obj.(type) {
	case *corev1.PersistentVolumeClaim:
		pvc.ObjectMeta.Namespace = ns
		// Check if exists
		found := &corev1.PersistentVolumeClaim{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, found); err != nil {
			if kerrors.IsNotFound(err) {
				// Create if not exists
				if err := c.Create(context.TODO(), pvc); err != nil {
					log.Fatal(err)
				}
				log.Printf("Successfully created PVC %s\n", pvc.ObjectMeta.Name)
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("PersistentVolumeClaim %s already exists\n", pvc.Name)
		}
	default:
		log.Fatalf("%s should contain PersistentVolumeClaim yaml (currently %s)\n", l2cv1.ConfigMapKeyPvc, reflect.TypeOf(obj).String())
	}

	// Get & Create Service
	svcString, ok := cm.Data[l2cv1.ConfigMapKeySvc]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.ConfigMapKeySvc, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(svcString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	switch svc := obj.(type) {
	case *corev1.Service:
		svc.ObjectMeta.Namespace = ns
		// Check if exists
		found := &corev1.Service{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found); err != nil {
			if kerrors.IsNotFound(err) {
				// Create if not exists
				if err := c.Create(context.TODO(), svc); err != nil {
					log.Fatal(err)
				}
				log.Printf("Successfully created Service %s\n", svc.ObjectMeta.Name)
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("Service %s already exists\n", svc.Name)
		}
	default:
		log.Fatalf("%s should contain Service yaml (currently %s)\n", l2cv1.ConfigMapKeySvc, reflect.TypeOf(obj).String())
	}

	// Get & Create Secret
	secretString, ok := cm.Data[l2cv1.ConfigMapKeySecret]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.ConfigMapKeySecret, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(secretString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	switch secret := obj.(type) {
	case *corev1.Secret:
		secret.ObjectMeta.Namespace = ns
		// Check if exists
		found := &corev1.Secret{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found); err != nil {
			if kerrors.IsNotFound(err) {
				// Create if not exists
				if err := c.Create(context.TODO(), secret); err != nil {
					log.Fatal(err)
				}
				log.Printf("Successfully created Secret %s\n", secret.ObjectMeta.Name)
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("Secret %s already exists\n", secret.Name)
		}
	default:
		log.Fatalf("%s should contain Secret yaml (currently %s)\n", l2cv1.ConfigMapKeySecret, reflect.TypeOf(obj).String())
	}

	// Get & Create Deployment
	deployString, ok := cm.Data[l2cv1.ConfigMapKeyDeploy]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.ConfigMapKeyDeploy, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(deployString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	var deployKey types.NamespacedName
	var deployLabels map[string]string
	switch deploy := obj.(type) {
	case *appsv1.Deployment:
		deploy.ObjectMeta.Namespace = ns
		// Check if exists
		found := &appsv1.Deployment{}
		deployKey = types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}
		deployLabels = deploy.ObjectMeta.Labels
		if err := c.Get(context.TODO(), deployKey, found); err != nil {
			if kerrors.IsNotFound(err) {
				// Create if not exists
				if err := c.Create(context.TODO(), deploy); err != nil {
					log.Fatal(err)
				}
				log.Printf("Successfully created Deployment %s\n", deploy.ObjectMeta.Name)
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("Deployment %s already exists\n", deploy.Name)
		}
	default:
		log.Fatalf("%s should contain Deployment yaml (currently %s)\n", l2cv1.ConfigMapKeyDeploy, reflect.TypeOf(obj).String())
	}

	log.Println("Successfully created all components")

	if strings.ToLower(os.Getenv("WAIT_UNTIL_RUNNING")) != "true" {
		os.Exit(0)
	}

	log.Printf("Waiting until deployment %+v gets into running state \n", deployKey)

	var labels []string
	for k, v := range deployLabels {
		labels = append(labels, k+"="+v)
	}
	label := strings.Join(labels, ",")

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	w, err := clientSet.AppsV1().Deployments(deployKey.Namespace).Watch(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		log.Fatal(err)
	}

	for event := range w.ResultChan() {
		switch dep := event.Object.(type) {
		case *appsv1.Deployment:
			if dep.Name == deployKey.Name {
				if dep.Status.ReadyReplicas > 0 {
					log.Println("Deployment is running!")
					os.Exit(0)
				} else {
					log.Println("Deployment is not ready yet")
				}
			}
		default:
			log.Printf("Object type is not Deployment (%+v)\n", event.Object)
		}
	}
}

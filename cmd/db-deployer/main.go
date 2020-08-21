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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
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
	pvcString, ok := cm.Data[l2cv1.DbConfigMapKeyPvc]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.DbConfigMapKeyPvc, configMapName)
	}
	obj, _, err := deserializer.Decode([]byte(pvcString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	pvc, isPvc := obj.(*corev1.PersistentVolumeClaim)
	if !isPvc {
		log.Fatalf("%s should contain PVC yaml (currently %s)\n", l2cv1.DbConfigMapKeyPvc, reflect.TypeOf(obj).String())
	}
	if err := utils.CheckAndCreateObject(pvc, nil, c, nil, false); err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfully created PVC %s\n", pvc.Name)

	// Get & Create Service
	svcString, ok := cm.Data[l2cv1.DbConfigMapKeySvc]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.DbConfigMapKeySvc, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(svcString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	svc, isSvc := obj.(*corev1.Service)
	if !isSvc {
		log.Fatalf("%s should contain Service yaml (currently %s)\n", l2cv1.DbConfigMapKeySvc, reflect.TypeOf(obj).String())
	}
	if err := utils.CheckAndCreateObject(svc, nil, c, nil, false); err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfully created Service %s\n", svc.Name)

	// Get & Create Secret
	secretString, ok := cm.Data[l2cv1.DbConfigMapKeySecret]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.DbConfigMapKeySecret, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(secretString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	secret, isSecret := obj.(*corev1.Secret)
	if !isSecret {
		log.Fatalf("%s should contain Secret yaml (currently %s)\n", l2cv1.DbConfigMapKeySecret, reflect.TypeOf(obj).String())
	}
	if err := utils.CheckAndCreateObject(secret, nil, c, nil, false); err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfully created Secret %s\n", secret.Name)

	// Get & Create Deployment
	deployString, ok := cm.Data[l2cv1.DbConfigMapKeyDeploy]
	if !ok {
		log.Fatalf("key %s does not exist in ConfigMap %s\n", l2cv1.DbConfigMapKeyDeploy, configMapName)
	}
	obj, _, err = deserializer.Decode([]byte(deployString), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	deploy, isDeploy := obj.(*appsv1.Deployment)
	if !isDeploy {
		log.Fatalf("%s should contain Deployment yaml (currently %s)\n", l2cv1.DbConfigMapKeyDeploy, reflect.TypeOf(obj).String())
	}
	if err := utils.CheckAndCreateObject(deploy, nil, c, nil, false); err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfully created Deployment %s\n", deploy.Name)

	deployKey := types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}
	deployLabels := deploy.ObjectMeta.Labels

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

	for {
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
}

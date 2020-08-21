package sonarqube

import (
	"fmt"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
)

const (
	DefaultAdminId = "admin"
	DefaultAdminPw = "admin"

	DefaultAnalyzerId = "analyzer"
	DefaultAnalyzerPw = "analyzer"

	DefaultResourceName = "l2c-managed-sonarqube"

	DefaultImage = "azssi/working:0.0.1" // TODO!!

	DefaultStorageClassName = "csi-cephfs-sc"
	DefaultStorageSize      = "1Gi"
)

const (
	SecretKeyAdminId = "adminId"
	SecretKeyAdminPw = "adminPw"
	SecretKeyToken   = "token"

	SecretKeyAnalyzerId    = "analyzerId"
	SecretKeyAnalyzerPw    = "analyzerPw"
	SecretKeyAnalyzerToken = "analyzerToken"
)

var log = logf.Log.WithName("sonarqube")

var depLabel = map[string]string{"app": "l2c-sonarqube"}
var label = map[string]string{"owned-by": "l2c-operator"}

type SonarQube struct {
	URL           string
	Token         string
	AnalyzerToken string

	AdminId string
	AdminPw string

	AnalyzerId string
	AnalyzerPw string

	ResourceName string
	Namespace    string

	Image string

	StorageClassName string
	StorageSize      string

	Ready     chan bool
	ClientSet *kubernetes.Clientset
}

func NewSonarQube() (*SonarQube, error) {
	// Namespace
	ns, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	// Resource name
	resourceName := os.Getenv("SONAR_RESOURCE_NAME")
	if resourceName == "" {
		resourceName = DefaultResourceName
	}

	// Image URL
	imageUrl := os.Getenv("SONAR_IMAGE")
	if imageUrl == "" {
		imageUrl = DefaultImage
	}

	// Storage settings
	storageClassName := os.Getenv("SONAR_STORAGE_CLASS_NAME")
	if storageClassName == "" {
		storageClassName = DefaultStorageClassName
	}

	storageSize := os.Getenv("SONAR_STORAGE_SIZE")
	if storageSize == "" {
		storageSize = DefaultStorageSize
	}

	// Create Client
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &SonarQube{
		ClientSet:        clientSet,
		URL:              fmt.Sprintf("http://%s.%s:9000", resourceName, ns),
		Namespace:        ns,
		ResourceName:     resourceName,
		Image:            imageUrl,
		StorageClassName: storageClassName,
		StorageSize:      storageSize,
		Ready:            make(chan bool),
	}, nil
}

func (s *SonarQube) Start() {
	// Secret - ID/PW/Token
	secretClient := s.ClientSet.CoreV1().Secrets(s.Namespace)
	secret, err := secretClient.Get(s.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating sonarqube secret...")
			if _, err := secretClient.Create(s.secret()); err != nil {
				log.Error(err, "cannot create secret")
				os.Exit(1)
			}
			if secret, err = secretClient.Get(s.ResourceName, metav1.GetOptions{}); err != nil {
				log.Error(err, "secret is not found")
				os.Exit(1)
			}
		} else {
			log.Error(err, "cannot get secret")
			os.Exit(1)
		}
	}
	adminId, exist := secret.Data[SecretKeyAdminId]
	if !exist {
		log.Error(fmt.Errorf("there is no secret key %s", SecretKeyAdminId), "")
		os.Exit(1)
	}
	adminPw, exist := secret.Data[SecretKeyAdminPw]
	if !exist {
		log.Error(fmt.Errorf("there is no secret key %s", SecretKeyAdminPw), "")
		os.Exit(1)
	}
	token, tokenExist := secret.Data[SecretKeyToken]

	s.AdminId = string(adminId)
	s.AdminPw = string(adminPw)
	if tokenExist {
		s.Token = string(token)
	}

	analyzerId, exist := secret.Data[SecretKeyAnalyzerId]
	if !exist {
		log.Error(fmt.Errorf("there is no secret key %s", SecretKeyAnalyzerId), "")
		os.Exit(1)
	}
	analyzerPw, exist := secret.Data[SecretKeyAnalyzerPw]
	if !exist {
		log.Error(fmt.Errorf("there is no secret key %s", SecretKeyAnalyzerPw), "")
		os.Exit(1)
	}
	analyzerToken, tokenExist := secret.Data[SecretKeyAnalyzerToken]

	s.AnalyzerId = string(analyzerId)
	s.AnalyzerPw = string(analyzerPw)
	if tokenExist {
		s.AnalyzerToken = string(analyzerToken)
	}

	// Get and create pvc if not exists
	pvcClient := s.ClientSet.CoreV1().PersistentVolumeClaims(s.Namespace)
	_, err = pvcClient.Get(s.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating sonarqube pvc..")
			if _, err := pvcClient.Create(s.pvc()); err != nil {
				log.Error(err, "cannot create pvc")
				os.Exit(1)
			}
		} else {
			log.Error(err, "cannot get pvc")
			os.Exit(1)
		}
	}

	// Get and create service if not exists
	svcClient := s.ClientSet.CoreV1().Services(s.Namespace)
	_, err = svcClient.Get(s.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating sonarqube service..")
			if _, err = svcClient.Create(s.service()); err != nil {
				log.Error(err, "cannot create service")
				os.Exit(1)
			}
		} else {
			log.Error(err, "cannot get service")
			os.Exit(1)
		}
	}

	// Get and create deployment if not exists
	dep := s.deployment()
	depClient := s.ClientSet.AppsV1().Deployments(s.Namespace)
	_, err = depClient.Get(s.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating sonarqube deployment..")
			if _, err = depClient.Create(dep); err != nil {
				log.Error(err, "cannot create deployment")
				os.Exit(1)
			}
		} else {
			log.Error(err, "cannot get deployment")
			os.Exit(1)
		}
	}

	// Watch deployment
	var labels []string
	for k, v := range dep.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	labelSelector := strings.Join(labels, ",")

	log.Info("Waiting until SonarQube Deployment is ready")

WatchLoop:
	for {
		w, err := s.ClientSet.AppsV1().Deployments(s.Namespace).Watch(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			log.Error(err, "")
			os.Exit(1)
		}

		for event := range w.ResultChan() {
			dep, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}

			if dep.Status.AvailableReplicas > 0 {
				break WatchLoop
			}
		}
	}

	log.Info("SonarQube deployment is ready!")

	// Update PW/Token
	if err := s.UpdateCred(); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Set default QualityGate
	if err := s.SetQualityGate(); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Register webhook
	if err := s.RegisterWebhook(); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	s.Ready <- true
}

func (s *SonarQube) UpdateCred() error {
	secretClient := s.ClientSet.CoreV1().Secrets(s.Namespace)
	// Update id/pw
	if s.AdminId == DefaultAdminId && s.AdminPw == DefaultAdminPw {
		secret, err := secretClient.Get(s.ResourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		adminIdBytes, exist := secret.Data[SecretKeyAdminId]
		if !exist {
			return fmt.Errorf("there is no secret key %s", SecretKeyAdminId)
		}
		s.AdminId = string(adminIdBytes)

		adminPwBytes, exist := secret.Data[SecretKeyAdminPw]
		if !exist {
			return fmt.Errorf("there is no secret key %s", SecretKeyAdminPw)
		}
		s.AdminPw = string(adminPwBytes)

		if s.AdminId == DefaultAdminId && s.AdminPw == DefaultAdminPw {
			newPw := utils.RandString(10)
			if err := s.ChangeAdminPassword(newPw); err != nil {
				return err
			}

			secret.Data[SecretKeyAdminPw] = []byte(newPw)
			if _, err := secretClient.Update(secret); err != nil {
				return err
			}

			s.AdminPw = newPw
		}
	}

	// Update Admin Token
	if s.Token == "" {
		secret, err := secretClient.Get(s.ResourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		token, exist := secret.Data[SecretKeyToken]
		if !exist {
			return fmt.Errorf("there is no secret key %s", SecretKeyToken)
		}

		s.Token = string(token)

		if s.Token == "" {
			newToken, err := s.GenerateToken(s.AdminId, s.AdminPw)
			if err != nil {
				return err
			}

			secret.Data[SecretKeyToken] = []byte(newToken)
			if _, err := secretClient.Update(secret); err != nil {
				return err
			}

			s.Token = newToken
		}
	}

	// Remove all permissions from default group (sonar-users)
	if err := s.RemoveAllGroupPermissions("sonar-users"); err != nil {
		return err
	}
	// Create analyzer account
	if err := s.CreateUser(s.AnalyzerId, s.AnalyzerPw); err != nil {
		return err
	}
	// Add analyzer permission to analyzer account
	if err := s.AddUserPermission(s.AnalyzerId, "scan"); err != nil {
		return err
	}

	// Update Analyzer Token
	if s.AnalyzerToken == "" {
		secret, err := secretClient.Get(s.ResourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		token, exist := secret.Data[SecretKeyAnalyzerToken]
		if !exist {
			return fmt.Errorf("there is no secret key %s", SecretKeyAnalyzerToken)
		}

		s.AnalyzerToken = string(token)

		if s.AnalyzerToken == "" {
			newToken, err := s.GenerateToken(s.AnalyzerId, s.AnalyzerPw)
			if err != nil {
				return err
			}

			secret.Data[SecretKeyAnalyzerToken] = []byte(newToken)
			if _, err := secretClient.Update(secret); err != nil {
				return err
			}

			s.AnalyzerToken = newToken
		}
	}

	return nil
}

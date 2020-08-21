package sonarqube

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	Port = 34335
)

type SonarServer struct {
	router *mux.Router

	sonar *SonarQube

	c client.Client

	clientCfg *rest.Config
	clientSet *kubernetes.Clientset
}

func NewServer(cl client.Client, s *SonarQube) *SonarServer {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	clSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	server := &SonarServer{
		c:         cl,
		clientSet: clSet,
		clientCfg: cfg,
		sonar:     s,
	}

	router := mux.NewRouter()
	router.PathPrefix("/webhook").HandlerFunc(server.WebhookHandleFunc)
	router.HandleFunc("/webhook", server.WebhookHandleFunc)

	server.router = router

	return server
}

func (s *SonarServer) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", Port)
	log.Info(fmt.Sprintf("SonarQube webhook is running on %s", addr))
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}

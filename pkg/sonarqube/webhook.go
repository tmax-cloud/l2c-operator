package sonarqube

import (
	"fmt"
	"net/http"
	"os"
)

const (
	Port = 34335
)

type Webhook struct {
	Handler *WebhookHandler
}

func NewWebhook() *Webhook {
	return &Webhook{
		Handler: &WebhookHandler{},
	}
}

func (w *Webhook) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", Port)
	log.Info(fmt.Sprintf("SonarQube webhook is running on %s", addr))
	if err := http.ListenAndServe(addr, w.Handler); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}

type WebhookHandler struct{}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//TODO
}

package sonarqube

import (
	"fmt"
	"net/http"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	WebhookName = "global-webhook"
)

func (s *SonarQube) RegisterWebhook() error {
	addr := fmt.Sprintf("http://%s:%d/webhook", utils.ApiServiceName(), Port)

	// First, get if webhook is already set correctly
	getResult := &tmaxv1.SonarWebhookResult{}
	if err := s.reqHttp(http.MethodPost, "/api/webhooks/list", nil, nil, getResult); err != nil {
		return err
	}

	for _, w := range getResult.Webhooks {
		if w.Name != WebhookName {
			continue
		}

		// Same name & Same addr -> don't need to do anything
		if w.URL == addr {
			return nil
		}

		// If same name & diff addr, update it
		if err := s.UpdateWebhook(w.Key, addr); err != nil {
			return err
		}
		return nil
	}

	// Register webhook
	data := map[string]string{
		"name": WebhookName,
		"url":  addr,
	}

	if err := s.reqHttp(http.MethodPost, "/api/webhooks/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Set SonarQube Webhook global-webhook as %s", addr))

	return nil
}

func (s *SonarQube) UpdateWebhook(key, uri string) error {
	data := map[string]string{
		"name":    WebhookName,
		"webhook": key,
		"url":     uri,
	}
	if err := s.reqHttp(http.MethodPost, "/api/webhooks/update", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Updates SonarQube Webhook global-webhook as %s", uri))

	return nil
}

package api

import (
	"github.com/crowdy/lm-cli/internal/model"
)

// WebhookAPI provides LINE webhook operations.
type WebhookAPI struct {
	Client *Client
}

func (a *WebhookAPI) Get() (*model.WebhookInfo, error) {
	var info model.WebhookInfo
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/channel/webhook/endpoint", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (a *WebhookAPI) Set(webhookURL string) error {
	body := map[string]string{"webhookEndpoint": webhookURL}
	return a.Client.Put(a.Client.BaseURL+"/v2/bot/channel/webhook/endpoint", body, nil)
}

func (a *WebhookAPI) Test() (*model.WebhookTestResponse, error) {
	var resp model.WebhookTestResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/channel/webhook/test", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

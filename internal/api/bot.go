package api

import (
	"github.com/crowdy/lm-cli/internal/model"
)

// BotAPI provides LINE bot info operations.
type BotAPI struct {
	Client *Client
}

func (a *BotAPI) GetInfo() (*model.BotInfo, error) {
	var info model.BotInfo
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/info", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (a *BotAPI) GetQuota() (*model.QuotaInfo, error) {
	var quota model.QuotaInfo
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/message/quota", &quota); err != nil {
		return nil, err
	}
	return &quota, nil
}

func (a *BotAPI) GetConsumption() (*model.ConsumptionInfo, error) {
	var consumption model.ConsumptionInfo
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/message/quota/consumption", &consumption); err != nil {
		return nil, err
	}
	return &consumption, nil
}

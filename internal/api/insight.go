package api

import (
	"net/url"

	"github.com/crowdy/lm-cli/internal/model"
)

// InsightAPI provides LINE insight operations.
type InsightAPI struct {
	Client *Client
}

func (a *InsightAPI) GetFollowers(date string) (*model.FollowerStats, error) {
	params := url.Values{}
	if date != "" {
		params.Set("date", date)
	}
	endpoint := a.Client.BaseURL + "/v2/bot/insight/followers"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	var resp model.FollowerStats
	if err := a.Client.Get(endpoint, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *InsightAPI) GetDelivery(msgType, date string) (*model.DeliveryStats, error) {
	params := url.Values{}
	params.Set("type", msgType)
	if date != "" {
		params.Set("date", date)
	}
	endpoint := a.Client.BaseURL + "/v2/bot/insight/message/delivery?" + params.Encode()
	var resp model.DeliveryStats
	if err := a.Client.Get(endpoint, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

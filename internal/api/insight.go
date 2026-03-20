package api

import (
	"fmt"

	"github.com/crowdy/lm-cli/internal/model"
)

// InsightAPI provides LINE insight operations.
type InsightAPI struct {
	Client *Client
}

func (a *InsightAPI) GetFollowers(date string) (*model.FollowerStats, error) {
	url := a.Client.BaseURL + "/v2/bot/insight/followers"
	if date != "" {
		url += "?date=" + date
	}
	var resp model.FollowerStats
	if err := a.Client.Get(url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *InsightAPI) GetDelivery(msgType, date string) (*model.DeliveryStats, error) {
	url := fmt.Sprintf("%s/v2/bot/insight/message/delivery?type=%s", a.Client.BaseURL, msgType)
	if date != "" {
		url += "&date=" + date
	}
	var resp model.DeliveryStats
	if err := a.Client.Get(url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

package api

import (
	"fmt"

	"github.com/crowdy/lm-cli/internal/model"
)

// AudienceAPI provides LINE audience group operations.
type AudienceAPI struct {
	Client *Client
}

func (a *AudienceAPI) Create(body any) (*model.CreateAudienceGroupResponse, error) {
	var resp model.CreateAudienceGroupResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/audienceGroup/upload", body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *AudienceAPI) Get(audienceGroupID int64) (*model.AudienceGroup, error) {
	var resp struct {
		AudienceGroup model.AudienceGroup `json:"audienceGroup"`
	}
	url := fmt.Sprintf("%s/v2/bot/audienceGroup/%d", a.Client.BaseURL, audienceGroupID)
	if err := a.Client.Get(url, &resp); err != nil {
		return nil, err
	}
	return &resp.AudienceGroup, nil
}

func (a *AudienceAPI) List() (*model.AudienceGroupsResponse, error) {
	var resp model.AudienceGroupsResponse
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/audienceGroup/list", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *AudienceAPI) Delete(audienceGroupID int64) error {
	url := fmt.Sprintf("%s/v2/bot/audienceGroup/%d", a.Client.BaseURL, audienceGroupID)
	return a.Client.Delete(url)
}

package api

import (
	"fmt"

	"github.com/crowdy/lm-cli/internal/model"
)

// UserAPI provides LINE user operations.
type UserAPI struct {
	Client *Client
}

func (a *UserAPI) GetProfile(userID string) (*model.UserProfile, error) {
	var profile model.UserProfile
	url := fmt.Sprintf("%s/v2/bot/profile/%s", a.Client.BaseURL, userID)
	if err := a.Client.Get(url, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (a *UserAPI) GetFollowers(limit int, start string) (*model.FollowerIDsResponse, error) {
	url := a.Client.BaseURL + "/v2/bot/followers/ids"
	if start != "" {
		url += "?start=" + start
	}
	if limit > 0 {
		if start != "" {
			url += fmt.Sprintf("&limit=%d", limit)
		} else {
			url += fmt.Sprintf("?limit=%d", limit)
		}
	}
	var resp model.FollowerIDsResponse
	if err := a.Client.Get(url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

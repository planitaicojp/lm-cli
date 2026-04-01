package api

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/crowdy/lm-cli/internal/model"
)

// UserAPI provides LINE user operations.
type UserAPI struct {
	Client *Client
}

func (a *UserAPI) GetProfile(userID string) (*model.UserProfile, error) {
	var profile model.UserProfile
	endpoint := fmt.Sprintf("%s/v2/bot/profile/%s", a.Client.BaseURL, userID)
	if err := a.Client.Get(endpoint, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (a *UserAPI) GetFollowers(limit int, start string) (*model.FollowerIDsResponse, error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	endpoint := a.Client.BaseURL + "/v2/bot/followers/ids"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	var resp model.FollowerIDsResponse
	if err := a.Client.Get(endpoint, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

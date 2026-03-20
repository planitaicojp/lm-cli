package api

import (
	"fmt"

	"github.com/crowdy/lm-cli/internal/model"
)

// GroupAPI provides LINE group operations.
type GroupAPI struct {
	Client *Client
}

func (a *GroupAPI) GetSummary(groupID string) (*model.GroupSummary, error) {
	var summary model.GroupSummary
	url := fmt.Sprintf("%s/v2/bot/group/%s/summary", a.Client.BaseURL, groupID)
	if err := a.Client.Get(url, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (a *GroupAPI) GetMembers(groupID string) (*model.GroupMembersResponse, error) {
	var resp model.GroupMembersResponse
	url := fmt.Sprintf("%s/v2/bot/group/%s/members/ids", a.Client.BaseURL, groupID)
	if err := a.Client.Get(url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *GroupAPI) Leave(groupID string) error {
	url := fmt.Sprintf("%s/v2/bot/group/%s/leave", a.Client.BaseURL, groupID)
	return a.Client.Delete(url)
}

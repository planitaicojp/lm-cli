package api

import (
	"github.com/crowdy/lm-cli/internal/model"
)

// MessageAPI provides LINE Messaging API message operations.
type MessageAPI struct {
	Client *Client
}

func (a *MessageAPI) Push(to string, messages []any) (*model.MessageResponse, error) {
	req := model.PushRequest{To: to, Messages: messages}
	var resp model.MessageResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/message/push", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *MessageAPI) Multicast(to []string, messages []any) (*model.MessageResponse, error) {
	req := model.MulticastRequest{To: to, Messages: messages}
	var resp model.MessageResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/message/multicast", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *MessageAPI) Broadcast(messages []any) (*model.MessageResponse, error) {
	req := model.BroadcastRequest{Messages: messages}
	var resp model.MessageResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/message/broadcast", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *MessageAPI) Narrowcast(req model.NarrowcastRequest) (*model.MessageResponse, error) {
	var resp model.MessageResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/message/narrowcast", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *MessageAPI) Reply(replyToken string, messages []any) (*model.MessageResponse, error) {
	req := model.ReplyRequest{ReplyToken: replyToken, Messages: messages}
	var resp model.MessageResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/message/reply", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

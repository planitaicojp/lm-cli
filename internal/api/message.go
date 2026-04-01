package api

import (
	"fmt"

	"github.com/crowdy/lm-cli/internal/model"
)

const multicastBatchSize = 500

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

// MulticastBatch splits userIDs into batches of 500 and sends each batch.
// The onBatch callback is called before each batch with (batchNumber, totalBatches).
func (a *MessageAPI) MulticastBatch(to []string, messages []any, onBatch func(batch, total int)) (*model.MessageResponse, error) {
	if len(to) <= multicastBatchSize {
		return a.Multicast(to, messages)
	}
	totalBatches := (len(to) + multicastBatchSize - 1) / multicastBatchSize
	var combined model.MessageResponse
	for i := 0; i < len(to); i += multicastBatchSize {
		end := i + multicastBatchSize
		if end > len(to) {
			end = len(to)
		}
		batchNum := i/multicastBatchSize + 1
		if onBatch != nil {
			onBatch(batchNum, totalBatches)
		}
		resp, err := a.Multicast(to[i:end], messages)
		if err != nil {
			return nil, fmt.Errorf("batch %d/%d failed: %w", batchNum, totalBatches, err)
		}
		combined.SentMessages = append(combined.SentMessages, resp.SentMessages...)
	}
	return &combined, nil
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

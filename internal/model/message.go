package model

// Message is the interface implemented by all LINE message types.
type Message interface {
	messageType() string
}

// TextMessage is a plain text message.
type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (m TextMessage) messageType() string { return "text" }

// NewTextMessage creates a TextMessage.
func NewTextMessage(text string) TextMessage {
	return TextMessage{Type: "text", Text: text}
}

// StickerMessage is a sticker message.
type StickerMessage struct {
	Type      string `json:"type"`
	PackageID string `json:"packageId"`
	StickerID string `json:"stickerId"`
}

func (m StickerMessage) messageType() string { return "sticker" }

// ImageMessage is an image message.
type ImageMessage struct {
	Type               string `json:"type"`
	OriginalContentURL string `json:"originalContentUrl"`
	PreviewImageURL    string `json:"previewImageUrl"`
}

func (m ImageMessage) messageType() string { return "image" }

// PushRequest is the request body for POST /v2/bot/message/push.
type PushRequest struct {
	To       string    `json:"to"`
	Messages []any     `json:"messages"`
}

// MulticastRequest is the request body for POST /v2/bot/message/multicast.
type MulticastRequest struct {
	To       []string  `json:"to"`
	Messages []any     `json:"messages"`
}

// BroadcastRequest is the request body for POST /v2/bot/message/broadcast.
type BroadcastRequest struct {
	Messages []any `json:"messages"`
}

// NarrowcastRequest is the request body for POST /v2/bot/message/narrowcast.
type NarrowcastRequest struct {
	Messages []any  `json:"messages"`
	Recipient any   `json:"recipient,omitempty"`
	Filter    any   `json:"filter,omitempty"`
	Limit     *int  `json:"limit,omitempty"`
}

// ReplyRequest is the request body for POST /v2/bot/message/reply.
type ReplyRequest struct {
	ReplyToken string `json:"replyToken"`
	Messages   []any  `json:"messages"`
}

// MessageResponse is the common response for message send endpoints.
type MessageResponse struct {
	SentMessages []SentMessage `json:"sentMessages"`
}

// SentMessage represents a message that was sent.
type SentMessage struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

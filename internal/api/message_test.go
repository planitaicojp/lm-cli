package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crowdy/lm-cli/internal/model"
)

func TestMessageAPI_Push(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/v2/bot/message/push" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			var req model.PushRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.To != "Uabc123" {
				t.Errorf("expected Uabc123, got %s", req.To)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.MessageResponse{
				SentMessages: []model.SentMessage{{ID: "msg1", Type: "text"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		resp, err := msgAPI.Push("Uabc123", []any{model.NewTextMessage("Hello")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.SentMessages) != 1 {
			t.Errorf("expected 1 sent message, got %d", len(resp.SentMessages))
		}
	})
}

func TestMessageAPI_Broadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/bot/message/broadcast" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.MessageResponse{
				SentMessages: []model.SentMessage{{ID: "msg1", Type: "text"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		resp, err := msgAPI.Broadcast([]any{model.NewTextMessage("Broadcast!")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.SentMessages) != 1 {
			t.Errorf("expected 1 sent message, got %d", len(resp.SentMessages))
		}
	})
}

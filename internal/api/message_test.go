package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func TestMessageAPI_Push_FlexMessage(t *testing.T) {
	t.Run("flex message structure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]any
			json.NewDecoder(r.Body).Decode(&req)
			msgs := req["messages"].([]any)
			if len(msgs) != 1 {
				t.Fatalf("expected 1 message, got %d", len(msgs))
			}
			msg := msgs[0].(map[string]any)
			if msg["type"] != "flex" {
				t.Errorf("expected type flex, got %v", msg["type"])
			}
			if msg["altText"] != "Test Alt" {
				t.Errorf("expected altText 'Test Alt', got %v", msg["altText"])
			}
			contents := msg["contents"].(map[string]any)
			if contents["type"] != "bubble" {
				t.Errorf("expected contents.type bubble, got %v", contents["type"])
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.MessageResponse{
				SentMessages: []model.SentMessage{{ID: "msg1", Type: "flex"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		flexMsg := map[string]any{
			"type":    "flex",
			"altText": "Test Alt",
			"contents": map[string]any{
				"type": "bubble",
				"body": map[string]any{
					"type":   "box",
					"layout": "vertical",
					"contents": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			},
		}

		resp, err := msgAPI.Push("Uabc123", []any{flexMsg})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.SentMessages) != 1 {
			t.Errorf("expected 1 sent message, got %d", len(resp.SentMessages))
		}
	})
}

func TestMessageAPI_MulticastBatch(t *testing.T) {
	t.Run("single_batch_under_500", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&reqCount, 1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sentMessages": []map[string]string{{"id": "msg1"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		ids := make([]string, 100)
		for i := range ids {
			ids[i] = "Utest"
		}

		resp, err := msgAPI.MulticastBatch(ids, []any{map[string]string{"type": "text", "text": "hi"}}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if atomic.LoadInt32(&reqCount) != 1 {
			t.Errorf("expected 1 request, got %d", reqCount)
		}
		if len(resp.SentMessages) != 1 {
			t.Errorf("expected 1 sent message, got %d", len(resp.SentMessages))
		}
	})

	t.Run("splits_at_501_users", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&reqCount, 1)
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			to := body["to"].([]any)
			if len(to) > 500 {
				t.Errorf("batch size %d exceeds 500", len(to))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sentMessages": []map[string]string{{"id": "msg1"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		ids := make([]string, 501)
		for i := range ids {
			ids[i] = "Utest"
		}

		var batchCalls int
		resp, err := msgAPI.MulticastBatch(ids, []any{map[string]string{"type": "text", "text": "hi"}}, func(batch, total int) {
			batchCalls++
			if total != 2 {
				t.Errorf("expected 2 total batches, got %d", total)
			}
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if atomic.LoadInt32(&reqCount) != 2 {
			t.Errorf("expected 2 requests, got %d", reqCount)
		}
		if batchCalls != 2 {
			t.Errorf("expected 2 batch callbacks, got %d", batchCalls)
		}
		if len(resp.SentMessages) != 2 {
			t.Errorf("expected 2 sent messages, got %d", len(resp.SentMessages))
		}
	})

	t.Run("exactly_500_single_batch", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&reqCount, 1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sentMessages": []map[string]string{{"id": "msg1"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		ids := make([]string, 500)
		for i := range ids {
			ids[i] = "Utest"
		}

		var batchCalls int
		_, err := msgAPI.MulticastBatch(ids, []any{map[string]string{"type": "text", "text": "hi"}}, func(batch, total int) {
			batchCalls++
			if total != 1 {
				t.Errorf("expected 1 total batch, got %d", total)
			}
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if atomic.LoadInt32(&reqCount) != 1 {
			t.Errorf("expected 1 request, got %d", reqCount)
		}
		if batchCalls != 1 {
			t.Errorf("expected 1 batch callback, got %d", batchCalls)
		}
	})

	t.Run("1500_users_three_batches", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&reqCount, 1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sentMessages": []map[string]string{{"id": "msg1"}},
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		msgAPI := &MessageAPI{Client: c}

		ids := make([]string, 1500)
		for i := range ids {
			ids[i] = "Utest"
		}

		var batchCalls int
		resp, err := msgAPI.MulticastBatch(ids, []any{map[string]string{"type": "text", "text": "hi"}}, func(batch, total int) {
			batchCalls++
			if total != 3 {
				t.Errorf("expected 3 total batches, got %d", total)
			}
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if atomic.LoadInt32(&reqCount) != 3 {
			t.Errorf("expected 3 requests, got %d", reqCount)
		}
		if batchCalls != 3 {
			t.Errorf("expected 3 batch callbacks, got %d", batchCalls)
		}
		if len(resp.SentMessages) != 3 {
			t.Errorf("expected 3 sent messages, got %d", len(resp.SentMessages))
		}
	})
}

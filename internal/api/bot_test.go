package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crowdy/lm-cli/internal/model"
)

func TestBotAPI_GetInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.BotInfo{
			UserID:      "U12345",
			BasicID:     "@linebot",
			DisplayName: "My Bot",
			ChatMode:    "bot",
		})
	}))
	defer srv.Close()

	c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
	botAPI := &BotAPI{Client: c}

	info, err := botAPI.GetInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UserID != "U12345" {
		t.Errorf("expected U12345, got %s", info.UserID)
	}
	if info.DisplayName != "My Bot" {
		t.Errorf("expected My Bot, got %s", info.DisplayName)
	}
}

func TestBotAPI_GetQuota(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.QuotaInfo{Type: "limited", Value: 500})
	}))
	defer srv.Close()

	c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
	botAPI := &BotAPI{Client: c}

	quota, err := botAPI.GetQuota()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quota.Value != 500 {
		t.Errorf("expected 500, got %d", quota.Value)
	}
}

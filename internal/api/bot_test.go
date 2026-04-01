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

func TestBotAPI_GetConsumption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.ConsumptionInfo{TotalUsage: 42})
	}))
	defer srv.Close()

	c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
	botAPI := &BotAPI{Client: c}

	consumption, err := botAPI.GetConsumption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumption.TotalUsage != 42 {
		t.Errorf("expected 42, got %d", consumption.TotalUsage)
	}
}

func TestBotUsageCalculation(t *testing.T) {
	tests := []struct {
		name      string
		quota     int
		used      int
		wantRem   int
		wantPct   float64
	}{
		{"normal", 500, 42, 458, 8.4},
		{"zero quota", 0, 0, 0, 0},
		{"over quota", 500, 600, 0, 120.0},
		{"full", 200, 200, 0, 100.0},
		{"one third", 300, 100, 200, 33.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := tt.quota - tt.used
			if remaining < 0 {
				remaining = 0
			}
			if remaining != tt.wantRem {
				t.Errorf("remaining: got %d, want %d", remaining, tt.wantRem)
			}

			var usagePct float64
			if tt.quota > 0 {
				usagePct = float64(int(float64(tt.used)/float64(tt.quota)*1000+0.5)) / 10
			}
			if usagePct != tt.wantPct {
				t.Errorf("usagePct: got %v, want %v", usagePct, tt.wantPct)
			}
		})
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

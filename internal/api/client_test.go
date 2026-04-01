package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"userId": "U123"})
		}))
		defer srv.Close()

		c := &Client{
			HTTP:    &http.Client{},
			Token:   "test-token",
			BaseURL: srv.URL,
		}

		var result map[string]string
		if err := c.Get(srv.URL+"/v2/bot/info", &result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["userId"] != "U123" {
			t.Errorf("expected U123, got %s", result["userId"])
		}
	})

	t.Run("401_returns_auth_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "bad", BaseURL: srv.URL}
		err := c.Get(srv.URL+"/v2/bot/info", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParseAPIError_LINEFormat(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantMsg    string
	}{
		{
			name:       "line_error_with_details",
			statusCode: 400,
			body:       `{"message":"The request body has 1 error(s)","details":[{"message":"May not be empty","property":"messages[0].text"}]}`,
			wantMsg:    "The request body has 1 error(s)",
		},
		{
			name:       "line_error_no_details",
			statusCode: 400,
			body:       `{"message":"Invalid channel access token"}`,
			wantMsg:    "Invalid channel access token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
			err := c.Get(srv.URL+"/test", nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			// Error message should contain the LINE API message
			if tt.wantMsg != "" {
				found := false
				// Use type assertion to check the message
				type msgErr interface{ Error() string }
				if e, ok := err.(msgErr); ok {
					if contains(e.Error(), tt.wantMsg) {
						found = true
					}
				}
				if !found {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewClient_HTTPSEnforcement(t *testing.T) {
	t.Run("rejects_http_endpoint", func(t *testing.T) {
		t.Setenv("LM_ENDPOINT", "http://evil.example.com")
		t.Setenv("LM_ALLOW_HTTP", "")
		_, err := NewClient("tok")
		if err == nil {
			t.Fatal("expected error for http endpoint, got nil")
		}
	})

	t.Run("allows_https_endpoint", func(t *testing.T) {
		t.Setenv("LM_ENDPOINT", "https://api.example.com")
		c, err := NewClient("tok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.BaseURL != "https://api.example.com" {
			t.Errorf("expected https://api.example.com, got %s", c.BaseURL)
		}
	})

	t.Run("allows_http_with_LM_ALLOW_HTTP", func(t *testing.T) {
		t.Setenv("LM_ENDPOINT", "http://localhost:8080")
		t.Setenv("LM_ALLOW_HTTP", "1")
		c, err := NewClient("tok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.BaseURL != "http://localhost:8080" {
			t.Errorf("expected http://localhost:8080, got %s", c.BaseURL)
		}
	})
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name   string
		header string
		wantGt bool
	}{
		{"empty", "", false},
		{"integer_seconds", "60", true},
		{"invalid", "abc", false},
		{"http_date_future", "Sun, 01 Jan 2034 00:00:00 GMT", true},
		{"http_date_past", "Mon, 01 Jan 2001 00:00:00 GMT", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := parseRetryAfter(tt.header)
			if tt.wantGt && d <= 0 {
				t.Errorf("expected positive duration for %q, got %v", tt.header, d)
			}
			if !tt.wantGt && d != 0 {
				t.Errorf("expected zero duration for %q, got %v", tt.header, d)
			}
		})
	}
}

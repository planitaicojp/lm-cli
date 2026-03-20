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

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRichMenuAPI_CreateAlias(t *testing.T) {
	t.Run("succeeds_with_empty_response_body", func(t *testing.T) {
		var gotMethod string
		var gotBody map[string]string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			// LINE API returns empty body for alias creation
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &RichMenuAPI{Client: c}

		body := map[string]string{
			"richMenuAliasId": "alias-1",
			"richMenuId":      "richmenu-123",
		}
		err := api.CreateAlias(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != http.MethodPost {
			t.Errorf("expected POST, got %s", gotMethod)
		}
		if gotBody["richMenuAliasId"] != "alias-1" {
			t.Errorf("expected alias-1, got %s", gotBody["richMenuAliasId"])
		}
	})

	t.Run("returns_error_on_400", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "alias already exists"})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &RichMenuAPI{Client: c}

		err := api.CreateAlias(map[string]string{"richMenuAliasId": "alias-1"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

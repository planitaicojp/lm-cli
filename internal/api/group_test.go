package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupAPI_Leave(t *testing.T) {
	t.Run("uses_POST_method", func(t *testing.T) {
		var gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &GroupAPI{Client: c}

		if err := api.Leave("C123"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != http.MethodPost {
			t.Errorf("expected POST, got %s", gotMethod)
		}
	})

	t.Run("correct_path", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &GroupAPI{Client: c}

		if err := api.Leave("Cabcdef"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/v2/bot/group/Cabcdef/leave"
		if gotPath != expected {
			t.Errorf("expected path %s, got %s", expected, gotPath)
		}
	})
}

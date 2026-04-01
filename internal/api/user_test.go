package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/crowdy/lm-cli/internal/model"
)

func TestUserAPI_GetFollowers(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.FollowerIDsResponse{
				UserIDs: []string{"U001", "U002"},
				Next:    "",
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &UserAPI{Client: c}

		resp, err := api.GetFollowers(0, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.UserIDs) != 2 {
			t.Errorf("expected 2 IDs, got %d", len(resp.UserIDs))
		}
		if resp.Next != "" {
			t.Errorf("expected empty next, got %s", resp.Next)
		}
	})

	t.Run("with pagination token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := r.URL.Query().Get("start")
			if start != "cursor1" {
				t.Errorf("expected start=cursor1, got %s", start)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.FollowerIDsResponse{
				UserIDs: []string{"U003"},
				Next:    "",
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &UserAPI{Client: c}

		resp, err := api.GetFollowers(0, "cursor1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.UserIDs) != 1 {
			t.Errorf("expected 1 ID, got %d", len(resp.UserIDs))
		}
	})

	t.Run("multi page auto-pagination", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := atomic.AddInt32(&reqCount, 1)
			next := ""
			if page < 3 {
				next = "cursor_next"
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.FollowerIDsResponse{
				UserIDs: []string{"U" + r.URL.Query().Get("start")},
				Next:    next,
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &UserAPI{Client: c}

		// Simulate --all pagination
		var allIDs []string
		cursor := ""
		for {
			resp, err := api.GetFollowers(0, cursor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			allIDs = append(allIDs, resp.UserIDs...)
			if resp.Next == "" {
				break
			}
			cursor = resp.Next
		}

		if len(allIDs) != 3 {
			t.Errorf("expected 3 IDs, got %d", len(allIDs))
		}
		if atomic.LoadInt32(&reqCount) != 3 {
			t.Errorf("expected 3 requests, got %d", reqCount)
		}
	})
}

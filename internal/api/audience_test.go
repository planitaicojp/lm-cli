package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/crowdy/lm-cli/internal/model"
)

func TestAudienceAPI_List(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.AudienceGroupsResponse{
				AudienceGroups: []model.AudienceGroup{
					{AudienceGroupID: 1, Description: "test"},
				},
				HasNextPage: false,
				TotalCount:  1,
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &AudienceAPI{Client: c}

		resp, err := api.List(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.AudienceGroups) != 1 {
			t.Errorf("expected 1 group, got %d", len(resp.AudienceGroups))
		}
	})

	t.Run("page parameter", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			if page != "2" {
				t.Errorf("expected page=2, got %s", page)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.AudienceGroupsResponse{
				AudienceGroups: []model.AudienceGroup{
					{AudienceGroupID: 2, Description: "page2"},
				},
				HasNextPage: false,
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &AudienceAPI{Client: c}

		resp, err := api.List(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.AudienceGroups[0].AudienceGroupID != 2 {
			t.Errorf("expected group ID 2, got %d", resp.AudienceGroups[0].AudienceGroupID)
		}
	})

	t.Run("multi page pagination", func(t *testing.T) {
		var reqCount int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := atomic.AddInt32(&reqCount, 1)
			hasNext := page < 3
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.AudienceGroupsResponse{
				AudienceGroups: []model.AudienceGroup{
					{AudienceGroupID: int64(page), Description: "group"},
				},
				HasNextPage: hasNext,
				TotalCount:  3,
			})
		}))
		defer srv.Close()

		c := &Client{HTTP: &http.Client{}, Token: "tok", BaseURL: srv.URL}
		api := &AudienceAPI{Client: c}

		// Simulate --all pagination
		var allGroups []model.AudienceGroup
		for p := 1; ; p++ {
			resp, err := api.List(p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			allGroups = append(allGroups, resp.AudienceGroups...)
			if !resp.HasNextPage {
				break
			}
		}

		if len(allGroups) != 3 {
			t.Errorf("expected 3 groups, got %d", len(allGroups))
		}
		if atomic.LoadInt32(&reqCount) != 3 {
			t.Errorf("expected 3 requests, got %d", reqCount)
		}
	})
}

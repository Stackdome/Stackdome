package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

func TestConfigHandler_Get(t *testing.T) {
	cases := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewConfigHandler(ConfigHandlerSpec{GitHubOAuthEnabled: tc.enabled})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
			rec := httptest.NewRecorder()

			h.Get(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			var body openapi.AppConfigResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body.GithubOauth == nil {
				t.Fatal("expected github_oauth to be present")
			}
			if *body.GithubOauth != tc.enabled {
				t.Fatalf("expected github_oauth=%v, got %v", tc.enabled, *body.GithubOauth)
			}
		})
	}
}

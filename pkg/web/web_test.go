package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_RootServesIndex(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "<title>Stackdome</title>") {
		t.Errorf("body did not contain <title>Stackdome</title>: %q", rr.Body.String())
	}
	if got := rr.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control: got %q, want %q", got, "no-cache")
	}
}

func TestHandler_UnknownPathFallsBackToIndex(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stacks/anything/here", nil)

	Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "<title>Stackdome</title>") {
		t.Errorf("SPA fallback didn't return index.html: %q", rr.Body.String())
	}
}

// TestHandler_HashedAssetsImmutable runs only when a real Vite build is
// present (i.e. dist/assets/ has files). On a fresh checkout it skips.
func TestHandler_HashedAssetsImmutable(t *testing.T) {
	sub, _ := fs.Sub(distFS, "dist")
	entries, err := fs.ReadDir(sub, "assets")
	if err != nil || len(entries) == 0 {
		t.Skip("no built assets in dist/assets/ — skipping (run vite build first)")
	}

	assetPath := "/assets/" + entries[0].Name()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, assetPath, nil)

	Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	want := "public, max-age=31536000, immutable"
	if got := rr.Header().Get("Cache-Control"); got != want {
		t.Errorf("Cache-Control: got %q, want %q", got, want)
	}
}

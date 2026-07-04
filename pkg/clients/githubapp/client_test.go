package githubapp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func testAppCredentials(t *testing.T) (*AppCredentials, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return &AppCredentials{
		AppID: 4242,
		Slug:  "stackdome-test",
		PEM:   string(pemData),
	}, key
}

// requireAppJWT decodes and verifies the Authorization app JWT, asserting the
// claim shape GitHub expects.
func requireAppJWT(t *testing.T, r *http.Request, key *rsa.PrivateKey, appID int64) {
	t.Helper()
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		t.Fatalf("expected bearer auth, got %q", header)
	}
	raw := strings.TrimPrefix(header, "Bearer ")

	claims := jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return &key.PublicKey, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("invalid app JWT: %v", err)
	}
	if claims.Issuer != strconv.FormatInt(appID, 10) {
		t.Fatalf("expected issuer %d, got %q", appID, claims.Issuer)
	}
	now := time.Now().Unix()
	if claims.IssuedAt > now {
		t.Fatalf("iat must be in the past, got %d", claims.IssuedAt)
	}
	if claims.IssuedAt < now-120 {
		t.Fatalf("iat too far in the past: %d", claims.IssuedAt)
	}
	if claims.ExpiresAt <= now || claims.ExpiresAt > now+10*60 {
		t.Fatalf("exp out of the expected ~9m window: %d", claims.ExpiresAt)
	}
}

func TestConvertManifestCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app-manifests/tmp-code/conversions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":             4242,
			"slug":           "stackdome-test",
			"pem":            "PEM-DATA",
			"webhook_secret": "hook-secret",
			"client_id":      "cid",
			"client_secret":  "csecret",
		})
	}))
	defer srv.Close()

	client := NewClient(ClientSpec{BaseURL: srv.URL})
	creds, err := client.ConvertManifestCode(context.Background(), "tmp-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AppID != 4242 || creds.Slug != "stackdome-test" || creds.PEM != "PEM-DATA" ||
		creds.WebhookSecret != "hook-secret" || creds.ClientID != "cid" || creds.ClientSecret != "csecret" {
		t.Fatalf("unexpected credentials %+v", creds)
	}
}

func TestListInstallationsPaginatesWithAppJWT(t *testing.T) {
	creds, key := testAppCredentials(t)

	pageOne := make([]map[string]any, 0, 100)
	for i := 0; i < 100; i++ {
		pageOne = append(pageOne, map[string]any{
			"id":                   i + 1,
			"account":              map[string]any{"login": fmt.Sprintf("acct-%d", i+1), "type": "Organization"},
			"repository_selection": "all",
		})
	}
	pageTwo := []map[string]any{
		{"id": 101, "account": map[string]any{"login": "acct-101", "type": "User"}, "repository_selection": "selected"},
	}

	var serverURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/installations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		requireAppJWT(t, r, key, creds.AppID)
		// go-github discovers the next page from the Link header rather than
		// from page-fullness, and omits page= on the first request.
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/app/installations?page=2&per_page=100>; rel="next"`, serverURL))
			_ = json.NewEncoder(w).Encode(pageOne)
		case "2":
			_ = json.NewEncoder(w).Encode(pageTwo)
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	serverURL = srv.URL
	defer srv.Close()

	client := NewClient(ClientSpec{BaseURL: srv.URL})
	installations, err := client.ListInstallations(context.Background(), creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(installations) != 101 {
		t.Fatalf("expected 101 installations across pages, got %d", len(installations))
	}
	last := installations[100]
	if last.ID != 101 || last.AccountLogin != "acct-101" || last.AccountType != "User" || last.RepositorySelection != "selected" {
		t.Fatalf("unexpected last installation %+v", last)
	}
}

func TestMintInstallationToken(t *testing.T) {
	creds, key := testAppCredentials(t)
	expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app/installations/77/access_tokens" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		requireAppJWT(t, r, key, creds.AppID)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "ghs_test_token",
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}))
	defer srv.Close()

	client := NewClient(ClientSpec{BaseURL: srv.URL})
	token, err := client.MintInstallationToken(context.Background(), creds, 77)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Value != "ghs_test_token" || !token.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected token %+v", token)
	}
}

func TestListInstallationReposFiltersByQuery(t *testing.T) {
	creds, _ := testAppCredentials(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/app/installations/"):
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token":      "ghs_repo_token",
				"expires_at": time.Now().Add(time.Hour).Format(time.RFC3339),
			})
		case r.URL.Path == "/installation/repositories":
			if got := r.Header.Get("Authorization"); got != "token ghs_repo_token" {
				t.Fatalf("expected installation token auth, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total_count": 2,
				"repositories": []map[string]any{
					{"full_name": "acme/api", "clone_url": "https://github.com/acme/api.git", "default_branch": "main", "private": true, "owner": map[string]any{"login": "acme"}},
					{"full_name": "acme/web", "clone_url": "https://github.com/acme/web.git", "default_branch": "main", "private": false, "owner": map[string]any{"login": "acme"}},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClient(ClientSpec{BaseURL: srv.URL})
	page, err := client.ListInstallationRepos(context.Background(), creds, 77, "api", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Repos) != 1 || page.Repos[0].FullName != "acme/api" || !page.Repos[0].Private {
		t.Fatalf("unexpected repos %+v", page.Repos)
	}
	if page.TotalCount != 2 || page.HasNext {
		t.Fatalf("unexpected paging %+v", page)
	}
}

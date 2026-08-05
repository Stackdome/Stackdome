package config

import "testing"

func TestGitHubRedirectURIDerivedFromExternalURL(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com")
	t.Setenv("GITHUB_REDIRECT_URI", "")

	c := NewApplicationConfig()
	c.LoadEnvVariables()

	want := "https://hub.example.com/auth/github/callback"
	if c.GitHubOAuth.RedirectURI != want {
		t.Fatalf("expected derived redirect %q, got %q", want, c.GitHubOAuth.RedirectURI)
	}
}

func TestGitHubRedirectURITrimsTrailingSlash(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com/")
	t.Setenv("GITHUB_REDIRECT_URI", "")

	c := NewApplicationConfig()
	c.LoadEnvVariables()

	want := "https://hub.example.com/auth/github/callback"
	if c.GitHubOAuth.RedirectURI != want {
		t.Fatalf("expected no double slash, got %q", c.GitHubOAuth.RedirectURI)
	}
}

func TestGitHubRedirectURIExplicitOverrideWins(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com")
	t.Setenv("GITHUB_REDIRECT_URI", "https://custom.example.com/callback")

	c := NewApplicationConfig()
	c.LoadEnvVariables()

	if c.GitHubOAuth.RedirectURI != "https://custom.example.com/callback" {
		t.Fatalf("explicit override should win, got %q", c.GitHubOAuth.RedirectURI)
	}
}

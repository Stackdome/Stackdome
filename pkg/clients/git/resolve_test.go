package git

import "testing"

func TestRepoHost(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{url: "https://github.com/acme/api", want: "github.com"},
		{url: "https://github.com/acme/api.git", want: "github.com"},
		{url: "https://gitlab.example.com:8443/acme/api.git", want: "gitlab.example.com"},
		{url: "ssh://git@bitbucket.example.com:7999/acme/api.git", want: "bitbucket.example.com"},
		{url: "git@github.com:acme/api.git", want: "github.com"},
		{url: "git@gitea.internal.example.com:acme/api.git", want: "gitea.internal.example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			if got := RepoHost(tc.url); got != tc.want {
				t.Fatalf("RepoHost(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}

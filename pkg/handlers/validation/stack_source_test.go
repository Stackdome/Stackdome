package validation

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
)

func gitSourceSpec(mutate func(*openapi.GitSource)) *openapi.SourceSpec {
	git := openapi.NewGitSource("https://github.com/acme/api")
	if mutate != nil {
		mutate(git)
	}
	return &openapi.SourceSpec{Git: git}
}

func TestValidateSourceRequiresExactlyOneVariant(t *testing.T) {
	if err := validateSource(nil); err == nil {
		t.Fatal("expected missing source to be rejected")
	}
	if err := validateSource(&openapi.SourceSpec{}); err == nil {
		t.Fatal("expected empty source to be rejected")
	}
	both := &openapi.SourceSpec{
		Git:   openapi.NewGitSource("https://github.com/acme/api"),
		Image: openapi.NewImageSource("redis:7"),
	}
	if err := validateSource(both); err == nil {
		t.Fatal("expected multiple variants to be rejected")
	}
	if err := validateSource(&openapi.SourceSpec{Image: openapi.NewImageSource("redis:7")}); err != nil {
		t.Fatalf("expected single variant to validate, got %v", err)
	}
}

func TestValidateGitSourceBranchTagExclusive(t *testing.T) {
	spec := gitSourceSpec(func(g *openapi.GitSource) {
		g.SetBranch("main")
		g.SetTag("v1.0.0")
	})
	if err := validateSource(spec); err == nil {
		t.Fatal("expected branch+tag to be rejected")
	}
}

func TestValidateGitSourceCommitRequiresBranchOrTag(t *testing.T) {
	spec := gitSourceSpec(func(g *openapi.GitSource) {
		g.SetCommit("abcdef1")
	})
	if err := validateSource(spec); err == nil {
		t.Fatal("expected bare commit to be rejected")
	}
	spec = gitSourceSpec(func(g *openapi.GitSource) {
		g.SetBranch("main")
		g.SetCommit("abcdef1")
	})
	if err := validateSource(spec); err != nil {
		t.Fatalf("expected commit with branch to validate, got %v", err)
	}
}

func TestValidateGitSourceCredentialExclusivity(t *testing.T) {
	spec := gitSourceSpec(func(g *openapi.GitSource) {
		g.SetCredentials(*openapi.NewInlineCredentials("alice", "s3cret"))
		g.SetIntegrationId("int-1")
	})
	if err := validateSource(spec); err == nil {
		t.Fatal("expected credentials+integration_id to be rejected")
	}

	push := openapi.NewPushTarget("ghcr.io/acme/api")
	push.SetCredentials(*openapi.NewInlineCredentials("alice", "s3cret"))
	push.SetRegistryCredentialsId("rc-1")
	spec = gitSourceSpec(func(g *openapi.GitSource) {
		g.SetPush(*push)
	})
	if err := validateSource(spec); err == nil {
		t.Fatal("expected push credentials+registry_credentials_id to be rejected")
	}

	img := openapi.NewImageSource("redis:7")
	img.SetCredentials(*openapi.NewInlineCredentials("alice", "s3cret"))
	img.SetRegistryCredentialsId("rc-1")
	if err := validateSource(&openapi.SourceSpec{Image: img}); err == nil {
		t.Fatal("expected image credentials+registry_credentials_id to be rejected")
	}
}

func TestValidateInlineCredentialsRequireUsernameAndPassword(t *testing.T) {
	spec := gitSourceSpec(func(g *openapi.GitSource) {
		g.SetCredentials(openapi.InlineCredentials{Username: "alice"})
	})
	if err := validateSource(spec); err == nil {
		t.Fatal("expected incomplete inline credentials to be rejected")
	}
}

func TestValidateVolumeSourceRequiresIdentifier(t *testing.T) {
	if err := validateSource(&openapi.SourceSpec{Volume: openapi.NewVolumeBuildSource()}); err == nil {
		t.Fatal("expected volume source without id or name to be rejected")
	}
	vol := openapi.NewVolumeBuildSource()
	vol.SetVolumeName("src")
	if err := validateSource(&openapi.SourceSpec{Volume: vol}); err != nil {
		t.Fatalf("expected volume source with name to validate, got %v", err)
	}
}

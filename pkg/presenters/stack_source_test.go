package presenters_test

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

func convertSingleResource(t *testing.T, source *openapi.SourceSpec) *models.StackResource {
	t.Helper()
	spec := openapi.NewStackSpec()
	spec.SetStackResources([]openapi.StackResource{{Name: "web", Source: source}})
	converted := presenters.ConvertStack(openapi.NewStack("demo", *spec))
	if len(converted.StackResources) != 1 {
		t.Fatalf("expected one resource, got %d", len(converted.StackResources))
	}
	return converted.StackResources[0]
}

func TestConvertSourceGitWithPushRoundTrip(t *testing.T) {
	git := openapi.NewGitSource("https://github.com/acme/api")
	git.SetBranch("main")
	git.SetCommit("abcdef1234")
	git.SetDockerfilePath("build/Dockerfile")
	git.SetBuildContext("./svc")
	push := openapi.NewPushTarget("ghcr.io/acme/api")
	push.SetRegistryCredentialsId("rc-1")
	git.SetPush(*push)

	resource := convertSingleResource(t, &openapi.SourceSpec{Git: git})

	bc := resource.BuildConfig
	if bc == nil || bc.SourceContext.Git == nil {
		t.Fatal("expected git build config")
	}
	if bc.SourceContext.Git.RepoURL != "https://github.com/acme/api" {
		t.Fatalf("unexpected repo url %q", bc.SourceContext.Git.RepoURL)
	}
	if bc.SourceRevision.Git == nil || bc.SourceRevision.Git.Branch != "main" || bc.SourceRevision.Git.Commit != "abcdef1234" {
		t.Fatalf("unexpected revision %+v", bc.SourceRevision.Git)
	}
	if bc.DockerfilePath != "build/Dockerfile" || bc.ContextPathWithinSource != "./svc" {
		t.Fatalf("unexpected paths %q %q", bc.DockerfilePath, bc.ContextPathWithinSource)
	}
	if bc.BuildImageRepository.UseInClusterRegistry || bc.BuildImageRepository.ExternalImageRef != "ghcr.io/acme/api" {
		t.Fatalf("unexpected push repository %+v", bc.BuildImageRepository)
	}
	if bc.PushRegistryCredentialID != "rc-1" {
		t.Fatalf("expected push registry credential id, got %q", bc.PushRegistryCredentialID)
	}

	presented := presenters.PresentStackResource(resource)
	if presented.Source == nil || presented.Source.Git == nil {
		t.Fatal("expected presented git source")
	}
	pg := presented.Source.Git
	if pg.RepoUrl != "https://github.com/acme/api" || pg.GetBranch() != "main" || pg.GetCommit() != "abcdef1234" {
		t.Fatalf("unexpected presented git %+v", pg)
	}
	if pg.Push == nil || pg.Push.Repository != "ghcr.io/acme/api" || pg.Push.GetRegistryCredentialsId() != "rc-1" {
		t.Fatalf("unexpected presented push %+v", pg.Push)
	}
}

func TestConvertSourceGitWithoutPushUsesInternalRegistry(t *testing.T) {
	git := openapi.NewGitSource("https://github.com/acme/api")
	git.SetBranch("main")

	resource := convertSingleResource(t, &openapi.SourceSpec{Git: git})
	bc := resource.BuildConfig
	if !bc.BuildImageRepository.UseInClusterRegistry {
		t.Fatal("expected in-cluster registry when push omitted")
	}
	if bc.DockerfilePath != models.DefaultDockerfilePath || bc.ContextPathWithinSource != models.DefaultBuildContext {
		t.Fatalf("expected defaults, got %q %q", bc.DockerfilePath, bc.ContextPathWithinSource)
	}

	presented := presenters.PresentStackResource(resource)
	if presented.Source.Git.Push != nil {
		t.Fatal("expected no push target for in-cluster builds")
	}
}

func TestConvertSourceImageRoundTrip(t *testing.T) {
	img := openapi.NewImageSource("redis:7")
	img.SetRegistryCredentialsId("rc-2")

	resource := convertSingleResource(t, &openapi.SourceSpec{Image: img})
	if resource.ImageConfig == nil || resource.ImageConfig.Image != "redis:7" {
		t.Fatalf("unexpected image config %+v", resource.ImageConfig)
	}
	if resource.ImageConfig.RegistryCredentialID != "rc-2" {
		t.Fatalf("expected registry credential id, got %q", resource.ImageConfig.RegistryCredentialID)
	}

	presented := presenters.PresentStackResource(resource)
	if presented.Source == nil || presented.Source.Image == nil {
		t.Fatal("expected presented image source")
	}
	if presented.Source.Image.Ref != "redis:7" || presented.Source.Image.GetRegistryCredentialsId() != "rc-2" {
		t.Fatalf("unexpected presented image %+v", presented.Source.Image)
	}
}

func TestConvertSourceVolumeRoundTrip(t *testing.T) {
	vol := openapi.NewVolumeBuildSource()
	vol.SetVolumeName("src")
	vol.SetCurrentVolumeHash("hash-1")
	vol.SetDockerfilePath("Dockerfile.dev")

	resource := convertSingleResource(t, &openapi.SourceSpec{Volume: vol})
	bc := resource.BuildConfig
	if bc == nil || bc.SourceContext.Volume == nil || bc.SourceContext.Volume.SourceVolumeName != "src" {
		t.Fatalf("unexpected volume build config %+v", bc)
	}
	if bc.SourceRevision.Volume == nil || bc.SourceRevision.Volume.CurrentVolumeHash != "hash-1" {
		t.Fatalf("unexpected volume revision %+v", bc.SourceRevision.Volume)
	}
	if !bc.BuildImageRepository.UseInClusterRegistry {
		t.Fatal("expected volume builds to target the internal registry")
	}

	presented := presenters.PresentStackResource(resource)
	if presented.Source == nil || presented.Source.Volume == nil {
		t.Fatal("expected presented volume source")
	}
	if presented.Source.Volume.GetVolumeName() != "src" || presented.Source.Volume.GetCurrentVolumeHash() != "hash-1" {
		t.Fatalf("unexpected presented volume %+v", presented.Source.Volume)
	}
	if presented.Source.Volume.GetDockerfilePath() != "Dockerfile.dev" {
		t.Fatalf("unexpected dockerfile %q", presented.Source.Volume.GetDockerfilePath())
	}
}

func TestPresentPreviewGitRepositoryMapsPublicFields(t *testing.T) {
	config := &models.StackPreviewConfig{
		ID:   "cfg-1",
		Name: "previews",
		GitRepository: models.PreviewGitRepository{
			RepoURL:       "https://github.com/acme/api",
			BaseBranch:    "main",
			IntegrationID: "int-1",
		},
	}
	presented := presenters.PresentStackPreviewConfig(config)
	if presented.GitRepository == nil {
		t.Fatal("expected presented git repository")
	}
	if presented.GitRepository.RepoUrl != "https://github.com/acme/api" {
		t.Fatalf("unexpected repo url %q", presented.GitRepository.RepoUrl)
	}
	if presented.GitRepository.GetIntegrationId() != "int-1" {
		t.Fatalf("expected integration id to round-trip, got %q", presented.GitRepository.GetIntegrationId())
	}
}

package builders

import (
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestShouldEnableTLS(t *testing.T) {
	tests := []struct {
		name string
		fqdn string
		want bool
	}{
		{"empty string", "", false},
		{"nip.io subdomain", "app.192-168-1-1.nip.io", false},
		{"sslip.io subdomain", "app.10-0-0-1.sslip.io", false},
		{"dot local", "myapp.local", false},
		{"dot localhost", "myapp.localhost", false},
		{"real domain", "app.example.com", true},
		{"subdomain", "api.staging.example.com", true},
		{"bare domain", "example.com", true},
		{"io TLD not matching nip.io", "myapp.io", true},
		{"domain ending in local but not .local suffix", "app.mylocal", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldEnableTLS(tt.fqdn)
			if got != tt.want {
				t.Errorf("shouldEnableTLS(%q) = %v, want %v", tt.fqdn, got, tt.want)
			}
		})
	}
}

func TestHasTLSPorts(t *testing.T) {
	tests := []struct {
		name string
		spec *corev1alpha1.StackResourceSpec
		want bool
	}{
		{
			"no ports",
			&corev1alpha1.StackResourceSpec{},
			false,
		},
		{
			"ports without TLS",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 8080, TLS: false},
					{Number: 3000, TLS: false},
				},
			},
			false,
		},
		{
			"single TLS port",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 443, TLS: true},
				},
			},
			true,
		},
		{
			"mixed ports",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 8080, TLS: false},
					{Number: 443, TLS: true},
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasTLSPorts(tt.spec)
			if got != tt.want {
				t.Errorf("hasTLSPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildImageRepositorySpec(t *testing.T) {
	t.Run("in-cluster registry", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true, ClusterRegistryName: "test-registry"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "orgid", "stack", "resource")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ClusterRegistryRef == nil || got.ClusterRegistryRef.Name != "test-registry" {
			t.Errorf("ClusterRegistryRef.Name = %v, want test-registry", got.ClusterRegistryRef)
		}
		if got.Repository != "orgid/stack/resource" {
			t.Errorf("Repository = %q, want %q", got.Repository, "orgid/stack/resource")
		}
		if got.External != nil {
			t.Error("expected External to be nil")
		}
		if got.Auth != nil {
			t.Error("expected Auth to be nil")
		}
	})

	t.Run("external docker.io", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "nginx"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "org", "stack", "res")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.External == nil || got.External.Host != "index.docker.io" {
			t.Errorf("External.Host = %v, want index.docker.io", got.External)
		}
		if got.Repository != "library/nginx" {
			t.Errorf("Repository = %q, want %q", got.Repository, "library/nginx")
		}
		if got.ClusterRegistryRef != nil {
			t.Error("expected ClusterRegistryRef to be nil")
		}
	})

	t.Run("external host with port", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "myregistry.io:5000/org/repo"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "org", "stack", "res")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.External == nil || got.External.Host != "myregistry.io:5000" {
			t.Errorf("External.Host = %v, want myregistry.io:5000", got.External)
		}
		if got.Repository != "org/repo" {
			t.Errorf("Repository = %q, want %q", got.Repository, "org/repo")
		}
	})

	t.Run("external insecure", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{InsecureRegistry: true, ExternalImageRef: "myregistry.io/org/repo"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "org", "stack", "res")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.External == nil || got.External.TLS == nil {
			t.Fatal("expected External.TLS to be set")
		}
		if !got.External.TLS.Insecure {
			t.Error("expected TLS.Insecure to be true")
		}
	})

	t.Run("external no push secret", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "myregistry.io/org/repo"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "org", "stack", "res")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth != nil {
			t.Error("expected Auth to be nil")
		}
	})
}

func TestBuildBuildSourceRevision(t *testing.T) {
	t.Run("branch + commit", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		rev := models.BuildSourceRevision{
			Git: &models.GitRevision{
				Branch: "main",
				Commit: "abc1234",
			},
		}
		got, err := b.buildBuildSourceRevision(rev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GitRepo == nil {
			t.Fatal("expected GitRepo to be set")
		}
		if got.GitRepo.Branch != "main" {
			t.Errorf("Branch = %q, want %q", got.GitRepo.Branch, "main")
		}
		if got.GitRepo.Commit != "abc1234" {
			t.Errorf("Commit = %q, want %q", got.GitRepo.Commit, "abc1234")
		}
	})

	t.Run("tag + commit", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		rev := models.BuildSourceRevision{
			Git: &models.GitRevision{Tag: "v1.0.0", Commit: "def5678"},
		}
		got, err := b.buildBuildSourceRevision(rev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GitRepo == nil {
			t.Fatal("expected GitRepo to be set")
		}
		if got.GitRepo.Tag != "v1.0.0" {
			t.Errorf("Tag = %q, want %q", got.GitRepo.Tag, "v1.0.0")
		}
		if got.GitRepo.Commit != "def5678" {
			t.Errorf("Commit = %q, want %q", got.GitRepo.Commit, "def5678")
		}
	})

	t.Run("volume hash", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		rev := models.BuildSourceRevision{
			Volume: &models.VolumeRevision{CurrentVolumeHash: "hash123"},
		}
		got, err := b.buildBuildSourceRevision(rev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Volume == nil {
			t.Fatal("expected Volume to be set")
		}
		if got.Volume.RevisionString != "hash123" {
			t.Errorf("RevisionString = %q, want %q", got.Volume.RevisionString, "hash123")
		}
	})

	t.Run("missing commit", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		rev := models.BuildSourceRevision{
			Git: &models.GitRevision{Branch: "main"},
		}
		_, err := b.buildBuildSourceRevision(rev)
		if err == nil {
			t.Fatal("expected error for missing commit")
		}
		if got := err.Error(); !strings.Contains(got, "commit SHA is required") {
			t.Errorf("error = %q, want it to contain %q", got, "commit SHA is required")
		}
	})
}

func TestBuildStackResourceCR_AnnotationMerge(t *testing.T) {
	builder := &clusterResourceBuilder{}

	t.Run("sets annotation when TLS ports exist", func(t *testing.T) {
		sr := &models.StackResource{
			Name:      "web",
			Namespace: "default",
			StackID:   "stack-1",
			Ports: []models.Port{
				{Number: 443, ExposedToPublic: true, ExposedFqdn: "app.example.com", Protocol: "http"},
			},
		}

		cr, err := builder.BuildStackResourceCR(sr, "test-stack", "test-org")
		if err != nil {
			t.Fatalf("BuildStackResourceCR() error = %v", err)
		}

		val, ok := cr.Annotations[corev1alpha1.ClusterIssuerAnnotation]
		if !ok {
			t.Fatal("expected cluster issuer annotation to be set")
		}
		if val != models.DefaultClusterIssuerName {
			t.Errorf("annotation value = %q, want %q", val, models.DefaultClusterIssuerName)
		}
	})

	t.Run("preserves existing labels when adding annotation", func(t *testing.T) {
		sr := &models.StackResource{
			Name:      "web",
			Namespace: "default",
			StackID:   "stack-1",
			Version:   3,
			Ports: []models.Port{
				{Number: 443, ExposedToPublic: true, ExposedFqdn: "app.example.com", Protocol: "http"},
			},
		}

		cr, err := builder.BuildStackResourceCR(sr, "test-stack", "test-org")
		if err != nil {
			t.Fatalf("BuildStackResourceCR() error = %v", err)
		}

		if cr.Labels[corev1alpha1.LabelStackID] != "stack-1" {
			t.Errorf("LabelStackID = %q, want %q", cr.Labels[corev1alpha1.LabelStackID], "stack-1")
		}
		if _, ok := cr.Annotations[corev1alpha1.ClusterIssuerAnnotation]; !ok {
			t.Fatal("expected cluster issuer annotation to be set")
		}
	})
}

// TestBuildCR_OrgCredentialSecretNames asserts that org-level (integration)
// registry credentials reference purpose-distinct cluster-secret names in the
// CR, matching exactly what the release worker syncs. A pull credential and a
// push credential on the same host must resolve to different names.
func TestBuildCR_OrgCredentialSecretNames(t *testing.T) {
	const host = "reg.example.com"

	t.Run("pull image auto-attach", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		resolver := mocks.NewMockCredentialResolver(ctrl)
		resolver.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "reg.example.com/project/app:v1", credentials.RegistryPurposePull, gomock.Any()).
			Return(&credentials.ResolvedRegistryCredential{
				Source:  credentials.SourceIntegration,
				Host:    host,
				Purpose: credentials.RegistryPurposePull,
			}, nil)

		b := &clusterResourceBuilder{credentialResolver: resolver}
		sr := &models.StackResource{
			ImageConfig: &models.ImageConfigSpec{Image: "reg.example.com/project/app:v1"},
		}
		got, err := b.buildStackResourceImageSpec(sr, "org-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.PullAuth == nil || got.PullAuth.DockerConfigAuth == nil {
			t.Fatal("expected PullAuth.DockerConfigAuth to be set")
		}
		want := credentials.ClusterSecretNameForRegistryHost(host, credentials.RegistryPurposePull)
		if got := got.PullAuth.DockerConfigAuth.SecretRef.Name; got != want {
			t.Errorf("pull SecretRef.Name = %q, want %q", got, want)
		}
	})

	t.Run("push repo auto-attach", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		resolver := mocks.NewMockCredentialResolver(ctrl)
		resolver.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "reg.example.com/project/app", credentials.RegistryPurposePush, gomock.Any()).
			Return(&credentials.ResolvedRegistryCredential{
				Source:  credentials.SourceIntegration,
				Host:    host,
				Purpose: credentials.RegistryPurposePush,
			}, nil)

		b := &clusterResourceBuilder{credentialResolver: resolver}
		cfg := &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "reg.example.com/project/app"},
		}
		got, err := b.buildImageRepositorySpec(cfg, "org-1", "stack", "res")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth == nil || got.Auth.DockerConfig == nil {
			t.Fatal("expected Auth.DockerConfig to be set")
		}
		want := credentials.ClusterSecretNameForRegistryHost(host, credentials.RegistryPurposePush)
		if got := got.Auth.DockerConfig.SecretRef.Name; got != want {
			t.Errorf("push SecretRef.Name = %q, want %q", got, want)
		}
		// Pull and push on the same host must not collide.
		pullName := credentials.ClusterSecretNameForRegistryHost(host, credentials.RegistryPurposePull)
		if want == pullName {
			t.Fatalf("push name %q collides with pull name", want)
		}
	})
}

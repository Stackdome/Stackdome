package builders

import (
	"context"
	"strings"
	"testing"

	pkgerrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

type stubSecretFetcher struct {
	secrets map[string]*models.Secret
}

func (s *stubSecretFetcher) InternalGetByID(_ context.Context, secretID string) (*models.Secret, *pkgerrors.ServiceError) {
	if sec, ok := s.secrets[secretID]; ok {
		return sec, nil
	}
	return nil, pkgerrors.NotFound("secret %s not found", secretID)
}

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
			ImageRepositoryUrl:   "orgid/stack/resource",
		}
		got, err := b.buildImageRepositorySpec(cfg)
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
			ImageRepositoryUrl:   "nginx",
			BuildImageRepository: models.BuildImageRepository{},
		}
		got, err := b.buildImageRepositorySpec(cfg)
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
			ImageRepositoryUrl:   "myregistry.io:5000/org/repo",
			BuildImageRepository: models.BuildImageRepository{},
		}
		got, err := b.buildImageRepositorySpec(cfg)
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
			ImageRepositoryUrl:   "myregistry.io/org/repo",
			BuildImageRepository: models.BuildImageRepository{InsecureRegistry: true},
		}
		got, err := b.buildImageRepositorySpec(cfg)
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

	t.Run("external with push secret", func(t *testing.T) {
		fetcher := &stubSecretFetcher{secrets: map[string]*models.Secret{
			"secret-1": {Name: "push-secret", Type: models.SecretTypeUsernamePassword},
		}}
		b := &clusterResourceBuilder{secretService: fetcher}
		cfg := &models.BuildConfigSpec{
			ImageRepositoryUrl:   "myregistry.io/org/repo",
			BuildImageRepository: models.BuildImageRepository{},
			RegistrySecretRef:    &models.SecretReference{SecretID: "secret-1"},
		}
		got, err := b.buildImageRepositorySpec(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth == nil || got.Auth.Basic == nil {
			t.Fatal("expected Auth.Basic to be set")
		}
		wantName := fetcher.secrets["secret-1"].ClusterSecretName()
		if got.Auth.Basic.SecretRef.Name != wantName {
			t.Errorf("SecretRef.Name = %q, want %q", got.Auth.Basic.SecretRef.Name, wantName)
		}
		if got.Auth.Basic.UsernameKey != models.UsernameSecretKey {
			t.Errorf("UsernameKey = %q, want %q", got.Auth.Basic.UsernameKey, models.UsernameSecretKey)
		}
		if got.Auth.Basic.PasswordKey != models.PasswordSecretKey {
			t.Errorf("PasswordKey = %q, want %q", got.Auth.Basic.PasswordKey, models.PasswordSecretKey)
		}
	})

	t.Run("external no push secret", func(t *testing.T) {
		b := &clusterResourceBuilder{}
		cfg := &models.BuildConfigSpec{
			ImageRepositoryUrl:   "myregistry.io/org/repo",
			BuildImageRepository: models.BuildImageRepository{},
		}
		got, err := b.buildImageRepositorySpec(cfg)
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

		cr, err := builder.BuildStackResourceCR(sr, "test-stack")
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

		cr, err := builder.BuildStackResourceCR(sr, "test-stack")
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

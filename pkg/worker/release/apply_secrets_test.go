package release

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	applySecretsTestOrgID     = "org-1"
	applySecretsTestNamespace = "ns-1"
	applySecretsTestStackID   = "stack-1"
)

func applySecretsTestClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme setup failed: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

func listSyncedSecrets(t *testing.T, c client.Client) map[string]*corev1.Secret {
	t.Helper()
	list := &corev1.SecretList{}
	if err := c.List(context.Background(), list, client.InNamespace(applySecretsTestNamespace)); err != nil {
		t.Fatalf("failed to list secrets: %v", err)
	}
	out := make(map[string]*corev1.Secret, len(list.Items))
	for i := range list.Items {
		out[list.Items[i].Name] = &list.Items[i]
	}
	return out
}

func dockerConfigAuths(t *testing.T, secret *corev1.Secret) map[string]map[string]string {
	t.Helper()
	raw, ok := secret.Data[corev1.DockerConfigJsonKey]
	if !ok {
		t.Fatalf("secret %q has no dockerconfigjson payload", secret.Name)
	}
	var parsed struct {
		Auths map[string]map[string]string `json:"auths"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("failed to unmarshal dockerconfigjson for %q: %v", secret.Name, err)
	}
	return parsed.Auths
}

// TestSyncHubSecrets_SecretRefPullPushMerge covers the collision where one hub
// secret backs both an image pull (host A) and a build push (host B). Both
// builders name the cluster secret secret.ClusterSecretName(); the pull and
// push dockerconfig auth entries must be merged into a single secret rather
// than the second write being skipped.
func TestSyncHubSecrets_SecretRefPullPushMerge(t *testing.T) {
	ctrl := gomock.NewController(t)

	hubSecret := &models.Secret{
		ID:   "hub-secret-1",
		Name: "shared",
		Type: models.SecretTypeDockerRegistry,
		Data: map[string]string{
			models.UsernameSecretKey: "shared-user",
			models.PasswordSecretKey: "shared-pass",
		},
	}
	secretRef := &models.SecretReference{SecretID: hubSecret.ID}

	resolver := mocks.NewMockCredentialResolver(ctrl)
	resolved := &credentials.ResolvedRegistryCredential{
		Source:   credentials.SourceSecretRef,
		Username: hubSecret.Data[models.UsernameSecretKey],
		Password: hubSecret.Data[models.PasswordSecretKey],
		SecretID: hubSecret.ID,
		DataHash: "hash-1",
		Secret:   hubSecret,
	}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), applySecretsTestOrgID, "reg-a.example.com/team/app:v1", credentials.RegistryPurposePull, gomock.Any()).
		Return(resolved, nil)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), applySecretsTestOrgID, "reg-b.example.com/svc", credentials.RegistryPurposePush, gomock.Any()).
		Return(resolved, nil)

	r := &applyReconciler{
		secretBuilder:      builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		credentialResolver: resolver,
		logger:             testLogger(),
	}

	stack := &models.Stack{
		ID:             applySecretsTestStackID,
		OrganisationID: applySecretsTestOrgID,
		Namespace:      applySecretsTestNamespace,
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "reg-a.example.com/team/app:v1", PullSecretRef: secretRef},
			},
			{
				// For an explicit push secret ref, syncHubSecrets derives the
				// dockerconfig auth URL from the real push repository
				// (ExternalImageRef, host B), not the resource name.
				Name: "builder",
				BuildConfig: &models.BuildConfigSpec{
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "reg-b.example.com/svc"},
					RegistrySecretRef:    secretRef,
				},
			},
		},
	}

	clusterClient := applySecretsTestClient(t)
	if err := r.syncHubSecrets(context.Background(), clusterClient, stack); err != nil {
		t.Fatalf("syncHubSecrets returned error: %v", err)
	}

	secrets := listSyncedSecrets(t, clusterClient)
	if len(secrets) != 1 {
		t.Fatalf("expected exactly 1 cluster secret, got %d: %v", len(secrets), secretNames(secrets))
	}

	wantName := hubSecret.ClusterSecretName()
	secret, ok := secrets[wantName]
	if !ok {
		t.Fatalf("expected secret %q, got %v", wantName, secretNames(secrets))
	}

	auths := dockerConfigAuths(t, secret)
	if _, ok := auths["https://reg-a.example.com"]; !ok {
		t.Errorf("expected pull auth entry for reg-a, got auths %v", authHosts(auths))
	}
	if _, ok := auths["https://reg-b.example.com"]; !ok {
		t.Errorf("expected push auth entry for reg-b, got auths %v", authHosts(auths))
	}
	if len(auths) != 2 {
		t.Errorf("expected 2 auth entries, got %d: %v", len(auths), authHosts(auths))
	}
}

// TestSyncHubSecrets_ExplicitPushSecretRefKeyedByRepo covers the explicit
// push-secret tier: a resource that builds and pushes to an external registry
// (ghcr.io) using an explicit RegistrySecretRef. The synced secret must be a
// dockerconfigjson keyed by the real push registry's auth URL — not Docker Hub,
// not the resource name — so it matches the DockerConfigAuth the CR builder
// emits and the agent authenticates with real credentials.
func TestSyncHubSecrets_ExplicitPushSecretRefKeyedByRepo(t *testing.T) {
	ctrl := gomock.NewController(t)

	hubSecret := &models.Secret{
		ID:   "hub-push-1",
		Name: "ghcr-push",
		Type: models.SecretTypeDockerRegistry,
		Data: map[string]string{
			models.UsernameSecretKey: "acme-bot",
			models.PasswordSecretKey: "ghp_token",
		},
	}
	secretRef := &models.SecretReference{SecretID: hubSecret.ID}

	resolver := mocks.NewMockCredentialResolver(ctrl)
	resolved := &credentials.ResolvedRegistryCredential{
		Source:   credentials.SourceSecretRef,
		Username: hubSecret.Data[models.UsernameSecretKey],
		Password: hubSecret.Data[models.PasswordSecretKey],
		SecretID: hubSecret.ID,
		DataHash: "hash-push",
		Secret:   hubSecret,
	}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), applySecretsTestOrgID, "ghcr.io/acme/api", credentials.RegistryPurposePush, gomock.Any()).
		Return(resolved, nil)

	r := &applyReconciler{
		secretBuilder:      builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		credentialResolver: resolver,
		logger:             testLogger(),
	}

	stack := &models.Stack{
		ID:             applySecretsTestStackID,
		OrganisationID: applySecretsTestOrgID,
		Namespace:      applySecretsTestNamespace,
		StackResources: []*models.StackResource{
			{
				Name: "api",
				BuildConfig: &models.BuildConfigSpec{
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "ghcr.io/acme/api"},
					RegistrySecretRef:    secretRef,
				},
			},
		},
	}

	clusterClient := applySecretsTestClient(t)
	if err := r.syncHubSecrets(context.Background(), clusterClient, stack); err != nil {
		t.Fatalf("syncHubSecrets returned error: %v", err)
	}

	secrets := listSyncedSecrets(t, clusterClient)
	if len(secrets) != 1 {
		t.Fatalf("expected exactly 1 cluster secret, got %d: %v", len(secrets), secretNames(secrets))
	}

	wantName := hubSecret.ClusterSecretName()
	secret, ok := secrets[wantName]
	if !ok {
		t.Fatalf("expected secret %q, got %v", wantName, secretNames(secrets))
	}
	if secret.Type != corev1.SecretTypeDockerConfigJson {
		t.Fatalf("expected dockerconfigjson secret type, got %q", secret.Type)
	}

	auths := dockerConfigAuths(t, secret)
	if _, ok := auths["https://ghcr.io"]; !ok {
		t.Errorf("expected auth entry keyed by ghcr auth URL, got %v", authHosts(auths))
	}
	if _, ok := auths["https://index.docker.io/v1/"]; ok {
		t.Errorf("must not key push auth by Docker Hub, got %v", authHosts(auths))
	}
	if len(auths) != 1 {
		t.Errorf("expected exactly 1 auth entry, got %d: %v", len(auths), authHosts(auths))
	}
	if got := auths["https://ghcr.io"]["username"]; got != "acme-bot" {
		t.Errorf("push auth username = %q, want acme-bot", got)
	}
}

// TestSyncHubSecrets_OrgPullPushDistinctNames covers an org-level pull
// credential and push credential on the SAME host but backed by different
// accounts. They must land as two distinct purpose-named cluster secrets, each
// carrying its own account, not collapse onto one host-only name.
func TestSyncHubSecrets_OrgPullPushDistinctNames(t *testing.T) {
	ctrl := gomock.NewController(t)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	pull := &credentials.ResolvedRegistryCredential{
		Source:       credentials.SourceIntegration,
		Username:     "pull-user",
		Password:     "pull-pass",
		Host:         "reg.example.com",
		Purpose:      credentials.RegistryPurposePull,
		CredentialID: "cred-pull",
		DataHash:     "hash-pull",
	}
	push := &credentials.ResolvedRegistryCredential{
		Source:       credentials.SourceIntegration,
		Username:     "push-user",
		Password:     "push-pass",
		Host:         "reg.example.com",
		Purpose:      credentials.RegistryPurposePush,
		CredentialID: "cred-push",
		DataHash:     "hash-push",
	}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), applySecretsTestOrgID, "reg.example.com/team/app:v1", credentials.RegistryPurposePull, gomock.Any()).
		Return(pull, nil)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), applySecretsTestOrgID, "reg.example.com/team/app", credentials.RegistryPurposePush, gomock.Any()).
		Return(push, nil)

	r := &applyReconciler{
		secretBuilder:      builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		credentialResolver: resolver,
		logger:             testLogger(),
	}

	stack := &models.Stack{
		ID:             applySecretsTestStackID,
		OrganisationID: applySecretsTestOrgID,
		Namespace:      applySecretsTestNamespace,
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "reg.example.com/team/app:v1"},
			},
			{
				Name: "builder",
				BuildConfig: &models.BuildConfigSpec{
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "reg.example.com/team/app"},
				},
			},
		},
	}

	clusterClient := applySecretsTestClient(t)
	if err := r.syncHubSecrets(context.Background(), clusterClient, stack); err != nil {
		t.Fatalf("syncHubSecrets returned error: %v", err)
	}

	secrets := listSyncedSecrets(t, clusterClient)
	if len(secrets) != 2 {
		t.Fatalf("expected exactly 2 cluster secrets, got %d: %v", len(secrets), secretNames(secrets))
	}

	pullName := credentials.ClusterSecretNameForRegistryHost(pull.Host, credentials.RegistryPurposePull)
	pushName := credentials.ClusterSecretNameForRegistryHost(push.Host, credentials.RegistryPurposePush)
	if pullName == pushName {
		t.Fatalf("pull and push secret names must differ, both %q", pullName)
	}

	pullSecret, ok := secrets[pullName]
	if !ok {
		t.Fatalf("expected pull secret %q, got %v", pullName, secretNames(secrets))
	}
	pushSecret, ok := secrets[pushName]
	if !ok {
		t.Fatalf("expected push secret %q, got %v", pushName, secretNames(secrets))
	}

	if got := dockerConfigAuths(t, pullSecret)["https://reg.example.com"]["username"]; got != "pull-user" {
		t.Errorf("pull secret account = %q, want pull-user", got)
	}
	if got := dockerConfigAuths(t, pushSecret)["https://reg.example.com"]["username"]; got != "push-user" {
		t.Errorf("push secret account = %q, want push-user", got)
	}
}

func secretNames(secrets map[string]*corev1.Secret) []string {
	names := make([]string, 0, len(secrets))
	for name := range secrets {
		names = append(names, name)
	}
	return names
}

func authHosts(auths map[string]map[string]string) []string {
	hosts := make([]string, 0, len(auths))
	for host := range auths {
		hosts = append(hosts, host)
	}
	return hosts
}

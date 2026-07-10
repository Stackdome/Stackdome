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

func applySecretsTestClient(t gomock.TestHelper) client.Client {
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

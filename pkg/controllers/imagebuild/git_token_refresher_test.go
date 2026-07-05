package imagebuild

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const testGitSecretName = "git-integration-github.com"

func refresherTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme setup failed: %v", err)
	}
	return scheme
}

func buildWithGitAuth(phase buildsv1alpha1.BuildPhase) *buildsv1alpha1.ImageBuild {
	return &buildsv1alpha1.ImageBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "build-1", Namespace: "ns-1"},
		Spec: buildsv1alpha1.ImageBuildSpec{
			ResourceName: "api",
			BuildContext: buildsv1alpha1.BuildContextSpec{
				ContextSource: &corev1alpha1.BuildContextSource{
					Git: &corev1alpha1.GitRepoSource{
						RepoUrl: "https://github.com/acme/api",
						Auth: &corev1alpha1.GitAuth{
							UsernamePasswordAuthRef: &corev1alpha1.CredentialSecretKeyPair{
								SecretRef:   corev1.SecretReference{Name: testGitSecretName},
								UsernameKey: models.UsernameSecretKey,
								PasswordKey: models.PasswordSecretKey,
							},
						},
					},
				},
			},
		},
		Status: buildsv1alpha1.ImageBuildStatus{Phase: phase},
	}
}

func tokenSecret(expiresAt time.Time) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testGitSecretName,
			Namespace: "ns-1",
			Annotations: map[string]string{
				models.GitTokenMintedAtAnnotation:  expiresAt.Add(-time.Hour).UTC().Format(time.RFC3339),
				models.GitTokenExpiresAtAnnotation: expiresAt.UTC().Format(time.RFC3339),
				models.GitIntegrationIDAnnotation:  "gi-app",
			},
		},
		Data: map[string][]byte{
			models.UsernameSecretKey: []byte(githubapp.CloneUsername),
			models.PasswordSecretKey: []byte("ghs_old"),
		},
	}
}

func newRefresherForTest(t *testing.T, now time.Time, objects ...client.Object) (*ImageBuildReconciler, *mocks.MockGitIntegrationService, client.Client) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	fakeClient := fake.NewClientBuilder().WithScheme(refresherTestScheme(t)).WithObjects(objects...).Build()
	gitIntegrations := mocks.NewMockGitIntegrationService(ctrl)
	r := &ImageBuildReconciler{
		Client:                fakeClient,
		GitIntegrationService: gitIntegrations,
		Logger:                logger.NewLoggerWithPrefix(context.Background(), "refresher-test"),
		clock:                 func() time.Time { return now },
	}
	return r, gitIntegrations, fakeClient
}

func TestRefresherTable(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name          string
		phase         buildsv1alpha1.BuildPhase
		expiresAt     time.Time
		expectMint    bool
		expectRequeue bool
	}{
		{
			name:          "terminal build never requeues",
			phase:         buildsv1alpha1.BuildPhaseSuccess,
			expiresAt:     now.Add(5 * time.Minute),
			expectMint:    false,
			expectRequeue: false,
		},
		{
			name:          "fresh token is a no-op with a scheduled requeue",
			phase:         buildsv1alpha1.BuildPhasePending,
			expiresAt:     now.Add(time.Hour),
			expectMint:    false,
			expectRequeue: true,
		},
		{
			name:          "near-expiry token is refreshed",
			phase:         buildsv1alpha1.BuildPhasePending,
			expiresAt:     now.Add(5 * time.Minute),
			expectMint:    true,
			expectRequeue: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			build := buildWithGitAuth(tc.phase)
			secret := tokenSecret(tc.expiresAt)
			r, gitIntegrations, fakeClient := newRefresherForTest(t, now, secret)

			if tc.expectMint {
				gitIntegrations.EXPECT().
					InternalMintCloneToken(gomock.Any(), "gi-app", "https://github.com/acme/api").
					Return(&models.GitHubAppMintResult{
						IntegrationID: "gi-app",
						Host:          "github.com",
						Token:         "ghs_new",
						MintedAt:      now,
						ExpiresAt:     now.Add(time.Hour),
					}, nil)
			}

			result, err := r.reconcileGitTokenRefresh(context.Background(), build, &models.ImageBuild{ID: "build-1"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.expectRequeue && result.RequeueAfter <= 0 {
				t.Fatalf("expected a requeue, got %+v", result)
			}
			if !tc.expectRequeue && result.RequeueAfter != 0 {
				t.Fatalf("expected no requeue, got %+v", result)
			}

			if tc.expectMint {
				updated := &corev1.Secret{}
				if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: testGitSecretName, Namespace: "ns-1"}, updated); err != nil {
					t.Fatalf("failed to get secret: %v", err)
				}
				if updated.Annotations[models.GitTokenExpiresAtAnnotation] != now.Add(time.Hour).Format(time.RFC3339) {
					t.Fatalf("expected refreshed expiry annotation, got %q", updated.Annotations[models.GitTokenExpiresAtAnnotation])
				}
				if got := updated.StringData[models.PasswordSecretKey]; got != "ghs_new" {
					t.Fatalf("expected refreshed token in secret, got %q", got)
				}
			}
		})
	}
}

func TestRefresherMintFailureRecordsConditionAndDoesNotCrash(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	build := buildWithGitAuth(buildsv1alpha1.BuildPhasePending)
	secret := tokenSecret(now.Add(5 * time.Minute))
	r, gitIntegrations, _ := newRefresherForTest(t, now, secret)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	dbBuilds := mocks.NewMockImageBuildService(ctrl)
	r.DBImageBuildService = dbBuilds

	gitIntegrations.EXPECT().
		InternalMintCloneToken(gomock.Any(), "gi-app", gomock.Any()).
		Return(nil, apperrors.NotFound("app uninstalled"))
	dbBuilds.EXPECT().
		InternalUpdateStatus(gomock.Any(), "build-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, status *models.ImageBuildStatus) *apperrors.ServiceError {
			if len(status.Conditions) == 0 || status.Conditions[0].Type != models.GitTokenRefreshFailedConditionType {
				t.Fatalf("expected a token refresh failure condition, got %+v", status.Conditions)
			}
			return nil
		})

	result, err := r.reconcileGitTokenRefresh(context.Background(), build, &models.ImageBuild{ID: "build-1"})
	if err != nil {
		t.Fatalf("expected mint failure to be swallowed, got %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("expected no requeue after mint failure, got %+v", result)
	}
}

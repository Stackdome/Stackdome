package imagebuild

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	// tokenRefreshWindow is how long before expiry a minted GitHub App token
	// is refreshed.
	tokenRefreshWindow = 15 * time.Minute
	// minTokenRefreshRequeue floors the requeue interval.
	minTokenRefreshRequeue = time.Minute
)

// reconcileGitTokenRefresh keeps minted GitHub App installation tokens fresh
// for in-flight builds: when the build's git-credentials secret is backed by
// an app token nearing expiry, a fresh token is minted and written back, and
// the build is requeued ahead of the next expiry.
func (r *ImageBuildReconciler) reconcileGitTokenRefresh(ctx context.Context, imageBuild *buildsv1alpha1.ImageBuild, dbBuild *models.ImageBuild) (ctrl.Result, error) {
	if isTerminalBuildPhase(imageBuild.Status.Phase) {
		return ctrl.Result{}, nil
	}
	if r.GitIntegrationService == nil {
		return ctrl.Result{}, nil
	}

	gitSource := gitSourceFromBuild(imageBuild)
	if gitSource == nil || gitSource.Auth == nil || gitSource.Auth.UsernamePasswordAuthRef == nil {
		return ctrl.Result{}, nil
	}
	secretName := gitSource.Auth.UsernamePasswordAuthRef.SecretRef.Name
	if secretName == "" {
		return ctrl.Result{}, nil
	}

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: secretName, Namespace: imageBuild.Namespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	expiresRaw := secret.Annotations[models.GitTokenExpiresAtAnnotation]
	if expiresRaw == "" {
		// Not backed by a minted token; nothing to refresh.
		return ctrl.Result{}, nil
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresRaw)
	if err != nil {
		r.Logger.Error(ctx, "build %s/%s: malformed token expiry annotation %q: %v", imageBuild.Namespace, imageBuild.Name, expiresRaw, err)
		return ctrl.Result{}, nil
	}

	now := r.clock()
	if remaining := expiresAt.Sub(now); remaining > tokenRefreshWindow {
		return ctrl.Result{RequeueAfter: floorRequeue(remaining - tokenRefreshWindow)}, nil
	}

	integrationID := secret.Annotations[models.GitIntegrationIDAnnotation]
	if integrationID == "" {
		r.Logger.Error(ctx, "build %s/%s: token-backed git secret '%s' has no integration annotation", imageBuild.Namespace, imageBuild.Name, secretName)
		return ctrl.Result{}, nil
	}

	mint, serr := r.GitIntegrationService.InternalMintCloneToken(ctx, integrationID, gitSource.RepoUrl)
	if serr != nil {
		// The app may have been uninstalled mid-build. Record the failure on
		// the build and stop refreshing; never crash the reconcile loop.
		r.Logger.Error(ctx, "build %s/%s: failed to refresh GitHub App token: %v", imageBuild.Namespace, imageBuild.Name, serr)
		r.recordTokenRefreshFailure(ctx, dbBuild, serr.Error())
		return ctrl.Result{}, nil
	}

	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}
	secret.StringData[models.UsernameSecretKey] = githubapp.CloneUsername
	secret.StringData[models.PasswordSecretKey] = mint.Token
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}
	secret.Annotations[models.GitTokenMintedAtAnnotation] = mint.MintedAt.UTC().Format(time.RFC3339)
	secret.Annotations[models.GitTokenExpiresAtAnnotation] = mint.ExpiresAt.UTC().Format(time.RFC3339)
	if err := r.Client.Update(ctx, secret); err != nil {
		return ctrl.Result{}, err
	}

	r.Logger.Info(ctx, "build %s/%s: refreshed GitHub App token for secret '%s'", imageBuild.Namespace, imageBuild.Name, secretName)
	return ctrl.Result{RequeueAfter: floorRequeue(time.Until(mint.ExpiresAt) - tokenRefreshWindow)}, nil
}

func (r *ImageBuildReconciler) recordTokenRefreshFailure(ctx context.Context, dbBuild *models.ImageBuild, message string) {
	if dbBuild == nil {
		return
	}
	status := dbBuild.Status
	if status == nil {
		status = &models.ImageBuildStatus{}
	}
	status.Conditions = append(status.Conditions, models.Condition{
		Type:               models.GitTokenRefreshFailedConditionType,
		Status:             string(metav1.ConditionTrue),
		LastTransitionTime: metav1.Now().Time,
		Reason:             models.GitTokenRefreshFailedConditionReason,
		Message:            message,
	})
	if serr := r.DBImageBuildService.InternalUpdateStatus(ctx, dbBuild.ID, status); serr != nil {
		r.Logger.Error(ctx, "failed to record token refresh failure for build '%s': %v", dbBuild.ID, serr)
	}
}

func gitSourceFromBuild(imageBuild *buildsv1alpha1.ImageBuild) *corev1alpha1.GitRepoSource {
	if imageBuild.Spec.BuildContext.ContextSource == nil {
		return nil
	}
	return imageBuild.Spec.BuildContext.ContextSource.Git
}

func isTerminalBuildPhase(phase buildsv1alpha1.BuildPhase) bool {
	switch phase {
	case buildsv1alpha1.BuildPhaseSuccess, buildsv1alpha1.BuildPhaseFailed, buildsv1alpha1.BuildPhaseCancelled:
		return true
	default:
		return false
	}
}

func floorRequeue(d time.Duration) time.Duration {
	if d < minTokenRefreshRequeue {
		return minTokenRefreshRequeue
	}
	return d
}

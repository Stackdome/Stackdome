package release

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackdeploy"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// syncHubSecrets ensures image pull, image push, and git credential secrets
// are synced to the target cluster namespace before the stack CR is applied.
func (r *applyReconciler) syncHubSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	if stack.HasImagePullSecrets() {
		for imageUrl, secretID := range stack.GetImagePullSecretIDMap() {
			secret, serr := r.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return fmt.Errorf("failed to get image pull secret '%s': %w", secretID, serr)
			}
			clusterSecret, err := r.secretBuilder.BuildDockerConfigJsonSecretForImage(ctx, secret, imageUrl)
			if err != nil {
				return fmt.Errorf("failed to build docker config for image '%s': %w", imageUrl, err)
			}
			clusterSecret.Namespace = stack.Namespace
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID
			if err := createOrUpdateSecret(ctx, clusterClient, clusterSecret); err != nil {
				return fmt.Errorf("failed to sync image pull secret '%s': %w", secretID, err)
			}
		}
	}

	if stack.HasImagePushSecrets() {
		for repoUrl, secretID := range stack.GetImagePushSecretIDMap() {
			secret, serr := r.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return fmt.Errorf("failed to get image push secret '%s': %w", secretID, serr)
			}
			clusterSecret, err := r.secretBuilder.BuildDockerConfigJsonSecretForRepository(ctx, secret, repoUrl)
			if err != nil {
				return fmt.Errorf("failed to build docker config for repo '%s': %w", repoUrl, err)
			}
			clusterSecret.Namespace = stack.Namespace
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID
			if err := createOrUpdateSecret(ctx, clusterClient, clusterSecret); err != nil {
				return fmt.Errorf("failed to sync image push secret '%s': %w", secretID, err)
			}
		}
	}

	if stack.HasGitCredentials() {
		for repoUrl, secretID := range stack.GetGitCredentialsMap() {
			secret, serr := r.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return fmt.Errorf("failed to get git credential secret '%s': %w", secretID, serr)
			}
			clusterSecret, err := r.secretBuilder.BuildGitCredentialsSecret(ctx, secret, repoUrl)
			if err != nil {
				return fmt.Errorf("failed to build git credential secret for repo '%s': %w", repoUrl, err)
			}
			clusterSecret.Namespace = stack.Namespace
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID
			if err := createOrUpdateSecret(ctx, clusterClient, clusterSecret); err != nil {
				return fmt.Errorf("failed to sync git credential secret '%s': %w", secretID, err)
			}
		}
	}

	return nil
}

// syncPostgresCredentialSecrets ensures postgres connection credentials are
// available as K8s secrets in the stack namespace.
func (r *applyReconciler) syncPostgresCredentialSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	pgConnections := stack.Connections.FromType(models.TopologyNodeTypePostgresAddon)
	if len(pgConnections) == 0 {
		return nil
	}

	for _, connection := range pgConnections {
		addonID := connection.From.Id
		database, superuser, err := stackdeploy.PostgresConnectionConfig(connection)
		if err != nil {
			return fmt.Errorf("failed to parse postgres connection config: %w", err)
		}

		creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, addonID, database, superuser)
		if credErr != nil {
			return fmt.Errorf("failed to get postgres credentials for addon '%s' db '%s': %w", addonID, database, credErr)
		}

		secretName := stackdeploy.PostgresCredentialSecretName(addonID, database)
		outputMap := creds.ToOutputMap()
		secretData := make(map[string][]byte, len(outputMap))
		for k, v := range outputMap {
			secretData[k] = []byte(v)
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: stack.Namespace,
				Labels: map[string]string{
					models.StackIDLabel: stack.ID,
				},
			},
			Data: secretData,
		}
		if err := createOrUpdateSecret(ctx, clusterClient, desired); err != nil {
			return fmt.Errorf("failed to sync postgres credential secret '%s': %w", secretName, err)
		}
	}

	return nil
}

// syncGenericSecrets ensures user-created secrets (Generic, Token, etc.) connected
// to the stack are available as K8s Opaque secrets in the namespace.
func (r *applyReconciler) syncGenericSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	secretConnections := stack.Connections.FromType(models.TopologyNodeTypeSecret)
	if len(secretConnections) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	for _, connection := range secretConnections {
		secretID := connection.From.Id
		if _, ok := seen[secretID]; ok {
			continue
		}
		seen[secretID] = struct{}{}

		secret, serr := r.secretService.InternalGetByID(ctx, secretID)
		if serr != nil {
			return fmt.Errorf("failed to get secret '%s': %w", secretID, serr)
		}

		secretData := make(map[string][]byte, len(secret.Data))
		for k, v := range secret.Data {
			secretData[k] = []byte(v)
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.ClusterSecretName(),
				Namespace: stack.Namespace,
				Labels: map[string]string{
					models.StackIDLabel: stack.ID,
				},
				Annotations: map[string]string{
					models.SecretDataHashAnnotation: secret.DataHash,
					models.SecretIDAnnotation:       secret.ID,
				},
			},
			Data: secretData,
		}
		if err := createOrUpdateSecret(ctx, clusterClient, desired); err != nil {
			return fmt.Errorf("failed to sync generic secret '%s': %w", secret.Name, err)
		}
	}

	return nil
}

func createOrUpdateSecret(ctx context.Context, clusterClient client.Client, desired *corev1.Secret) error {
	existing := &corev1.Secret{}
	err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			return clusterClient.Create(ctx, desired)
		}
		return err
	}
	desired.ResourceVersion = existing.ResourceVersion
	return clusterClient.Update(ctx, desired)
}

func failRelease(ctx context.Context, svc releaseService, log logger.Logger, release *models.StackRelease, msg string) {
	log.Errorf("release %s failed: %s", release.ID, msg)
	if _, err := svc.MarkFailed(ctx, release.ID, msg, nil); err != nil {
		log.Errorf("failed to mark release %s as failed: %v", release.ID, err)
	}
}

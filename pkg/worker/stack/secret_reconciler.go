package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type secretReconciler struct {
	clusterManager       clustermanager.ClusterManager
	secretService        secretService
	resourceUsageService resourceUsageService
	secretBuilder        builders.SecretBuilder
	logger               logger.Logger
}

type SecretReconcilerSpec struct {
	ClusterManager       clustermanager.ClusterManager
	SecretService        secretService
	ResourceUsageService resourceUsageService
	Logger               logger.Logger
}

func NewSecretReconciler(spec SecretReconcilerSpec) *secretReconciler {
	return &secretReconciler{
		clusterManager:       spec.ClusterManager,
		secretService:        spec.SecretService,
		resourceUsageService: spec.ResourceUsageService,
		logger:               spec.Logger,
		secretBuilder: builders.NewSecretBuilder(
			builders.SecretBuilderSpec{
				SecretFetcher: spec.SecretService,
			},
		),
	}
}

func (s *secretReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	clusterClient, cerr := s.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client for cluster %s: %w", stack.ClusterID, cerr)
	}
	reconcileFns := []func(context.Context, client.Client, *models.Stack) (subReconcilerResult, error){
		s.reconcileImagePullSecrets,
		s.reconcileImagePushSecrets,
		s.reconcileGitCredentials,
		s.reconcileDirectConfigUsage,
		s.removeUnusedSecrets,
	}
	for _, fn := range reconcileFns {
		result, err := fn(ctx, clusterClient, stack)
		if err != nil {
			return resultNil, fmt.Errorf("failed to reconcile secrets for stack '%s': %w", stack.ID, err)
		}
		if result != resultNil {
			return result, nil
		}
	}
	return resultNil, nil
}

func (s *secretReconciler) reconcileImagePullSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) (subReconcilerResult, error) {
	if stack.HasImagePullSecrets() {
		for imageUrl, secretID := range stack.GetImagePullSecretIDMap() {
			secret, serr := s.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return resultNil, fmt.Errorf("failed to get secret with ID '%s': %w", secretID, serr)
			}
			clusterSecret, berr := s.secretBuilder.BuildDockerConfigJsonSecretForImage(ctx, secret, imageUrl)
			if berr != nil {
				return resultNil, fmt.Errorf("failed to build docker config JSON for image '%s': %w", imageUrl, berr)
			}
			clusterSecret.Namespace = stack.Namespace
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID

			existingSecret := &corev1.Secret{}
			if err := clusterClient.Get(ctx, client.ObjectKey{
				Name:      clusterSecret.Name,
				Namespace: stack.Namespace,
			}, existingSecret); err != nil {
				if k8sapierrors.IsNotFound(err) {
					s.logger.Infof("creating secret '%s' in namespace '%s'", clusterSecret.Name, stack.Namespace)
					return resultNil, clusterClient.Create(ctx, clusterSecret)
				}
				return resultNil, fmt.Errorf("failed to get secret '%s' in namespace '%s': %w", clusterSecret.Name, stack.Namespace, err)
			}
			if existingSecret.Annotations[models.SecretDataHashAnnotation] != secret.DataHash {
				clusterSecret.ResourceVersion = existingSecret.ResourceVersion
				s.logger.Infof("updating secret '%s' in namespace '%s'", clusterSecret.Name, stack.Namespace)
				return resultNil, clusterClient.Update(ctx, clusterSecret)
			}
		}
	}
	return resultNil, nil
}

func (s *secretReconciler) reconcileImagePushSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) (subReconcilerResult, error) {
	if stack.HasImagePushSecrets() {
		for repoUrl, secretID := range stack.GetImagePushSecretIDMap() {
			secret, serr := s.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return resultNil, fmt.Errorf("failed to get secret with ID '%s': %w", secretID, serr)
			}
			clusterSecret, berr := s.secretBuilder.BuildDockerConfigJsonSecretForRepository(ctx, secret, repoUrl)
			if berr != nil {
				return resultNil, fmt.Errorf("failed to build docker config JSON for repository '%s': %w", repoUrl, berr)
			}
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID
			clusterSecret.Namespace = stack.Namespace

			existingSecret := &corev1.Secret{}
			if err := clusterClient.Get(ctx, client.ObjectKey{
				Name:      clusterSecret.Name,
				Namespace: stack.Namespace,
			}, existingSecret); err != nil {
				if k8sapierrors.IsNotFound(err) {
					return resultNil, clusterClient.Create(ctx, clusterSecret)
				}
				return resultNil, fmt.Errorf("failed to get secret '%s' in namespace '%s': %w", clusterSecret.Name, stack.Namespace, err)
			}
			if existingSecret.Annotations[models.SecretDataHashAnnotation] != secret.DataHash {
				clusterSecret.ResourceVersion = existingSecret.ResourceVersion
				return resultNil, clusterClient.Update(ctx, clusterSecret)
			}
		}
	}

	return resultNil, nil
}

func (s *secretReconciler) reconcileGitCredentials(ctx context.Context, clusterClient client.Client, stack *models.Stack) (subReconcilerResult, error) {
	if stack.HasGitCredentials() {
		for repoUrl, secretID := range stack.GetGitCredentialsMap() {
			secret, serr := s.secretService.InternalGetByID(ctx, secretID)
			if serr != nil {
				return resultNil, fmt.Errorf("failed to get secret with ID '%s': %w", secretID, serr)
			}
			clusterSecret, berr := s.secretBuilder.BuildGitCredentialsSecret(ctx, secret, repoUrl)
			if berr != nil {
				return resultNil, fmt.Errorf("failed to build git credentials secret for repository '%s': %w", repoUrl, berr)
			}
			clusterSecret.Annotations[models.SecretDataHashAnnotation] = secret.DataHash
			clusterSecret.Annotations[models.SecretIDAnnotation] = secret.ID
			clusterSecret.Labels[models.StackIDLabel] = stack.ID
			clusterSecret.Namespace = stack.Namespace

			existingSecret := &corev1.Secret{}
			if err := clusterClient.Get(ctx, client.ObjectKey{
				Name:      clusterSecret.Name,
				Namespace: stack.Namespace,
			}, existingSecret); err != nil {
				if k8sapierrors.IsNotFound(err) {
					return resultNil, clusterClient.Create(ctx, clusterSecret)
				}
				return resultNil, fmt.Errorf("failed to get secret '%s' in namespace '%s': %w", clusterSecret.Name, stack.Namespace, err)
			}
			if existingSecret.Annotations[models.SecretDataHashAnnotation] != secret.DataHash {
				clusterSecret.ResourceVersion = existingSecret.ResourceVersion
				return resultNil, clusterClient.Update(ctx, clusterSecret)
			}
		}
	}
	return resultNil, nil
}

func (s *secretReconciler) removeUnusedSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) (subReconcilerResult, error) {
	secretList := &corev1.SecretList{}
	if err := clusterClient.List(ctx, secretList, client.InNamespace(stack.Namespace), client.MatchingLabels{models.StackIDLabel: stack.ID}); err != nil {
		return resultNil, fmt.Errorf("failed to list secrets in namespace '%s': %w", stack.Namespace, err)
	}

	currentSecretsInUse := stack.SecretsInUse()
	s.logger.Infof("current secrets in use by stack: %v", currentSecretsInUse)

	secretsToDeleteInCluster := lo.Filter(secretList.Items, func(secret corev1.Secret, _ int) bool {
		secretID, ok := secret.Annotations[models.SecretIDAnnotation]
		if ok && !lo.Contains(currentSecretsInUse, secretID) {
			return true
		}
		return false
	})

	for _, secret := range secretsToDeleteInCluster {
		s.logger.Infof("deleting secret '%s' in namespace '%s'", secret.Name, stack.Namespace)
		if err := clusterClient.Delete(ctx, &secret, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			if k8sapierrors.IsNotFound(err) {
				continue
			}
			return resultNil, fmt.Errorf("failed to delete secret '%s' in namespace '%s': %w", secret.Name, stack.Namespace, err)
		}
	}

	existingUsages, err := s.resourceUsageService.GetByStackID(ctx, stack.ID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get resource usages for stack '%s': %w", stack.ID, err)
	}
	directConfigSecretIDs := stack.DirectConfigSecretsInUse()

	for _, usage := range existingUsages {
		if usage.ResourceType != "secret" {
			continue
		}
		if !lo.Contains(directConfigSecretIDs, usage.ResourceID) {
			s.logger.Infof("deleting resource usage for secret ID '%s' in stack '%s'", usage.ResourceID, stack.ID)
			if err := s.resourceUsageService.Delete(ctx, usage.ResourceType, usage.ResourceID, usage.StackID); err != nil {
				return resultNil, fmt.Errorf("failed to delete resource usage for secret ID '%s' in stack '%s': %w", usage.ResourceID, stack.ID, err)
			}
		}
	}

	return resultNil, nil
}

// reconcileDirectConfigUsage tracks secrets referenced by direct config fields
// (pull secrets, push secrets, git secrets) in the resource_usages table for
// delete protection. Connection-sourced secrets are tracked in stack_connections.
func (s *secretReconciler) reconcileDirectConfigUsage(ctx context.Context, _ client.Client, stack *models.Stack) (subReconcilerResult, error) {
	directConfigSecretIDs := stack.DirectConfigSecretsInUse()
	existingUsages, err := s.resourceUsageService.GetByStackID(ctx, stack.ID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get resource usages for stack '%s': %w", stack.ID, err)
	}
	existingSecretIDs := make(map[string]struct{})
	for _, usage := range existingUsages {
		if usage.ResourceType == "secret" {
			existingSecretIDs[usage.ResourceID] = struct{}{}
		}
	}

	for _, secretID := range directConfigSecretIDs {
		if _, exists := existingSecretIDs[secretID]; !exists {
			s.logger.Infof("creating resource usage for direct-config secret '%s' in stack '%s'", secretID, stack.ID)
			if err := s.resourceUsageService.Create(ctx, &models.ResourceUsage{
				ResourceType: "secret",
				ResourceID:   secretID,
				ConsumerType: "stack",
				ConsumerID:   stack.ID,
				StackID:      stack.ID,
			}); err != nil {
				return resultNil, fmt.Errorf("failed to create resource usage for secret '%s' in stack '%s': %w", secretID, stack.ID, err)
			}
		}
	}
	return resultNil, nil
}

func (s *secretReconciler) Name() string {
	return "secret-reconciler"
}

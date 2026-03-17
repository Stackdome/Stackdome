package postgresaddon

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type secretReconciler struct {
	secretService  secretService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

func newSecretReconciler(spec PostgresAddonWorkerSpec) *secretReconciler {
	return &secretReconciler{
		secretService:  spec.SecretService,
		clusterManager: spec.ClusterManager,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-secret"),
	}
}

func (r *secretReconciler) Name() string { return "secret" }

func (r *secretReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	if addon.Initialization.ImportFromExternal == nil {
		return resultNil, nil
	}

	imp := addon.Initialization.ImportFromExternal
	if imp.PasswordSecretID == "" {
		return resultNil, nil
	}

	clusterClient, err := r.clusterManager.GetClient(addon.ClusterID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", err)
	}

	// InternalGetByID returns the secret with decrypted data
	secret, serr := r.secretService.InternalGetByID(ctx, imp.PasswordSecretID)
	if serr != nil {
		return resultNil, fmt.Errorf("failed to get import password secret: %w", serr)
	}

	secretName := addon.ImportPasswordSecretName()
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: addon.Namespace,
			Labels: map[string]string{
				models.PostgresAddonIDLabel: addon.ID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte(secret.Data["password"]),
		},
	}

	existing := &corev1.Secret{}
	if err := clusterClient.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: addon.Namespace,
	}, existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			r.logger.Infof("Creating import password secret '%s'", secretName)
			return resultNil, clusterClient.Create(ctx, k8sSecret)
		}
		return resultNil, fmt.Errorf("failed to get secret '%s': %w", secretName, err)
	}

	return resultNil, nil
}

package postgresaddon

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	barmanapi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"
	barmancloudv1 "github.com/cloudnative-pg/plugin-barman-cloud/api/v1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type objectStoreDependencyReconciler struct {
	objectStoreService objectStoreService
	secretService      secretService
	clusterManager     clustermanager.ClusterManager
	logger             logger.Logger
}

func newObjectStoreDependencyReconciler(spec PostgresAddonWorkerSpec) *objectStoreDependencyReconciler {
	return &objectStoreDependencyReconciler{
		objectStoreService: spec.ObjectStoreService,
		secretService:      spec.SecretService,
		clusterManager:     spec.ClusterManager,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-objectstore"),
	}
}

func (r *objectStoreDependencyReconciler) Name() string { return "objectstore-dependency" }

func (r *objectStoreDependencyReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	if addon.BackupConfig.ObjectStoreID == "" {
		return resultNil, nil
	}

	objectStore, err := r.objectStoreService.GetByID(ctx, addon.BackupConfig.ObjectStoreID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get object store '%s': %w", addon.BackupConfig.ObjectStoreID, err)
	}

	clusterClient, cerr := r.clusterManager.GetClient(addon.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", cerr)
	}

	for _, deployed := range objectStore.Status.DeployedClusters {
		if deployed.ClusterID == addon.ClusterID && deployed.Namespace == addon.Namespace {
			return resultNil, nil
		}
	}

	r.logger.Infof("Deploying ObjectStore '%s' to namespace '%s'", objectStore.Name, addon.Namespace)

	if err := r.ensureCredentialSecret(ctx, clusterClient, objectStore, addon.Namespace); err != nil {
		return resultNil, fmt.Errorf("failed to ensure credential secret: %w", err)
	}

	if err := r.createOrUpdateObjectStoreCR(ctx, clusterClient, objectStore, addon.Namespace); err != nil {
		return resultNil, fmt.Errorf("failed to create ObjectStore CR: %w", err)
	}

	objectStore.Status.DeployedClusters = append(objectStore.Status.DeployedClusters, models.DeployedClusterInfo{
		ClusterID: addon.ClusterID,
		Namespace: addon.Namespace,
	})
	if serr := r.objectStoreService.UpdateStatus(ctx, objectStore.ID, objectStore.Status); serr != nil {
		return resultNil, fmt.Errorf("failed to update object store status: %w", serr)
	}

	return resultNil, nil
}

func credentialSecretName(objectStoreName string) string {
	return fmt.Sprintf("objectstore-%s-credentials", objectStoreName)
}

// ensureCredentialSecret resolves app-level secrets and creates a K8s Secret
// in the target namespace that the barman-cloud ObjectStore CR references.
func (r *objectStoreDependencyReconciler) ensureCredentialSecret(
	ctx context.Context,
	clusterClient client.Client,
	objectStore *models.ObjectStore,
	namespace string,
) error {
	secretData, err := r.resolveCredentialData(ctx, objectStore)
	if err != nil {
		return err
	}

	secretName := credentialSecretName(objectStore.Name)
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"stackdome.io/object-store": objectStore.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	existing := &corev1.Secret{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(k8sSecret), existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			return clusterClient.Create(ctx, k8sSecret)
		}
		return fmt.Errorf("failed to get existing secret '%s': %w", secretName, err)
	}

	existing.Data = secretData
	return clusterClient.Update(ctx, existing)
}

// resolveCredentialData fetches referenced app secrets and builds the K8s Secret data map.
func (r *objectStoreDependencyReconciler) resolveCredentialData(
	ctx context.Context,
	objectStore *models.ObjectStore,
) (map[string][]byte, error) {
	data := make(map[string][]byte)
	cfg := objectStore.Configuration

	if cfg.S3Credentials != nil {
		accessKey, err := r.fetchSecretValue(ctx, cfg.S3Credentials.AccessKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve S3 access key: %w", err)
		}
		data["ACCESS_KEY_ID"] = []byte(accessKey)

		secretKey, err := r.fetchSecretValue(ctx, cfg.S3Credentials.SecretAccessKey)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve S3 secret key: %w", err)
		}
		data["ACCESS_SECRET_KEY"] = []byte(secretKey)

		if cfg.S3Credentials.Region != "" {
			data["REGION"] = []byte(cfg.S3Credentials.Region)
		}
	}

	if cfg.AzureCredentials != nil {
		connStr, err := r.fetchSecretValue(ctx, cfg.AzureCredentials.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve Azure connection string: %w", err)
		}
		data["AZURE_CONNECTION_STRING"] = []byte(connStr)

		if cfg.AzureCredentials.StorageAccountName != "" {
			data["AZURE_STORAGE_ACCOUNT"] = []byte(cfg.AzureCredentials.StorageAccountName)
		}
	}

	if cfg.GCSCredentials != nil {
		creds, err := r.fetchSecretValue(ctx, cfg.GCSCredentials.ServiceAccountCredentials)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve GCS credentials: %w", err)
		}
		data["GOOGLE_APPLICATION_CREDENTIALS"] = []byte(creds)
	}

	return data, nil
}

// fetchSecretValue fetches a decrypted secret value from the app database.
func (r *objectStoreDependencyReconciler) fetchSecretValue(ctx context.Context, ref models.SecretReference) (string, error) {
	// InternalGetByID returns the secret with decrypted data
	secret, serr := r.secretService.InternalGetByID(ctx, ref.SecretID)
	if serr != nil {
		return "", fmt.Errorf("failed to get secret '%s': %w", ref.SecretID, serr)
	}
	val, ok := secret.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found in secret '%s'", ref.Key, ref.SecretID)
	}
	return val, nil
}

func (r *objectStoreDependencyReconciler) createOrUpdateObjectStoreCR(
	ctx context.Context,
	clusterClient client.Client,
	objectStore *models.ObjectStore,
	namespace string,
) error {
	cr := r.buildObjectStoreCR(objectStore, namespace)

	existing := &barmancloudv1.ObjectStore{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(cr), existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			return clusterClient.Create(ctx, cr)
		}
		return fmt.Errorf("failed to get existing ObjectStore CR '%s': %w", objectStore.Name, err)
	}

	existing.Spec = cr.Spec
	return clusterClient.Update(ctx, existing)
}

func (r *objectStoreDependencyReconciler) buildObjectStoreCR(
	objectStore *models.ObjectStore,
	namespace string,
) *barmancloudv1.ObjectStore {
	secretName := credentialSecretName(objectStore.Name)
	cfg := objectStore.Configuration

	cr := &barmancloudv1.ObjectStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectStore.Name,
			Namespace: namespace,
			Labels: map[string]string{
				"stackdome.io/object-store": objectStore.Name,
			},
		},
		Spec: barmancloudv1.ObjectStoreSpec{
			Configuration: barmanapi.BarmanObjectStoreConfiguration{
				DestinationPath: objectStore.DestinationPath,
			},
			RetentionPolicy: objectStore.RetentionPolicy,
		},
	}

	if cfg.S3Credentials != nil {
		cr.Spec.Configuration.BarmanCredentials = barmanapi.BarmanCredentials{
			AWS: &barmanapi.S3Credentials{
				AccessKeyIDReference: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
					Key:                  "ACCESS_KEY_ID",
				},
				SecretAccessKeyReference: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
					Key:                  "ACCESS_SECRET_KEY",
				},
			},
		}
		if cfg.S3Credentials.Region != "" {
			cr.Spec.Configuration.AWS.RegionReference = &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
				Key:                  "REGION",
			}
		}
		if cfg.S3Credentials.Endpoint != "" {
			cr.Spec.Configuration.EndpointURL = cfg.S3Credentials.Endpoint
		}
	}

	if cfg.AzureCredentials != nil {
		cr.Spec.Configuration.BarmanCredentials = barmanapi.BarmanCredentials{
			Azure: &barmanapi.AzureCredentials{
				ConnectionString: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
					Key:                  "AZURE_CONNECTION_STRING",
				},
				StorageAccount: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
					Key:                  "AZURE_STORAGE_ACCOUNT",
				},
			},
		}
	}

	if cfg.GCSCredentials != nil {
		cr.Spec.Configuration.BarmanCredentials = barmanapi.BarmanCredentials{
			Google: &barmanapi.GoogleCredentials{
				ApplicationCredentials: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
					Key:                  "GOOGLE_APPLICATION_CREDENTIALS",
				},
			},
		}
	}

	return cr
}

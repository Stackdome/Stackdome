package builders

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	// DefaultImageCatalogName is the name of the CNPG ImageCatalog resource
	// that must exist in the target namespace before a PostgresCluster can be created.
	DefaultImageCatalogName = "postgres-catalog"
)

// PostgresClusterBuildContext provides pre-resolved references needed
// to build a PostgresCluster CR. The caller is responsible for resolving
// database IDs to cluster resource names before invoking the builder.
type PostgresClusterBuildContext struct {
	// BackupObjectStoreName is the resolved name of the barman ObjectStore
	// K8s resource used for backups. Required when BackupConfig.ObjectStoreID is set.
	BackupObjectStoreName string

	// RestoreObjectStoreName is the resolved name of the barman ObjectStore
	// K8s resource used for restore. Required when restoring from object store.
	RestoreObjectStoreName string

	// RecoverySourceClusterName is the name of the PostgresCluster CR that
	// originally archived backups to the object store. Used as the barman-cloud
	// serverName when restoring from object store.
	RecoverySourceClusterName string

	// RestoreBackupName is the resolved CNPG Backup CR name to restore from.
	// Required when restoring from a specific backup.
	RestoreBackupName string
}

type PostgresClusterBuilder interface {
	BuildPostgresClusterCR(addon *models.PostgresAddon, buildCtx PostgresClusterBuildContext) (*addonsv1alpha1.PostgresCluster, error)
	// BuildImportPasswordSecret builds the K8s Secret that holds the password for
	// importing from an external database. The returned secret must be created in
	// the cluster before the PostgresCluster CR is applied.
	BuildImportPasswordSecret(addon *models.PostgresAddon, password string) *corev1.Secret
}

type postgresClusterBuilder struct{}

func NewPostgresClusterBuilder() PostgresClusterBuilder {
	return &postgresClusterBuilder{}
}

func (b *postgresClusterBuilder) BuildImportPasswordSecret(addon *models.PostgresAddon, password string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon.ImportPasswordSecretName(),
			Namespace: addon.Namespace,
			Labels: map[string]string{
				models.PostgresAddonIDLabel: addon.ID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"password": password,
		},
	}
}

func (b *postgresClusterBuilder) BuildPostgresClusterCR(addon *models.PostgresAddon, buildCtx PostgresClusterBuildContext) (*addonsv1alpha1.PostgresCluster, error) {
	resourceReqs, err := buildPGResourceRequirements(addon.Resources)
	if err != nil {
		return nil, fmt.Errorf("failed to build resource requirements: %w", err)
	}

	minorVersion := addon.PostgresVersion.Minor
	if minorVersion == 0 {
		minorVersion = 1
	}

	cr := &addonsv1alpha1.PostgresCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon.Name,
			Namespace: addon.Namespace,
			Labels: map[string]string{
				models.PostgresAddonIDLabel: addon.ID,
			},
		},
		Spec: addonsv1alpha1.PostgresClusterSpec{
			Instances: addon.Instances.Count,
			ReplicasSpec: addonsv1alpha1.ReplicasSpec{
				SynchronousReplicaDataDurability: addonsv1alpha1.PreferredDataDurability,
			},
			PostgreSQLSpec: &addonsv1alpha1.PostgreSQLSpec{
				ImageCatalogRef: &addonsv1alpha1.ImageCatalogRef{
					Name: DefaultImageCatalogName,
				},
				PostgreSQLMajorVersion:     addon.PostgresVersion.Major,
				PostgreSQLMinorVersion:     minorVersion,
				EnableMajorVersionUpgrades: addon.PostgresVersion.EnableMajorVersionUpgrade,
				EnableMinorVersionUpgrades: addon.PostgresVersion.EnableMinorVersionUpgrade,
				PostgresConf:               addon.Configuration.Parameters,
			},
			EnableSuperuserAccess: addon.Configuration.EnableSuperuserAccess,
			StorageSpec: &addonsv1alpha1.StorageSpec{
				Size:             addon.Storage.Size,
				StorageClassName: addon.Storage.StorageClass,
			},
			ResourceSpec:       resourceReqs,
			Databases:          buildPGDatabaseSpecs(addon.Databases),
			HibernationEnabled: addon.LifecycleConfig.HibernationEnabled,
		},
	}

	if addon.Instances.Placement != nil {
		cr.Spec.InstancePlacementSpec = buildPGPlacementSpec(addon.Instances.Placement)
	}

	if addon.LifecycleConfig.FencingEnabled {
		cr.Spec.FencingSpec = &addonsv1alpha1.FencingSpec{
			FenceCluster: true,
		}
	}

	if addon.BackupConfig.ObjectStoreID != "" {
		cr.Spec.ClusterBackupSpec = buildPGBackupSpec(addon, buildCtx)
	}

	bootstrapSpec := buildPGBootstrapSpec(addon, buildCtx)
	if bootstrapSpec != nil {
		cr.Spec.BootstrapSpec = bootstrapSpec
	}

	return cr, nil
}

func buildPGResourceRequirements(res models.PostgresResources) (corev1.ResourceRequirements, error) {
	reqs := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}

	if res.CPU.Request != "" {
		reqs.Requests[corev1.ResourceCPU] = resource.MustParse(res.CPU.Request)
	}
	if res.CPU.Limit != "" {
		reqs.Limits[corev1.ResourceCPU] = resource.MustParse(res.CPU.Limit)
	}
	if res.Memory.Request != "" {
		reqs.Requests[corev1.ResourceMemory] = resource.MustParse(res.Memory.Request)
	}
	if res.Memory.Limit != "" {
		reqs.Limits[corev1.ResourceMemory] = resource.MustParse(res.Memory.Limit)
	}

	return reqs, nil
}

func buildPGDatabaseSpecs(databases []models.PostgresAddonDatabase) []addonsv1alpha1.DatabaseSpec {
	if len(databases) == 0 {
		return nil
	}

	specs := make([]addonsv1alpha1.DatabaseSpec, len(databases))
	for i, db := range databases {
		dbSpec := addonsv1alpha1.DatabaseSpec{
			Name: db.Name,
		}
		if len(db.Extensions) > 0 {
			dbSpec.Extensions = make([]addonsv1alpha1.ExtensionSpec, len(db.Extensions))
			for j, ext := range db.Extensions {
				dbSpec.Extensions[j] = addonsv1alpha1.ExtensionSpec{
					Name: addonsv1alpha1.PgExtension(ext),
				}
			}
		}
		specs[i] = dbSpec
	}
	return specs
}

func buildPGPlacementSpec(placement *models.PostgresInstancePlacement) *addonsv1alpha1.InstancePlacementSpec {
	spec := &addonsv1alpha1.InstancePlacementSpec{
		TopologyKey:     placement.TopologyKey,
		PlacementPolicy: addonsv1alpha1.InstancePlacementPolicy(placement.Policy),
		NodeSelector:    placement.NodeSelector,
	}

	if len(placement.Tolerations) > 0 {
		spec.Tolerations = make([]corev1.Toleration, len(placement.Tolerations))
		for i, t := range placement.Tolerations {
			spec.Tolerations[i] = corev1.Toleration{
				Key:      t.Key,
				Operator: corev1.TolerationOperator(t.Operator),
				Value:    t.Value,
				Effect:   corev1.TaintEffect(t.Effect),
			}
		}
	}

	return spec
}

func buildPGBackupSpec(addon *models.PostgresAddon, buildCtx PostgresClusterBuildContext) *addonsv1alpha1.ClusterBackupSpec {
	spec := &addonsv1alpha1.ClusterBackupSpec{
		WalArchivingEnabled: addon.BackupConfig.WALArchiving,
		ObjectStoreName:     buildCtx.BackupObjectStoreName,
	}

	if addon.BackupConfig.Schedule != "" {
		spec.ScheduledBaseBackupSpec = &addonsv1alpha1.ScheduledBaseBackupSpec{
			Schedule: addon.BackupConfig.Schedule,
			Enabled:  true,
		}
	}

	if addon.BackupRequestedAt != nil {
		spec.LastBaseBackupRequestedAt = &metav1.Time{Time: addon.BackupRequestedAt.UTC()}
	}

	return spec
}

func buildPGBootstrapSpec(addon *models.PostgresAddon, buildCtx PostgresClusterBuildContext) *addonsv1alpha1.BootstrapSpec {
	init := addon.Initialization

	switch {
	case init.ImportFromExternal != nil:
		return buildImportBootstrapSpec(addon, init.ImportFromExternal)

	case init.RestoreFromObjectStore != nil:
		return buildObjectStoreRecoverySpec(init.RestoreFromObjectStore, buildCtx)

	case init.RestoreFromBackup != nil:
		return &addonsv1alpha1.BootstrapSpec{
			RecoverySpec: &addonsv1alpha1.RecoverySpec{
				BackupSpec: &addonsv1alpha1.BackupSpec{
					BackupName: buildCtx.RestoreBackupName,
				},
			},
		}
	}

	return nil
}

func buildImportBootstrapSpec(addon *models.PostgresAddon, ext *models.PostgresImportFromExternal) *addonsv1alpha1.BootstrapSpec {
	sslMode := "disable"
	if ext.SslMode != nil {
		sslMode = *ext.SslMode
	}

	return &addonsv1alpha1.BootstrapSpec{
		InitDbSpec: &addonsv1alpha1.InitDbSpec{
			Import: &addonsv1alpha1.ImportSpec{
				Databases: ext.DatabasesToImport,
				SourceClusterSpec: &addonsv1alpha1.ExternalClusterSpec{
					Name:    addon.ExternalClusterRefName(),
					Host:    ext.Host,
					User:    ext.Username,
					Port:    ext.Port,
					DbName:  ext.Database,
					SslMode: sslMode,
					Password: corev1alpha1.CredentialSecret{
						SecretRef: corev1.SecretReference{
							Name: addon.ImportPasswordSecretName(),
						},
						Key: models.PasswordSecretKey,
					},
				},
			},
		},
	}
}

func buildObjectStoreRecoverySpec(restore *models.PostgresRestoreFromObjectStore, buildCtx PostgresClusterBuildContext) *addonsv1alpha1.BootstrapSpec {
	recoverySpec := &addonsv1alpha1.RecoveryFromObjectStoreSpec{
		ObjectStoreName:   buildCtx.RestoreObjectStoreName,
		SourceClusterName: buildCtx.RecoverySourceClusterName,
	}

	if restore.RecoveryTargetTime != nil {
		recoverySpec.RecoveryTarget = &addonsv1alpha1.RecoveryTargetSpec{
			RecoveryTargetTime: restore.RecoveryTargetTime.UTC().Format("2006-01-02 15:04:05.000000+00"),
		}
	}

	return &addonsv1alpha1.BootstrapSpec{
		RecoverySpec: &addonsv1alpha1.RecoverySpec{
			ObjectStoreSpec: recoverySpec,
		},
	}
}

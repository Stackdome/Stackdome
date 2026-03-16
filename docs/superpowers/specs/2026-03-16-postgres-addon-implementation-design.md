# PostgresAddon Implementation Design

**Date:** 2026-03-16
**Status:** Draft
**Author:** Claude

## Overview

This document specifies the implementation of PostgresAddon cluster integration, including CR creation, status synchronization, credential access, and Stack consumption.

## Goals

1. Create PostgresCluster CRs in clusters when PostgresAddon is created
2. Create ObjectStore CRs (barman-cloud) when referenced by PostgresAddon backup config
3. Sync status from cluster back to database (including backup records)
4. Provide JIT credential access without storing credentials in API server
5. Enable Stacks to consume PostgresAddon credentials via env vars
6. Track addon usage by stacks for deletion protection

## Non-Goals

- Separate ObjectStore worker (ObjectStore CR creation handled by PostgresAddonWorker)
- Storing credentials in API server database
- Auto-deletion of ObjectStore CR when PostgresAddon is deleted

## Breaking Changes

**PostgresAddonStatus restructure:** The existing `PostgresConnectionInfo` struct (containing `Credentials` with `Username`/`Password` fields) will be replaced with `PostgresAddonConnectionInfo` that stores K8s secret *names* instead of actual credentials. This is a deliberate security improvement - credentials are fetched JIT from the cluster instead of being stored in the API server database.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           API Server                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌──────────────────┐    ┌───────────────────┐  │
│  │ PostgresAddon   │    │ ObjectStore      │    │ PostgresBackup    │  │
│  │ Service (CRUD)  │    │ Service (CRUD)   │    │ Service           │  │
│  └────────┬────────┘    └──────────────────┘    └─────────┬─────────┘  │
│           │                                               │             │
│  ┌────────▼────────────────────────────────────┐          │             │
│  │ PostgresAddonWorker                         │          │             │
│  │  - DeprovisionReconciler                    │          │             │
│  │  - ValidationReconciler                     │          │             │
│  │  - NamespaceReconciler                      │          │             │
│  │  - ObjectStoreDependencyReconciler          │          │             │
│  │  - SecretReconciler                         │          │             │
│  │  - PostgresClusterReconciler                │          │             │
│  │  - LifecycleReconciler                      │          │             │
│  └────────┬────────────────────────────────────┘          │             │
│           │                                               │             │
│  ┌────────▼───────────────────────────────────────────────▼──────────┐  │
│  │                    Cluster Manager                                 │  │
│  └────────┬───────────────────────────────────────────────┬──────────┘  │
│           │                                               │             │
│  ┌────────▼────────┐                             ┌────────▼──────────┐  │
│  │ PostgresAddon   │                             │ PostgresBackup    │  │
│  │ Controller      │                             │ Controller        │  │
│  └─────────────────┘                             └───────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                        ┌───────────────────┐
                        │   Cluster         │
                        │  ┌─────────────┐  │
                        │  │ ObjectStore │  │
                        │  │ (barman)    │  │
                        │  └─────────────┘  │
                        │  ┌─────────────┐  │
                        │  │ Postgres    │  │
                        │  │ Cluster CR  │  │
                        │  └──────┬──────┘  │
                        │         │         │
                        │  ┌──────▼──────┐  │
                        │  │ CNPG        │  │
                        │  │ Operator    │  │
                        │  └──────┬──────┘  │
                        │         │         │
                        │  ┌──────▼──────┐  │
                        │  │ PG Pods +   │  │
                        │  │ K8s Secrets │  │
                        │  └─────────────┘  │
                        └───────────────────┘
```

---

## Component Details

### 1. PostgresAddonWorker

**Location:** `pkg/worker/postgresaddon/`

**Purpose:** Reconciles PostgresAddon models to cluster resources.

#### Worker Structure

```go
type postgresAddonWorker struct {
    postgresAddonService  postgresAddonService
    objectStoreService    objectStoreService
    namespaceService      namespaceService
    secretService         secretService
    clusterManager        clustermanager.ClusterManager
    subReconcilers        []subReconciler
    worker.BaseWorker
}
```

#### Sub-reconcilers (Execution Order)

| Order | Reconciler | Responsibility |
|-------|------------|----------------|
| 1 | DeprovisionReconciler | Handle deletion - delete CR, clean up |
| 2 | ValidationReconciler | Validate spec before cluster operations |
| 3 | NamespaceReconciler | Ensure namespace exists in cluster |
| 4 | ObjectStoreDependencyReconciler | Create ObjectStore CR if backup config references one |
| 5 | SecretReconciler | Create secrets for external DB import credentials |
| 6 | PostgresClusterReconciler | Create/update PostgresCluster CR |
| 7 | LifecycleReconciler | Handle backup/hibernate/fence triggers |

#### GetInput Query

```go
func (w *postgresAddonWorker) GetInput(ctx context.Context) ([]worker.Operand, error) {
    return w.postgresAddonService.InternalList(ctx,
        "status->>'state' IN ?",
        []string{"Pending", "Error", "Deleting"},
    )
}
```

#### Sub-reconciler Error Handling

Each sub-reconciler can return:
- `resultNil` - Continue to next sub-reconciler
- `resultStop` - Stop processing, don't requeue (terminal state)
- `resultRequeue` - Stop processing, requeue immediately
- `resultRequeueAfter(duration)` - Stop processing, requeue after delay
- `error` - Stop processing, requeue with backoff

**Error behavior by sub-reconciler:**

| Sub-reconciler | On Error | Notes |
|----------------|----------|-------|
| DeprovisionReconciler | Requeue with backoff | Retries deletion until success |
| ValidationReconciler | Stop, set Error state | No retry - user must fix spec |
| NamespaceReconciler | Requeue with backoff | Transient cluster errors |
| ObjectStoreDependencyReconciler | Requeue with backoff | ObjectStore CR creation may fail transiently |
| SecretReconciler | Requeue with backoff | Transient cluster errors |
| PostgresClusterReconciler | Requeue with backoff | CR creation/update may fail transiently |
| LifecycleReconciler | Requeue with backoff | Lifecycle operations may fail transiently |

---

### 2. PostgresCluster CR Builder

**Location:** `pkg/builders/postgres_cluster_builder.go`

**Interface:**

```go
type PostgresClusterBuilder interface {
    BuildPostgresClusterCR(addon *models.PostgresAddon) (*addonsv1alpha1.PostgresCluster, error)
}
```

**Field Mapping:**

| PostgresAddon Model | PostgresCluster CR |
|---------------------|-------------------|
| `Instances.Count` | `spec.instances` |
| `PostgresVersion.Major/Minor` | `spec.postgreSQLSpec.postgreSQLVersion` |
| `Storage.Size/Class` | `spec.storageSpec.size/storageClassName` |
| `Resources.CPU/Memory` | `spec.resourceSpec.requests/limits` |
| `Configuration.SuperuserAccess` | `spec.enableSuperuserAccess` |
| `Configuration.Parameters` | `spec.postgreSQLSpec.postgresConf` |
| `BackupConfig.ObjectStoreID` | `spec.clusterBackupSpec.objectStoreName` (resolved) |
| `BackupConfig.Schedule` | `spec.clusterBackupSpec.scheduledBaseBackupSpec.schedule` |
| `BackupConfig.WALArchivingEnabled` | `spec.clusterBackupSpec.walArchivingEnabled` |
| `LifecycleConfig.Hibernation` | `spec.hibernationEnabled` |
| `LifecycleConfig.Fencing` | `spec.fencingSpec.fenceCluster` |
| `Instances.Placement.TopologyKey` | `spec.instancePlacementSpec.topologyKey` |
| `Instances.Placement.Policy` | `spec.instancePlacementSpec.placementPolicy` |
| `Instances.Placement.Tolerations` | `spec.instancePlacementSpec.tolerations` |
| `Instances.Placement.NodeSelector` | `spec.instancePlacementSpec.nodeSelector` |
| `Databases[]` | `spec.databases[]` |
| `Initialization.*` | `spec.bootstrapSpec.*` |

---

### 3. ObjectStore CR Builder

**Location:** `pkg/builders/objectstore_builder.go`

**Note:** The ObjectStore CR is from the barman-cloud project (`github.com/cloudnative-pg/barman-cloud/api/v1`), not a stackdome CRD. This is a third-party CRD that CNPG uses for backup storage configuration.

**Interface:**

```go
import barmancloudv1 "github.com/cloudnative-pg/barman-cloud/api/v1"

type ObjectStoreBuilder interface {
    BuildObjectStoreCR(ctx context.Context, objectStore *models.ObjectStore, namespace string) (*barmancloudv1.ObjectStore, error)
    BuildCredentialSecret(ctx context.Context, objectStore *models.ObjectStore, namespace string) (*corev1.Secret, error)
}
```

**Field Mapping:**

| API Server Model | Barman ObjectStore CR |
|------------------|----------------------|
| `S3Credentials.Region` | `spec.configuration.s3.region` |
| `S3Credentials.Endpoint` | `spec.configuration.s3.endpointURL` |
| `AzureCredentials.*` | `spec.configuration.azureCredentials.*` |
| `GCSCredentials.*` | `spec.configuration.googleCredentials.*` |
| `DestinationPath` | `spec.configuration.destinationPath` |
| `RetentionPolicy` | `spec.retentionPolicy` |

---

### 4. ObjectStore CR Lifecycle

ObjectStore CRs are tied to the ObjectStore DB record, not PostgresAddon.

**Flow:**

```
ObjectStore Created (DB):
  → No cluster action

PostgresAddon Created (references ObjectStore):
  → ObjectStoreDependencyReconciler creates ObjectStore CR in cluster (if not exists)
  → Updates ObjectStore status with deployed cluster info

PostgresAddon Deleted:
  → ObjectStore CR stays in cluster

ObjectStore Deleted (DB):
  → Clean up ObjectStore CR from all deployed clusters
  → Validation: Fail if referenced by any PostgresAddon
```

**ObjectStore Status Addition:**

The existing `ObjectStoreStatus` struct has `State`, `Message`, and `Conditions` fields. We add `DeployedClusters` to track where the CR has been deployed:

```go
// Existing fields preserved, new field added
type ObjectStoreStatus struct {
    State            string                `json:"state,omitempty"`
    Message          string                `json:"message,omitempty"`
    Conditions       []Condition           `json:"conditions,omitempty"`
    DeployedClusters []DeployedClusterInfo `json:"deployedClusters,omitempty"`  // NEW
}

type DeployedClusterInfo struct {
    ClusterID string `json:"clusterId"`
    Namespace string `json:"namespace"`
}
```

**Deletion Validation:**

```go
func (s *objectStoreService) Delete(ctx context.Context, id string) error {
    // Check not referenced by any PostgresAddon
    if referenced, _ := s.isReferencedByAddon(ctx, id); referenced {
        return errors.BadRequest("ObjectStore is in use by PostgresAddon")
    }

    // Clean up CRs from all deployed clusters
    objectStore, _ := s.store.GetByID(ctx, id)
    for _, deployed := range objectStore.Status.DeployedClusters {
        client := s.clusterManager.GetClient(deployed.ClusterID)
        // Delete ObjectStore CR and credential secrets
    }

    return s.store.Delete(ctx, id)
}
```

---

## Controllers (Status Synchronization)

### 5. PostgresAddon Controller (Enhanced)

**Location:** `pkg/controllers/postgres_addon/postgres_addon_controller.go`

**Enhancement:** Full status mapping including connection info and K8s secret names.

```go
func (r *postgresAddonReconciler) mapToPostgresAddonStatus(cr *addonsv1alpha1.PostgresCluster) *models.PostgresAddonStatus {
    status := &models.PostgresAddonStatus{
        State:      cr.Status.Phase,
        Conditions: convertConditions(cr.Status.Conditions),
    }

    // Map connection info if outputs are available
    if cr.Status.Outputs != nil {
        status.ConnectionInfo = &models.PostgresAddonConnectionInfo{
            WriteService: cr.Status.Outputs.WriteService,
            ReadService:  cr.Status.Outputs.ReadService,
            ClusterSecrets: &models.PostgresAddonClusterSecrets{
                SuperuserSecret:     cr.Status.Outputs.SuperUserCredentialSecret,  // *string in CRD
                UserSecrets:         cr.Status.Outputs.UserCredentialSecretNames,  // map[string]string in CRD
                CACertificateSecret: cr.Status.Outputs.ClientCASecret,
            },
        }

        // Map cluster connection info if available
        if cr.Status.Outputs.ClusterConnection != nil {
            status.ConnectionInfo.Host = cr.Status.Outputs.ClusterConnection.Host
            status.ConnectionInfo.Port = cr.Status.Outputs.ClusterConnection.Port
            status.ConnectionInfo.SSLMode = cr.Status.Outputs.ClusterConnection.SSLMode
        }

        status.Databases = mapDatabaseInfo(cr.Status.Outputs.Databases)
    }

    return status
}
```

**Status Model Additions:**

```go
// pkg/models/postgres_addon.go

type PostgresAddonConnectionInfo struct {
    Host           string                       `json:"host"`
    Port           int32                        `json:"port"`
    SSLMode        string                       `json:"sslMode"`
    WriteService   string                       `json:"writeService"`
    ReadService    string                       `json:"readService"`
    ClusterSecrets *PostgresAddonClusterSecrets `json:"clusterSecrets,omitempty"`
}

type PostgresAddonClusterSecrets struct {
    SuperuserSecret     *string           `json:"superuserSecret,omitempty"`
    UserSecrets         map[string]string `json:"userSecrets,omitempty"`
    CACertificateSecret string            `json:"caCertificateSecret,omitempty"`
}
```

---

### 6. PostgresBackup Controller (New)

**Location:** `pkg/controllers/postgres_backup/postgres_backup_controller.go`

**Purpose:** Watch CNPG Backup CRs and sync to PostgresBackup records in DB.

```go
type postgresBackupReconciler struct {
    Client                client.Client
    Log                   logger.Logger
    PostgresBackupService services.PostgresBackupService
    PostgresAddonService  services.PostgresAddonService
}

func (r *postgresBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    backup := &cnpgv1.Backup{}
    r.Client.Get(ctx, req.NamespacedName, backup)

    addonID := r.findAddonIDFromBackup(backup)

    existing, err := r.PostgresBackupService.GetByName(ctx, addonID, backup.Name)

    if err.Is404() {
        r.PostgresBackupService.Create(ctx, &models.PostgresBackup{
            PostgresAddonID: addonID,
            Name:            backup.Name,
            Type:            mapBackupType(backup),
            Phase:           string(backup.Status.Phase),
            StartedAt:       backup.Status.StartedAt,
            StoppedAt:       backup.Status.StoppedAt,
            Size:            backup.Status.BackupSize,
        })
    } else {
        r.PostgresBackupService.UpdateStatus(ctx, existing.ID, ...)
    }
}
```

---

## JIT Credentials Endpoint

### 7. Credentials API

**Endpoints:**

```
GET /api/v1/organizations/{org_id}/addons/postgres/{id}/credentials/{database}
GET /api/v1/organizations/{org_id}/addons/postgres/{id}/credentials?superuser=true
```

**Response:**

```json
{
  "database": "myapp",
  "host": "pg-myaddon-17.namespace.svc.cluster.local",
  "port": 5432,
  "username": "myapp_owner",
  "password": "generated-password-here",
  "sslMode": "verify-full",
  "connectionString": "postgresql://myapp_owner:xxx@host:5432/myapp?sslmode=verify-full",
  "caCertificate": "-----BEGIN CERTIFICATE-----\n..."
}
```

**Service Implementation:**

```go
func (s *postgresAddonService) GetCredentials(
    ctx context.Context,
    addonID string,
    database string,
    superuser bool,
) (*models.PostgresCredentials, error) {
    addon, _ := s.store.GetByID(ctx, addonID)
    client, _ := s.clusterManager.GetClient(addon.ClusterID)

    var secretName string
    if superuser {
        secretName = *addon.Status.ConnectionInfo.ClusterSecrets.SuperuserSecret
    } else {
        secretName = addon.Status.ConnectionInfo.ClusterSecrets.UserSecrets[database]
    }

    // Read secret from cluster (JIT)
    secret := &corev1.Secret{}
    client.Get(ctx, client.ObjectKey{
        Name:      secretName,
        Namespace: addon.Namespace,
    }, secret)

    caSecret := &corev1.Secret{}
    client.Get(ctx, ..., caSecret)

    return &models.PostgresCredentials{
        Database:         database,
        Host:             addon.Status.ConnectionInfo.Host,
        Port:             addon.Status.ConnectionInfo.Port,
        Username:         string(secret.Data["username"]),
        Password:         string(secret.Data["password"]),
        SSLMode:          addon.Status.ConnectionInfo.SSLMode,
        ConnectionString: buildConnectionString(...),
        CACertificate:    string(caSecret.Data["ca.crt"]),
    }, nil
}
```

**Authorization:**

| Action | Permission Required |
|--------|---------------------|
| Get database credentials | `Execute` on PostgresAddon |
| Get superuser credentials | `Execute` on PostgresAddon + addon must have `Configuration.SuperuserAccess: true` |

**Superuser access logic:**
1. User must have `Execute` permission on the PostgresAddon resource
2. The PostgresAddon must have `Configuration.SuperuserAccess` enabled (set at creation time)
3. If addon has `SuperuserAccess: false`, superuser credentials endpoint returns 403

```go
func (s *postgresAddonService) GetCredentials(..., superuser bool) error {
    if superuser {
        if !addon.Configuration.SuperuserAccess {
            return errors.Forbidden("superuser access not enabled for this addon")
        }
    }
    // ... fetch credentials
}
```

---

## Stack Consumption

### 8. EnvFromAddons

Stacks consume addon credentials via explicit field-to-env-var mapping.

**Model Changes:**

```go
// pkg/models/stack_resource.go

type ExecutionConfig struct {
    Command            []string                 `json:"command,omitempty"`
    Args               []string                 `json:"args,omitempty"`
    Env                []EnvVar                 `json:"env,omitempty"`
    EnvVarsFromSecrets []EnvSecretReference     `json:"env_vars_from_secrets,omitempty"`
    EnvFromAddons      []AddonEnvSource         `json:"env_from_addons,omitempty"`
}

// Only one addon-specific field should be set
type AddonEnvSource struct {
    Postgres *PostgresAddonEnvSource `json:"postgres,omitempty"`
    // Future:
    // Redis   *RedisAddonEnvSource   `json:"redis,omitempty"`
    // Kafka   *KafkaAddonEnvSource   `json:"kafka,omitempty"`
}

type PostgresAddonEnvSource struct {
    AddonID    string            `json:"addon_id"`
    Database   string            `json:"database"`
    EnvMapping map[string]string `json:"env_mapping"`
}
```

**Available Fields for Postgres:**

| Field | Description |
|-------|-------------|
| `host` | Write service hostname |
| `port` | Port (5432) |
| `username` | Database owner username |
| `password` | Password from K8s secret |
| `database` | Database name |
| `sslmode` | SSL mode (verify-full) |
| `connectionString` | Full `postgresql://...` URL |
| `caCertificate` | CA cert for SSL |

**Example Usage:**

```json
{
  "name": "my-app",
  "stackResources": [{
    "name": "api",
    "imageConfig": { "image": "myapp:latest" },
    "executionConfig": {
      "envFromAddons": [{
        "postgres": {
          "addonId": "uuid-of-addon",
          "database": "myapp",
          "envMapping": {
            "host": "DB_HOST",
            "port": "DB_PORT",
            "username": "DB_USER",
            "password": "DB_PASS",
            "database": "DB_NAME",
            "connectionString": "DATABASE_URL"
          }
        }
      }]
    }
  }]
}
```

**Result in container:**

```
DB_HOST=pg-myaddon-17-rw.namespace.svc.cluster.local
DB_PORT=5432
DB_USER=myapp_owner
DB_PASS=xxx
DB_NAME=myapp
DATABASE_URL=postgresql://myapp_owner:xxx@host:5432/myapp?sslmode=verify-full
```

**Validation:**

```go
var PostgresAddonEnvFields = []string{
    "host", "port", "username", "password",
    "database", "sslmode", "connectionString", "caCertificate",
}

func validatePostgresEnvMapping(mapping map[string]string) error {
    for field := range mapping {
        if !contains(PostgresAddonEnvFields, field) {
            return fmt.Errorf("unknown postgres field: %s", field)
        }
    }
    return nil
}
```

**Runtime Behavior and Error Handling:**

The Stack worker resolves addon references at **Stack reconciliation time** (not container startup):

1. **Addon validation** - Stack creation/update validates addon exists and is in same organization
2. **Addon readiness check** - Stack worker waits for addon to reach `Ready` state before proceeding
3. **Secret mounting** - Worker configures StackResource CR to mount the K8s secret created by CNPG
4. **Credential resolution** - Actual credential values are resolved by the cluster-agent at pod creation time

**Error scenarios:**

| Scenario | Behavior |
|----------|----------|
| Addon doesn't exist | Stack creation fails validation |
| Addon in different org | Stack creation fails validation |
| Addon in `Pending` state | Stack worker requeues, waits for addon to be Ready |
| Addon in `Error` state | Stack worker requeues with backoff |
| Database not found in addon | Stack creation fails validation |
| Cluster unreachable | Stack remains in Pending, retries on next reconcile |

```go
func (r *addonEnvReconciler) Reconcile(ctx context.Context, stack *models.Stack) (result, error) {
    for _, resource := range stack.StackResources {
        for _, addonEnv := range resource.ExecutionConfig.EnvFromAddons {
            if addonEnv.Postgres != nil {
                addon, err := r.postgresAddonService.GetByID(ctx, addonEnv.Postgres.AddonID)
                if err != nil {
                    return resultNil, err
                }

                // Wait for addon to be ready
                if addon.Status.State != "Ready" {
                    return resultRequeue, nil
                }

                // Validate database exists
                if !addon.HasDatabase(addonEnv.Postgres.Database) {
                    return resultNil, fmt.Errorf("database %s not found in addon", addonEnv.Postgres.Database)
                }

                // Resolve and add to resource spec
                // ...
            }
        }
    }
    return resultNil, nil
}
```

---

## Addon Usage Tracking

### 9. AddonType Definition

```go
// pkg/models/addon_type.go
type AddonType string

const (
    AddonTypePostgres AddonType = "postgres"
    // Future addon types:
    // AddonTypeRedis   AddonType = "redis"
    // AddonTypeKafka   AddonType = "kafka"
)
```

### 10. AddonUsage Association Table

Track which addons are used by which stacks for deletion protection.

**Model:**

```go
// pkg/models/addon_usage.go
type AddonUsage struct {
    AddonType       AddonType `gorm:"not null"`
    AddonID         string    `gorm:"not null"`
    StackID         string    `gorm:"not null"`
    StackResourceID string    `gorm:"not null"`
}
```

**Store Interface:**

```go
// pkg/stores/addon_usage_store.go
type AddonUsageStore interface {
    Create(ctx context.Context, usage *AddonUsage) error
    Delete(ctx context.Context, addonType AddonType, addonID, stackID, resourceID string) error
    GetByAddonID(ctx context.Context, addonType AddonType, addonID string) ([]*AddonUsage, error)
    GetByStackID(ctx context.Context, stackID string) ([]*AddonUsage, error)
    IsAddonInUse(ctx context.Context, addonType AddonType, addonID string) (bool, error)
}
```

**Usage in Deletion:**

```go
func (s *postgresAddonService) Delete(ctx context.Context, id string) error {
    inUse, _ := s.addonUsageStore.IsAddonInUse(ctx, AddonTypePostgres, id)
    if inUse {
        return errors.BadRequest("PostgresAddon is in use by one or more stacks")
    }
    // Proceed with deletion
}
```

**Stack Worker Reconciliation:**

```go
func (r *addonEnvReconciler) reconcileAddonUsage(ctx context.Context, stack *models.Stack) error {
    existingUsages, _ := r.addonUsageStore.GetByStackID(ctx, stack.ID)
    newUsages := r.computeAddonUsages(stack)

    // Add new usages
    for _, usage := range newUsages {
        if !contains(existingUsages, usage) {
            r.addonUsageStore.Create(ctx, usage)
        }
    }

    // Remove stale usages
    for _, existing := range existingUsages {
        if !contains(newUsages, existing) {
            r.addonUsageStore.Delete(ctx, existing.AddonType, existing.AddonID, existing.StackID, existing.StackResourceID)
        }
    }
    return nil
}
```

**Migration:**

```sql
CREATE TABLE addon_usages (
    addon_type VARCHAR(50) NOT NULL,
    addon_id UUID NOT NULL,
    stack_id UUID NOT NULL,
    stack_resource_id UUID NOT NULL,
    PRIMARY KEY (addon_type, addon_id, stack_id, stack_resource_id)
);

CREATE INDEX idx_addon_usages_addon ON addon_usages(addon_type, addon_id);
CREATE INDEX idx_addon_usages_stack ON addon_usages(stack_id);
```

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `pkg/worker/postgresaddon/postgres_addon_worker.go` | Main worker |
| `pkg/worker/postgresaddon/deprovision_reconciler.go` | Deletion handling |
| `pkg/worker/postgresaddon/validation_reconciler.go` | Spec validation |
| `pkg/worker/postgresaddon/namespace_reconciler.go` | Namespace creation |
| `pkg/worker/postgresaddon/objectstore_dependency_reconciler.go` | ObjectStore CR creation |
| `pkg/worker/postgresaddon/secret_reconciler.go` | External DB secrets |
| `pkg/worker/postgresaddon/postgres_cluster_reconciler.go` | PostgresCluster CR |
| `pkg/worker/postgresaddon/lifecycle_reconciler.go` | Backup/hibernate/fence |
| `pkg/worker/postgresaddon/interfaces.go` | Service interfaces |
| `pkg/models/addon_type.go` | AddonType enum definition |
| `pkg/builders/postgres_cluster_builder.go` | CR builder |
| `pkg/builders/objectstore_builder.go` | ObjectStore CR builder |
| `pkg/controllers/postgres_backup/postgres_backup_controller.go` | Backup sync |
| `pkg/worker/stack/addon_env_reconciler.go` | Stack addon env resolution |
| `pkg/models/addon_usage.go` | Addon usage model |
| `pkg/stores/addon_usage_store.go` | Addon usage store interface |
| `pkg/stores/pgstore/addon_usage.go` | Addon usage store implementation |

### Modified Files

| File | Changes |
|------|---------|
| `pkg/models/postgres_addon.go` | Add ConnectionInfo, ClusterSecrets structs |
| `pkg/models/object_store.go` | Add DeployedClusters to status |
| `pkg/models/stack_resource.go` | Add EnvFromAddons, AddonEnvSource structs |
| `pkg/services/postgres_addon_service.go` | Add GetCredentials method |
| `pkg/services/object_store_service.go` | Add deployment tracking, deletion cleanup |
| `pkg/handlers/postgres_addon_handler.go` | Add GetCredentials handler |
| `pkg/controllers/postgres_addon/postgres_addon_controller.go` | Enhanced status mapping |
| `pkg/validator/stack/stack_validator.go` | Add addon env validation |
| `pkg/worker/stack/stack_worker.go` | Add AddonEnvReconciler |
| `cmd/server/routes.go` | Add credentials endpoint |
| `config/openapi/stackdome_api.yaml` | Add credentials endpoint, EnvFromAddons schema |

---

## Migrations

### Migration 1: ObjectStore Status

```go
func addObjectStoreStatus() *gormigrate.Migration {
    return &gormigrate.Migration{
        ID: "202603161000",
        Migrate: func(db *gorm.DB) error {
            return db.Exec(`
                ALTER TABLE object_stores
                ADD COLUMN IF NOT EXISTS status JSONB DEFAULT '{}'
            `).Error
        },
    }
}
```

### Migration 2: Addon Usages Table

```go
func createAddonUsagesTable() *gormigrate.Migration {
    return &gormigrate.Migration{
        ID: "202603161001",
        Migrate: func(db *gorm.DB) error {
            return db.Exec(`
                CREATE TABLE addon_usages (
                    addon_type VARCHAR(50) NOT NULL,
                    addon_id UUID NOT NULL,
                    stack_id UUID NOT NULL,
                    stack_resource_id UUID NOT NULL,
                    PRIMARY KEY (addon_type, addon_id, stack_id, stack_resource_id)
                );
                CREATE INDEX idx_addon_usages_addon ON addon_usages(addon_type, addon_id);
                CREATE INDEX idx_addon_usages_stack ON addon_usages(stack_id);
            `).Error
        },
    }
}
```

---

## Testing

### Unit Tests

- PostgresClusterBuilder field mapping
- ObjectStoreBuilder field mapping
- Credential service JIT fetch
- EnvMapping validation
- AddonEnvSource validation
- AddonUsage CRUD operations

### Integration Tests

- PostgresAddon creation → PostgresCluster CR created
- ObjectStore referenced → ObjectStore CR created
- Status sync from cluster → DB updated
- Backup created → PostgresBackup record created
- Credentials endpoint → returns JIT credentials
- Stack with envFromAddons → env vars injected
- Addon deletion blocked when in use by stack
- Stack deletion → addon usage records removed

---

## Security Considerations

1. **Credentials never stored in API server** - JIT fetch from cluster only
2. **Superuser access requires elevated permission** - Separate authz check
3. **ObjectStore credentials** - Stored encrypted in API server Secrets, synced to K8s secrets
4. **Cluster connectivity required** - Credentials unavailable if cluster unreachable
5. **Addon usage tracking** - Prevents accidental deletion of in-use addons

---

## Open Questions

None at this time.

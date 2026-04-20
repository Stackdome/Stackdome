# PostgresAddon Implementation Design

**Date:** 2026-03-16
**Status:** Implemented (2026-04-12)
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

**Migration Strategy:** Since the status field is JSONB and re-populated from the cluster on every reconciliation, this is a safe breaking change:
1. Update the model structs in code
2. Deploy the new code - existing status data becomes partially unparseable
3. PostgresAddonController will overwrite status with new structure on next reconciliation
4. No data migration needed - cluster is the source of truth for status

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
│  │  - NamespaceReconciler                      │          │             │
│  │  - ObjectStoreDependencyReconciler          │          │             │
│  │  - SecretReconciler                         │          │             │
│  │  - PostgresClusterReconciler                │          │             │
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
| 1 | DeprovisionReconciler | Handle deletion — check addon usage, delete CR, delete from DB |
| 2 | NamespaceReconciler | Ensure namespace exists in cluster |
| 3 | ObjectStoreDependencyReconciler | Deploy barman-cloud ObjectStore CR and credential K8s Secret |
| 4 | SecretReconciler | Create K8s Secret for external DB import passwords |
| 5 | PostgresClusterReconciler | Build context, create/update PostgresCluster CR |

**Note:** ValidationReconciler and LifecycleReconciler were removed. Lifecycle operations (hibernation, fencing, backup triggers) are handled via CR spec fields set by the PostgresClusterReconciler's builder. The cluster-agent operator handles the actual transitions.

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
| NamespaceReconciler | Requeue with backoff | Transient cluster errors |
| ObjectStoreDependencyReconciler | Requeue with backoff | ObjectStore CR creation may fail transiently |
| SecretReconciler | Requeue with backoff | Transient cluster errors |
| PostgresClusterReconciler | Requeue with backoff | CR creation/update may fail transiently |

---

### 2. PostgresCluster CR Builder

**Location:** `pkg/builders/postgres_cluster_builder.go`

**Interface:**

```go
type PostgresClusterBuildContext struct {
    BackupObjectStoreName     string // Resolved barman ObjectStore name for backups
    RestoreObjectStoreName    string // Resolved barman ObjectStore name for restore
    RecoverySourceClusterName string // PostgresCluster CR name that archived backups (barman-cloud serverName)
    RestoreBackupName         string // CNPG Backup CR name to restore from
}

type PostgresClusterBuilder interface {
    BuildPostgresClusterCR(addon *models.PostgresAddon, buildCtx PostgresClusterBuildContext) (*addonsv1alpha1.PostgresCluster, error)
    BuildImportPasswordSecret(addon *models.PostgresAddon, password string) *corev1.Secret
}
```

The `PostgresClusterReconciler` resolves DB IDs to cluster resource names in a `buildContext()` method, then passes the pre-resolved context to the pure builder function. Key resolution:
- `BackupObjectStoreName`: Looked up via `objectStoreService.GetByID(addon.BackupConfig.ObjectStoreID)`
- `RecoverySourceClusterName`: Looked up via `postgresAddonService.GetPostgresAddon(addon.Initialization.RestoreFromObjectStore.SourcePostgresAddonID)` — uses `sourceAddon.Name` (which equals the CR name)

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

**Location:** Inline in `pkg/worker/postgresaddon/objectstore_dependency_reconciler.go`

**Note:** The ObjectStore CR is from the barman-cloud plugin project (`github.com/cloudnative-pg/plugin-barman-cloud/api/v1`), not a stackdome CRD. The `barmancloudv1` scheme is registered in `pkg/clustermanager/manager.go` so the cluster client can operate on ObjectStore CRs.

**Credential Resolution Chain:**
1. App-level encrypted secrets (SecretReference with SecretID + Key) are decrypted via `InternalGetByID`
2. Raw credential values are created as a K8s Secret (`objectstore-<name>-credentials`) in the target namespace
3. The ObjectStore CR references the K8s Secret via `SecretKeySelector`

**Provider Support:**

| Provider | Secret Keys | CR Fields |
|----------|-------------|-----------|
| S3 | `ACCESS_KEY_ID`, `ACCESS_SECRET_KEY`, optional `REGION` | `spec.configuration.s3`, custom `endpointURL` for S3-compatible stores |
| Azure | `AZURE_CONNECTION_STRING`, `AZURE_STORAGE_ACCOUNT` | `spec.configuration.azureCredentials` |
| GCS | `GOOGLE_APPLICATION_CREDENTIALS` (service account JSON key) | `spec.configuration.googleCredentials` |

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
                UserSecrets:         cr.Status.Outputs.UserCredentialSecrets,      // map[string]string (json: userCredentialSecretNames)
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

    addonID := r.findAddonIDFromBackup(ctx, backup)
    if addonID == "" {
        r.Log.Infof("backup %s not associated with any PostgresAddon, skipping", backup.Name)
        return ctrl.Result{}, nil
    }

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

// findAddonIDFromBackup resolves the PostgresAddon ID from a CNPG Backup.
// It looks up the owning CNPG Cluster, then finds the PostgresCluster CR that created it,
// and extracts the addon ID from its labels.
func (r *postgresBackupReconciler) findAddonIDFromBackup(ctx context.Context, backup *cnpgv1.Backup) string {
    // CNPG Backup references the cluster name in spec.cluster.name
    clusterName := backup.Spec.Cluster.Name

    // Find the PostgresCluster CR that owns this CNPG Cluster
    // The CNPG Cluster is named: <postgres-cluster-cr-name>-<major-version>
    // We need to find PostgresCluster CRs in this namespace and match
    pgClusterList := &addonsv1alpha1.PostgresClusterList{}
    r.Client.List(ctx, pgClusterList, client.InNamespace(backup.Namespace))

    for _, pgCluster := range pgClusterList.Items {
        if pgCluster.CnpgClusterName() == clusterName {
            return pgCluster.Labels[models.PostgresAddonIDLabel]
        }
    }
    return ""
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

**Service Interface Addition:**

```go
// pkg/services/postgres_addon_service.go
type PostgresAddonService interface {
    // ... existing methods ...
    GetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, error)
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

**Error Handling:**

| Scenario | HTTP Status | Error Message |
|----------|-------------|---------------|
| Addon not found | 404 | "PostgresAddon not found" |
| Database not in addon | 404 | "Database 'X' not found in addon" |
| Addon not ready | 503 | "Addon not ready, status: X" |
| Cluster unreachable | 503 | "Unable to connect to cluster" |
| Secret not found in cluster | 500 | "Credential secret not found in cluster" |
| Superuser access disabled | 403 | "Superuser access not enabled for this addon" |
| Permission denied | 403 | "Permission denied" |

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

// HasEnvFromAddons returns true if any addon env sources are configured
func (s *StackResource) HasEnvFromAddons() bool {
    return s.ExecutionConfig != nil && len(s.ExecutionConfig.EnvFromAddons) > 0
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
3. **Secret mounting** - Worker configures StackResource CR with addon secret references
4. **Credential resolution** - Cluster-agent resolves credentials at pod creation time

**StackResource CR Secret Mounting:**

The Stack worker adds addon credential information to the StackResource CR spec. The cluster-agent then:
1. Reads the user credential secret name from PostgresCluster CR status
2. Mounts that secret into the pod using `envFrom` or individual `env` entries
3. Maps addon fields to env vars based on the `EnvMapping` configuration

```go
// Stack worker adds this to StackResource CR spec
type AddonSecretReference struct {
    AddonType      string            `json:"addonType"`
    SecretName     string            `json:"secretName"`     // K8s secret name in same namespace
    EnvMapping     map[string]string `json:"envMapping"`     // field -> env var name
}

// StackResource CR spec (in cluster-agent)
type StackResourceSpec struct {
    // ... existing fields ...
    AddonSecrets []AddonSecretReference `json:"addonSecrets,omitempty"`
}
```

The cluster-agent's StackResource controller reads `addonSecrets` and configures the pod spec with appropriate `env` entries that reference the K8s secrets created by CNPG.

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
    ID              string    `gorm:"primary_key;default:gen_random_uuid()"`
    AddonType       AddonType `gorm:"not null;index:idx_addon_usages_addon"`
    AddonID         string    `gorm:"not null;index:idx_addon_usages_addon"`
    StackID         string    `gorm:"not null;index:idx_addon_usages_stack"`
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    addon_type VARCHAR(50) NOT NULL,
    addon_id UUID NOT NULL,
    stack_id UUID NOT NULL,
    stack_resource_id UUID NOT NULL,
    UNIQUE (addon_type, addon_id, stack_id, stack_resource_id)
);

CREATE INDEX idx_addon_usages_addon ON addon_usages(addon_type, addon_id);
CREATE INDEX idx_addon_usages_stack ON addon_usages(stack_id);
```

---

## Files to Create/Modify

### New Files (Implemented)

| File | Purpose |
|------|---------|
| `pkg/worker/postgresaddon/postgres_addon_worker.go` | Main worker with reconciler chain |
| `pkg/worker/postgresaddon/deprovision_reconciler.go` | Deletion: usage check, CR delete, DB delete |
| `pkg/worker/postgresaddon/namespace_reconciler.go` | Ensures namespace exists in cluster |
| `pkg/worker/postgresaddon/objectstore_dependency_reconciler.go` | Deploys barman-cloud ObjectStore CR + credential Secret |
| `pkg/worker/postgresaddon/secret_reconciler.go` | Creates K8s Secret for external DB import passwords |
| `pkg/worker/postgresaddon/postgres_cluster_reconciler.go` | Builds context, creates/updates PostgresCluster CR |
| `pkg/worker/postgresaddon/types.go` | Sub-reconciler result types and flow control |
| `pkg/worker/postgresaddon/interfaces.go` | Narrow service interfaces for dependency injection |
| `pkg/models/addon_usage.go` | AddonUsage model and AddonType enum |
| `pkg/builders/postgres_cluster_builder.go` | PostgresCluster CR builder (pure function, no I/O) |
| `pkg/stores/addon_usage_store.go` | AddonUsage store interface |
| `pkg/stores/pgstore/addon_usage.go` | AddonUsage pgstore implementation |
| `pkg/db/migrations/202603161001_create_addon_usages_table.go` | Migration for addon_usages table |

### Removed Files

| File | Reason |
|------|--------|
| `pkg/worker/postgresaddon/validation_reconciler.go` | Never created — validation handled at API layer |
| `pkg/worker/postgresaddon/lifecycle_reconciler.go` | Removed — lifecycle ops handled via CR spec fields in builder |
| `pkg/builders/objectstore_builder.go` | Never created — ObjectStore CR building is inline in the reconciler |

### Modified Files (Implemented)

| File | Changes |
|------|---------|
| `pkg/models/postgres_addon.go` | Restructured status for JIT credentials, added `ExternalClusterRefName()`, `ImportPasswordSecretName()`, `HasDatabase()` |
| `pkg/models/object_store.go` | Added `DeployedClusters` to status |
| `pkg/models/stack_resource.go` | Added `EnvFromAddons`, `AddonEnvSource`, `PostgresAddonEnvSource` types |
| `pkg/services/postgres_addon_service.go` | Added `GetCredentials`, `SecretService` dependency, import password validation, deletion protection |
| `pkg/services/object_store_service.go` | Added `UpdateStatus` method |
| `pkg/stores/object_store_store.go` | Added `UpdateStatus` to interface |
| `pkg/stores/pgstore/object_store.go` | Added `UpdateStatus` implementation |
| `pkg/handlers/postgres_addon_handler.go` | Added `GetCredentials` handler |
| `pkg/controllers/postgres_addon/postgres_addon_controller.go` | Full status mapping with hash comparison |
| `pkg/clustermanager/manager.go` | Registered `addonsv1alpha1` and `barmancloudv1` schemes |
| `cmd/server/routes.go` | Added `GET /{id}/credentials/{database}` route |
| `cmd/environment/development_environment.go` | Wired worker, `ClusterManager` + `SecretService` injection |
| `cmd/environment/test_environment.go` | Same wiring as development |

---

## Migrations

### Migration 1: Add DeployedClusters to ObjectStore Status

**File:** `pkg/db/migrations/202603161000_add_deployed_clusters_to_object_store_status.go`

The ObjectStore table already has a `status` JSONB column. No schema change needed - the `DeployedClusters` field will be stored within the existing JSONB when the model is updated. This is a code-only change.

### Migration 2: Create Addon Usages Table

**File:** `pkg/db/migrations/202603161001_create_addon_usages_table.go`

```go
package migrations

import (
    "fmt"

    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"
)

func createAddonUsagesTable() *gormigrate.Migration {
    type AddonUsage struct {
        ID              string `gorm:"primary_key;default:gen_random_uuid()"`
        AddonType       string `gorm:"not null"`
        AddonID         string `gorm:"not null"`
        StackID         string `gorm:"not null"`
        StackResourceID string `gorm:"not null"`
    }
    return &gormigrate.Migration{
        ID: "202603161001_create_addon_usages_table",
        Migrate: func(tx *gorm.DB) error {
            if err := tx.Migrator().AutoMigrate(&AddonUsage{}); err != nil {
                return fmt.Errorf("error running addon usages migration: %w", err)
            }
            // Add unique constraint
            if err := tx.Exec(
                "ALTER TABLE addon_usages ADD CONSTRAINT uq_addon_usages UNIQUE (addon_type, addon_id, stack_id, stack_resource_id)").Error; err != nil {
                return fmt.Errorf("error adding unique constraint on addon_usages: %w", err)
            }
            // Add indexes
            if err := tx.Exec(
                "CREATE INDEX IF NOT EXISTS idx_addon_usages_addon ON addon_usages(addon_type, addon_id)").Error; err != nil {
                return fmt.Errorf("error creating addon_usages addon index: %w", err)
            }
            if err := tx.Exec(
                "CREATE INDEX IF NOT EXISTS idx_addon_usages_stack ON addon_usages(stack_id)").Error; err != nil {
                return fmt.Errorf("error creating addon_usages stack index: %w", err)
            }
            return nil
        },
    }
}
```

### Migration 3: Add EnvFromAddons to Stack Resources

**File:** `pkg/db/migrations/202603161002_add_env_from_addons_to_stack_resources.go`

The `execution_config` column is already JSONB. Adding the `env_from_addons` field to `ExecutionConfig` struct is a code-only change - no schema migration needed.

### Registration

Add migrations to `pkg/db/migrations/migrations.go`:

```go
// In GetMigrations() function
createAddonUsagesTable(),
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

## Deferred Items

See `docs/plans/postgres-addon-deferred-items.md` for the executable plan.

## Open Questions

None at this time.

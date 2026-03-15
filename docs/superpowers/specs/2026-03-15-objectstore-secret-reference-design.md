# ObjectStore SecretReference Design

**Date:** 2026-03-15
**Status:** Approved
**Scope:** Fix ObjectStore model to use SecretReference for credentials

## Overview

The ObjectStore model has been updated to use `SecretReference` for credential fields instead of raw strings. This design documents the changes needed across OpenAPI spec, presenters, validators, and services to propagate this change.

## Design Decisions

1. **API accepts SecretReference only** - No inline credential values. Users must create secrets first, then reference them.
2. **ObjectStore CRs created on PostgresAddon reference** - Not immediately on API creation.
3. **Azure: Connection string only** - Simplified from multiple auth options.
4. **Explicit SecretReference objects** - Each credential field specifies `{secret_id, key}`.

## SecretReference Structure

```go
type SecretReference struct {
    SecretID string `json:"secret_id"` // UUID of the Stackdome secret
    Key      string `json:"key"`       // Key within the secret containing the value
}
```

## OpenAPI Schema Changes

### New SecretReference Schema

```yaml
SecretReference:
  type: object
  required:
    - secret_id
    - key
  properties:
    secret_id:
      type: string
      format: uuid
      description: UUID of the Stackdome secret
    key:
      type: string
      description: Key within the secret containing the value
```

### Updated Credential Schemas

```yaml
S3Credentials:
  type: object
  required:
    - access_key_id
    - secret_access_key
    - region
  properties:
    access_key_id:
      $ref: '#/components/schemas/SecretReference'
    secret_access_key:
      $ref: '#/components/schemas/SecretReference'
    region:
      type: string
    endpoint_url:
      type: string

AzureCredentials:
  type: object
  required:
    - connection_string
  properties:
    connection_string:
      $ref: '#/components/schemas/SecretReference'
    storage_account_name:
      type: string

GCSCredentials:
  type: object
  required:
    - service_account_credentials
  properties:
    service_account_credentials:
      $ref: '#/components/schemas/SecretReference'
```

## Model Layer

The model in `pkg/models/object_store.go` already has the correct structure:

```go
type S3Credentials struct {
    AccessKeyID     SecretReference `json:"accessKeyId"`
    SecretAccessKey SecretReference `json:"secretAccessKey"`
    Region          string          `json:"region"`
    Endpoint        string          `json:"endpoint,omitempty"`
}

type AzureCredentials struct {
    ConnectionString   SecretReference `json:"connectionString"`
    StorageAccountName string          `json:"storageAccountName"`
}

type GCSCredentials struct {
    ServiceAccountCredentials SecretReference `json:"serviceAccountCredentials"`
}
```

## Presenter Changes

File: `pkg/presenters/object_store.go`

Add helper functions for SecretReference conversion:

```go
func convertSecretReference(in openapi.SecretReference) models.SecretReference {
    return models.SecretReference{
        SecretID: in.GetSecretId(),
        Key:      in.GetKey(),
    }
}

func presentSecretReference(in models.SecretReference) openapi.SecretReference {
    ref := openapi.SecretReference{}
    ref.SetSecretId(in.SecretID)
    ref.SetKey(in.Key)
    return ref
}
```

Update `convertObjectStoreConfiguration` and `presentObjectStoreConfiguration` to use these helpers for each credential field.

## Validator Changes

File: `pkg/validator/objectstore/object_store_validator.go`

Add SecretReference validation helper:

```go
func (v *objectStoreValidator) validateSecretReference(ref models.SecretReference, fieldName string) *errors.ServiceError {
    if ref.SecretID == "" {
        return errors.BadRequest("%s secret_id cannot be empty", fieldName)
    }
    if ref.Key == "" {
        return errors.BadRequest("%s key cannot be empty", fieldName)
    }
    return nil
}
```

Update credential validation methods to validate SecretReference fields instead of string comparisons.

## Service Layer Changes

File: `pkg/services/object_store_service.go`

### New Dependency

Add `SecretService` dependency to validate that referenced secrets and keys exist:

```go
type ObjectStoreServiceSpec struct {
    SessionFactory db.SessionFactory
    SecretService  SecretService
    Logger         logger.Logger
}
```

### Secret Validation

```go
func (s *objectStoreService) validateSecretReference(ctx context.Context, ref models.SecretReference, fieldName string) *errors.ServiceError {
    secret, err := s.secretService.GetSecret(ctx, ref.SecretID)
    if err != nil {
        if err.Is404() {
            return errors.BadRequest("%s: secret with ID '%s' does not exist", fieldName, ref.SecretID)
        }
        return err
    }

    // Check if the key exists in the secret
    keyFound := false
    for _, k := range secret.Keys {
        if k == ref.Key {
            keyFound = true
            break
        }
    }
    if !keyFound {
        return errors.BadRequest("%s: key '%s' does not exist in secret '%s'", fieldName, ref.Key, secret.Name)
    }

    return nil
}

func (s *objectStoreService) validateSecretReferences(ctx context.Context, config models.ObjectStoreConfiguration) *errors.ServiceError {
    if config.S3Credentials != nil {
        if err := s.validateSecretReference(ctx, config.S3Credentials.AccessKeyID, "S3 access key ID"); err != nil {
            return err
        }
        if err := s.validateSecretReference(ctx, config.S3Credentials.SecretAccessKey, "S3 secret access key"); err != nil {
            return err
        }
    }
    if config.AzureCredentials != nil {
        if err := s.validateSecretReference(ctx, config.AzureCredentials.ConnectionString, "Azure connection string"); err != nil {
            return err
        }
    }
    if config.GCSCredentials != nil {
        if err := s.validateSecretReference(ctx, config.GCSCredentials.ServiceAccountCredentials, "GCS service account credentials"); err != nil {
            return err
        }
    }
    return nil
}
```

Call `validateSecretReferences` in both `Create` and `Update` methods before persisting.

## Implementation Order

1. **OpenAPI Spec** - `config/openapi/stackdome_api.yaml`
2. **Regenerate OpenAPI Client** - Run code generator
3. **Model Verification** - `pkg/models/object_store.go`
4. **Presenter** - `pkg/presenters/object_store.go`
5. **Validator** - `pkg/validator/objectstore/object_store_validator.go`
6. **Service** - `pkg/services/object_store_service.go`
7. **Environment/Wiring** - Inject SecretService dependency
8. **Verify** - Run `make binary`

## Testing Considerations

- Unit tests for presenter conversion functions
- Unit tests for validator with valid/invalid SecretReferences
- Integration tests verifying secret existence validation
- Existing ObjectStore integration tests may need fixture updates

## Future Work

After this fix is complete:
1. Create PostgresAddon worker for CR creation
2. ObjectStore CR creation when referenced by PostgresAddon
3. Status reconciliation from cluster back to database

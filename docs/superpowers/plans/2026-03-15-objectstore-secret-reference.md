# ObjectStore SecretReference Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix ObjectStore to use SecretReference for credentials instead of raw strings.

**Architecture:** Update OpenAPI spec to define SecretReference schema, regenerate client code, then update presenter/validator/service layers to handle SecretReference objects. Service layer validates that referenced secrets and keys exist.

**Tech Stack:** Go, OpenAPI 3.0, OpenAPI Generator, GORM

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `config/openapi/stackdome_api.yaml` | Modify | Add SecretReference schema, update credential schemas |
| `pkg/api/openapi/*` | Regenerate | Generated OpenAPI client code |
| `pkg/presenters/object_store.go` | Modify | Convert between API and model SecretReference |
| `pkg/validator/objectstore/object_store_validator.go` | Modify | Validate SecretReference fields |
| `pkg/services/object_store_service.go` | Modify | Add SecretService dependency, validate secrets exist |
| `cmd/environment/development_environment.go` | Modify | Wire SecretService into ObjectStoreService |
| `cmd/environment/test_environment.go` | Modify | Wire SecretService into ObjectStoreService |

---

## Chunk 1: OpenAPI Schema Updates

### Task 1: Add SecretReference Schema to OpenAPI Spec

**Files:**
- Modify: `config/openapi/stackdome_api.yaml`

- [ ] **Step 1: Add SecretReference schema**

Add after line ~3445 (after GCSCredentials schema, before WALConfiguration):

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

- [ ] **Step 2: Update S3Credentials schema**

Replace lines ~3409-3427:

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
          description: Reference to secret containing AWS access key ID
        secret_access_key:
          $ref: '#/components/schemas/SecretReference'
          description: Reference to secret containing AWS secret access key
        region:
          type: string
          description: AWS region
        endpoint_url:
          type: string
          description: Custom S3 endpoint URL for S3-compatible storage
```

- [ ] **Step 3: Update AzureCredentials schema**

Replace lines ~3429-3436:

```yaml
    AzureCredentials:
      type: object
      required:
        - connection_string
      properties:
        connection_string:
          $ref: '#/components/schemas/SecretReference'
          description: Reference to secret containing Azure Blob Storage connection string
        storage_account_name:
          type: string
          description: Azure storage account name
```

- [ ] **Step 4: Update GCSCredentials schema**

Replace lines ~3438-3445:

```yaml
    GCSCredentials:
      type: object
      required:
        - service_account_credentials
      properties:
        service_account_credentials:
          $ref: '#/components/schemas/SecretReference'
          description: Reference to secret containing GCS service account credentials JSON
```

- [ ] **Step 5: Commit OpenAPI changes**

```bash
git add config/openapi/stackdome_api.yaml
git commit -m "feat(openapi): Add SecretReference schema for ObjectStore credentials"
```

---

### Task 2: Regenerate OpenAPI Client

**Files:**
- Regenerate: `pkg/api/openapi/*`

- [ ] **Step 1: Run OpenAPI generator**

```bash
make generate
```

Expected: New files generated including `model_secret_reference.go`

- [ ] **Step 2: Verify SecretReference model was generated**

```bash
ls pkg/api/openapi/model_secret_reference.go
```

Expected: File exists

- [ ] **Step 3: Verify S3Credentials now uses SecretReference**

```bash
grep -A5 "type S3Credentials struct" pkg/api/openapi/model_s3_credentials.go
```

Expected: `AccessKeyId` and `SecretAccessKey` fields are `*SecretReference` type

- [ ] **Step 4: Commit generated code**

```bash
git add pkg/api/openapi/
git commit -m "chore: Regenerate OpenAPI client with SecretReference"
```

---

## Chunk 2: Presenter Layer Updates

### Task 3: Update Object Store Presenter

**Files:**
- Modify: `pkg/presenters/object_store.go`

- [ ] **Step 1: Add SecretReference conversion helpers**

Add after the imports, before `ConvertObjectStore`:

```go
// convertSecretReference converts API SecretReference to domain model
func convertSecretReference(in openapi.SecretReference) models.SecretReference {
	return models.SecretReference{
		SecretID: in.GetSecretId(),
		Key:      in.GetKey(),
	}
}

// presentSecretReference converts domain model to API SecretReference
func presentSecretReference(in models.SecretReference) openapi.SecretReference {
	ref := openapi.SecretReference{}
	ref.SetSecretId(in.SecretID)
	ref.SetKey(in.Key)
	return ref
}
```

- [ ] **Step 2: Rewrite convertObjectStoreConfiguration**

Replace the entire `convertObjectStoreConfiguration` function:

```go
func convertObjectStoreConfiguration(in openapi.ObjectStoreConfiguration) models.ObjectStoreConfiguration {
	res := models.ObjectStoreConfiguration{}

	if in.S3Credentials != nil {
		res.S3Credentials = &models.S3Credentials{
			AccessKeyID:     convertSecretReference(in.S3Credentials.GetAccessKeyId()),
			SecretAccessKey: convertSecretReference(in.S3Credentials.GetSecretAccessKey()),
			Region:          in.S3Credentials.GetRegion(),
			Endpoint:        in.S3Credentials.GetEndpointUrl(),
		}
	}

	if in.AzureCredentials != nil {
		res.AzureCredentials = &models.AzureCredentials{
			ConnectionString:   convertSecretReference(in.AzureCredentials.GetConnectionString()),
			StorageAccountName: in.AzureCredentials.GetStorageAccountName(),
		}
	}

	if in.GcsCredentials != nil {
		res.GCSCredentials = &models.GCSCredentials{
			ServiceAccountCredentials: convertSecretReference(in.GcsCredentials.GetServiceAccountCredentials()),
		}
	}

	return res
}
```

- [ ] **Step 3: Rewrite presentObjectStoreConfiguration**

Replace the entire `presentObjectStoreConfiguration` function:

```go
func presentObjectStoreConfiguration(in models.ObjectStoreConfiguration) openapi.ObjectStoreConfiguration {
	res := openapi.ObjectStoreConfiguration{}

	if in.S3Credentials != nil {
		s3 := openapi.S3Credentials{}
		s3.SetAccessKeyId(presentSecretReference(in.S3Credentials.AccessKeyID))
		s3.SetSecretAccessKey(presentSecretReference(in.S3Credentials.SecretAccessKey))
		s3.SetRegion(in.S3Credentials.Region)
		if in.S3Credentials.Endpoint != "" {
			s3.SetEndpointUrl(in.S3Credentials.Endpoint)
		}
		res.SetS3Credentials(s3)
	}

	if in.AzureCredentials != nil {
		azure := openapi.AzureCredentials{}
		azure.SetConnectionString(presentSecretReference(in.AzureCredentials.ConnectionString))
		if in.AzureCredentials.StorageAccountName != "" {
			azure.SetStorageAccountName(in.AzureCredentials.StorageAccountName)
		}
		res.SetAzureCredentials(azure)
	}

	if in.GCSCredentials != nil {
		gcs := openapi.GCSCredentials{}
		gcs.SetServiceAccountCredentials(presentSecretReference(in.GCSCredentials.ServiceAccountCredentials))
		res.SetGcsCredentials(gcs)
	}

	return res
}
```

- [ ] **Step 4: Run go fmt**

```bash
go fmt ./pkg/presenters/object_store.go
```

- [ ] **Step 5: Commit presenter changes**

```bash
git add pkg/presenters/object_store.go
git commit -m "feat(presenter): Update ObjectStore to handle SecretReference"
```

---

## Chunk 3: Validator Layer Updates

### Task 4: Update Object Store Validator

**Files:**
- Modify: `pkg/validator/objectstore/object_store_validator.go`

- [ ] **Step 1: Add validateSecretReference helper**

Add after `validateBasicFields` function:

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

- [ ] **Step 2: Rewrite validateS3Configuration**

Replace the entire function:

```go
func (v *objectStoreValidator) validateS3Configuration(s3 *models.S3Credentials) *errors.ServiceError {
	if err := v.validateSecretReference(s3.AccessKeyID, "S3 access key ID"); err != nil {
		return err
	}

	if err := v.validateSecretReference(s3.SecretAccessKey, "S3 secret access key"); err != nil {
		return err
	}

	if s3.Region == "" {
		return errors.BadRequest("S3 region cannot be empty")
	}

	// Validate region format
	regionRegex := regexp.MustCompile(`^[a-z]([a-z0-9-]*[a-z0-9])?$`)
	if !regionRegex.MatchString(s3.Region) {
		return errors.BadRequest("S3 region must be a valid AWS region format")
	}

	// Validate endpoint if specified
	if s3.Endpoint != "" {
		if !strings.HasPrefix(s3.Endpoint, "http://") && !strings.HasPrefix(s3.Endpoint, "https://") {
			return errors.BadRequest("S3 endpoint must be a valid URL with http:// or https:// prefix")
		}
	}

	return nil
}
```

- [ ] **Step 3: Rewrite validateAzureConfiguration**

Replace the entire function:

```go
func (v *objectStoreValidator) validateAzureConfiguration(azure *models.AzureCredentials) *errors.ServiceError {
	if err := v.validateSecretReference(azure.ConnectionString, "Azure connection string"); err != nil {
		return err
	}

	return nil
}
```

- [ ] **Step 4: Rewrite validateGCSConfiguration**

Replace the entire function:

```go
func (v *objectStoreValidator) validateGCSConfiguration(gcs *models.GCSCredentials) *errors.ServiceError {
	if err := v.validateSecretReference(gcs.ServiceAccountCredentials, "GCS service account credentials"); err != nil {
		return err
	}

	return nil
}
```

- [ ] **Step 5: Run go fmt**

```bash
go fmt ./pkg/validator/objectstore/object_store_validator.go
```

- [ ] **Step 6: Commit validator changes**

```bash
git add pkg/validator/objectstore/object_store_validator.go
git commit -m "feat(validator): Update ObjectStore to validate SecretReference"
```

---

## Chunk 4: Service Layer Updates

### Task 5: Add SecretService Dependency to ObjectStoreService

**Files:**
- Modify: `pkg/services/object_store_service.go`

- [ ] **Step 1: Add SecretService to ObjectStoreServiceSpec**

Update the struct (around line 27):

```go
type ObjectStoreServiceSpec struct {
	SessionFactory db.SessionFactory
	SecretService  SecretService
	Logger         logger.Logger
}
```

- [ ] **Step 2: Add secretService field to objectStoreService**

Update the struct (around line 32):

```go
type objectStoreService struct {
	objectStoreStore stores.ObjectStoreStore
	secretService    SecretService
	validator        validator.ObjectStoreValidator
	logger           logger.Logger
}
```

- [ ] **Step 3: Update NewObjectStoreService constructor**

Update the function (around line 38):

```go
func NewObjectStoreService(spec ObjectStoreServiceSpec) ObjectStoreService {
	return &objectStoreService{
		objectStoreStore: pgstore.NewObjectStoreStore(pgstore.ObjectStoreStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		secretService: spec.SecretService,
		validator:     objectstore.NewObjectStoreValidator(),
		logger:        spec.Logger,
	}
}
```

- [ ] **Step 4: Add validateSecretReference method**

Add after `NewObjectStoreService`:

```go
func (s *objectStoreService) validateSecretReference(ctx context.Context, ref models.SecretReference, fieldName string) *errors.ServiceError {
	secret, err := s.secretService.GetByID(ctx, ref.SecretID)
	if err != nil {
		if err.Is404() {
			return errors.BadRequest("%s: secret with ID '%s' does not exist", fieldName, ref.SecretID)
		}
		return err
	}

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
```

- [ ] **Step 5: Add validateSecretReferences method**

Add after `validateSecretReference`:

```go
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

- [ ] **Step 6: Update Create method to validate secrets**

In the `Create` method, add after validator check (around line 53):

```go
// Validate secret references exist
if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
	return nil, err
}
```

- [ ] **Step 7: Update Update method to validate secrets**

In the `Update` method, add after validator check (around line 98):

```go
// Validate secret references exist
if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
	return nil, err
}
```

- [ ] **Step 8: Run go fmt**

```bash
go fmt ./pkg/services/object_store_service.go
```

- [ ] **Step 9: Commit service changes**

```bash
git add pkg/services/object_store_service.go
git commit -m "feat(service): Add SecretService dependency to ObjectStoreService"
```

---

### Task 6: Wire SecretService in Environment Files

**Files:**
- Modify: `cmd/environment/development_environment.go`
- Modify: `cmd/environment/test_environment.go`

- [ ] **Step 1: Update development_environment.go**

Find the `objectStoreService` initialization (around line 340) and update:

```go
objectStoreService := services.NewObjectStoreService(services.ObjectStoreServiceSpec{
	SessionFactory: d.DBSession,
	SecretService:  secretService,
	Logger:         d.Logger,
})
```

- [ ] **Step 2: Update test_environment.go**

Find the `objectStoreService` initialization (around line 327) and update:

```go
objectStoreService := services.NewObjectStoreService(services.ObjectStoreServiceSpec{
	SessionFactory: t.DBSession,
	SecretService:  secretService,
	Logger:         t.Logger,
})
```

- [ ] **Step 3: Run go fmt**

```bash
go fmt ./cmd/environment/development_environment.go ./cmd/environment/test_environment.go
```

- [ ] **Step 4: Commit wiring changes**

```bash
git add cmd/environment/development_environment.go cmd/environment/test_environment.go
git commit -m "feat(env): Wire SecretService into ObjectStoreService"
```

---

## Chunk 5: Verification

### Task 7: Verify Build and Tests

**Files:**
- All modified files

- [ ] **Step 1: Run make binary to verify compilation**

```bash
make binary
```

Expected: Build succeeds with no errors

- [ ] **Step 2: Run go vet**

```bash
go vet ./...
```

Expected: No errors

- [ ] **Step 3: Run existing tests**

```bash
go test ./pkg/presenters/... ./pkg/validator/... ./pkg/services/... -v
```

Expected: Tests pass (some may need fixture updates)

- [ ] **Step 4: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: Address any test/lint issues from SecretReference changes"
```

---

## Summary

This plan implements the ObjectStore SecretReference changes in 7 tasks across 5 chunks:

1. **Chunk 1:** OpenAPI schema updates and code regeneration
2. **Chunk 2:** Presenter layer updates
3. **Chunk 3:** Validator layer updates
4. **Chunk 4:** Service layer updates with SecretService integration
5. **Chunk 5:** Build verification

Total estimated steps: 35 atomic actions with frequent commits.

package objectstore

import (
	"context"
	"regexp"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
)

type objectStoreValidator struct{}

func NewObjectStoreValidator() validator.ObjectStoreValidator {
	return &objectStoreValidator{}
}

func (v *objectStoreValidator) ValidateForCreate(ctx context.Context, spec *models.ObjectStore) *errors.ServiceError {
	if err := v.validateBasicFields(spec); err != nil {
		return err
	}

	if err := v.validateConfiguration(spec); err != nil {
		return err
	}

	return nil
}

func (v *objectStoreValidator) ValidateForUpdate(ctx context.Context, existing *models.ObjectStore, spec *models.ObjectStore) *errors.ServiceError {
	// Only validate immutable field changes
	// Full validation is done by the service after preserving immutable fields
	return v.validateImmutableFields(existing, spec)
}

func (v *objectStoreValidator) validateBasicFields(spec *models.ObjectStore) *errors.ServiceError {
	if spec.Name == "" {
		return errors.BadRequest("Object store name cannot be empty")
	}

	if spec.OrganisationID == "" {
		return errors.BadRequest("Object store organisation ID cannot be empty")
	}

	// Validate name format (DNS-1123 subdomain)
	nameRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !nameRegex.MatchString(spec.Name) {
		return errors.BadRequest("Object store name must be a valid DNS subdomain (lowercase letters, numbers, and hyphens)")
	}

	if len(spec.Name) > 63 {
		return errors.BadRequest("Object store name cannot be longer than 63 characters")
	}

	// Validate retention policy format
	if spec.RetentionPolicy != "" {
		retentionRegex := regexp.MustCompile(`^[0-9]+[dwmy]$`)
		if !retentionRegex.MatchString(spec.RetentionPolicy) {
			return errors.BadRequest("Retention policy must be in format '<number><unit>' where unit is d (days), w (weeks), m (months), or y (years)")
		}
	}

	return nil
}

func (v *objectStoreValidator) validateSecretReference(ref models.SecretReference, fieldName string) *errors.ServiceError {
	if ref.SecretID == "" {
		return errors.BadRequest("%s secret_id cannot be empty", fieldName)
	}
	if ref.Key == "" {
		return errors.BadRequest("%s key cannot be empty", fieldName)
	}
	return nil
}

func (v *objectStoreValidator) validateConfiguration(spec *models.ObjectStore) *errors.ServiceError {
	credentialCount := 0

	if spec.Configuration.S3Credentials != nil {
		credentialCount++
		if err := v.validateS3Configuration(spec.Configuration.S3Credentials); err != nil {
			return err
		}
	}

	if spec.Configuration.AzureCredentials != nil {
		credentialCount++
		if err := v.validateAzureConfiguration(spec.Configuration.AzureCredentials); err != nil {
			return err
		}
	}

	if spec.Configuration.GCSCredentials != nil {
		credentialCount++
		if err := v.validateGCSConfiguration(spec.Configuration.GCSCredentials); err != nil {
			return err
		}
	}

	if credentialCount == 0 {
		return errors.BadRequest("At least one credential configuration (S3, Azure, or GCS) must be specified")
	}

	if credentialCount > 1 {
		return errors.BadRequest("Only one credential configuration (S3, Azure, or GCS) can be specified")
	}

	if err := v.validateDestinationPath(spec); err != nil {
		return err
	}

	return nil
}

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

func (v *objectStoreValidator) validateAzureConfiguration(azure *models.AzureCredentials) *errors.ServiceError {
	if err := v.validateSecretReference(azure.ConnectionString, "Azure connection string"); err != nil {
		return err
	}

	return nil
}

func (v *objectStoreValidator) validateGCSConfiguration(gcs *models.GCSCredentials) *errors.ServiceError {
	if err := v.validateSecretReference(gcs.ServiceAccountCredentials, "GCS service account credentials"); err != nil {
		return err
	}

	return nil
}

func (v *objectStoreValidator) validateDestinationPath(spec *models.ObjectStore) *errors.ServiceError {
	path := spec.DestinationPath
	if path == "" {
		return errors.BadRequest("Destination path cannot be empty")
	}

	switch {
	case spec.Configuration.S3Credentials != nil:
		if !strings.HasPrefix(path, "s3://") {
			return errors.BadRequest("S3 destination path must start with 's3://' (e.g., s3://bucket-name/path)")
		}
		bucket := strings.TrimPrefix(path, "s3://")
		bucket = strings.TrimSuffix(bucket, "/")
		bucket = strings.SplitN(bucket, "/", 2)[0]
		if bucket == "" {
			return errors.BadRequest("S3 destination path must include a bucket name (e.g., s3://bucket-name/)")
		}

	case spec.Configuration.AzureCredentials != nil:
		azurePattern := regexp.MustCompile(`^https?://[a-z0-9]+\.[a-z]+\.core\.windows\.net/[^/]+/.+`)
		if !azurePattern.MatchString(path) {
			return errors.BadRequest("Azure destination path must match format: http(s)://<account>.<service>.core.windows.net/<container>/<blob>")
		}

	case spec.Configuration.GCSCredentials != nil:
		if !strings.HasPrefix(path, "gs://") {
			return errors.BadRequest("GCS destination path must start with 'gs://' (e.g., gs://bucket-name/path)")
		}
		bucket := strings.TrimPrefix(path, "gs://")
		bucket = strings.TrimSuffix(bucket, "/")
		bucket = strings.SplitN(bucket, "/", 2)[0]
		if bucket == "" {
			return errors.BadRequest("GCS destination path must include a bucket name (e.g., gs://bucket-name/)")
		}
	}

	return nil
}

func (v *objectStoreValidator) validateImmutableFields(existing *models.ObjectStore, spec *models.ObjectStore) *errors.ServiceError {
	// Only check name if it's set in the spec (non-empty)
	if spec.Name != "" && existing.Name != spec.Name {
		return errors.BadRequest("Object store name cannot be changed")
	}

	// Only check OrganisationID if it's set in the spec (non-empty)
	if spec.OrganisationID != "" && existing.OrganisationID != spec.OrganisationID {
		return errors.BadRequest("Object store organisation ID cannot be changed")
	}

	return nil
}

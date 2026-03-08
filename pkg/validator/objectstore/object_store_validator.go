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
	// Validate immutable fields
	if err := v.validateImmutableFields(existing, spec); err != nil {
		return err
	}

	// Run create validation on the new spec
	if err := v.ValidateForCreate(ctx, spec); err != nil {
		return err
	}

	return nil
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

	return nil
}

func (v *objectStoreValidator) validateS3Configuration(s3 *models.S3Credentials) *errors.ServiceError {
	if s3.AccessKeyID == "" {
		return errors.BadRequest("S3 access key ID cannot be empty")
	}

	if s3.SecretAccessKey == "" {
		return errors.BadRequest("S3 secret access key cannot be empty")
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
		// Basic URL validation
		if !strings.HasPrefix(s3.Endpoint, "http://") && !strings.HasPrefix(s3.Endpoint, "https://") {
			return errors.BadRequest("S3 endpoint must be a valid URL with http:// or https:// prefix")
		}
	}

	return nil
}

func (v *objectStoreValidator) validateAzureConfiguration(azure *models.AzureCredentials) *errors.ServiceError {
	if azure.ConnectionString == "" {
		return errors.BadRequest("Azure connection string cannot be empty")
	}

	// Basic validation of connection string format
	if !strings.Contains(azure.ConnectionString, "AccountName=") || !strings.Contains(azure.ConnectionString, "AccountKey=") {
		return errors.BadRequest("Azure connection string must contain AccountName and AccountKey")
	}

	return nil
}

func (v *objectStoreValidator) validateGCSConfiguration(gcs *models.GCSCredentials) *errors.ServiceError {
	if gcs.ServiceAccountKey == "" {
		return errors.BadRequest("GCS service account key cannot be empty")
	}

	// Validate service account key format (should be valid JSON)
	if !strings.HasPrefix(strings.TrimSpace(gcs.ServiceAccountKey), "{") {
		return errors.BadRequest("GCS service account key must be valid JSON")
	}

	return nil
}

func (v *objectStoreValidator) validateImmutableFields(existing *models.ObjectStore, spec *models.ObjectStore) *errors.ServiceError {
	if existing.Name != spec.Name {
		return errors.BadRequest("Object store name cannot be changed")
	}

	if existing.OrganisationID != spec.OrganisationID {
		return errors.BadRequest("Object store organisation ID cannot be changed")
	}

	return nil
}

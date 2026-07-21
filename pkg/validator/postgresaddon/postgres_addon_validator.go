package postgresaddon

import (
	"context"
	"regexp"
	"strings"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator"
	gocron "github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/validation"
)

type postgresAddonValidator struct{}

func NewPostgresAddonValidator() validator.PostgresAddonValidator {
	return &postgresAddonValidator{}
}

func (v *postgresAddonValidator) ValidateForCreate(ctx context.Context, spec *models.PostgresAddon) *errors.ServiceError {
	if err := v.validateBasicFields(spec); err != nil {
		return err
	}

	if err := v.validateConfiguration(spec); err != nil {
		return err
	}

	if err := v.validateResources(spec); err != nil {
		return err
	}

	if err := v.validateInstances(spec); err != nil {
		return err
	}

	if err := v.validateStorage(spec); err != nil {
		return err
	}

	if err := v.validateBackupConfig(spec); err != nil {
		return err
	}

	if err := v.validateInitialization(spec); err != nil {
		return err
	}

	if err := v.validateDatabases(spec); err != nil {
		return err
	}

	return nil
}

func (v *postgresAddonValidator) ValidateForUpdate(ctx context.Context, existing *models.PostgresAddon, spec *models.PostgresAddon) *errors.ServiceError {
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

func (v *postgresAddonValidator) validateBasicFields(spec *models.PostgresAddon) *errors.ServiceError {
	if spec.Name == "" {
		return errors.BadRequest("PostgreSQL addon name cannot be empty")
	}

	if spec.OrganisationID == "" {
		return errors.BadRequest("PostgreSQL addon organisation ID cannot be empty")
	}

	if spec.UserID == "" {
		return errors.BadRequest("PostgreSQL addon user ID cannot be empty")
	}

	if spec.ProjectID == "" {
		return errors.BadRequest("PostgreSQL addon project ID cannot be empty")
	}

	// Validate name format (DNS-1123 subdomain)
	nameRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !nameRegex.MatchString(spec.Name) {
		return errors.BadRequest("PostgreSQL addon name must be a valid DNS subdomain (lowercase letters, numbers, and hyphens)")
	}

	if len(spec.Name) > models.MaxAddonNameLength {
		return errors.BadRequest("PostgreSQL addon name cannot be longer than %d characters", models.MaxAddonNameLength)
	}

	return nil
}

func (v *postgresAddonValidator) validateConfiguration(spec *models.PostgresAddon) *errors.ServiceError {
	if spec.PostgresVersion.Major == 0 {
		return errors.BadRequest("PostgreSQL version major cannot be zero")
	}

	if spec.PostgresVersion.Major < 10 || spec.PostgresVersion.Major > 17 {
		return errors.BadRequest("PostgreSQL version major must be between 10 and 17")
	}

	if spec.PostgresVersion.Minor < 0 {
		return errors.BadRequest("PostgreSQL version minor cannot be negative")
	}

	// Validate parameters if provided
	if spec.Configuration.Parameters != nil {
		for key, value := range spec.Configuration.Parameters {
			if strings.TrimSpace(key) == "" {
				return errors.BadRequest("PostgreSQL configuration parameter key cannot be empty")
			}
			if strings.TrimSpace(value) == "" {
				return errors.BadRequest("PostgreSQL configuration parameter value for key '%s' cannot be empty", key)
			}
		}
	}

	return nil
}

func (v *postgresAddonValidator) validateResources(spec *models.PostgresAddon) *errors.ServiceError {
	// Resources are optional, so if not provided, skip validation
	if spec.Resources.CPU.Request == "" && spec.Resources.CPU.Limit == "" &&
		spec.Resources.Memory.Request == "" && spec.Resources.Memory.Limit == "" {
		return nil
	}

	// If any resource is specified, validate all of them
	if spec.Resources.CPU.Request == "" {
		return errors.BadRequest("CPU request cannot be empty when resources are specified")
	}

	if spec.Resources.CPU.Limit == "" {
		return errors.BadRequest("CPU limit cannot be empty when resources are specified")
	}

	if spec.Resources.Memory.Request == "" {
		return errors.BadRequest("Memory request cannot be empty when resources are specified")
	}

	if spec.Resources.Memory.Limit == "" {
		return errors.BadRequest("Memory limit cannot be empty when resources are specified")
	}

	// Validate resource format (Kubernetes resource format)
	resourceRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?[a-zA-Z]*$`)

	if !resourceRegex.MatchString(spec.Resources.CPU.Request) {
		return errors.BadRequest("CPU request must be in valid Kubernetes resource format (e.g., '100m', '1', '2.5')")
	}

	if !resourceRegex.MatchString(spec.Resources.CPU.Limit) {
		return errors.BadRequest("CPU limit must be in valid Kubernetes resource format (e.g., '100m', '1', '2.5')")
	}

	memoryRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?[KMGT]i?$`)
	if !memoryRegex.MatchString(spec.Resources.Memory.Request) {
		return errors.BadRequest("Memory request must be in valid Kubernetes memory format (e.g., '512Mi', '1Gi', '2G')")
	}

	if !memoryRegex.MatchString(spec.Resources.Memory.Limit) {
		return errors.BadRequest("Memory limit must be in valid Kubernetes memory format (e.g., '512Mi', '1Gi', '2G')")
	}

	return nil
}

func (v *postgresAddonValidator) validateInstances(spec *models.PostgresAddon) *errors.ServiceError {
	if spec.Instances.Count < 1 {
		return errors.BadRequest("Number of instances must be at least 1")
	}

	if spec.Instances.Count > 5 {
		return errors.BadRequest("Number of instances cannot exceed 5")
	}

	// Validate placement if specified
	if spec.Instances.Placement != nil {
		if spec.Instances.Placement.NodeSelector != nil {
			for key := range spec.Instances.Placement.NodeSelector {
				if strings.TrimSpace(key) == "" {
					return errors.BadRequest("Node selector key cannot be empty")
				}
			}
		}

		// Validate tolerations if specified
		for _, toleration := range spec.Instances.Placement.Tolerations {
			if toleration.Key == "" {
				return errors.BadRequest("Toleration key cannot be empty")
			}

			validOperators := []string{"Equal", "Exists"}
			if !lo.Contains(validOperators, toleration.Operator) {
				return errors.BadRequest("Toleration operator must be one of: %s", strings.Join(validOperators, ", "))
			}

			if toleration.Operator == "Equal" && toleration.Value == "" {
				return errors.BadRequest("Toleration value cannot be empty when operator is 'Equal'")
			}

			validEffects := []string{"NoSchedule", "PreferNoSchedule", "NoExecute"}
			if !lo.Contains(validEffects, toleration.Effect) {
				return errors.BadRequest("Toleration effect must be one of: %s", strings.Join(validEffects, ", "))
			}
		}
	}

	return nil
}

func (v *postgresAddonValidator) validateStorage(spec *models.PostgresAddon) *errors.ServiceError {
	if spec.Storage.Size == "" {
		return errors.BadRequest("Storage size cannot be empty")
	}

	// Validate storage size format
	sizeRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?[KMGT]i?$`)
	if !sizeRegex.MatchString(spec.Storage.Size) {
		return errors.BadRequest("Storage size must be in valid Kubernetes storage format (e.g., '10Gi', '100Mi', '1Ti')")
	}

	if spec.Storage.StorageClass == "" {
		return errors.BadRequest("Storage class cannot be empty")
	}

	return nil
}

func (v *postgresAddonValidator) validateBackupConfig(spec *models.PostgresAddon) *errors.ServiceError {
	if spec.BackupConfig.WALArchiving && spec.BackupConfig.ObjectStoreID == "" {
		return errors.BadRequest("WAL archiving requires an object store to be configured")
	}

	// Validate schedule format if provided
	if spec.BackupConfig.Schedule != "" {
		parser := gocron.NewParser(gocron.Second | gocron.Minute | gocron.Hour | gocron.Dom | gocron.Month | gocron.Dow)
		if _, err := parser.Parse(spec.BackupConfig.Schedule); err != nil {
			return errors.BadRequest("Backup schedule must be in valid cron format: %s", err.Error())
		}
	}

	return nil
}

func (v *postgresAddonValidator) validateInitialization(spec *models.PostgresAddon) *errors.ServiceError {
	// Check if initialization is empty struct
	if (spec.Initialization == models.PostgresInitialization{}) {
		return nil // Initialization is optional
	}

	initMethodCount := 0

	if spec.Initialization.ImportFromExternal != nil {
		initMethodCount++
		if err := v.validateImportFromExternal(spec.Initialization.ImportFromExternal); err != nil {
			return err
		}
	}

	if spec.Initialization.RestoreFromBackup != nil {
		initMethodCount++
		if err := v.validateRestoreFromBackup(spec.Initialization.RestoreFromBackup); err != nil {
			return err
		}
	}

	if spec.Initialization.RestoreFromObjectStore != nil {
		initMethodCount++
		if err := v.validateRestoreFromObjectStore(spec.Initialization.RestoreFromObjectStore); err != nil {
			return err
		}
	}

	if initMethodCount > 1 {
		return errors.BadRequest("Only one initialization method can be specified")
	}

	return nil
}

func (v *postgresAddonValidator) validateImportFromExternal(import_ *models.PostgresImportFromExternal) *errors.ServiceError {
	if import_.Host == "" {
		return errors.BadRequest("External import host cannot be empty")
	}

	if import_.Port < 1 || import_.Port > 65535 {
		return errors.BadRequest("External import port must be between 1 and 65535")
	}

	if import_.Username == "" {
		return errors.BadRequest("External import username cannot be empty")
	}

	if import_.Database == "" {
		return errors.BadRequest("External import database cannot be empty")
	}

	if import_.PasswordSecretID == "" {
		return errors.BadRequest("External import password secret ID cannot be empty")
	}

	return nil
}

func (v *postgresAddonValidator) validateRestoreFromBackup(restore *models.PostgresRestoreFromBackup) *errors.ServiceError {
	if restore.BackupID == "" {
		return errors.BadRequest("Backup ID cannot be empty for restore from backup")
	}

	return nil
}

func (v *postgresAddonValidator) validateRestoreFromObjectStore(restore *models.PostgresRestoreFromObjectStore) *errors.ServiceError {
	if restore.ObjectStoreID == "" {
		return errors.BadRequest("Object store ID cannot be empty for restore from object store")
	}

	if restore.SourcePostgresAddonID == "" {
		return errors.BadRequest("Source PostgreSQL addon ID cannot be empty for restore from object store")
	}

	return nil
}

func (v *postgresAddonValidator) validateDatabases(spec *models.PostgresAddon) *errors.ServiceError {
	// Databases are optional
	if len(spec.Databases) == 0 {
		return nil
	}

	databaseNames := make(map[string]bool)

	for _, db := range spec.Databases {
		if db.Name == "" {
			return errors.BadRequest("Database name cannot be empty")
		}

		// Check for duplicate database names
		if databaseNames[db.Name] {
			return errors.BadRequest("Duplicate database name '%s'", db.Name)
		}
		databaseNames[db.Name] = true

		if errs := validation.IsDNS1123Subdomain(db.Name); len(errs) > 0 {
			return errors.BadRequest("Database name '%s' is not a valid DNS subdomain: %s", db.Name, strings.Join(errs, "; "))
		}
		for _, ext := range db.Extensions {
			if !lo.Contains(models.SupportedPostgresExtensions, ext) {
				return errors.BadRequest("Unsupported PostgreSQL extension '%s' for database '%s'", ext, db.Name)
			}
		}
	}

	return nil
}

func (v *postgresAddonValidator) validateImmutableFields(existing *models.PostgresAddon, spec *models.PostgresAddon) *errors.ServiceError {
	if existing.Name != spec.Name {
		return errors.BadRequest("PostgreSQL addon name cannot be changed")
	}

	if existing.OrganisationID != spec.OrganisationID {
		return errors.BadRequest("PostgreSQL addon organisation ID cannot be changed")
	}

	if existing.PostgresVersion.Major != spec.PostgresVersion.Major || existing.PostgresVersion.Minor != spec.PostgresVersion.Minor {
		return errors.BadRequest("PostgreSQL version cannot be changed")
	}

	if existing.Storage.Size != spec.Storage.Size {
		return errors.BadRequest("Storage size cannot be changed")
	}

	if existing.Storage.StorageClass != spec.Storage.StorageClass {
		return errors.BadRequest("Storage class cannot be changed")
	}

	// Initialization is immutable after creation - check if both are non-empty structs
	existingInitEmpty := (existing.Initialization == models.PostgresInitialization{})
	specInitEmpty := (spec.Initialization == models.PostgresInitialization{})

	if !existingInitEmpty && !specInitEmpty {
		return errors.BadRequest("Initialization configuration cannot be changed after creation")
	}

	if existingInitEmpty != specInitEmpty {
		return errors.BadRequest("Initialization configuration cannot be changed after creation")
	}

	return nil
}

package postgresaddon

import (
	"context"
	"strings"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func cloudPostgresAddonValidator() *postgresAddonValidator {
	return &postgresAddonValidator{externalImportDisabled: true}
}

func validPostgresAddonSpec(name string) *models.PostgresAddon {
	return &models.PostgresAddon{
		Name:            name,
		OrganisationID:  "org-1",
		UserID:          "user-1",
		ProjectID:       "project-1",
		PostgresVersion: models.PostgresVersion{Major: 16},
		Instances:       models.PostgresInstances{Count: 1},
		Storage:         models.PostgresStorage{Size: "10Gi", StorageClass: "standard"},
	}
}

var _ = Describe("ValidateForCreate name length", func() {
	It("accepts a name of MaxAddonNameLength characters", func() {
		spec := validPostgresAddonSpec(strings.Repeat("a", models.MaxAddonNameLength))

		err := NewPostgresAddonValidator(PostgresAddonValidatorSpec{}).ValidateForCreate(context.Background(), spec)

		Expect(err).To(BeNil())
	})

	It("rejects a name one character over MaxAddonNameLength", func() {
		spec := validPostgresAddonSpec(strings.Repeat("a", models.MaxAddonNameLength+1))

		err := NewPostgresAddonValidator(PostgresAddonValidatorSpec{}).ValidateForCreate(context.Background(), spec)

		Expect(err).NotTo(BeNil())
	})
})

var _ = Describe("Cloud mode initialization validation", func() {
	It("rejects an external import with a stable field error", func() {
		spec := validPostgresAddonSpec("postgres")
		spec.Initialization.ImportFromExternal = &models.PostgresImportFromExternal{
			Host: "database.example.com", Port: 5432, Database: "app", Username: "app", PasswordSecretID: "secret-1",
		}

		err := cloudPostgresAddonValidator().ValidateForCreate(context.Background(), spec)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(errors.ErrorValidation))
		details, ok := err.Details.(errors.ValidationErrorDetails)
		Expect(ok).To(BeTrue())
		Expect(details.Errors).To(ConsistOf(errors.FieldError{
			Field:   ExternalImportDisabledField,
			Code:    errors.VErrPostgresExternalImportDisabled,
			Message: ExternalImportDisabledMessage,
		}))
	})

	DescribeTable("keeps supported initialization methods enabled",
		func(initialization models.PostgresInitialization) {
			spec := validPostgresAddonSpec("postgres")
			spec.Initialization = initialization

			Expect(cloudPostgresAddonValidator().ValidateForCreate(context.Background(), spec)).To(BeNil())
		},
		Entry("ordinary create", models.PostgresInitialization{}),
		Entry("backup restore", models.PostgresInitialization{
			RestoreFromBackup: &models.PostgresRestoreFromBackup{BackupID: "backup-1"},
		}),
		Entry("object-store restore", models.PostgresInitialization{
			RestoreFromObjectStore: &models.PostgresRestoreFromObjectStore{
				ObjectStoreID: "store-1", SourcePostgresAddonID: "postgres-1",
			},
		}),
	)
})

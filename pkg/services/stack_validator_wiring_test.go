package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/glebarez/sqlite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// sqliteSessionFactory backs the fat-path wiring tests with a real in-memory
// database so the validator's raw-store seams are exercised exactly as
// NewStackService wires them in production.
type sqliteSessionFactory struct{ db *gorm.DB }

func newSQLiteSessionFactory(t *testing.T, ddlStatements ...string) *sqliteSessionFactory {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	for _, ddl := range ddlStatements {
		if err := gdb.Exec(ddl).Error; err != nil {
			t.Fatalf("failed to execute DDL: %v", err)
		}
	}
	return &sqliteSessionFactory{db: gdb}
}

func (f *sqliteSessionFactory) Init(*config.DatabaseConfig) {}
func (f *sqliteSessionFactory) DirectDB() *sql.DB           { d, _ := f.db.DB(); return d }
func (f *sqliteSessionFactory) New(context.Context) *gorm.DB {
	return f.db.Session(&gorm.Session{})
}
func (f *sqliteSessionFactory) CheckConnection() error { return nil }
func (f *sqliteSessionFactory) Close() error           { d, _ := f.db.DB(); return d.Close() }

const secretsTableDDL = `
	CREATE TABLE IF NOT EXISTS secrets (
		id TEXT PRIMARY KEY,
		organisation_id TEXT NOT NULL,
		project_id TEXT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		type TEXT NOT NULL,
		encrypted_data TEXT NOT NULL,
		keys TEXT,
		data_hash TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)
`

// newProductionWiredStackValidator builds the stack validator exactly as
// NewStackService wires it, then reaches through the (package-internal)
// stackService to hand it back. SecretService is deliberately nil: the fat
// path's env secret_key_ref existence check must run against the raw secret
// store, never the RBAC-enforcing SecretService, so a caller holding only
// stacks:write can apply the same payload the thin per-resource path
// accepts. If the wiring ever regresses to routing the check through
// SecretService, construction (or the check itself) fails loudly instead of
// silently reintroducing a secrets:read requirement.
func newProductionWiredStackValidator(t *testing.T, sf db.SessionFactory) validator.StackValidator {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc := NewStackService(StackServiceSpec{
		SessionFactory:        sf,
		CredentialResolver:    mocks.NewMockCredentialResolver(ctrl),
		GitIntegrationService: mocks.NewMockGitIntegrationService(ctrl),
	}).(*stackService)
	return svc.stackValidator
}

func envSecretStack(orgID, secretName string) *models.Stack {
	return &models.Stack{
		Name:           "test-stack",
		OrganisationID: orgID,
		UserID:         "user-1",
		StackResources: []*models.StackResource{{
			Name:        "web",
			ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
			ExecutionConfig: &models.ExecutionConfig{
				Env: []models.EnvVar{{
					Name:         "DB_PASSWORD",
					SecretKeyRef: &models.EnvSecretRef{SecretName: secretName, Key: "password"},
				}},
			},
		}},
	}
}

func seedSecret(t *testing.T, sf *sqliteSessionFactory, id, orgID, name string) {
	t.Helper()
	secret := &models.Secret{
		ID:             id,
		OrganisationID: orgID,
		UserID:         "user-1",
		Name:           name,
		Type:           models.SecretTypeGeneric,
		EncryptedData:  "irrelevant",
		DataHash:       "hash-1",
		Keys:           models.SecretKeys{"password"},
	}
	if err := sf.New(context.Background()).Create(secret).Error; err != nil {
		t.Fatalf("failed to seed secret: %v", err)
	}
}

// TestFatPathEnvSecretValidation_NoSecretsReadRequired asserts the fat
// (whole-stack) validation path resolves env secret_key_ref references via
// the raw org-scoped store: an existing same-org secret validates without
// any SecretService (and therefore without any secrets:read permission
// check), matching the thin per-resource path wired in cmd/environment.
func TestFatPathEnvSecretValidation_NoSecretsReadRequired(t *testing.T) {
	sf := newSQLiteSessionFactory(t, secretsTableDDL)
	seedSecret(t, sf, "sec-1", "org-1", "db-creds")

	v := newProductionWiredStackValidator(t, sf)

	if err := v.ValidateForCreate(context.Background(), envSecretStack("org-1", "db-creds")); err != nil {
		t.Fatalf("expected env secret_key_ref to validate without secrets:read, got %v", err)
	}
}

// TestFatPathEnvSecretValidation_CrossOrgSecretNotFound asserts the raw
// lookup stays org-scoped: a secret that only exists in another organisation
// behaves as not-found rather than leaking cross-org existence.
func TestFatPathEnvSecretValidation_CrossOrgSecretNotFound(t *testing.T) {
	sf := newSQLiteSessionFactory(t, secretsTableDDL)
	seedSecret(t, sf, "sec-1", "org-other", "db-creds")

	v := newProductionWiredStackValidator(t, sf)

	err := v.ValidateForCreate(context.Background(), envSecretStack("org-1", "db-creds"))
	if err == nil {
		t.Fatal("expected a validation error for a cross-org secret, got nil")
	}
	details, ok := err.Details.(errors.ValidationErrorDetails)
	if !ok {
		t.Fatalf("expected errors.ValidationErrorDetails, got %#v", err.Details)
	}
	if len(details.Errors) != 1 {
		t.Fatalf("expected exactly 1 field error, got %d: %#v", len(details.Errors), details.Errors)
	}
	fe := details.Errors[0]
	if fe.Code != errors.VErrSecretNotFound {
		t.Fatalf("expected %s, got %s", errors.VErrSecretNotFound, fe.Code)
	}
	if got, want := fe.Field, "spec.stack_resources[0].execution_config.env[0].secret_key_ref.secret_name"; got != want {
		t.Fatalf("unexpected field: got %q want %q", got, want)
	}
}

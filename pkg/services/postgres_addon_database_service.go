package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type PostgresAddonDatabaseService interface {
	Create(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	CreateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresAddonDatabase, *errors.ServiceError)
	GetByName(ctx context.Context, postgresAddonID, name string) (*models.PostgresAddonDatabase, *errors.ServiceError)
	Update(ctx context.Context, id string, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresAddonDatabase, *errors.ServiceError)
	ValidateDatabaseExists(ctx context.Context, databaseID string) (bool, *errors.ServiceError)
}

type PostgresAddonDatabaseServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

type postgresAddonDatabaseService struct {
	databaseStore stores.PostgresAddonDatabaseStore
	// TODO: Implement PostgresAddonValidator
	// validator     validator.PostgresAddonValidator
	logger logger.Logger
}

func NewPostgresAddonDatabaseService(spec PostgresAddonDatabaseServiceSpec) PostgresAddonDatabaseService {
	return &postgresAddonDatabaseService{
		databaseStore: pgstore.NewPostgresAddonDatabaseStore(pgstore.PostgresAddonDatabaseStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		// validator: postgres_addon.NewPostgresAddonValidator(), // TODO: implement validator
		logger: spec.Logger,
	}
}

func (s *postgresAddonDatabaseService) Create(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	// Validate database configuration
	if err := s.validateDatabase(database); err != nil {
		return nil, err
	}

	createdDatabase, err := s.databaseStore.Create(ctx, database)
	if err != nil {
		return nil, err
	}

	return createdDatabase, nil
}

func (s *postgresAddonDatabaseService) CreateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	// Validate database configuration
	if err := s.validateDatabase(database); err != nil {
		return nil, err
	}

	createdDatabase, err := s.databaseStore.CreateWithTx(ctx, database)
	if err != nil {
		return nil, err
	}

	return createdDatabase, nil
}

func (s *postgresAddonDatabaseService) GetByID(ctx context.Context, ID string) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	database, err := s.databaseStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return database, nil
}

func (s *postgresAddonDatabaseService) GetByName(ctx context.Context, postgresAddonID, name string) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	database, err := s.databaseStore.GetByName(ctx, postgresAddonID, name)
	if err != nil {
		return nil, err
	}
	return database, nil
}

func (s *postgresAddonDatabaseService) Update(ctx context.Context, id string, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	existingDatabase, err := s.databaseStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Preserve immutable fields
	database.ID = existingDatabase.ID
	database.PostgresAddonID = existingDatabase.PostgresAddonID

	// Validate updated database configuration
	if err := s.validateDatabase(database); err != nil {
		return nil, err
	}

	updatedDatabase, err := s.databaseStore.Update(ctx, database)
	if err != nil {
		return nil, err
	}

	return updatedDatabase, nil
}

func (s *postgresAddonDatabaseService) UpdateWithTx(ctx context.Context, id string, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	existingDatabase, err := s.databaseStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Preserve immutable fields
	database.ID = existingDatabase.ID
	database.PostgresAddonID = existingDatabase.PostgresAddonID

	// Validate updated database configuration
	if err := s.validateDatabase(database); err != nil {
		return nil, err
	}

	updatedDatabase, err := s.databaseStore.UpdateWithTx(ctx, database)
	if err != nil {
		return nil, err
	}

	return updatedDatabase, nil
}

func (s *postgresAddonDatabaseService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := s.databaseStore.Delete(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *postgresAddonDatabaseService) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	if err := s.databaseStore.DeleteWithTx(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *postgresAddonDatabaseService) ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresAddonDatabase, *errors.ServiceError) {
	databases, err := s.databaseStore.ListByPostgresAddon(ctx, postgresAddonID)
	if err != nil {
		return nil, err
	}
	return databases, nil
}

func (s *postgresAddonDatabaseService) ValidateDatabaseExists(ctx context.Context, databaseID string) (bool, *errors.ServiceError) {
	return s.databaseStore.ValidateDatabaseExists(ctx, databaseID)
}

func (s *postgresAddonDatabaseService) validateDatabase(database *models.PostgresAddonDatabase) *errors.ServiceError {
	// Basic validation
	if database.Name == "" {
		return errors.BadRequest("database name is required")
	}

	// Validate PostgreSQL database naming rules
	if len(database.Name) > 63 {
		return errors.BadRequest("database name must be 63 characters or less")
	}

	// Validate extensions if provided
	if database.Extensions != nil {
		for _, extension := range database.Extensions {
			if err := s.validatePostgresExtension(extension); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *postgresAddonDatabaseService) validatePostgresExtension(extension string) *errors.ServiceError {
	// List of supported PostgreSQL extensions
	supportedExtensions := map[string]bool{
		"vector": true,
		// Add more supported extensions as needed
	}

	if !supportedExtensions[extension] {
		return errors.BadRequest("unsupported PostgreSQL extension: %s", extension)
	}

	return nil
}

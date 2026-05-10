package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator/postgresaddon"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PostgresAddonService interface {
	CreatePostgresAddon(ctx context.Context, postgresAddon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	GetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	GetPostgresAddonByName(ctx context.Context, organisationID, name string) (*models.PostgresAddon, *errors.ServiceError)
	UpdatePostgresAddon(ctx context.Context, id string, postgresAddon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	DeletePostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	ListPostgresAddonsByOrganisation(ctx context.Context, organisationID string) ([]*models.PostgresAddon, *errors.ServiceError)
	ListPostgresAddonsForCurrentUser(ctx context.Context, orgID string) ([]*models.PostgresAddon, *errors.ServiceError)
	ListPostgresAddonsByTeamID(ctx context.Context, teamID string) ([]*models.PostgresAddon, *errors.ServiceError)
	UpdatePostgresAddonStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError

	// Lifecycle operations
	TriggerBackup(ctx context.Context, id string) *errors.ServiceError
	TriggerHibernate(ctx context.Context, id string, enabled bool) *errors.ServiceError
	TriggerFence(ctx context.Context, id string, enabled bool) *errors.ServiceError

	// Backup operations
	ListBackups(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError)

	// Credentials
	GetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)

	// Internal operations
	InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	InternalGetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, id string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError)

	InjectClusterManager(clusterManager clustermanager.ClusterManager)

	BackgroundJobEnqueuerInjectable
	ClusterResourceServiceInjectable
}

type PostgresAddonServiceSpec struct {
	SessionFactory        db.SessionFactory
	ObjectStoreService    ObjectStoreService
	ClusterService        ClusterService
	NamespaceService      NamespaceService
	SecretService         SecretService
	PostgresBackupService PostgresBackupService
	ClusterManager        clustermanager.ClusterManager
	TeamService           TeamService
	Logger                logger.Logger
	Permissions           auth.PermissionService
}

type postgresAddonService struct {
	postgresAddonStore stores.PostgresAddonStore
	addonUsageStore    stores.AddonUsageStore
	databaseService    PostgresAddonDatabaseService
	backupService      PostgresBackupService
	namespaceService   NamespaceService
	clusterService     ClusterService
	objectStoreService ObjectStoreService
	secretService      SecretService
	teamService        TeamService
	clusterManager     clustermanager.ClusterManager
	validator          validator.PostgresAddonValidator
	logger             logger.Logger
	sessionFactory     db.SessionFactory
	permissions        auth.PermissionService

	BackgroundJobEnqueuerDep
	ClusterResourceServiceDeps
}

func NewPostgresAddonService(spec PostgresAddonServiceSpec) PostgresAddonService {
	postgresAddonStore := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{
		SessionFactory: spec.SessionFactory,
	})

	databaseService := NewPostgresAddonDatabaseService(PostgresAddonDatabaseServiceSpec{
		SessionFactory: spec.SessionFactory,
		Logger:         spec.Logger,
	})

	addonUsageStore := pgstore.NewAddonUsageStore(pgstore.AddonUsageStoreSpec{
		SessionFactory: spec.SessionFactory,
	})

	return &postgresAddonService{
		postgresAddonStore: postgresAddonStore,
		addonUsageStore:    addonUsageStore,
		databaseService:    databaseService,
		backupService:      spec.PostgresBackupService,
		clusterService:     spec.ClusterService,
		namespaceService:   spec.NamespaceService,
		objectStoreService: spec.ObjectStoreService,
		secretService:      spec.SecretService,
		teamService:        spec.TeamService,
		clusterManager:     spec.ClusterManager,
		validator:          postgresaddon.NewPostgresAddonValidator(),
		logger:             spec.Logger,
		sessionFactory:     spec.SessionFactory,
		permissions:        spec.Permissions,
	}
}

func (s *postgresAddonService) InjectClusterManager(clusterManager clustermanager.ClusterManager) {
	s.clusterManager = clusterManager
}

func (s *postgresAddonService) CreatePostgresAddon(ctx context.Context, postgresAddon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, postgresAddon.TeamID, auth.ResourceAddonsPostgres, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	// Validate input
	if err := s.validator.ValidateForCreate(ctx, postgresAddon); err != nil {
		return nil, err
	}

	namespace, err := s.namespaceService.PrepareNamespaceForAddon(ctx, postgresAddon, postgresAddon.OrganisationID)
	if err != nil {
		return nil, errors.GeneralError("failed to prepare namespace for PostgreSQL addon: %s", err.Error())
	}

	cluster, err := s.clusterService.GetClusterForOrg(ctx, postgresAddon.OrganisationID)
	if err != nil {
		return nil, errors.GeneralError("failed to get cluster for organisation '%s': %s", postgresAddon.OrganisationID, err.Error())
	}
	postgresAddon.ClusterID = cluster.ID
	postgresAddon.Namespace = namespace.Name

	// Check if PostgreSQL addon with same name already exists
	existingPostgresAddon, _ := s.GetPostgresAddonByName(ctx, postgresAddon.OrganisationID, postgresAddon.Name)
	if existingPostgresAddon != nil {
		return nil, errors.Conflict("PostgreSQL addon with name '%s' already exists", postgresAddon.Name)
	}

	// Validate BackupConfig object store if specified
	if postgresAddon.BackupConfig.ObjectStoreID != "" {
		exists, err := s.objectStoreService.ValidateObjectStoreExists(ctx, postgresAddon.BackupConfig.ObjectStoreID)
		if err != nil {
			return nil, errors.GeneralError("failed to validate backup object store: %s", err.Error())
		}
		if !exists {
			return nil, errors.BadRequest("backup object store with ID '%s' does not exist", postgresAddon.BackupConfig.ObjectStoreID)
		}
	}

	// Validate Initialization if specified
	if (postgresAddon.Initialization != models.PostgresInitialization{}) {
		if postgresAddon.Initialization.RestoreFromBackup != nil {
			exists, err := s.backupService.ValidateBackupExists(ctx, postgresAddon.Initialization.RestoreFromBackup.BackupID)
			if err != nil {
				return nil, errors.GeneralError("failed to validate initialization backup: %s", err.Error())
			}
			if !exists {
				return nil, errors.BadRequest("initialization backup with ID '%s' does not exist", postgresAddon.Initialization.RestoreFromBackup.BackupID)
			}
		}

		if postgresAddon.Initialization.ImportFromExternal != nil && postgresAddon.Initialization.ImportFromExternal.PasswordSecretID != "" {
			valid, missingKeys, err := s.secretService.ValidateSecretHasKeys(ctx, postgresAddon.Initialization.ImportFromExternal.PasswordSecretID, []string{models.PasswordSecretKey})
			if err != nil {
				return nil, errors.GeneralError("failed to validate import password secret: %s", err.Error())
			}
			if !valid {
				return nil, errors.BadRequest("import password secret must contain key(s): %v", missingKeys)
			}
		}

		if postgresAddon.Initialization.RestoreFromObjectStore != nil {
			exists, err := s.objectStoreService.ValidateObjectStoreExists(ctx, postgresAddon.Initialization.RestoreFromObjectStore.ObjectStoreID)
			if err != nil {
				return nil, errors.GeneralError("failed to validate initialization object store: %s", err.Error())
			}
			if !exists {
				return nil, errors.BadRequest("initialization object store with ID '%s' does not exist", postgresAddon.Initialization.RestoreFromObjectStore.ObjectStoreID)
			}
		}
	}

	// Set default lifecycle config if not provided
	if (postgresAddon.LifecycleConfig == models.PostgresLifecycleConfig{}) {
		postgresAddon.LifecycleConfig = models.PostgresLifecycleConfig{
			HibernationEnabled: false,
			FencingEnabled:     false,
		}
	}

	// Set default database as 'app' if no databases specified or default database not specified.
	if len(postgresAddon.Databases) == 0 || !postgresAddon.DefaultDatabaseSpecified() {
		postgresAddon.Databases = append(postgresAddon.Databases, models.PostgresAddonDatabase{
			Name: models.DefaultDatabaseName,
		})
	}

	// Set initial status
	postgresAddon.Status = models.PostgresAddonStatus{
		State:   models.PostgresAddonStatePending,
		Message: "PostgreSQL addon is being created",
	}

	var createdPostgresAddon *models.PostgresAddon
	err = s.postgresAddonStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Create namespace
		createdNamepace, err := s.namespaceService.CreateInDBWithTx(ctx, namespace)
		if err != nil {
			return errors.GeneralError("failed to create namespace for PostgreSQL addon: %s", err.Error())
		}
		postgresAddon.NamespaceID = createdNamepace.ID
		// Create PostgreSQL addon within transaction
		var createErr *errors.ServiceError
		createdPostgresAddon, createErr = s.postgresAddonStore.CreateWithTx(ctx, postgresAddon)
		if createErr != nil {
			return createErr
		}
		// Create databases within same transaction
		for _, db := range postgresAddon.Databases {
			db.PostgresAddonID = createdPostgresAddon.ID
			_, createErr := s.databaseService.CreateWithTx(ctx, &db)
			if createErr != nil {
				return createErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get the complete addon with databases loaded
	createdPostgresAddon, err = s.GetPostgresAddon(ctx, createdPostgresAddon.ID)
	if err != nil {
		return nil, err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{
		ID: createdPostgresAddon.ID,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for postgres addon '%s': %s", createdPostgresAddon.Name, err.Error())
	}

	return createdPostgresAddon, nil
}

func (s *postgresAddonService) GetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError) {
	addon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, addon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return addon, nil
}

func (s *postgresAddonService) InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError) {
	return s.postgresAddonStore.GetByID(ctx, id)
}

func (s *postgresAddonService) GetPostgresAddonByName(ctx context.Context, organisationID, name string) (*models.PostgresAddon, *errors.ServiceError) {
	return s.postgresAddonStore.GetByName(ctx, organisationID, name)
}

func (s *postgresAddonService) UpdatePostgresAddon(ctx context.Context, id string, postgresAddon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	existingPostgresAddon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingPostgresAddon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	postgresAddon.ClusterID = existingPostgresAddon.ClusterID
	postgresAddon.NamespaceID = existingPostgresAddon.NamespaceID
	postgresAddon.Namespace = existingPostgresAddon.Namespace
	postgresAddon.OrganisationID = existingPostgresAddon.OrganisationID
	postgresAddon.TeamID = existingPostgresAddon.TeamID

	// Validate update using validator
	if err := s.validator.ValidateForUpdate(ctx, existingPostgresAddon, postgresAddon); err != nil {
		return nil, err
	}

	// Validate BackupConfig object store if specified
	if postgresAddon.BackupConfig.ObjectStoreID != "" {
		exists, err := s.objectStoreService.ValidateObjectStoreExists(ctx, postgresAddon.BackupConfig.ObjectStoreID)
		if err != nil {
			return nil, errors.GeneralError("failed to validate backup object store: %s", err.Error())
		}
		if !exists {
			return nil, errors.BadRequest("backup object store with ID '%s' does not exist", postgresAddon.BackupConfig.ObjectStoreID)
		}
	}

	var updatedPostgresAddon *models.PostgresAddon
	err = s.postgresAddonStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Update PostgreSQL addon within transaction
		var updateErr *errors.ServiceError
		updatedPostgresAddon, updateErr = s.postgresAddonStore.UpdateWithTx(ctx, postgresAddon)
		if updateErr != nil {
			return updateErr
		}

		// Handle database changes with proper diff logic within same transaction
		updateErr = s.updateDatabasesWithDiff(ctx, updatedPostgresAddon.ID, existingPostgresAddon.Databases, postgresAddon.Databases)
		if updateErr != nil {
			return updateErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{
		ID: updatedPostgresAddon.ID,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for postgres addon '%s': %s", updatedPostgresAddon.Name, err.Error())
	}

	return updatedPostgresAddon, nil
}

func (s *postgresAddonService) updateDatabasesWithDiff(ctx context.Context, postgresAddonID string, existingDatabases, newDatabases []models.PostgresAddonDatabase) *errors.ServiceError {
	// Create maps for efficient lookup
	existingMap := make(map[string]models.PostgresAddonDatabase)
	for _, db := range existingDatabases {
		existingMap[db.Name] = db
	}

	newMap := make(map[string]models.PostgresAddonDatabase)
	for _, db := range newDatabases {
		newMap[db.Name] = db
	}

	// Handle creates and updates
	for _, newDb := range newDatabases {
		if existingDb, exists := existingMap[newDb.Name]; exists {
			// Update existing database if extensions changed
			if !s.extensionsEqual(existingDb.Extensions, newDb.Extensions) {
				newDb.ID = existingDb.ID
				newDb.PostgresAddonID = postgresAddonID
				_, err := s.databaseService.UpdateWithTx(ctx, existingDb.ID, &newDb)
				if err != nil {
					return err
				}
			}
		} else {
			// Create new database
			newDb.PostgresAddonID = postgresAddonID
			_, err := s.databaseService.CreateWithTx(ctx, &newDb)
			if err != nil {
				return err
			}
		}
	}

	// Handle deletes
	for _, existingDb := range existingDatabases {
		if _, exists := newMap[existingDb.Name]; !exists {
			err := s.databaseService.DeleteWithTx(ctx, existingDb.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *postgresAddonService) extensionsEqual(a, b models.PostgresExtensions) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, ext := range a {
		aMap[ext] = true
	}

	for _, ext := range b {
		if !aMap[ext] {
			return false
		}
	}

	return true
}

func (s *postgresAddonService) DeletePostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError) {
	postgresAddon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, postgresAddon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionDelete); permErr != nil {
		return nil, permErr
	}

	// Check addon usage before deletion
	inUse, usageErr := s.addonUsageStore.IsAddonInUse(ctx, models.AddonTypePostgres, id)
	if usageErr != nil {
		return nil, errors.GeneralError("failed to check addon usage: %v", usageErr)
	}
	if inUse {
		return nil, errors.BadRequest("addon is in use by one or more stacks and cannot be deleted")
	}

	// Mark for deletion
	postgresAddon.Status.State = models.PostgresAddonStateDeleting
	postgresAddon.Status.Message = "PostgreSQL addon is being deleted"

	err = s.postgresAddonStore.UpdateStatus(ctx, id, &postgresAddon.Status)
	if err != nil {
		return nil, errors.GeneralError("failed to update PostgreSQL addon status for deletion: %s", err.Error())
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{
		ID: id,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for postgres addon '%s': %s", postgresAddon.Name, err.Error())
	}

	return postgresAddon, nil
}

func (s *postgresAddonService) ListPostgresAddonsByOrganisation(ctx context.Context, organisationID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceAddonsPostgres, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.postgresAddonStore.ListByOrganisation(ctx, organisationID)
}

func (s *postgresAddonService) ListPostgresAddonsForCurrentUser(ctx context.Context, orgID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}

	if identity.IsOrgAdmin() {
		return s.postgresAddonStore.ListByOrganisation(ctx, orgID)
	}

	memberships, serr := s.teamService.InternalListUserTeams(ctx, identity.UserID, orgID)
	if serr != nil {
		return nil, serr
	}

	var allowedTeamIDs []string
	for _, m := range memberships {
		if permErr := s.permissions.Check(ctx, m.TeamID, auth.ResourceAddonsPostgres, "", auth.ActionList); permErr == nil {
			allowedTeamIDs = append(allowedTeamIDs, m.TeamID)
		}
	}

	return s.postgresAddonStore.ListByTeamIDs(ctx, allowedTeamIDs)
}

func (s *postgresAddonService) ListPostgresAddonsByTeamID(ctx context.Context, teamID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceAddonsPostgres, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.postgresAddonStore.ListByTeamID(ctx, teamID)
}

func (s *postgresAddonService) UpdatePostgresAddonStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError {
	return s.postgresAddonStore.UpdateStatus(ctx, id, status)
}

// Lifecycle operations - accept updated models and persist lifecycle config changes
func (s *postgresAddonService) RequestPostgresAddonBackup(ctx context.Context, id string) *errors.ServiceError {
	now := time.Now().UTC()
	return s.postgresAddonStore.UpdateBackupRequestedAt(ctx, id, &now)
}

// Internal operations
func (s *postgresAddonService) InternalDeleteFromDB(ctx context.Context, id string) *errors.ServiceError {
	err := s.postgresAddonStore.Delete(ctx, id)
	if err != nil {
		if err.Is404() {
			return nil // Already deleted
		}
		return errors.GeneralError("failed to delete PostgreSQL addon with ID '%s' from database: %s", id, err.Error())
	}
	return nil
}

func (s *postgresAddonService) InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError) {
	return s.postgresAddonStore.InternalList(ctx, query, args...)
}

// Composition services
func (s *postgresAddonService) DatabaseService() PostgresAddonDatabaseService {
	return s.databaseService
}

// TODO: Implement PostgresAddonBackupService
// func (s *postgresAddonService) BackupService() PostgresAddonBackupService {
//	return s.backupService
// }

// TriggerBackup triggers a backup for the postgres addon
func (s *postgresAddonService) TriggerBackup(ctx context.Context, id string) *errors.ServiceError {
	addon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if permErr := s.permissions.Check(ctx, addon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionWrite); permErr != nil {
		return permErr
	}

	// Update the backup requested timestamp
	now := time.Now()
	if err := s.postgresAddonStore.UpdateBackupRequestedAt(ctx, id, &now); err != nil {
		return err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{
		ID: id,
	}); err != nil {
		return errors.GeneralError("failed to enqueue backup job for postgres addon '%s': %s", id, err.Error())
	}
	return nil
}

// TriggerHibernate triggers hibernation for the postgres addon
func (s *postgresAddonService) TriggerHibernate(ctx context.Context, id string, enabled bool) *errors.ServiceError {
	// Get the current postgres addon
	postgresAddon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if permErr := s.permissions.Check(ctx, postgresAddon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionWrite); permErr != nil {
		return permErr
	}

	// Update lifecycle config
	postgresAddon.LifecycleConfig.HibernationEnabled = enabled

	_, err = s.postgresAddonStore.Update(ctx, postgresAddon)
	if err != nil {
		return err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{ID: id}); err != nil {
		return errors.GeneralError("failed to enqueue hibernation job for postgres addon '%s': %s", postgresAddon.Name, err.Error())
	}
	return nil
}

// TriggerFence triggers fencing for the postgres addon
func (s *postgresAddonService) TriggerFence(ctx context.Context, id string, enabled bool) *errors.ServiceError {
	// Get the current postgres addon
	postgresAddon, err := s.postgresAddonStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if permErr := s.permissions.Check(ctx, postgresAddon.TeamID, auth.ResourceAddonsPostgres, id, auth.ActionWrite); permErr != nil {
		return permErr
	}

	// Update lifecycle config
	postgresAddon.LifecycleConfig.FencingEnabled = enabled

	_, err = s.postgresAddonStore.Update(ctx, postgresAddon)
	if err != nil {
		return err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PostgresAddon{ID: id}); err != nil {
		return errors.GeneralError("failed to enqueue fencing job for postgres addon '%s': %s", postgresAddon.Name, err.Error())
	}
	return nil
}

// ListBackups lists all backups for a postgres addon
func (s *postgresAddonService) ListBackups(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError) {
	addon, err := s.postgresAddonStore.GetByID(ctx, postgresAddonID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, addon.TeamID, auth.ResourceAddonsPostgres, postgresAddonID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.backupService.ListByPostgresAddon(ctx, postgresAddonID)
}

func (s *postgresAddonService) GetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError) {
	addon, err := s.postgresAddonStore.GetByID(ctx, addonID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, addon.TeamID, auth.ResourceAddonsPostgres, addonID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.getCredentialsForAddon(ctx, addon, database, superuser)
}

func (s *postgresAddonService) InternalGetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError) {
	addon, err := s.postgresAddonStore.GetByID(ctx, addonID)
	if err != nil {
		return nil, err
	}
	return s.getCredentialsForAddon(ctx, addon, database, superuser)
}

func (s *postgresAddonService) getCredentialsForAddon(ctx context.Context, addon *models.PostgresAddon, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError) {
	ok := models.IsConditionTrue(addon.Status.Conditions, string(models.PostgresAddonConditionReadyOnce))
	if !ok {
		return nil, errors.BadRequest("credentials not available until addon has been ready at least once")
	}

	if addon.Status.ConnectionInfo == nil || addon.Status.ConnectionInfo.ClusterSecrets == nil {
		return nil, errors.BadRequest("addon connection info not available yet")
	}

	if superuser {
		if !addon.Configuration.EnableSuperuserAccess {
			return nil, errors.Forbidden("superuser access not enabled for this addon")
		}
		if addon.Status.ConnectionInfo.ClusterSecrets.SuperuserSecret == nil {
			return nil, errors.BadRequest("superuser secret not available yet")
		}
	} else {
		if !addon.HasDatabase(database) {
			return nil, errors.NotFound("database '%s' not found in addon", database)
		}

		_, found := lo.Find(addon.Status.Databases, func(d models.PostgresDatabaseInfo) bool {
			return d.Name == database
		})
		if !found {
			return nil, errors.NotFound("database '%s' not applied", database)
		}
	}

	clusterClient, cerr := s.clusterManager.GetClient(addon.ClusterID)
	if cerr != nil {
		return nil, errors.GeneralError("unable to connect to cluster: %v", cerr)
	}

	var secretName string
	if superuser {
		secretName = *addon.Status.ConnectionInfo.ClusterSecrets.SuperuserSecret
	} else {
		// Find the database owner from status, then look up the owner's secret.
		// CNPG creates credential secrets per database owner, not per database name.
		owner := ""
		for _, dbInfo := range addon.Status.Databases {
			if dbInfo.Name == database {
				owner = dbInfo.Owner
				break
			}
		}
		if owner == "" {
			return nil, errors.NotFound("credential secret for database '%s' not found", database)
		}
		var ok bool
		secretName, ok = addon.Status.ConnectionInfo.ClusterSecrets.UserSecrets[owner]
		if !ok {
			return nil, errors.NotFound("credential secret for database '%s' not found", database)
		}
	}

	// Fetch secret from cluster (JIT - never stored in API server DB)
	secret := &corev1.Secret{}
	if cerr := clusterClient.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: addon.Namespace,
	}, secret); cerr != nil {
		return nil, errors.GeneralError("credential secret not found in cluster: %v", cerr)
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])

	dbName := database
	if superuser && database == "" {
		dbName = "postgres"
	}

	creds := &models.PostgresCredentials{
		Database: dbName,
		Host:     addon.Status.ConnectionInfo.Host,
		Port:     addon.Status.ConnectionInfo.Port,
		Username: username,
		Password: password,
		SSLMode:  addon.Status.ConnectionInfo.SSLMode,
		ConnectionString: fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			username, password,
			addon.Status.ConnectionInfo.Host, addon.Status.ConnectionInfo.Port,
			dbName, addon.Status.ConnectionInfo.SSLMode,
		),
	}

	// Fetch CA certificate if available
	caSecretName := addon.Status.ConnectionInfo.ClusterSecrets.CACertificateSecret
	if caSecretName != "" {
		caSecret := &corev1.Secret{}
		if cerr := clusterClient.Get(ctx, client.ObjectKey{
			Name:      caSecretName,
			Namespace: addon.Namespace,
		}, caSecret); cerr == nil {
			creds.CACertificate = string(caSecret.Data["ca.crt"])
		}
	}

	return creds, nil
}

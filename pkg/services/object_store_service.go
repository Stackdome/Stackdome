package services

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/Stackdome/stackdome/pkg/validator/objectstore"
	barmancloudv1 "github.com/cloudnative-pg/plugin-barman-cloud/api/v1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectStoreService interface {
	Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError)
	Update(ctx context.Context, id string, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string) ([]*models.ObjectStore, *errors.ServiceError)
	ListObjectStoresForCurrentUser(ctx context.Context, orgID string) ([]*models.ObjectStore, *errors.ServiceError)
	ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError)
	TestConnection(ctx context.Context, objectStoreID string) *errors.ServiceError
	UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError
}

type ObjectStoreServiceSpec struct {
	SessionFactory db.SessionFactory
	SecretService  SecretService
	TeamService    TeamService
	ClusterManager clustermanager.ClusterManager
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

type objectStoreService struct {
	objectStoreStore stores.ObjectStoreStore
	secretService    SecretService
	teamService      TeamService
	clusterManager   clustermanager.ClusterManager
	validator        validator.ObjectStoreValidator
	permissions      auth.PermissionService
	logger           logger.Logger
}

func NewObjectStoreService(spec ObjectStoreServiceSpec) ObjectStoreService {
	return &objectStoreService{
		objectStoreStore: pgstore.NewObjectStoreStore(pgstore.ObjectStoreStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		secretService:  spec.SecretService,
		teamService:    spec.TeamService,
		clusterManager: spec.ClusterManager,
		validator:      objectstore.NewObjectStoreValidator(),
		permissions:    spec.Permissions,
		logger:         spec.Logger,
	}
}

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

func (s *objectStoreService) Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, objectStore.TeamID, auth.ResourceObjectStores, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	// Validate input
	if err := s.validator.ValidateForCreate(ctx, objectStore); err != nil {
		return nil, err
	}

	// Validate secret references exist
	if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
		return nil, err
	}

	// Set default retention policy if not provided
	if objectStore.RetentionPolicy == "" {
		objectStore.RetentionPolicy = models.DefaultObjectStoreRetentionPolicy
	}

	createdObjectStore, err := s.objectStoreStore.Create(ctx, objectStore)
	if err != nil {
		return nil, err
	}

	return createdObjectStore, nil
}

func (s *objectStoreService) GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError) {
	objectStore, err := s.objectStoreStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, objectStore.TeamID, auth.ResourceObjectStores, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return objectStore, nil
}

func (s *objectStoreService) InternalGetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError) {
	return s.objectStoreStore.GetByID(ctx, ID)
}

func (s *objectStoreService) GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError) {
	objectStore, err := s.objectStoreStore.GetByName(ctx, organisationID, name)
	if err != nil {
		return nil, err
	}
	return objectStore, nil
}

func (s *objectStoreService) Update(ctx context.Context, id string, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	existingObjectStore, err := s.objectStoreStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingObjectStore.TeamID, auth.ResourceObjectStores, id, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	// Validate immutable field changes before overwriting
	if err := s.validator.ValidateForUpdate(ctx, existingObjectStore, objectStore); err != nil {
		return nil, err
	}

	// Preserve immutable fields (must be done AFTER validation check)
	objectStore.ID = existingObjectStore.ID
	objectStore.OrganisationID = existingObjectStore.OrganisationID
	objectStore.Name = existingObjectStore.Name

	// Validate the full spec including configuration
	if err := s.validator.ValidateForCreate(ctx, objectStore); err != nil {
		return nil, err
	}

	// Validate secret references exist
	if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
		return nil, err
	}

	updatedObjectStore, err := s.objectStoreStore.Update(ctx, objectStore)
	if err != nil {
		return nil, err
	}

	return updatedObjectStore, nil
}

func (s *objectStoreService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	objectStore, err := s.objectStoreStore.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if permErr := s.permissions.Check(ctx, objectStore.TeamID, auth.ResourceObjectStores, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}

	inUse, err := s.objectStoreStore.IsReferencedByAddon(ctx, ID)
	if err != nil {
		return err
	}
	if inUse {
		return errors.Conflict("object store is in use by one or more PostgreSQL addons and cannot be deleted")
	}

	for _, deployed := range objectStore.Status.DeployedClusters {
		s.cleanupFromCluster(ctx, objectStore, deployed)
	}

	return s.objectStoreStore.Delete(ctx, ID)
}

func (s *objectStoreService) cleanupFromCluster(ctx context.Context, objectStore *models.ObjectStore, deployed models.DeployedClusterInfo) {
	clusterClient, clientErr := s.clusterManager.GetClient(deployed.ClusterID)
	if clientErr != nil {
		s.logger.Errorf("failed to get client for cluster %s during ObjectStore cleanup: %v", deployed.ClusterID, clientErr)
		return
	}

	osCR := &barmancloudv1.ObjectStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectStore.Name,
			Namespace: deployed.Namespace,
		},
	}
	if deleteErr := clusterClient.Delete(ctx, osCR); deleteErr != nil && !k8sapierrors.IsNotFound(deleteErr) {
		s.logger.Errorf("failed to delete ObjectStore CR %s from cluster %s: %v", objectStore.Name, deployed.ClusterID, deleteErr)
	}

	credSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("objectstore-%s-credentials", objectStore.Name),
			Namespace: deployed.Namespace,
		},
	}
	if deleteErr := clusterClient.Delete(ctx, credSecret, client.PropagationPolicy(metav1.DeletePropagationBackground)); deleteErr != nil && !k8sapierrors.IsNotFound(deleteErr) {
		s.logger.Errorf("failed to delete credential secret for ObjectStore %s from cluster %s: %v", objectStore.Name, deployed.ClusterID, deleteErr)
	}
}

func (s *objectStoreService) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceObjectStores, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	objectStores, err := s.objectStoreStore.ListByOrganisation(ctx, organisationID)
	if err != nil {
		return nil, err
	}
	return objectStores, nil
}

func (s *objectStoreService) ListByTeamID(ctx context.Context, teamID string) ([]*models.ObjectStore, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceObjectStores, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.objectStoreStore.ListByTeamID(ctx, teamID)
}

func (s *objectStoreService) ListObjectStoresForCurrentUser(ctx context.Context, orgID string) ([]*models.ObjectStore, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}

	if identity.IsOrgAdmin() {
		return s.objectStoreStore.ListByOrganisation(ctx, orgID)
	}

	memberships, serr := s.teamService.InternalListUserTeams(ctx, identity.UserID, orgID)
	if serr != nil {
		return nil, serr
	}

	var allowedTeamIDs []string
	for _, m := range memberships {
		if permErr := s.permissions.Check(ctx, m.TeamID, auth.ResourceObjectStores, "", auth.ActionList); permErr == nil {
			allowedTeamIDs = append(allowedTeamIDs, m.TeamID)
		}
	}

	return s.objectStoreStore.ListByTeamIDs(ctx, allowedTeamIDs)
}

func (s *objectStoreService) ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError) {
	return s.objectStoreStore.ValidateObjectStoreExists(ctx, objectStoreID)
}

func (s *objectStoreService) TestConnection(ctx context.Context, objectStoreID string) *errors.ServiceError {
	objectStore, err := s.objectStoreStore.GetByID(ctx, objectStoreID)
	if err != nil {
		return err
	}

	// TODO: Implement actual connection testing based on credential type
	// This would involve:
	// - For S3: Test ListObjects operation
	// - For Azure: Test blob container access
	// - For GCS: Test bucket access

	s.logger.Info(ctx, "Object store connection test not yet implemented", "objectStoreId", objectStore.ID)
	return nil
}

func (s *objectStoreService) UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError {
	return s.objectStoreStore.UpdateStatus(ctx, id, status)
}

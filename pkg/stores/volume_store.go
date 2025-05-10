package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type VolumeStore interface {
	Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Volume, *errors.ServiceError)
	UpdateGitRepoSourceRevision(ctx context.Context, id string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError)
	UpdateGitRepoSourceRevisionWithTx(ctx context.Context, id string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError)
	UpdateRemoteDirSourceHash(ctx context.Context, id string, remoteDirHash string) (*models.Volume, *errors.ServiceError)
	UpdateRemoteDirSourceHashWithTx(ctx context.Context, id string, remoteDirHash string) (*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.VolumeStatus) *errors.ServiceError
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	GetByUserID(ctx context.Context, id string) ([]*models.Volume, *errors.ServiceError)
	GetByVolumeNameAndNamespace(ctx context.Context, volumeName string, namespace string) (*models.Volume, *errors.ServiceError)
	AtomicExecutor
}

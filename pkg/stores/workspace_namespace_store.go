package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type WorkspaceNamespaceStore interface {
	Create(context.Context, *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError)
	CreateBatch(context.Context, []*models.WorkspaceNamespace) ([]*models.WorkspaceNamespace, *errors.ServiceError)
	CreateBatchWithTx(context.Context, []*models.WorkspaceNamespace) ([]*models.WorkspaceNamespace, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceNamespace, *errors.ServiceError)
	GetByNamespace(ctx context.Context, namespace string) (*models.WorkspaceNamespace, *errors.ServiceError)
	GetByWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.WorkspaceNamespace, *errors.ServiceError)
	Update(ctx context.Context, workspaceName string, userID string, spec *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, workspaceName string, userID string, spec *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError)
}

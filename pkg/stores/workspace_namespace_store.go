package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
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

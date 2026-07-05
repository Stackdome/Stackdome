package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type StackConnectionStore interface {
	ListByStackID(ctx context.Context, stackID string) (models.StackConnections, *errors.ServiceError)
	CreateWithTx(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	DeleteWithTx(ctx context.Context, stackID, connectionID string) *errors.ServiceError
	ReplaceByStackIDWithTx(ctx context.Context, stackID string, connections models.StackConnections) *errors.ServiceError
	IsNodeReferenced(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error)
	IsNodeReferencedAsSource(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error)
	IsNodeReferencedAsTarget(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error)
}

package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=stack_store.go -destination=../mocks/mock_stack_store.go -package=mocks

type StackStore interface {
	Create(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Stack, *errors.ServiceError)
	GetByName(ctx context.Context, Name string, userID string) (*models.Stack, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateRevision(ctx context.Context, id string, revision string) *errors.ServiceError
	UpdateConnectionsWithTx(ctx context.Context, id string, connections models.StackConnections) *errors.ServiceError
	CreateConnectionWithTx(ctx context.Context, id string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	UpdateConnectionWithTx(ctx context.Context, id string, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	DeleteConnectionWithTx(ctx context.Context, id string, connectionID string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.StackStatus) *errors.ServiceError
	UpdateForDelete(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Delete(ctx context.Context, id string) *errors.ServiceError
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError)
	ListByTeamIDs(ctx context.Context, teamIDs []string) ([]*models.Stack, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError)
	AtomicExecutor
}

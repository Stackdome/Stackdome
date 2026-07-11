package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=stack_store.go -destination=../mocks/mock_stack_store.go -package=mocks

type StackStore interface {
	Create(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Stack, *errors.ServiceError)
	// LockByID takes a row-level lock on the stack (SELECT ... FOR UPDATE).
	// Only meaningful inside WithTransaction; serializes concurrent mutations
	// that must observe each other (e.g. duplicate-name checks on create).
	LockByID(ctx context.Context, id string) *errors.ServiceError
	// GetByNameAndProjectID resolves a stack by name within a project scope. Stack
	// names are unique per project, so this is the canonical lookup for
	// name-addressed operations (e.g. declarative apply).
	GetByNameAndProjectID(ctx context.Context, name string, projectID string) (*models.Stack, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateRevision(ctx context.Context, id string, revision string) *errors.ServiceError
	UpdateConnectionsWithTx(ctx context.Context, id string, connections models.StackConnections) *errors.ServiceError
	CreateConnectionWithTx(ctx context.Context, id string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	UpdateConnectionWithTx(ctx context.Context, id string, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	DeleteConnectionWithTx(ctx context.Context, id string, connectionID string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateShellWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.StackStatus) *errors.ServiceError
	UpdateForDelete(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Delete(ctx context.Context, id string) *errors.ServiceError
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError)
	ListByProjectID(ctx context.Context, projectID string) ([]*models.Stack, *errors.ServiceError)
	ListByProjectIDs(ctx context.Context, projectIDs []string) ([]*models.Stack, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError)
	AtomicExecutor
}

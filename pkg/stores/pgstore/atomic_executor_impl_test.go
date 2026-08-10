package pgstore

import (
	"context"
	"database/sql"
	stderrors "errors"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/glebarez/sqlite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type atomicTestSessionFactory struct {
	database *gorm.DB
}

func newAtomicTestSessionFactory(t *testing.T) *atomicTestSessionFactory {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	return &atomicTestSessionFactory{database: database}
}

func (f *atomicTestSessionFactory) Init(*config.DatabaseConfig) {}
func (f *atomicTestSessionFactory) DirectDB() *sql.DB {
	database, _ := f.database.DB()
	return database
}
func (f *atomicTestSessionFactory) New(ctx context.Context) *gorm.DB {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return f.database.WithContext(ctx)
}
func (f *atomicTestSessionFactory) CheckConnection() error { return nil }
func (f *atomicTestSessionFactory) Close() error {
	database, _ := f.database.DB()
	return database.Close()
}

type failingCommitPool struct {
	gorm.ConnPool
	transaction gorm.TxCommitter
	err         error
}

func (p *failingCommitPool) Commit() error   { return p.err }
func (p *failingCommitPool) Rollback() error { return p.transaction.Rollback() }

func failTransactionCommit(t *testing.T, ctx context.Context, commitErr error) {
	t.Helper()
	tx := db.TxFromContext(ctx)
	if tx == nil {
		t.Fatal("transaction missing from callback context")
	}
	transaction, ok := tx.Statement.ConnPool.(gorm.TxCommitter)
	if !ok {
		t.Fatalf("transaction connection %T cannot commit", tx.Statement.ConnPool)
	}
	tx.Statement.ConnPool = &failingCommitPool{
		ConnPool:    tx.Statement.ConnPool,
		transaction: transaction,
		err:         commitErr,
	}
}

func TestAtomicExecutorRunsRollbackHooksInLIFOOrderOnFunctionError(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	originalErr := apperrors.GeneralError("callback failed")
	var order []string

	got := executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(context.Context) error {
			order = append(order, "first")
			return nil
		}); err != nil {
			t.Fatalf("register first rollback hook: %v", err)
		}
		if err := db.OnRollback(ctx, func(context.Context) error {
			order = append(order, "second")
			return nil
		}); err != nil {
			t.Fatalf("register second rollback hook: %v", err)
		}
		return originalErr
	})

	if got != originalErr {
		t.Fatalf("expected original callback error, got %v", got)
	}
	if want := []string{"second", "first"}; !equalStrings(order, want) {
		t.Fatalf("rollback order = %v, want %v", order, want)
	}
}

func TestAtomicExecutorRunsOuterRollbackHooksAfterNestedTransactionFailure(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	originalErr := apperrors.GeneralError("outer callback failed")
	var compensated bool

	got := executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if nestedErr := executor.WithTransaction(ctx, func(nestedCtx context.Context) *apperrors.ServiceError {
			if err := db.OnRollback(nestedCtx, func(context.Context) error {
				compensated = true
				return nil
			}); err != nil {
				t.Fatalf("register nested rollback hook: %v", err)
			}
			return nil
		}); nestedErr != nil {
			t.Fatalf("nested transaction failed: %v", nestedErr)
		}
		if compensated {
			t.Fatal("nested executor ran rollback hook before the outer transaction finished")
		}
		return originalErr
	})

	if got != originalErr {
		t.Fatalf("expected original callback error, got %v", got)
	}
	if !compensated {
		t.Fatal("outer transaction did not run the nested rollback hook")
	}
}

func TestAtomicExecutorRunsRollbackHooksAfterCommitFailure(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	commitErr := stderrors.New("commit failed")
	var compensated bool

	got := executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(context.Context) error {
			compensated = true
			return nil
		}); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		failTransactionCommit(t, ctx, commitErr)
		return nil
	})

	if got == nil || got.Error() != "error: failed to commit transaction: commit failed" {
		t.Fatalf("expected commit error, got %v", got)
	}
	if !compensated {
		t.Fatal("commit failure did not run rollback hook")
	}
}

func TestAtomicExecutorDoesNotRunRollbackHooksAfterSuccessfulCommit(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	var compensated bool

	got := executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(context.Context) error {
			compensated = true
			return nil
		}); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		return nil
	})

	if got != nil {
		t.Fatalf("successful transaction returned %v", got)
	}
	if compensated {
		t.Fatal("successful commit ran rollback compensation")
	}
}

func TestAtomicExecutorDoesNotCompensateWhenPostCommitHookPanics(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	var compensated bool

	defer func() {
		if recovered := recover(); recovered != "post-commit panic" {
			t.Fatalf("recovered %v, want post-commit panic", recovered)
		}
		if compensated {
			t.Fatal("post-commit panic ran rollback compensation after commit")
		}
	}()

	_ = executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(context.Context) error {
			compensated = true
			return nil
		}); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		db.OnPostCommit(ctx, func() { panic("post-commit panic") })
		return nil
	})
}

func TestAtomicExecutorProvidesUsableContextToRollbackHooks(t *testing.T) {
	type requestMarkerKey struct{}

	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	requestCtx, cancel := context.WithCancel(context.Background())
	markerKey := requestMarkerKey{}
	requestCtx = context.WithValue(requestCtx, markerKey, "request-value")
	originalErr := apperrors.GeneralError("callback failed")

	got := executor.WithTransaction(requestCtx, func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(compensationCtx context.Context) error {
			if db.TxFromContext(compensationCtx) != nil {
				t.Fatal("compensation context contains completed transaction")
			}
			if compensationCtx.Err() != nil {
				t.Fatalf("compensation context is cancelled: %v", compensationCtx.Err())
			}
			if value := compensationCtx.Value(markerKey); value != "request-value" {
				t.Fatalf("compensation context lost request value: %v", value)
			}
			return nil
		}); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		cancel()
		return originalErr
	})

	if got != originalErr {
		t.Fatalf("expected original callback error, got %v", got)
	}
}

func TestAtomicExecutorRunsRollbackHooksBeforeRepanicking(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	var compensated bool

	defer func() {
		if recovered := recover(); recovered != "callback panic" {
			t.Fatalf("recovered %v, want callback panic", recovered)
		}
		if !compensated {
			t.Fatal("panic did not run rollback hook")
		}
	}()

	_ = executor.WithTransaction(context.Background(), func(ctx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(ctx, func(context.Context) error {
			compensated = true
			return nil
		}); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		panic("callback panic")
	})
}

func TestAtomicExecutorLogsCompensationFailuresWithoutMaskingOriginalError(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	ctx := logger.AddLoggerToContext(context.Background(), log)
	originalErr := apperrors.GeneralError("callback failed")
	compensationErr := stderrors.New("compensation failed")

	log.EXPECT().Error(gomock.Any(), "transaction rollback compensation failed: %v", compensationErr)
	got := executor.WithTransaction(ctx, func(txCtx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(txCtx, func(context.Context) error { return compensationErr }); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		return originalErr
	})

	if got != originalErr {
		t.Fatalf("compensation failure masked original error: got %v", got)
	}
}

func TestAtomicExecutorRecoversCompensationPanicWithoutMaskingOriginalError(t *testing.T) {
	factory := newAtomicTestSessionFactory(t)
	t.Cleanup(func() { _ = factory.Close() })
	executor := NewAtomicExecutor(factory)
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	ctx := logger.AddLoggerToContext(context.Background(), log)
	originalErr := apperrors.GeneralError("callback failed")

	log.EXPECT().Error(gomock.Any(), "transaction rollback compensation failed: %v", gomock.Any()).
		Do(func(_ context.Context, _ string, err error) {
			if err.Error() != "rollback hook panicked: compensation panic" {
				t.Fatalf("unexpected compensation panic error: %v", err)
			}
		})
	got := executor.WithTransaction(ctx, func(txCtx context.Context) *apperrors.ServiceError {
		if err := db.OnRollback(txCtx, func(context.Context) error { panic("compensation panic") }); err != nil {
			t.Fatalf("register rollback hook: %v", err)
		}
		return originalErr
	})

	if got != originalErr {
		t.Fatalf("compensation panic masked original error: got %v", got)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

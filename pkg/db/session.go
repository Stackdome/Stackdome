package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"gorm.io/gorm"
)

type xactionCtxKey string

const (
	transactionCtxKey xactionCtxKey = "transaction"
	postCommitCtxKey  xactionCtxKey = "postCommitHooks"
)

type PostCommitHooks struct {
	hooks         []func()
	rollbackHooks []func(context.Context) error
}

func (h *PostCommitHooks) Run() {
	for _, fn := range h.hooks {
		fn()
	}
}

// RunRollback runs compensations in reverse registration order. Every hook is
// attempted; a hook error or panic is returned to the transaction owner for
// logging and never prevents the remaining compensations from running.
func (h *PostCommitHooks) RunRollback(ctx context.Context) []error {
	var errs []error
	for index := len(h.rollbackHooks) - 1; index >= 0; index-- {
		if err := runRollbackHook(ctx, h.rollbackHooks[index]); err != nil {
			errs = append(errs, err)
		}
	}
	h.rollbackHooks = nil
	return errs
}

func runRollbackHook(ctx context.Context, hook func(context.Context) error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("rollback hook panicked: %v", recovered)
		}
	}()
	return hook(ctx)
}

func CtxWithPostCommitHooks(ctx context.Context) (context.Context, *PostCommitHooks) {
	h := &PostCommitHooks{}
	return context.WithValue(ctx, postCommitCtxKey, h), h
}

func postCommitHooksFromCtx(ctx context.Context) *PostCommitHooks {
	h, _ := ctx.Value(postCommitCtxKey).(*PostCommitHooks)
	return h
}

// OnPostCommit registers fn to run after the outermost transaction commits.
// If called outside a transaction, fn runs immediately.
// Hooks are fire-and-forget: errors are logged by the caller inside fn,
// not propagated, because the transaction is already committed.
func OnPostCommit(ctx context.Context, fn func()) {
	if h := postCommitHooksFromCtx(ctx); h != nil {
		h.hooks = append(h.hooks, fn)
		return
	}
	fn()
}

// OnRollback registers a compensation owned by the outermost transaction.
// The callback receives a non-cancelled context without the completed
// transaction, so it can safely perform database or external cleanup after a
// rollback. Registering outside an atomic transaction is an error.
func OnRollback(ctx context.Context, fn func(context.Context) error) error {
	if h := postCommitHooksFromCtx(ctx); h != nil {
		h.rollbackHooks = append(h.rollbackHooks, fn)
		return nil
	}
	return fmt.Errorf("rollback hook requires an outer transaction")
}

type SessionFactory interface {
	Init(*config.DatabaseConfig)
	DirectDB() *sql.DB
	New(ctx context.Context) *gorm.DB
	CheckConnection() error
	Close() error
}

func CtxWithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionCtxKey, tx)
}

func TxFromContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(transactionCtxKey).(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}

type completedTransaction struct{}

// ContextAfterTransaction preserves request values for compensation while
// removing cancellation and the completed transaction handle.
func ContextAfterTransaction(ctx context.Context) context.Context {
	// A new value for the same key shadows the completed *gorm.DB stored by the
	// transaction context. TxFromContext deliberately ignores non-*gorm.DB values.
	return context.WithValue(context.WithoutCancel(ctx), transactionCtxKey, completedTransaction{})
}

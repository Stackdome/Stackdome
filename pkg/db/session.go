package db

import (
	"context"
	"database/sql"

	"github.com/Stackdome/stackdome/config"
	"gorm.io/gorm"
)

type xactionCtxKey string

const (
	transactionCtxKey xactionCtxKey = "transaction"
	postCommitCtxKey  xactionCtxKey = "postCommitHooks"
)

type PostCommitHooks struct {
	hooks []func()
}

func (h *PostCommitHooks) Run() {
	for _, fn := range h.hooks {
		fn()
	}
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

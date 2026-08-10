package pgstore

import (
	"context"
	"database/sql"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type atomicExecutor struct {
	sessionFactory db.SessionFactory
}

func NewAtomicExecutor(sf db.SessionFactory) stores.AtomicExecutor {
	return &atomicExecutor{sessionFactory: sf}
}

func (a *atomicExecutor) WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError {
	var (
		tx        *gorm.DB
		txCtx     context.Context
		isOwner   bool
		hooks     *db.PostCommitHooks
		committed bool
	)
	tx = db.TxFromContext(ctx)
	if tx == nil {
		tx = a.sessionFactory.New(ctx).Begin()
		if tx.Error != nil {
			return errors.GeneralError("failed to begin transaction: %s", tx.Error.Error())
		}
		txCtx = db.CtxWithTransaction(ctx, tx)
		txCtx, hooks = db.CtxWithPostCommitHooks(txCtx)
		isOwner = true
	} else {
		txCtx = ctx
	}

	defer func() {
		if r := recover(); r != nil {
			if isOwner && !committed {
				a.rollbackAndCompensate(txCtx, tx, hooks)
			}
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if isOwner {
			a.rollbackAndCompensate(txCtx, tx, hooks)
		}
		return err
	}

	if isOwner {
		if err := tx.Commit().Error; err != nil {
			a.rollbackAndCompensate(txCtx, tx, hooks)
			return errors.GeneralError("failed to commit transaction: %s", err.Error())
		}
		committed = true
		hooks.Run()
	}
	return nil
}

func (a *atomicExecutor) rollbackAndCompensate(ctx context.Context, tx *gorm.DB, hooks *db.PostCommitHooks) {
	log := logger.GetLoggerFromContext(ctx)
	if err := tx.Rollback().Error; err != nil && !stderrors.Is(err, sql.ErrTxDone) {
		log.Error(ctx, "failed to roll back transaction: %v", err)
	}
	compensationCtx := db.ContextAfterTransaction(ctx)
	for _, err := range hooks.RunRollback(compensationCtx) {
		log.Error(compensationCtx, "transaction rollback compensation failed: %v", err)
	}
}

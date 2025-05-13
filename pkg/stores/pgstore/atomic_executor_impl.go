package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"gorm.io/gorm"
)

type atomicExecutor struct {
	sessionFactory db.SessionFactory
}

func (a *atomicExecutor) WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError {
	var (
		tx    *gorm.DB
		txCtx context.Context
	)
	tx = db.TxFromContext(ctx)
	if tx == nil {
		tx = a.sessionFactory.New(ctx).Begin()
		txCtx = db.CtxWithTransaction(ctx, tx)
	} else {
		txCtx = ctx
	}

	// recover and rollback on panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.GeneralError("failed to commit transaction: %s", err.Error())
	}
	return nil
}

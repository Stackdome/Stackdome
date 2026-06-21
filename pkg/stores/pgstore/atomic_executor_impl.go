package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
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
		tx      *gorm.DB
		txCtx   context.Context
		isOwner bool
	)
	tx = db.TxFromContext(ctx)
	if tx == nil {
		tx = a.sessionFactory.New(ctx).Begin()
		txCtx = db.CtxWithTransaction(ctx, tx)
		isOwner = true
	} else {
		txCtx = ctx
	}

	defer func() {
		if r := recover(); r != nil {
			if isOwner {
				tx.Rollback()
			}
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if isOwner {
			tx.Rollback()
		}
		return err
	}

	if isOwner {
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return errors.GeneralError("failed to commit transaction: %s", err.Error())
		}
	}
	return nil
}

package db

import (
	"context"
	"database/sql"

	"github.com/ashishmax31/stackdome-api-server/config"
	"gorm.io/gorm"
)

type xactionCtxKey string

const (
	transactionCtxKey xactionCtxKey = "transaction"
)

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

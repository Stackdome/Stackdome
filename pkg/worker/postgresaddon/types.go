package postgresaddon

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
)

type subReconcilerResult struct {
	resultNil          bool
	resultStop         bool
	resultRequeue      bool
	resultRequeueAfter *time.Duration
}

var (
	resultNil          = subReconcilerResult{resultNil: true}
	resultStop         = subReconcilerResult{resultStop: true}
	resultRequeue      = subReconcilerResult{resultRequeue: true}
	resultRequeueAfter = func(t time.Duration) subReconcilerResult {
		return subReconcilerResult{resultRequeueAfter: &t}
	}
)

type subReconciler interface {
	Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error)
	Name() string
}

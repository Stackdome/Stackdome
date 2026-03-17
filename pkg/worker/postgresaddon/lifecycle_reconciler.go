package postgresaddon

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type lifecycleReconciler struct {
	logger logger.Logger
}

func newLifecycleReconciler(_ PostgresAddonWorkerSpec) *lifecycleReconciler {
	return &lifecycleReconciler{
		logger: logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-lifecycle"),
	}
}

func (r *lifecycleReconciler) Name() string { return "lifecycle" }

func (r *lifecycleReconciler) Reconcile(_ context.Context, _ *models.PostgresAddon) (subReconcilerResult, error) {
	return resultNil, nil
}

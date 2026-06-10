package stack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackdeploy"
)

const (
	deployResolverReconcilerName         = "deploy-resolver-reconciler"
	dependencyNotReadyRequeueInterval    = 30 * time.Second
)

type deployResolverReconciler struct {
	resolver *stackdeploy.Resolver
}

type DeployResolverReconcilerSpec struct {
	VolumeService        volumeService
	PostgresAddonService postgresAddonService
	SecretService        secretService
}

func NewDeployResolverReconciler(spec DeployResolverReconcilerSpec) *deployResolverReconciler {
	return &deployResolverReconciler{
		resolver: stackdeploy.NewResolver(stackdeploy.ResolverSpec{
			VolumeService:        spec.VolumeService,
			PostgresAddonService: spec.PostgresAddonService,
			SecretService:        spec.SecretService,
		}),
	}
}

func (r *deployResolverReconciler) Name() string {
	return deployResolverReconcilerName
}

func (r *deployResolverReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp != nil {
		return resultNil, nil
	}

	effective, err := r.resolver.Resolve(ctx, stack)
	if err != nil {
		var notReady stackdeploy.DependencyNotReadyError
		if errors.As(err, &notReady) {
			return resultRequeueAfter(dependencyNotReadyRequeueInterval), nil
		}
		return resultNil, fmt.Errorf("failed to resolve effective stack: %w", err)
	}

	*stack = *effective
	return resultNil, nil
}

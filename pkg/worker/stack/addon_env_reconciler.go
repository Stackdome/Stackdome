package stack

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

const addonReadinessRequeueInterval = 30 * time.Second

type addonEnvReconciler struct {
	postgresAddonService postgresAddonService
	addonUsageService    addonUsageService
	logger               logger.Logger
}

type AddonEnvReconcilerSpec struct {
	PostgresAddonService postgresAddonService
	AddonUsageService    addonUsageService
}

func NewAddonEnvReconciler(spec AddonEnvReconcilerSpec) *addonEnvReconciler {
	return &addonEnvReconciler{
		postgresAddonService: spec.PostgresAddonService,
		addonUsageService:    spec.AddonUsageService,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "stack-addon-env-reconciler"),
	}
}

func (r *addonEnvReconciler) Name() string {
	return "addon-env-reconciler"
}

func (r *addonEnvReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp != nil {
		return resultNil, nil
	}

	for _, resource := range stack.StackResources {
		if resource.ExecutionConfig == nil || len(resource.ExecutionConfig.EnvFromAddons) == 0 {
			continue
		}

		resolvedEnvVars, requeueResult, err := r.resolveAddonEnvVars(ctx, stack.ID, resource)
		if err != nil {
			return resultNil, fmt.Errorf("failed to resolve addon env vars for resource '%s': %w", resource.Name, err)
		}
		if requeueResult != nil {
			return *requeueResult, nil
		}

		if len(resolvedEnvVars) == 0 {
			continue
		}

		resource.ExecutionConfig.Env = appendWithoutDuplicates(resource.ExecutionConfig.Env, resolvedEnvVars)

		if err := r.syncAddonUsages(ctx, stack.ID, resource); err != nil {
			return resultNil, fmt.Errorf("failed to sync addon usages for resource '%s': %w", resource.Name, err)
		}
	}

	return resultNil, nil
}

func (r *addonEnvReconciler) resolveAddonEnvVars(ctx context.Context, stackID string, resource *models.StackResource) ([]models.EnvVar, *subReconcilerResult, error) {
	var envVars []models.EnvVar

	for _, addonSource := range resource.ExecutionConfig.EnvFromAddons {
		if addonSource.Postgres == nil {
			continue
		}
		pg := addonSource.Postgres

		creds, credErr := r.postgresAddonService.GetCredentials(ctx, pg.AddonID, pg.Database, false)
		if credErr != nil {
			// Check if credentials were resolved in a prior reconciliation.
			// If yes, the CR already has valid env vars — proceed without blocking.
			previouslyResolved, lookupErr := r.addonUsageService.ExistsByStackResourceAndAddon(ctx, stackID, resource.ID, pg.AddonID)
			if lookupErr != nil {
				return nil, nil, fmt.Errorf("failed to check addon usage for addon '%s': %w", pg.AddonID, lookupErr)
			}

			if previouslyResolved {
				r.logger.Infof("addon '%s' credentials unavailable but previously resolved, proceeding with existing CR values", pg.AddonID)
				continue
			}

			// First deploy — addon must be ready before we can proceed
			r.logger.Infof("addon '%s' not available for first-time credential resolution, will requeue in %s", pg.AddonID, addonReadinessRequeueInterval)
			result := resultRequeueAfter(addonReadinessRequeueInterval)
			return nil, &result, nil
		}

		fieldMap := creds.ToFieldMap()
		for credField, envName := range pg.EnvMapping {
			value, ok := fieldMap[credField]
			if !ok {
				return nil, nil, fmt.Errorf("unknown credential field '%s' in env mapping for addon '%s'", credField, pg.AddonID)
			}
			envVars = append(envVars, models.EnvVar{
				Name:  envName,
				Value: value,
			})
		}
	}

	return envVars, nil, nil
}

func (r *addonEnvReconciler) syncAddonUsages(ctx context.Context, stackID string, resource *models.StackResource) error {
	for _, addonSource := range resource.ExecutionConfig.EnvFromAddons {
		if addonSource.Postgres == nil {
			continue
		}
		if err := r.addonUsageService.Create(ctx, &models.AddonUsage{
			AddonType:       models.AddonTypePostgres,
			AddonID:         addonSource.Postgres.AddonID,
			StackID:         stackID,
			StackResourceID: resource.ID,
		}); err != nil {
			return fmt.Errorf("failed to create addon usage for addon '%s': %w", addonSource.Postgres.AddonID, err)
		}
	}
	return nil
}

// appendWithoutDuplicates appends new env vars, skipping any whose name already exists.
func appendWithoutDuplicates(existing []models.EnvVar, newVars []models.EnvVar) []models.EnvVar {
	nameSet := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		nameSet[v.Name] = struct{}{}
	}
	for _, v := range newVars {
		if _, exists := nameSet[v.Name]; !exists {
			existing = append(existing, v)
			nameSet[v.Name] = struct{}{}
		}
	}
	return existing
}

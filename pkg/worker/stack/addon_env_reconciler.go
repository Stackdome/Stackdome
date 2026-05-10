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

	var desiredAddonUsages []*models.AddonUsage
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

		for _, addonSource := range resource.ExecutionConfig.EnvFromAddons {
			if addonSource.Postgres != nil {
				desiredAddonUsages = append(desiredAddonUsages, &models.AddonUsage{
					AddonType:       models.AddonTypePostgres,
					AddonID:         addonSource.Postgres.AddonID,
					StackID:         stack.ID,
					StackResourceID: resource.ID,
				})
			}
		}

		if len(resolvedEnvVars) > 0 {
			resource.ExecutionConfig.Env = appendWithoutDuplicates(resource.ExecutionConfig.Env, resolvedEnvVars)
		}
	}

	if err := r.syncAddonUsages(ctx, stack.ID, desiredAddonUsages); err != nil {
		return resultNil, fmt.Errorf("failed to sync addon usages: %w", err)
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

		creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, pg.AddonID, pg.Database, pg.Superuser)
		if credErr != nil {
			r.logger.Errorf("failed to fetch db: '%s' credentials, got err: %s", pg.Database, credErr.Error())
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

		// TODO: Pass K8s secret references to the cluster-agent instead of resolving
		// credentials as plain env var values (see docs/plans/postgres-addon-improvements.md #8).
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

func addonUsageKey(resourceID, addonID string, addonType models.AddonType) string {
	return resourceID + ":" + addonID + ":" + string(addonType)
}

func (r *addonEnvReconciler) syncAddonUsages(ctx context.Context, stackID string, desiredAddonUsages []*models.AddonUsage) error {
	existing, err := r.addonUsageService.GetByStackID(ctx, stackID)
	if err != nil {
		return fmt.Errorf("failed to list addon usages for stack '%s': %w", stackID, err)
	}

	existingKeys := make(map[string]struct{}, len(existing))
	for _, u := range existing {
		existingKeys[addonUsageKey(u.StackResourceID, u.AddonID, u.AddonType)] = struct{}{}
	}

	desiredKeys := make(map[string]struct{}, len(desiredAddonUsages))
	for _, u := range desiredAddonUsages {
		key := addonUsageKey(u.StackResourceID, u.AddonID, u.AddonType)
		if _, ok := desiredKeys[key]; ok {
			continue
		}
		desiredKeys[key] = struct{}{}
		if _, ok := existingKeys[key]; !ok {
			if err := r.addonUsageService.Create(ctx, u); err != nil {
				return fmt.Errorf("failed to create addon usage for addon '%s': %w", u.AddonID, err)
			}
		}
	}

	for _, u := range existing {
		if _, ok := desiredKeys[addonUsageKey(u.StackResourceID, u.AddonID, u.AddonType)]; !ok {
			if err := r.addonUsageService.Delete(ctx, u.AddonType, u.AddonID, stackID, u.StackResourceID); err != nil {
				return fmt.Errorf("failed to delete stale addon usage for addon '%s': %w", u.AddonID, err)
			}
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

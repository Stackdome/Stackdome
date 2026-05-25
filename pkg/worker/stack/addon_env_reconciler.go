package stack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

const addonReadinessRequeueInterval = 30 * time.Second

type addonEnvReconciler struct {
	postgresAddonService postgresAddonService
	secretService        secretOutputService
	addonUsageService    addonUsageService
	logger               logger.Logger
}

type secretOutputService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type AddonEnvReconcilerSpec struct {
	PostgresAddonService postgresAddonService
	SecretService        secretOutputService
	AddonUsageService    addonUsageService
}

func NewAddonEnvReconciler(spec AddonEnvReconcilerSpec) *addonEnvReconciler {
	return &addonEnvReconciler{
		postgresAddonService: spec.PostgresAddonService,
		secretService:        spec.SecretService,
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
		hasLegacyAddonEnv := resource.ExecutionConfig != nil && len(resource.ExecutionConfig.EnvFromAddons) > 0
		hasConnectionEnv := hasEnvConnections(stack, resource.Name)
		hasSelfOutputEnv := hasSelfOutputEnvVars(resource)
		if !hasLegacyAddonEnv && !hasConnectionEnv && !hasSelfOutputEnv {
			continue
		}

		if hasSelfOutputEnv {
			if err := resolveSelfOutputEnvVars(resource); err != nil {
				return resultNil, fmt.Errorf("failed to resolve self output env vars for resource '%s': %w", resource.Name, err)
			}
		}

		resolvedEnvVars, usages, requeueResult, err := r.resolveAddonEnvVars(ctx, stack.ID, stack, resource)
		if err != nil {
			return resultNil, fmt.Errorf("failed to resolve addon env vars for resource '%s': %w", resource.Name, err)
		}
		if requeueResult != nil {
			return *requeueResult, nil
		}
		desiredAddonUsages = append(desiredAddonUsages, usages...)

		if len(resolvedEnvVars) > 0 {
			if resource.ExecutionConfig == nil {
				resource.ExecutionConfig = &models.ExecutionConfig{}
			}
			mergedEnv, mergeErr := appendWithoutDuplicates(resource.ExecutionConfig.Env, resolvedEnvVars)
			if mergeErr != nil {
				return resultNil, fmt.Errorf("failed to merge addon env vars for resource '%s': %w", resource.Name, mergeErr)
			}
			resource.ExecutionConfig.Env = mergedEnv
		}
	}

	if err := r.syncAddonUsages(ctx, stack.ID, desiredAddonUsages); err != nil {
		return resultNil, fmt.Errorf("failed to sync addon usages: %w", err)
	}

	return resultNil, nil
}

func (r *addonEnvReconciler) resolveAddonEnvVars(ctx context.Context, stackID string, stack *models.Stack, resource *models.StackResource) ([]models.EnvVar, []*models.AddonUsage, *subReconcilerResult, error) {
	var envVars []models.EnvVar
	var desiredAddonUsages []*models.AddonUsage

	for _, addonSource := range envFromAddonSources(resource) {
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
				return nil, nil, nil, fmt.Errorf("failed to check addon usage for addon '%s': %w", pg.AddonID, lookupErr)
			}

			if previouslyResolved {
				r.logger.Infof("addon '%s' credentials unavailable but previously resolved, proceeding with existing CR values", pg.AddonID)
				continue
			}

			// First deploy — addon must be ready before we can proceed
			r.logger.Infof("addon '%s' not available for first-time credential resolution, will requeue in %s", pg.AddonID, addonReadinessRequeueInterval)
			result := resultRequeueAfter(addonReadinessRequeueInterval)
			return nil, nil, &result, nil
		}

		// TODO: Pass K8s secret references to the cluster-agent instead of resolving
		// credentials as plain env var values (see docs/plans/postgres-addon-improvements.md #8).
		fieldMap := creds.ToOutputMap()
		for credField, envName := range pg.EnvMapping {
			value, ok := fieldMap[credField]
			if !ok {
				return nil, nil, nil, fmt.Errorf("unknown credential field '%s' in env mapping for addon '%s'", credField, pg.AddonID)
			}
			envVars = append(envVars, models.EnvVar{
				Name:  envName,
				Value: value,
			})
		}
		desiredAddonUsages = append(desiredAddonUsages, &models.AddonUsage{
			AddonType:       models.AddonTypePostgres,
			AddonID:         pg.AddonID,
			StackID:         stack.ID,
			StackResourceID: resource.ID,
		})
	}

	resourceMap := stack.ResourcesMap()
	for _, connection := range postgresEnvConnectionsForResource(stack, resource.Name) {
		resolvedEnvVars, usages, requeueResult, err := r.resolvePostgresConnectionEnvVars(ctx, stackID, resource, connection)
		if err != nil {
			return nil, nil, nil, err
		}
		if requeueResult != nil {
			return nil, nil, requeueResult, nil
		}
		envVars = append(envVars, resolvedEnvVars...)
		desiredAddonUsages = append(desiredAddonUsages, usages...)
	}

	for _, connection := range stackResourceEnvConnectionsForResource(stack, resource.Name) {
		sourceResource, ok := resourceMap[connection.From.Name]
		if !ok {
			return nil, nil, nil, fmt.Errorf("stack resource connection '%s' references unknown resource '%s'", connection.Id, connection.From.Name)
		}
		resolvedEnvVars, err := resolveConnectionMappings(connection, sourceResource.ToOutputMap())
		if err != nil {
			return nil, nil, nil, err
		}
		envVars = append(envVars, resolvedEnvVars...)
	}

	for _, connection := range secretEnvConnectionsForResource(stack, resource.Name) {
		if r.secretService == nil {
			return nil, nil, nil, fmt.Errorf("secret service is not configured for resolving connection '%s'", connection.Id)
		}
		secret, serviceErr := r.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to fetch secret '%s' for connection '%s': %w", connection.From.Id, connection.Id, serviceErr)
		}
		resolvedEnvVars, err := resolveConnectionMappings(connection, secret.ToOutputMap())
		if err != nil {
			return nil, nil, nil, err
		}
		envVars = append(envVars, resolvedEnvVars...)
	}

	return envVars, desiredAddonUsages, nil, nil
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

func envFromAddonSources(resource *models.StackResource) []models.AddonEnvSource {
	if resource.ExecutionConfig == nil {
		return nil
	}
	return resource.ExecutionConfig.EnvFromAddons
}

func hasSelfOutputEnvVars(resource *models.StackResource) bool {
	if resource.ExecutionConfig == nil {
		return false
	}
	for _, envVar := range resource.ExecutionConfig.Env {
		if envVar.SelfOutput != "" {
			return true
		}
	}
	return false
}

func resolveSelfOutputEnvVars(resource *models.StackResource) error {
	if resource.ExecutionConfig == nil {
		return nil
	}
	outputs := resource.ToOutputMap()
	for i := range resource.ExecutionConfig.Env {
		if resource.ExecutionConfig.Env[i].SelfOutput == "" {
			continue
		}
		value, ok := outputs[resource.ExecutionConfig.Env[i].SelfOutput]
		if !ok {
			return fmt.Errorf("unknown self output '%s'", resource.ExecutionConfig.Env[i].SelfOutput)
		}
		resource.ExecutionConfig.Env[i].Value = value
	}
	return nil
}

func hasEnvConnections(stack *models.Stack, resourceName string) bool {
	return len(envConnectionsForResource(stack, resourceName)) > 0
}

func envConnectionsForResource(stack *models.Stack, resourceName string) []models.StackConnection {
	var connections []models.StackConnection
	for _, connection := range stack.Connections {
		if connection.Kind != models.ConnectionKindEnv {
			continue
		}
		if connection.To.Type != models.TopologyNodeTypeStackResource || connection.To.Name != resourceName {
			continue
		}
		connections = append(connections, connection)
	}
	return connections
}

func postgresEnvConnectionsForResource(stack *models.Stack, resourceName string) []models.StackConnection {
	return filterEnvConnectionsBySourceType(envConnectionsForResource(stack, resourceName), models.TopologyNodeTypePostgresAddon)
}

func stackResourceEnvConnectionsForResource(stack *models.Stack, resourceName string) []models.StackConnection {
	return filterEnvConnectionsBySourceType(envConnectionsForResource(stack, resourceName), models.TopologyNodeTypeStackResource)
}

func secretEnvConnectionsForResource(stack *models.Stack, resourceName string) []models.StackConnection {
	return filterEnvConnectionsBySourceType(envConnectionsForResource(stack, resourceName), models.TopologyNodeTypeSecret)
}

func filterEnvConnectionsBySourceType(connections []models.StackConnection, sourceType models.TopologyNodeType) []models.StackConnection {
	var filtered []models.StackConnection
	for _, connection := range connections {
		if connection.From.Type == sourceType {
			filtered = append(filtered, connection)
		}
	}
	return filtered
}

func postgresConnectionConfig(connection models.StackConnection) (database string, superuser bool, err error) {
	if rawDatabase, ok := connection.Config["database"]; ok {
		value, ok := rawDatabase.(string)
		if !ok {
			return "", false, fmt.Errorf("connection '%s' config.database must be a string", connection.Id)
		}
		database = value
	}

	if rawScope, ok := connection.Config["credential_scope"]; ok {
		scope, ok := rawScope.(string)
		if !ok {
			return "", false, fmt.Errorf("connection '%s' config.credential_scope must be a string", connection.Id)
		}
		superuser = scope == "superuser"
	}
	if rawSuperuser, ok := connection.Config["superuser"]; ok {
		value, ok := rawSuperuser.(bool)
		if !ok {
			return "", false, fmt.Errorf("connection '%s' config.superuser must be a boolean", connection.Id)
		}
		superuser = value
	}

	return database, superuser, nil
}

func (r *addonEnvReconciler) resolvePostgresConnectionEnvVars(
	ctx context.Context,
	stackID string,
	resource *models.StackResource,
	connection models.StackConnection,
) ([]models.EnvVar, []*models.AddonUsage, *subReconcilerResult, error) {
	addonID := connection.From.Id
	database, superuser, err := postgresConnectionConfig(connection)
	if err != nil {
		return nil, nil, nil, err
	}

	creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, addonID, database, superuser)
	if credErr != nil {
		r.logger.Errorf("failed to fetch postgres addon '%s' credentials, got err: %s", addonID, credErr.Error())
		previouslyResolved, lookupErr := r.addonUsageService.ExistsByStackResourceAndAddon(ctx, stackID, resource.ID, addonID)
		if lookupErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to check addon usage for addon '%s': %w", addonID, lookupErr)
		}
		if previouslyResolved {
			r.logger.Infof("addon '%s' credentials unavailable but previously resolved, proceeding with existing CR values", addonID)
			return nil, nil, nil, nil
		}

		r.logger.Infof("addon '%s' not available for first-time credential resolution, will requeue in %s", addonID, addonReadinessRequeueInterval)
		result := resultRequeueAfter(addonReadinessRequeueInterval)
		return nil, nil, &result, nil
	}

	envVars, err := resolveConnectionMappings(connection, creds.ToOutputMap())
	if err != nil {
		return nil, nil, nil, err
	}

	return envVars, []*models.AddonUsage{
		{
			AddonType:       models.AddonTypePostgres,
			AddonID:         addonID,
			StackID:         stackID,
			StackResourceID: resource.ID,
		},
	}, nil, nil
}

func resolveConnectionMappings(connection models.StackConnection, outputs map[string]string) ([]models.EnvVar, error) {
	envVars := make([]models.EnvVar, 0, len(connection.Mappings))
	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return nil, fmt.Errorf("connection '%s' target type '%s' is not supported for env resolution", connection.Id, mapping.Target.Type)
		}
		value, err := resolveConnectionValue(mapping.Value, outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve connection '%s' mapping for env '%s': %w", connection.Id, mapping.Target.Name, err)
		}
		envVars = append(envVars, models.EnvVar{
			Name:  mapping.Target.Name,
			Value: value,
		})
	}
	return envVars, nil
}

func resolveConnectionValue(valueRef models.ValueRef, outputs map[string]string) (string, error) {
	if valueRef.Output != "" {
		value, ok := outputs[valueRef.Output]
		if !ok {
			return "", fmt.Errorf("unknown output '%s'", valueRef.Output)
		}
		return value, nil
	}

	if valueRef.Template != "" {
		resolved := valueRef.Template
		for key, ref := range valueRef.Values {
			value, ok := outputs[ref.Output]
			if !ok {
				return "", fmt.Errorf("unknown output '%s' for template key '%s'", ref.Output, key)
			}
			resolved = strings.ReplaceAll(resolved, "{{ "+key+" }}", value)
			resolved = strings.ReplaceAll(resolved, "{{"+key+"}}", value)
		}
		return resolved, nil
	}

	return "", fmt.Errorf("value ref is missing output or template")
}

// appendWithoutDuplicates appends new env vars, rejecting any whose name already exists.
func appendWithoutDuplicates(existing []models.EnvVar, newVars []models.EnvVar) ([]models.EnvVar, error) {
	nameSet := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		nameSet[v.Name] = struct{}{}
	}
	for _, v := range newVars {
		if _, exists := nameSet[v.Name]; exists {
			return nil, fmt.Errorf("duplicate env var name '%s'", v.Name)
		}
		existing = append(existing, v)
		nameSet[v.Name] = struct{}{}
	}
	return existing, nil
}

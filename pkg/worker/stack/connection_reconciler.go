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

type connectionReconciler struct {
	volumeService        volumeService
	postgresAddonService postgresAddonService
	secretService        secretOutputService
	logger               logger.Logger
}

type secretOutputService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type postgresAddonService interface {
	InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	InternalGetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)
}

type ConnectionReconcilerSpec struct {
	VolumeService        volumeService
	PostgresAddonService postgresAddonService
	SecretService        secretOutputService
}

func NewConnectionReconciler(spec ConnectionReconcilerSpec) *connectionReconciler {
	return &connectionReconciler{
		volumeService:        spec.VolumeService,
		postgresAddonService: spec.PostgresAddonService,
		secretService:        spec.SecretService,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "stack-connection-reconciler"),
	}
}

func (r *connectionReconciler) Name() string {
	return "connection-reconciler"
}

func (r *connectionReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp != nil {
		return resultNil, nil
	}

	if len(stack.Connections) == 0 && !hasAnySelfOutputEnvVars(stack) {
		return resultNil, nil
	}

	if stack.Connections.HasAnyKind(models.ConnectionKindVolumeMount, models.ConnectionKindBuildArtifactSource) {
		if err := r.loadVolumes(ctx, stack); err != nil {
			return resultNil, err
		}
		resourceMap := stack.ResourcesMap()
		volumeMap := r.buildVolumeMap(stack)

		if err := r.resolveVolumeMountConnections(stack, resourceMap, volumeMap); err != nil {
			return resultNil, err
		}
		if err := r.resolveBuildArtifactSourceConnections(stack, resourceMap, volumeMap); err != nil {
			return resultNil, err
		}
	}

	if result, err := r.resolveEnvConnections(ctx, stack); err != nil {
		return resultNil, err
	} else if !result.resultNil {
		return result, nil
	}

	return resultNil, nil
}

// --- Volume mount and build artifact source resolution ---

func (r *connectionReconciler) loadVolumes(ctx context.Context, stack *models.Stack) error {
	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, stack.ID)
	if serr != nil {
		return fmt.Errorf("failed to list volumes for stack '%s': %w", stack.ID, serr)
	}
	stack.Volumes = volumes
	return nil
}

func (r *connectionReconciler) buildVolumeMap(stack *models.Stack) map[string]*models.Volume {
	volumeMap := make(map[string]*models.Volume, len(stack.Volumes))
	for _, v := range stack.Volumes {
		volumeMap[v.Name] = v
	}
	return volumeMap
}

func (r *connectionReconciler) resolveVolumeMountConnections(
	stack *models.Stack,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
) error {
	for _, connection := range stack.Connections.OfKind(models.ConnectionKindVolumeMount) {
		volume, ok := volumeMap[connection.From.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown volume '%s'", connection.Id, connection.From.Name)
		}

		resource, ok := resourceMap[connection.To.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown stack resource '%s'", connection.Id, connection.To.Name)
		}

		mountPath, err := connection.RequiredConfigString("mount_path")
		if err != nil {
			return fmt.Errorf("volume_mount connection '%s' has invalid config: %w", connection.Id, err)
		}
		subPath, _, err := connection.ConfigString("sub_path")
		if err != nil {
			return fmt.Errorf("volume_mount connection '%s' has invalid config: %w", connection.Id, err)
		}

		mount := &models.VolumeMount{
			StackID:          stack.ID,
			StackResourceID:  resource.ID,
			SourceVolumeName: volume.Name,
			SourceVolumeID:   volume.ID,
			SourceVolumeType: volume.VolumeSourceType(),
			SourceSubPath:    subPath,
			TargetPath:       mountPath,
		}

		resource.VolumeMounts = append(resource.VolumeMounts, mount)
		r.logger.Infof("resolved volume_mount connection '%s': volume '%s' → resource '%s' at '%s'", connection.Id, volume.Name, resource.Name, mountPath)
	}

	return nil
}

func (r *connectionReconciler) resolveBuildArtifactSourceConnections(
	stack *models.Stack,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
) error {
	for _, connection := range stack.Connections.OfKind(models.ConnectionKindBuildArtifactSource) {
		if _, ok := resourceMap[connection.From.Name]; !ok {
			return fmt.Errorf("build_artifact_source connection '%s' references unknown stack resource '%s'", connection.Id, connection.From.Name)
		}

		volume, ok := volumeMap[connection.To.Name]
		if !ok {
			return fmt.Errorf("build_artifact_source connection '%s' references unknown volume '%s'", connection.Id, connection.To.Name)
		}

		sourcePath, err := connection.RequiredConfigString("source_path")
		if err != nil {
			return fmt.Errorf("build_artifact_source connection '%s' has invalid config: %w", connection.Id, err)
		}
		destinationPath, _, err := connection.ConfigString("destination_path")
		if err != nil {
			return fmt.Errorf("build_artifact_source connection '%s' has invalid config: %w", connection.Id, err)
		}

		if volume.VolumeSource == nil {
			volume.VolumeSource = &models.VolumeSource{}
		}

		volume.VolumeSource.BuildSource = append(volume.VolumeSource.BuildSource, models.BuildArtifactSource{
			ResourceName:    connection.From.Name,
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
		})
		r.logger.Infof("resolved build_artifact_source connection '%s': resource '%s' → volume '%s'", connection.Id, connection.From.Name, volume.Name)
	}

	return nil
}

// --- Env connection resolution ---

func (r *connectionReconciler) resolveEnvConnections(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	for _, resource := range stack.StackResources {
		envConnections := stack.Connections.EnvForStackResource(resource.Name)
		hasSelfOutputEnv := hasSelfOutputEnvVars(resource)
		if len(envConnections) == 0 && !hasSelfOutputEnv {
			continue
		}

		if hasSelfOutputEnv {
			if err := resolveSelfOutputEnvVars(resource); err != nil {
				return resultNil, fmt.Errorf("failed to resolve self output env vars for resource '%s': %w", resource.Name, err)
			}
		}

		resolvedEnvVars, requeueResult, err := r.resolveConnectionEnvVars(ctx, stack, resource, envConnections)
		if err != nil {
			return resultNil, fmt.Errorf("failed to resolve connection env vars for resource '%s': %w", resource.Name, err)
		}
		if requeueResult != nil {
			return *requeueResult, nil
		}

		if len(resolvedEnvVars) > 0 {
			if resource.ExecutionConfig == nil {
				resource.ExecutionConfig = &models.ExecutionConfig{}
			}
			mergedEnv, mergeErr := appendWithoutDuplicates(resource.ExecutionConfig.Env, resolvedEnvVars)
			if mergeErr != nil {
				return resultNil, fmt.Errorf("failed to merge connection env vars for resource '%s': %w", resource.Name, mergeErr)
			}
			resource.ExecutionConfig.Env = mergedEnv
		}
	}

	return resultNil, nil
}

func (r *connectionReconciler) resolveConnectionEnvVars(
	ctx context.Context,
	stack *models.Stack,
	resource *models.StackResource,
	connections models.StackConnections,
) ([]models.EnvVar, *subReconcilerResult, error) {
	var envVars []models.EnvVar

	resourceMap := stack.ResourcesMap()
	for _, connection := range connections.FromType(models.TopologyNodeTypePostgresAddon) {
		resolvedEnvVars, requeueResult, err := r.resolvePostgresConnectionEnvVars(ctx, connection)
		if err != nil {
			return nil, nil, err
		}
		if requeueResult != nil {
			return nil, requeueResult, nil
		}
		envVars = append(envVars, resolvedEnvVars...)
	}

	for _, connection := range connections.FromType(models.TopologyNodeTypeStackResource) {
		sourceResource, ok := resourceMap[connection.From.Name]
		if !ok {
			return nil, nil, fmt.Errorf("stack resource connection '%s' references unknown resource '%s'", connection.Id, connection.From.Name)
		}
		resolvedEnvVars, err := resolveConnectionMappings(connection, sourceResource.ToOutputMap())
		if err != nil {
			return nil, nil, err
		}
		envVars = append(envVars, resolvedEnvVars...)
	}

	for _, connection := range connections.FromType(models.TopologyNodeTypeSecret) {
		if r.secretService == nil {
			return nil, nil, fmt.Errorf("secret service is not configured for resolving connection '%s'", connection.Id)
		}
		secret, serviceErr := r.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil {
			return nil, nil, fmt.Errorf("failed to fetch secret '%s' for connection '%s': %w", connection.From.Id, connection.Id, serviceErr)
		}
		resolvedEnvVars, err := resolveConnectionMappings(connection, secret.ToOutputMap())
		if err != nil {
			return nil, nil, err
		}
		envVars = append(envVars, resolvedEnvVars...)
	}

	return envVars, nil, nil
}

func (r *connectionReconciler) resolvePostgresConnectionEnvVars(
	ctx context.Context,
	connection models.StackConnection,
) ([]models.EnvVar, *subReconcilerResult, error) {
	addonID := connection.From.Id
	database, superuser, err := postgresConnectionConfig(connection)
	if err != nil {
		return nil, nil, err
	}

	creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, addonID, database, superuser)
	if credErr != nil {
		r.logger.Infof("addon '%s' credentials unavailable, will requeue in %s", addonID, addonReadinessRequeueInterval)
		result := resultRequeueAfter(addonReadinessRequeueInterval)
		return nil, &result, nil
	}

	envVars, err := resolveConnectionMappings(connection, creds.ToOutputMap())
	if err != nil {
		return nil, nil, err
	}

	return envVars, nil, nil
}

// --- Helper functions ---

func hasAnySelfOutputEnvVars(stack *models.Stack) bool {
	for _, resource := range stack.StackResources {
		if hasSelfOutputEnvVars(resource) {
			return true
		}
	}
	return false
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

func postgresConnectionConfig(connection models.StackConnection) (database string, superuser bool, err error) {
	if value, ok, err := connection.ConfigString("database"); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.Id, err)
	} else if ok {
		database = value
	}

	if scope, ok, err := connection.ConfigString("credential_scope"); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.Id, err)
	} else if ok {
		superuser = scope == "superuser"
	}
	if value, ok, err := connection.ConfigBool("superuser"); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.Id, err)
	} else if ok {
		superuser = value
	}

	return database, superuser, nil
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

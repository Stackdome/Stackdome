package stackdeploy

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func resolveSelfOutputEnvVars(stack *models.Stack) error {
	for _, resource := range stack.StackResources {
		if err := resolveSelfOutputsForResource(resource); err != nil {
			return fmt.Errorf("failed to resolve self output env vars for resource '%s': %w", resource.Name, err)
		}
	}
	return nil
}

func resolveSelfOutputsForResource(resource *models.StackResource) error {
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

func (r *Resolver) resolveEnvConnections(ctx context.Context, stack *models.Stack) error {
	for _, resource := range stack.StackResources {
		envConnections := stack.Connections.EnvForStackResource(resource.Name)
		if len(envConnections) == 0 {
			continue
		}

		resolvedEnvVars, err := r.resolveConnectionEnvVars(ctx, stack, envConnections)
		if err != nil {
			return fmt.Errorf("failed to resolve connection env vars for resource '%s': %w", resource.Name, err)
		}

		if len(resolvedEnvVars) > 0 {
			if resource.ExecutionConfig == nil {
				resource.ExecutionConfig = &models.ExecutionConfig{}
			}
			mergedEnv, mergeErr := appendWithoutDuplicates(resource.ExecutionConfig.Env, resolvedEnvVars)
			if mergeErr != nil {
				return fmt.Errorf("failed to merge connection env vars for resource '%s': %w", resource.Name, mergeErr)
			}
			resource.ExecutionConfig.Env = mergedEnv
		}
	}
	return nil
}

func (r *Resolver) resolveConnectionEnvVars(
	ctx context.Context,
	stack *models.Stack,
	connections models.StackConnections,
) ([]models.EnvVar, error) {
	var envVars []models.EnvVar

	resourceMap := stack.ResourcesMap()
	for _, connection := range connections.FromType(models.TopologyNodeTypePostgresAddon) {
		resolved, err := r.resolvePostgresConnectionEnvVars(ctx, connection)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, resolved...)
	}

	for _, connection := range connections.FromType(models.TopologyNodeTypeStackResource) {
		sourceResource, ok := resourceMap[connection.From.Name]
		if !ok {
			return nil, fmt.Errorf("stack resource connection '%s' references unknown resource '%s'", connection.ID, connection.From.Name)
		}
		resolved, err := resolveConnectionMappings(connection, sourceResource.ToOutputMap())
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, resolved...)
	}

	for _, connection := range connections.FromType(models.TopologyNodeTypeSecret) {
		secret, serviceErr := r.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil {
			return nil, fmt.Errorf("failed to fetch secret '%s' for connection '%s': %w", connection.From.Id, connection.ID, serviceErr)
		}
		resolved, err := resolveSecretConnectionMappings(connection, secret)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, resolved...)
	}

	return envVars, nil
}

func (r *Resolver) resolvePostgresConnectionEnvVars(
	ctx context.Context,
	connection models.StackConnection,
) ([]models.EnvVar, error) {
	addonID := connection.From.Id
	database, superuser, err := PostgresConnectionConfig(connection)
	if err != nil {
		return nil, err
	}

	creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, addonID, database, superuser)
	if credErr != nil {
		if credErr.Is404() {
			return nil, fmt.Errorf("postgres addon '%s' not found", addonID)
		}
		return nil, DependencyNotReadyError{
			Message: fmt.Sprintf("postgres addon '%s' credentials are not ready", addonID),
		}
	}

	secretName := PostgresCredentialSecretName(addonID, database)
	return resolvePostgresSecretRefMappings(connection, secretName, creds.ToOutputMap())
}

func PostgresConnectionConfig(connection models.StackConnection) (database string, superuser bool, err error) {
	if value, ok, err := connection.ConfigString(string(models.ConnectionConfigKeyDatabase)); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.ID, err)
	} else if ok {
		database = value
	}

	if scope, ok, err := connection.ConfigString(string(models.ConnectionConfigKeyCredentialScope)); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.ID, err)
	} else if ok {
		superuser = scope == "superuser"
	}
	if value, ok, err := connection.ConfigBool(string(models.ConnectionConfigKeySuperuser)); err != nil {
		return "", false, fmt.Errorf("connection '%s' has invalid config: %w", connection.ID, err)
	} else if ok {
		superuser = value
	}

	return database, superuser, nil
}

func resolveSecretConnectionMappings(connection models.StackConnection, secret *models.Secret) ([]models.EnvVar, error) {
	outputs := secret.ToOutputMap()
	envVars := make([]models.EnvVar, 0, len(connection.Mappings))
	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return nil, fmt.Errorf("connection '%s' target type '%s' is not supported for env resolution", connection.ID, mapping.Target.Type)
		}
		if mapping.Value.Output == "" {
			return nil, fmt.Errorf("connection '%s' mapping for env '%s': secret connections require a direct output key, not a template", connection.ID, mapping.Target.Name)
		}
		if _, ok := outputs[mapping.Value.Output]; !ok {
			return nil, fmt.Errorf("connection '%s' mapping for env '%s': unknown output '%s'", connection.ID, mapping.Target.Name, mapping.Value.Output)
		}
		envVars = append(envVars, models.EnvVar{
			Name: mapping.Target.Name,
			SecretKeyRef: &models.EnvSecretRef{
				SecretName: secret.ClusterSecretName(),
				Key:        mapping.Value.Output,
			},
		})
	}
	return envVars, nil
}

func resolvePostgresSecretRefMappings(connection models.StackConnection, secretName string, outputs map[string]string) ([]models.EnvVar, error) {
	envVars := make([]models.EnvVar, 0, len(connection.Mappings))
	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return nil, fmt.Errorf("connection '%s' target type '%s' is not supported for env resolution", connection.ID, mapping.Target.Type)
		}
		if mapping.Value.Output == "" {
			return nil, fmt.Errorf("connection '%s' mapping for env '%s': postgres connections require a direct output key, not a template", connection.ID, mapping.Target.Name)
		}
		if _, ok := outputs[mapping.Value.Output]; !ok {
			return nil, fmt.Errorf("connection '%s' mapping for env '%s': unknown output '%s'", connection.ID, mapping.Target.Name, mapping.Value.Output)
		}
		envVars = append(envVars, models.EnvVar{
			Name: mapping.Target.Name,
			SecretKeyRef: &models.EnvSecretRef{
				SecretName: secretName,
				Key:        mapping.Value.Output,
			},
		})
	}
	return envVars, nil
}

func resolveConnectionMappings(connection models.StackConnection, outputs map[string]string) ([]models.EnvVar, error) {
	envVars := make([]models.EnvVar, 0, len(connection.Mappings))
	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return nil, fmt.Errorf("connection '%s' target type '%s' is not supported for env resolution", connection.ID, mapping.Target.Type)
		}
		value, err := resolveConnectionValue(mapping.Value, outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve connection '%s' mapping for env '%s': %w", connection.ID, mapping.Target.Name, err)
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
		data := make(map[string]string, len(valueRef.Values))
		for key, ref := range valueRef.Values {
			value, ok := outputs[ref.Output]
			if !ok {
				return "", fmt.Errorf("unknown output '%s' for template key '%s'", ref.Output, key)
			}
			data[key] = value
		}
		return ResolveTemplate(valueRef.Template, data)
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
			return nil, fmt.Errorf("duplicate env var name '%s': check connections that target this resource for overlapping mappings", v.Name)
		}
		existing = append(existing, v)
		nameSet[v.Name] = struct{}{}
	}
	return existing, nil
}

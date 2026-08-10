package stackdeploy

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
)

func (r *Resolver) resolveVolumeConnections(ctx context.Context, stack *models.Stack) error {
	if !stack.Connections.HasAnyKind(models.ConnectionKindVolumeMount) {
		return nil
	}

	if err := r.loadVolumes(ctx, stack); err != nil {
		return err
	}
	resourceMap := stack.ResourcesMap()
	volumeMap := buildVolumeMap(stack)

	return resolveVolumeMountConnections(stack, resourceMap, volumeMap)
}

func (r *Resolver) loadVolumes(ctx context.Context, stack *models.Stack) error {
	if stack.Volumes != nil {
		return nil // snapshot already provides volumes, skip DB query
	}
	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, stack.ID)
	if serr != nil {
		return fmt.Errorf("failed to list volumes for stack '%s': %w", stack.ID, serr)
	}
	stack.Volumes = volumes
	return nil
}

func buildVolumeMap(stack *models.Stack) map[string]*models.Volume {
	volumeMap := make(map[string]*models.Volume, len(stack.Volumes))
	for _, v := range stack.Volumes {
		volumeMap[v.Name] = v
	}
	return volumeMap
}

func resolveVolumeMountConnections(
	stack *models.Stack,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
) error {
	for _, connection := range stack.Connections.OfKind(models.ConnectionKindVolumeMount) {
		volume, ok := volumeMap[connection.From.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown volume '%s'", connection.ID, connection.From.Name)
		}

		resource, ok := resourceMap[connection.To.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown stack resource '%s'", connection.ID, connection.To.Name)
		}

		mountPath, err := connection.RequiredConfigString(string(models.ConnectionConfigKeyMountPath))
		if err != nil {
			return fmt.Errorf("volume_mount connection '%s' has invalid config: %w", connection.ID, err)
		}
		subPath, _, err := connection.ConfigString(string(models.ConnectionConfigKeySubPath))
		if err != nil {
			return fmt.Errorf("volume_mount connection '%s' has invalid config: %w", connection.ID, err)
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
	}

	return nil
}

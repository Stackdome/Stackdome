package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type connectionReconciler struct {
	volumeService volumeService
	logger        logger.Logger
}

type ConnectionReconcilerSpec struct {
	VolumeService volumeService
}

func NewConnectionReconciler(spec ConnectionReconcilerSpec) *connectionReconciler {
	return &connectionReconciler{
		volumeService: spec.VolumeService,
		logger:        logger.NewLoggerWithPrefix(context.Background(), "stack-connection-reconciler"),
	}
}

func (r *connectionReconciler) Name() string {
	return "connection-reconciler"
}

func (r *connectionReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp != nil {
		return resultNil, nil
	}

	hasVolumeMounts := false
	hasBuildArtifactSources := false
	for _, connection := range stack.Connections {
		if connection.Kind == models.ConnectionKindVolumeMount {
			hasVolumeMounts = true
		}
		if connection.Kind == models.ConnectionKindBuildArtifactSource {
			hasBuildArtifactSources = true
		}
	}
	if !hasVolumeMounts && !hasBuildArtifactSources {
		return resultNil, nil
	}

	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, stack.ID)
	if serr != nil {
		return resultNil, fmt.Errorf("failed to list volumes for stack '%s': %w", stack.ID, serr)
	}
	volumeMap := make(map[string]*models.Volume, len(volumes))
	for _, v := range volumes {
		volumeMap[v.Name] = v
	}
	stack.Volumes = volumes

	resourceMap := stack.ResourcesMap()

	if hasVolumeMounts {
		if err := r.resolveVolumeMountConnections(stack, resourceMap, volumeMap); err != nil {
			return resultNil, err
		}
	}

	if hasBuildArtifactSources {
		if err := r.resolveBuildArtifactSourceConnections(stack, resourceMap, volumeMap); err != nil {
			return resultNil, err
		}
	}

	return resultNil, nil
}

func (r *connectionReconciler) resolveVolumeMountConnections(
	stack *models.Stack,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
) error {
	for _, connection := range stack.Connections {
		if connection.Kind != models.ConnectionKindVolumeMount {
			continue
		}

		volume, ok := volumeMap[connection.From.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown volume '%s'", connection.Id, connection.From.Name)
		}

		resource, ok := resourceMap[connection.To.Name]
		if !ok {
			return fmt.Errorf("volume_mount connection '%s' references unknown stack resource '%s'", connection.Id, connection.To.Name)
		}

		mountPath, _ := stringFromConfig(connection.Config, "mount_path")
		subPath, _ := stringFromConfig(connection.Config, "sub_path")

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
	for _, connection := range stack.Connections {
		if connection.Kind != models.ConnectionKindBuildArtifactSource {
			continue
		}

		if _, ok := resourceMap[connection.From.Name]; !ok {
			return fmt.Errorf("build_artifact_source connection '%s' references unknown stack resource '%s'", connection.Id, connection.From.Name)
		}

		volume, ok := volumeMap[connection.To.Name]
		if !ok {
			return fmt.Errorf("build_artifact_source connection '%s' references unknown volume '%s'", connection.Id, connection.To.Name)
		}

		sourcePath, _ := stringFromConfig(connection.Config, "source_path")
		destinationPath, _ := stringFromConfig(connection.Config, "destination_path")

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

func stringFromConfig(config map[string]interface{}, key string) (string, bool) {
	v, ok := config[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

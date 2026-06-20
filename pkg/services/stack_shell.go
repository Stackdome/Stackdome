package services

import "github.com/ashishmax31/stackdome-api-server/pkg/models"

// stackShellFrom returns a stack spec with child resources and volumes stripped.
// The stack row must exist before volumes or resources are persisted.
func stackShellFrom(spec *models.Stack) models.Stack {
	shell := *spec
	shell.Volumes = nil
	shell.StackResources = nil
	return shell
}

package services

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestStackShellFrom(t *testing.T) {
	spec := &models.Stack{
		Name: "demo",
		Volumes: []*models.Volume{
			{Name: "data"},
		},
		StackResources: []*models.StackResource{
			{Name: "web"},
		},
	}

	shell := stackShellFrom(spec)

	assert.Equal(t, "demo", shell.Name)
	assert.Nil(t, shell.Volumes)
	assert.Nil(t, shell.StackResources)
}

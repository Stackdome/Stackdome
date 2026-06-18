package services

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyStackResourcePortDefaults(t *testing.T) {
	resource := &models.StackResource{
		Ports: models.Ports{
			{Number: 80, ExposedToPublic: true},
			{Number: 443, ExposedToPublic: true, Protocol: "https"},
		},
	}

	applyStackResourcePortDefaults(resource)

	assert.Equal(t, "http", resource.Ports[0].Protocol)
	assert.Equal(t, "https", resource.Ports[1].Protocol)
}

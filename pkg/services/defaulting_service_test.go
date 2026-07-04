package services

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyStackSettingsDefaults_NilSettings(t *testing.T) {
	stack := &models.Stack{}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	assert.NotNil(t, result.Settings)
	assert.Equal(t, models.DefaultReleaseRetentionLimit, result.Settings.ReleaseRetentionLimit)
	assert.Equal(t, models.DefaultMinSuccessfulReleases, result.Settings.MinSuccessfulReleases)
	assert.Equal(t, models.DefaultDeployTimeoutMinutes, result.Settings.DeployTimeoutMinutes)
}

func TestApplyStackSettingsDefaults_PartialSettings(t *testing.T) {
	stack := &models.Stack{
		Settings: &models.StackSettings{
			ReleaseRetentionLimit: 20,
		},
	}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	assert.Equal(t, 20, result.Settings.ReleaseRetentionLimit)
	assert.Equal(t, models.DefaultMinSuccessfulReleases, result.Settings.MinSuccessfulReleases)
	assert.Equal(t, models.DefaultDeployTimeoutMinutes, result.Settings.DeployTimeoutMinutes)
}

func TestApplyStackSettingsDefaults_AllSet(t *testing.T) {
	stack := &models.Stack{
		Settings: &models.StackSettings{
			ReleaseRetentionLimit: 30,
			MinSuccessfulReleases: 10,
			DeployTimeoutMinutes:  45,
		},
	}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	assert.Equal(t, 30, result.Settings.ReleaseRetentionLimit)
	assert.Equal(t, 10, result.Settings.MinSuccessfulReleases)
	assert.Equal(t, 45, result.Settings.DeployTimeoutMinutes)
}

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

package services

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyStackSettingsDefaults_NilSettings(t *testing.T) {
	stack := &models.Stack{}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	assert.NotNil(t, result.Settings)
	assert.Equal(t, models.DefaultReleaseRetentionLimit, result.Settings.ReleaseRetentionLimit)
	assert.Equal(t, models.DefaultMinSuccessfulReleases, result.Settings.MinSuccessfulReleases)
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
}

func TestApplyStackSettingsDefaults_AllSet(t *testing.T) {
	stack := &models.Stack{
		Settings: &models.StackSettings{
			ReleaseRetentionLimit: 30,
			MinSuccessfulReleases: 10,
		},
	}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	assert.Equal(t, 30, result.Settings.ReleaseRetentionLimit)
	assert.Equal(t, 10, result.Settings.MinSuccessfulReleases)
}

func TestNormalizeStackResourceReplicas(t *testing.T) {
	int32Ptr := func(v int32) *int32 { return &v }

	t.Run("job replicas hard-coded to 1 regardless of input", func(t *testing.T) {
		resource := &models.StackResource{
			WorkloadType: models.WorkloadTypeJob,
			Replicas:     int32Ptr(5),
		}
		normalizeStackResourceReplicas(resource)
		require.NotNil(t, resource.Replicas)
		assert.Equal(t, singleInstanceReplicas, *resource.Replicas)
	})

	t.Run("cronjob replicas set to 1 even when unset", func(t *testing.T) {
		resource := &models.StackResource{
			WorkloadType: models.WorkloadTypeCronJob,
		}
		normalizeStackResourceReplicas(resource)
		require.NotNil(t, resource.Replicas)
		assert.Equal(t, singleInstanceReplicas, *resource.Replicas)
	})

	t.Run("service replicas untouched", func(t *testing.T) {
		resource := &models.StackResource{
			WorkloadType: models.WorkloadTypeService,
			Replicas:     int32Ptr(3),
		}
		normalizeStackResourceReplicas(resource)
		require.NotNil(t, resource.Replicas)
		assert.Equal(t, int32(3), *resource.Replicas)
	})

	t.Run("worker with unset replicas stays unset", func(t *testing.T) {
		resource := &models.StackResource{
			WorkloadType: models.WorkloadTypeWorker,
		}
		normalizeStackResourceReplicas(resource)
		assert.Nil(t, resource.Replicas)
	})
}

func TestPopulateDefaultValues_NormalizesJobReplicas(t *testing.T) {
	replicas := int32(7)
	stack := &models.Stack{
		StackResources: []*models.StackResource{
			{Name: "batch", WorkloadType: models.WorkloadTypeJob, Replicas: &replicas},
			{Name: "web", WorkloadType: models.WorkloadTypeService, Replicas: &replicas},
		},
	}
	svc := NewStackDefaultingService()
	result, err := svc.PopulateDefaultValues(stack)
	assert.NoError(t, err)
	require.NotNil(t, result.StackResources[0].Replicas)
	assert.Equal(t, singleInstanceReplicas, *result.StackResources[0].Replicas)
	require.NotNil(t, result.StackResources[1].Replicas)
	assert.Equal(t, int32(7), *result.StackResources[1].Replicas)
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

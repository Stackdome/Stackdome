package services

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeStackResourceSubdomainPrefix_stableForStackIDAndName(t *testing.T) {
	stackID := "stack-abc"
	name := "web"
	port := 8080

	first := EncodeStackResourceSubdomainPrefix(stackID, name, port)
	second := EncodeStackResourceSubdomainPrefix(stackID, name, port)

	assert.Equal(t, first, second)
	assert.Len(t, first, 16)
	assert.NotEqual(t,
		EncodeStackResourceSubdomainPrefix(stackID, name, 9090),
		EncodeStackResourceSubdomainPrefix(stackID, "api", port),
	)
}

func TestAssignExposedPortFQDNs_generatedPrefix(t *testing.T) {
	ports := models.Ports{
		{Number: 8080, ExposedToPublic: true},
	}

	AssignExposedPortFQDNs("stack-1", "web", "example.com", ports)

	require.NotEmpty(t, ports[0].GeneratedSubdomainPrefix)
	assert.Contains(t, ports[0].ExposedFqdn, "web.example.com")
	assert.Contains(t, ports[0].ExposedFqdn, ports[0].GeneratedSubdomainPrefix)
}

func TestAssignExposedPortFQDNs_customPrefix(t *testing.T) {
	ports := models.Ports{
		{Number: 443, ExposedToPublic: true, SubdomainPrefix: "api"},
	}

	AssignExposedPortFQDNs("stack-1", "web", "example.com", ports)

	assert.Empty(t, ports[0].GeneratedSubdomainPrefix)
	assert.Equal(t, "api.example.com", ports[0].ExposedFqdn)
}

func TestAssignExposedPortFQDNs_skipsPrivatePorts(t *testing.T) {
	ports := models.Ports{
		{Number: 8080, ExposedToPublic: false},
	}

	AssignExposedPortFQDNs("stack-1", "web", "example.com", ports)

	assert.Empty(t, ports[0].ExposedFqdn)
	assert.Empty(t, ports[0].GeneratedSubdomainPrefix)
}

func TestStackResourceHasExposedPorts(t *testing.T) {
	assert.False(t, stackResourceHasExposedPorts(models.Ports{{Number: 80}}))
	assert.True(t, stackResourceHasExposedPorts(models.Ports{{Number: 80, ExposedToPublic: true}}))
}

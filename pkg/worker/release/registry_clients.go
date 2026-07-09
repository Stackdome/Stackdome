package release

import (
	"strings"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/credentials"
)

//go:generate mockgen -source=registry_clients.go -destination=registry_clients_mock_test.go -package=release

// registryClientProvider builds registry clients from resolved credentials so
// release-time image/push checks can be faked in tests.
type registryClientProvider interface {
	ClientFor(resolved *credentials.ResolvedRegistryCredential) (clients.RegistryClient, error)
}

type defaultRegistryClientProvider struct{}

func (defaultRegistryClientProvider) ClientFor(resolved *credentials.ResolvedRegistryCredential) (clients.RegistryClient, error) {
	if resolved != nil && resolved.Username != "" && resolved.Password != "" {
		return clients.NewRegistryClientWithAuth(resolved.Username, resolved.Password)
	}
	return clients.NewRegistryClientAnonymous()
}

// registryHostForRef extracts the normalized registry host from an
// image/repository ref, falling back to the raw ref when it can't be parsed.
func registryHostForRef(ref string) string {
	host, err := clients.NormalizeRegistryHost(ref)
	if err != nil {
		return ref
	}
	return host
}

// isClusterLocalRegistryRef reports whether ref points at an in-cluster
// registry; those aren't reachable from the hub.
func isClusterLocalRegistryRef(ref string) bool {
	return strings.HasSuffix(registryHostForRef(ref), clients.ClusterLocalRegistrySuffix)
}

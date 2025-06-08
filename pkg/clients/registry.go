package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type RegistryClient interface {
	CheckImage(ctx context.Context, imageRef string) (bool, error)
}

// registryClient provides functionalities to interact with a container registry.
type registryClient struct {
	auth authn.Authenticator
}

func NewRegistryClientWithAuth(username, password string) (RegistryClient, error) {
	var auth authn.Authenticator
	if username != "" && password != "" {
		auth = &authn.Basic{
			Username: username,
			Password: password,
		}
	} else {
		return nil, fmt.Errorf("username and password must be provided for authentication")
	}
	return &registryClient{
		auth: auth,
	}, nil
}

func NewRegistryClientAnonymous() (RegistryClient, error) {
	return &registryClient{
		auth: authn.Anonymous,
	}, nil
}

// CheckImage checks if the specified image exists and is pullable
func (ic *registryClient) CheckImage(ctx context.Context, imageRef string) (bool, error) {
	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, fmt.Errorf("invalid image reference: %v", err)
	}

	_, err = remote.Get(ref, remote.WithAuth(ic.auth), remote.WithContext(ctx))
	if err != nil {
		if isAuthError(err) {
			return false, fmt.Errorf("authentication failed for image")
		} else if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check image: %v", err)
	}

	_, err = remote.Image(ref, remote.WithAuth(ic.auth), remote.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("image exists but not pullable: %v", err)
	}

	return true, nil
}

// Helper function to check if error is authentication related
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "403")
}

// Helper function to check if error is not found related
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "name unknown") ||
		strings.Contains(errStr, "manifest_unknown: manifest unknown; unknown tag")
}

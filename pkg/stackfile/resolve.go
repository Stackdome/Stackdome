package stackfile

import (
	"context"
	"fmt"

	openapi "github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

type Resolver interface {
	ResolveSecretByName(ctx context.Context, name string) (string, error)
	ResolveAddonByName(ctx context.Context, addonType, name string) (string, error)
}

func ResolveStack(ctx context.Context, stack *openapi.Stack, resolver Resolver) error {
	for i := range stack.Spec.StackResources {
		res := &stack.Spec.StackResources[i]
		if res.BuildSpec == nil || res.BuildSpec.SourceContext.GitRepo == nil {
			continue
		}
		gitSecret := res.BuildSpec.SourceContext.GitRepo.GitSecret
		if gitSecret == nil || gitSecret.SecretId == "" {
			continue
		}
		id, err := resolver.ResolveSecretByName(ctx, gitSecret.SecretId)
		if err != nil {
			return fmt.Errorf("git secret %q not found for resource %q", gitSecret.SecretId, res.Name)
		}
		gitSecret.SecretId = id
	}

	for i := range stack.Spec.Connections {
		conn := &stack.Spec.Connections[i]

		switch conn.From.Type {
		case "secret":
			if conn.From.Name == nil || *conn.From.Name == "" {
				continue
			}
			id, err := resolver.ResolveSecretByName(ctx, *conn.From.Name)
			if err != nil {
				return fmt.Errorf("secret %q not found", *conn.From.Name)
			}
			conn.From.Id = &id
			conn.From.Name = nil

		case "addon/postgres":
			if conn.From.Name == nil || *conn.From.Name == "" {
				continue
			}
			id, err := resolver.ResolveAddonByName(ctx, "postgres", *conn.From.Name)
			if err != nil {
				return fmt.Errorf("postgres addon %q not found", *conn.From.Name)
			}
			conn.From.Id = &id
			conn.From.Name = nil

		default:
			if isAddonType(conn.From.Type) && conn.From.Name != nil {
				return fmt.Errorf("unsupported addon type in connection: %s", conn.From.Type)
			}
		}
	}
	return nil
}

func isAddonType(t string) bool {
	return len(t) > 6 && t[:6] == "addon/"
}

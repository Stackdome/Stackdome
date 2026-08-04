package controllers

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=release_resolve.go -destination=../mocks/mock_release_resolver.go -package=mocks

type ReleaseResolver interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetLatestByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
}

// ResolveEventRelease picks the release a CR's state belongs to:
//
//   - The CR's release-id annotation is authoritative — the apply step stamps
//     it on every deploy, and the agent copies it onto the resources it creates.
//   - Record on that release while it is active, or terminal but still the
//     latest: a cert can finish issuing, or a build report land, after the
//     release is done.
//   - Superseded (a newer sequence exists): nil. That state describes an old
//     rollout.
//   - No annotation (pre-annotation CR): fall back to the active release.
func ResolveEventRelease(ctx context.Context, resolver ReleaseResolver, log logger.Logger, stackID, releaseID string) *models.StackRelease {
	if releaseID == "" {
		active, serr := resolver.InternalGetActiveByStackID(ctx, stackID)
		if serr != nil {
			log.Error(ctx, "failed to look up active release for stack %s: %v", stackID, serr)
			return nil
		}
		return active
	}

	release, serr := resolver.InternalGet(ctx, releaseID)
	if serr != nil {
		if !serr.Is404() {
			log.Error(ctx, "failed to look up release %s for stack %s: %v", releaseID, stackID, serr)
		}
		return nil
	}
	if release.State.Active() {
		return release
	}

	latest, serr := resolver.InternalGetLatestByStackID(ctx, stackID)
	if serr != nil {
		log.Error(ctx, "failed to look up latest release for stack %s: %v", stackID, serr)
		return nil
	}
	if latest == nil || latest.ID != release.ID {
		log.Debug(ctx, "CR for stack %s carries superseded release %s; skipping events", stackID, releaseID)
		return nil
	}
	return release
}

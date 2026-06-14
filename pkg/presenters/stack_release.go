package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentStackRelease(r *models.StackRelease) map[string]interface{} {
	result := map[string]interface{}{
		"id":                r.ID,
		"stack_id":          r.StackID,
		"sequence":          r.Sequence,
		"state":             string(r.State),
		"message":           r.Message,
		"snapshot_revision": r.SnapshotRevision,
		"manifest_revision": r.ManifestRevision,
		"renderer_version":  r.RendererVersion,
		"created_by":        r.CreatedBy,
		"created_at":        r.CreatedAt,
		"updated_at":        r.UpdatedAt,
	}
	if r.RenderedAt != nil {
		result["rendered_at"] = r.RenderedAt
	}
	if r.CompletedAt != nil {
		result["completed_at"] = r.CompletedAt
	}
	result["cause"] = map[string]interface{}{
		"kind":   string(r.Cause.Kind),
		"detail": r.Cause.Detail,
	}
	if r.Outcome != nil {
		result["outcome"] = r.Outcome
	}
	if r.Pins.Resources != nil {
		result["pins"] = r.Pins
	}
	return result
}

func PresentStackReleaseList(releases []*models.StackRelease) map[string]interface{} {
	items := make([]map[string]interface{}, len(releases))
	for i, r := range releases {
		items[i] = PresentStackRelease(r)
	}
	return map[string]interface{}{
		"items": items,
		"total": len(releases),
	}
}

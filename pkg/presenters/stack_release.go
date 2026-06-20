package presenters

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentStackRelease(r *models.StackRelease) openapi.StackRelease {
	state := openapi.StackReleaseState(r.State)
	result := openapi.StackRelease{
		Id:               &r.ID,
		StackId:          &r.StackID,
		Sequence:         ptr.To(int32(r.Sequence)),
		State:            &state,
		Message:          &r.Message,
		SnapshotRevision: &r.SnapshotRevision,
		ManifestRevision: &r.ManifestRevision,
		RendererVersion:  &r.RendererVersion,
		CreatedBy:        &r.CreatedBy,
		CreatedAt:        &r.CreatedAt,
		UpdatedAt:        &r.UpdatedAt,
		RenderedAt:       r.RenderedAt,
		CompletedAt:      r.CompletedAt,
	}

	causeKind := openapi.ReleaseCauseKind(r.Cause.Kind)
	result.Cause = &openapi.ReleaseCause{
		Kind:   &causeKind,
		Detail: &r.Cause.Detail,
	}

	if r.Outcome != nil {
		result.Outcome = presentReleaseOutcome(r.Outcome)
	}

	if r.Pins.Resources != nil {
		result.Pins = presentReleasePins(&r.Pins)
	}

	return result
}

func presentReleasePins(p *models.ReleasePins) *openapi.ReleasePins {
	if p == nil || p.Resources == nil {
		return nil
	}
	resources := make(map[string]openapi.ResourcePins, len(p.Resources))
	for name, rp := range p.Resources {
		resources[name] = openapi.ResourcePins{
			GitSha:      ptr.To(rp.GitSHA),
			VolumeHash:  ptr.To(rp.VolumeHash),
			ImageDigest: ptr.To(rp.ImageDigest),
		}
	}
	return &openapi.ReleasePins{Resources: &resources}
}

func presentReleaseOutcome(o *models.ReleaseOutcome) *openapi.ReleaseOutcome {
	if o == nil {
		return nil
	}
	result := &openapi.ReleaseOutcome{
		Duration: ptr.To(fmt.Sprintf("%s", o.Duration)),
	}
	if o.Resources != nil {
		resources := make(map[string]openapi.ResourceOutcome, len(o.Resources))
		for name, ro := range o.Resources {
			resources[name] = openapi.ResourceOutcome{
				Phase:         ptr.To(string(ro.Phase)),
				ReadyReplicas: &ro.ReadyReplicas,
				Replicas:      &ro.Replicas,
				Message:       &ro.Message,
			}
		}
		result.Resources = &resources
	}
	return result
}

func PresentStackReleaseList(releases []*models.StackRelease) openapi.StackReleaseList {
	items := make([]openapi.StackRelease, len(releases))
	for i, r := range releases {
		items[i] = PresentStackRelease(r)
	}
	return openapi.StackReleaseList{
		Items: items,
		Total: ptr.To(int32(len(releases))),
	}
}

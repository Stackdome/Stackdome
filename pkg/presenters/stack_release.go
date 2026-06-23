package presenters

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
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

func PresentStackReleaseDetail(r *models.StackRelease) openapi.StackReleaseDetail {
	base := PresentStackRelease(r)
	detail := openapi.StackReleaseDetail{
		Id:               base.Id,
		StackId:          base.StackId,
		Sequence:         base.Sequence,
		State:            base.State,
		Message:          base.Message,
		Cause:            base.Cause,
		SnapshotRevision: base.SnapshotRevision,
		ManifestRevision: base.ManifestRevision,
		RendererVersion:  base.RendererVersion,
		Pins:             base.Pins,
		Outcome:          base.Outcome,
		CreatedBy:        base.CreatedBy,
		CreatedAt:        base.CreatedAt,
		UpdatedAt:        base.UpdatedAt,
		RenderedAt:       base.RenderedAt,
		CompletedAt:      base.CompletedAt,
		Snapshot:         presentStackReleaseSnapshot(&r.Snapshot),
	}
	return detail
}

func presentStackReleaseSnapshot(s *models.StackSnapshot) *openapi.StackReleaseSnapshot {
	if s == nil {
		return nil
	}
	labels := s.Stack.Labels.ToMap()
	annotations := s.Stack.Annotations.ToMap()
	snap := &openapi.StackReleaseSnapshot{
		Stack: &openapi.StackReleaseSnapshotStack{
			Id:             &s.Stack.ID,
			OrganisationId: &s.Stack.OrganisationID,
			TeamId:         &s.Stack.TeamID,
			ClusterId:      &s.Stack.ClusterID,
			UserId:         &s.Stack.UserID,
			Name:           &s.Stack.Name,
			NamespaceId:    &s.Stack.NamespaceID,
			Namespace:      &s.Stack.Namespace,
			Labels:         &labels,
			Annotations:    &annotations,
		},
		Resources:   PresentStackResourceList(s.Resources),
		Volumes:     PresentVolumeList(s.Volumes, false),
		Connections: PresentStackConnections(s.Connections),
		CapturedAt:  &s.CapturedAt,
	}
	return snap
}

func PresentStackReleaseList(result *stores.PaginatedResult[*models.StackRelease]) openapi.StackReleaseList {
	items := make([]openapi.StackRelease, len(result.Items))
	for i, r := range result.Items {
		items[i] = PresentStackRelease(r)
	}
	return openapi.StackReleaseList{
		Items:      items,
		Total:      ptr.To(int32(result.Total)),
		Page:       ptr.To(int32(result.Page)),
		PageSize:   ptr.To(int32(result.PageSize)),
		TotalPages: ptr.To(int32(result.TotalPages)),
	}
}

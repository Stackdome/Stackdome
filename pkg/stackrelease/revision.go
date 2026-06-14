package stackrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/util/rand"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

// RendererVersion is bumped whenever the rendering logic changes in a way that
// could produce different CRs from the same input.
const RendererVersion = "2026-06-14.1"

var printer = spew.ConfigState{
	Indent:         " ",
	SortKeys:       true,
	DisableMethods: true,
	SpewKeys:       true,
}

// ComputeResourceRevision computes a deterministic revision hash for a
// rendered StackResource CR, using the same fnv+spew approach as the
// cluster resource builder.
func ComputeResourceRevision(cr *corev1alpha1.StackResource) string {
	hasher := fnv.New64a()
	printer.Fprintf(hasher, "%#v", cr.Spec)
	printer.Fprintf(hasher, "%#v", cr.Labels)
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum64()))
}

// ComputeManifestRevision computes a single revision hash over the entire
// release manifest, combining all resource revisions.
func ComputeManifestRevision(m *models.ReleaseManifest) string {
	hasher := fnv.New64a()
	printer.Fprintf(hasher, "%#v", m.StackCR)
	printer.Fprintf(hasher, "%#v", m.ResourceNames)
	printer.Fprintf(hasher, "%#v", m.ResourceRevisions)
	printer.Fprintf(hasher, "%#v", m.VolumeBuildSources)
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum64()))
}

// HashJSON returns a hex-encoded SHA-256 of the JSON representation of v.
func HashJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("hash json marshal: %w", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}

package services

import (
	"strings"

	"github.com/Stackdome/stackdome/pkg/models"
)

// statefulImages are datastores that corrupt when two instances write the same
// data directory, so they must never be rolled with surge. Keyed by the final
// path segment of the image reference, or by the final two segments where the
// product name sits in the penultimate one.
var statefulImages = map[string]struct{}{
	"postgres":             {},
	"postgresql":           {},
	"mysql":                {},
	"mariadb":              {},
	"percona-server-mysql": {},
	"mssql":                {},
	"mssql/server":         {},
	"mongo":                {},
	"mongodb":              {},
	"redis":                {},
	"valkey":               {},
	"keydb":                {},
	"rabbitmq":             {},
	"kafka":                {},
	"zookeeper":            {},
	"nats":                 {},
	"elasticsearch":        {},
	"opensearch":           {},
	"cassandra":            {},
	"scylla":               {},
	"clickhouse":           {},
	"clickhouse-server":    {},
	"minio":                {},
	"etcd":                 {},
	"influxdb":             {},
	"neo4j":                {},
	"couchdb":              {},
	"couchbase":            {},
	"prometheus":           {},
	"grafana":              {},
	"otel-lgtm":            {},
	"gitea":                {},
}

// applyStatefulWorkloadDefault promotes a known datastore image to
// StatefulService. A Deployment rolls with surge, so the replacement pod mounts
// the data directory while the old one is still writing it; a StatefulSet
// terminates before it replaces. Only an unset or Service type is promoted.
//
// Deliberately keyed on the image rather than on the presence of a volume:
// plenty of application workloads keep state on a volume and tolerate an
// overlapping rollout, and promoting those would trade their zero-downtime
// deploys for nothing.
func applyStatefulWorkloadDefault(resource *models.StackResource) {
	if resource.ImageConfig == nil {
		return
	}
	if resource.WorkloadType != "" && resource.WorkloadType != models.WorkloadTypeService {
		return
	}
	for _, key := range imageNameKeys(resource.ImageConfig.Image) {
		if _, ok := statefulImages[key]; ok {
			resource.WorkloadType = models.WorkloadTypeStatefulService
			return
		}
	}
}

// imageNameKeys reduces an image reference to its final path segment and, when
// there is one, its final two segments — vendors such as
// mcr.microsoft.com/mssql/server put the product name in the penultimate
// segment. The registry host, the tag and the digest are dropped.
func imageNameKeys(image string) []string {
	ref := image
	if idx := strings.Index(ref, "@"); idx >= 0 {
		ref = ref[:idx]
	}
	segments := strings.Split(ref, "/")
	last := segments[len(segments)-1]
	if idx := strings.Index(last, ":"); idx >= 0 {
		last = last[:idx]
	}
	last = strings.ToLower(last)
	if len(segments) < 2 {
		return []string{last}
	}
	return []string{last, strings.ToLower(segments[len(segments)-2]) + "/" + last}
}

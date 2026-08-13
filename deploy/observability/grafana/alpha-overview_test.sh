#!/bin/sh
set -eu

dashboard=${1:-deploy/observability/grafana/alpha-overview.json}
excluded_routes='/health|/metrics|.*stream.*'

jq -e --arg excluded_routes "$excluded_routes" '
  . as $dashboard |
  def panel($title): ([$dashboard.panels[] | select(.title == $title)][0] // null);
  def require($condition; $message): if $condition then true else error($message) end;
  require(.time.from == "now-3h"; "dashboard must default to the last three hours") and
  require(.refresh == "15s"; "dashboard must refresh every 15 seconds") and
  require([.panels[].id] | all(. != null) and (length == (unique | length)); "panel IDs must be present and unique") and
  require(["Request rate", "5xx error rate", "In-flight requests", "Failed stacks", "Worker queue depth", "Snapshot collection", "Average latency", "p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title) != null); "overview and aggregate latency panels must be present") and
  require(panel("Cloud compute access").type == "row"; "cloud compute access row must be present") and
  require(["Total entitlements issued", "Expired entitlements", "Active leases", "Reserved capacity", "Cleanup backlog", "Lease breakdown by state"] | all(. as $title | panel($title) != null); "cloud compute access panels must be present") and
  require((panel("Total entitlements issued").targets[0].expr | contains("stackdome_cloud_compute_entitlements_issued") and contains("max by (source)")); "total entitlements must aggregate replica-safe source gauges") and
  require((panel("Expired entitlements").targets[0].expr | contains("stackdome_cloud_compute_entitlements_expired") and contains("max by (source)")); "expired entitlements must aggregate replica-safe source gauges") and
  require(panel("Active leases").targets[0].expr | contains("stackdome_cloud_shared_compute_leases{state=\"active\"}"); "active leases must select the active lifecycle state") and
  require((panel("Reserved capacity").targets[0].expr | contains("stackdome_cloud_shared_compute_leases{state!=\"cleaned\"}") and contains("max by (state)")); "reserved capacity must count every non-cleaned lease without multiplying replicas") and
  require(panel("Reserved capacity").targets[1].expr | contains("stackdome_cloud_shared_compute_capacity_limit"); "reserved capacity must show the configured limit") and
  require((panel("Cleanup backlog").targets[0].expr | contains("state=~\"cleanup_pending|cleaning|error\"") and contains("max by (state)")); "cleanup backlog must include pending, cleaning, and error leases") and
  require(panel("Lease breakdown by state").targets[0].expr | contains("max by (state) (stackdome_cloud_shared_compute_leases)"); "lease breakdown must remain replica-safe and grouped by state") and
  require(["Total entitlements issued", "Expired entitlements", "Active leases", "Reserved capacity", "Cleanup backlog"] | all(. as $title | (panel($title).targets[0].expr | contains("or vector(0)") | not)); "cloud snapshot panels must show missing data instead of converting collection failures to zero") and
  require(["Average latency", "p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title).targets[0].expr | contains($excluded_routes)); "aggregate latency panels must exclude internal and streaming routes") and
  require(panel("Average latency").targets[0].expr | contains("_sum") and contains("_count"); "average latency must divide histogram sum by count") and
  require(["p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title).targets[0].expr | contains("sum by (le)")); "percentiles must aggregate histogram buckets by le") and
  require(["5xx error rate", "Worker queue depth", "Snapshot collection", "p95 latency"] | all(. as $title | panel($title).options.colorMode == "value"); "healthy overview states must not fill cards with background color") and
  require(panel("Failed stacks").options.colorMode == "background"; "failed stacks must retain the strongest incident emphasis") and
  require([.panels[] | select(.type != "row") | .description] | all(type == "string" and length > 0); "every data panel must have a description")
' "$dashboard" >/dev/null

#!/bin/sh
set -eu

dashboard=${1:-deploy/observability/grafana/alpha-overview.json}
excluded_routes='/health|/metrics|.*stream.*'

jq -e --arg excluded_routes "$excluded_routes" '
  . as $dashboard |
  def panel($title): first($dashboard.panels[] | select(.title == $title));
  def require($condition; $message): if $condition then true else error($message) end;
  require(.time.from == "now-3h"; "dashboard must default to the last three hours") and
  require(.refresh == "15s"; "dashboard must refresh every 15 seconds") and
  require([.panels[].id] | all(. != null) and (length == (unique | length)); "panel IDs must be present and unique") and
  require(["Request rate", "5xx error rate", "In-flight requests", "Failed stacks", "Worker queue depth", "Snapshot collection", "Average latency", "p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title) != null); "overview and aggregate latency panels must be present") and
  require(["Average latency", "p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title).targets[0].expr | contains($excluded_routes)); "aggregate latency panels must exclude internal and streaming routes") and
  require(panel("Average latency").targets[0].expr | contains("_sum") and contains("_count"); "average latency must divide histogram sum by count") and
  require(["p50 latency", "p95 latency", "p99 latency"] | all(. as $title | panel($title).targets[0].expr | contains("sum by (le)")); "percentiles must aggregate histogram buckets by le") and
  require(["5xx error rate", "Worker queue depth", "Snapshot collection", "p95 latency"] | all(. as $title | panel($title).options.colorMode == "value"); "healthy overview states must not fill cards with background color") and
  require(panel("Failed stacks").options.colorMode == "background"; "failed stacks must retain the strongest incident emphasis") and
  require([.panels[] | select(.type != "row") | .description] | all(type == "string" and length > 0); "every data panel must have a description")
' "$dashboard" >/dev/null

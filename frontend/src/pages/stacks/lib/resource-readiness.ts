// cluster-agent api/core/v1alpha1/stack_resource_types.go:21 — rollout states
// are Pending | Ready | Degraded | Failed; comparisons are case-insensitive to
// match statusVariant()'s handling of server-provided state strings.
const RESOURCE_STATE_READY = "ready";

/** True when a resource's live rollout state allows log/metrics streaming. */
export function isResourceReady(state?: string | null): boolean {
  return (state ?? "").trim().toLowerCase() === RESOURCE_STATE_READY;
}

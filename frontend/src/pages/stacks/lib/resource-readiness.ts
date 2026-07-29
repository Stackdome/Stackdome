// cluster-agent api/core/v1alpha1/stack_resource_types.go:21 — rollout states
// are Pending | Ready | Degraded | Failed; comparisons are case-insensitive to
// match statusVariant()'s handling of server-provided state strings.
const RESOURCE_STATE_READY = "ready";

/** True when a resource's live rollout state allows log/metrics streaming. */
export function isResourceReady(state?: string | null): boolean {
  return (state ?? "").trim().toLowerCase() === RESOURCE_STATE_READY;
}

/** Named resources whose streams may open. With no live_status at all (still
 *  loading), readiness is unknown: fail open rather than blocking streams
 *  that may be fine. */
export function readyResourceNames(
  resources: ReadonlyArray<{ name?: string }>,
  liveStatusResources?: Record<string, { state?: string }> | null,
): string[] {
  const names = resources.map((r) => r.name).filter((name): name is string => !!name);
  if (!liveStatusResources) return names;
  return names.filter((name) => isResourceReady(liveStatusResources[name]?.state));
}

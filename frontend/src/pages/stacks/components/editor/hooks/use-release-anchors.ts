import { useEffect } from "react";
import type { Stack } from "@/pages/stacks/types";
import { isTerminal } from "@/pages/stacks/components/editor/tabs/deployments/release-states";
import type { useReleaseDetail } from "@/pages/stacks/components/editor/tabs/deployments/use-release-detail";

interface ReleaseSummary {
  id?: string;
  state?: string;
}

/** The three release anchors the editor page needs:
 * — baseline: the latest release (the config last shipped via Deploy), falling
 *   back to the live (currently converged) release until the releases list loads.
 *   The diff baseline is pinned to its snapshot, not the autosaved server state.
 * — current: stack.current_release — what's actually serving traffic; source for
 *   live per-resource status (public ingress endpoints in the header).
 * — status: the active non-terminal release when a deploy is under way (so the
 *   canvas/drawer reflect the in-flight rollout, not stale current-release data),
 *   else the live current_release. Distinct from current, which stays pinned to
 *   what's serving traffic for the header's PUBLIC row. */
export function useReleaseAnchors({
  stack,
  releases,
  activeRelease,
  releaseDetail,
}: {
  stack: Stack | null;
  releases: ReleaseSummary[];
  activeRelease: ReleaseSummary | undefined;
  releaseDetail: ReturnType<typeof useReleaseDetail>;
}) {
  const baselineReleaseId = activeRelease?.id ?? stack?.current_release?.id;
  useEffect(() => {
    if (baselineReleaseId) releaseDetail.ensure(baselineReleaseId);
  }, [baselineReleaseId, releaseDetail]);
  const deployedSnapshot = releaseDetail.peek(baselineReleaseId).data?.snapshot;

  const currentReleaseId = stack?.current_release?.id;
  useEffect(() => {
    if (currentReleaseId) releaseDetail.ensure(currentReleaseId);
  }, [currentReleaseId, releaseDetail]);
  const currentReleaseDetail = releaseDetail.peek(currentReleaseId).data;

  const nonTerminalRelease = releases.find((r) => !isTerminal(r.state));
  const statusReleaseId = nonTerminalRelease?.id ?? stack?.current_release?.id;
  const statusReleaseState = nonTerminalRelease?.state ?? stack?.current_release?.state;
  // Refetch on every id/state change (mount, and each transition — including the
  // terminal one, since statusReleaseState is a dep) plus a 5s poll while non-terminal.
  // Depend on the stable refresh callback, NOT the releaseDetail object (rebuilt every
  // render): refresh has no cached-data short-circuit, so an object dep would loop.
  const refreshRelease = releaseDetail.refresh;
  useEffect(() => {
    if (statusReleaseId) refreshRelease(statusReleaseId);
  }, [statusReleaseId, statusReleaseState, refreshRelease]);
  useEffect(() => {
    if (!statusReleaseId || isTerminal(statusReleaseState)) return;
    const t = setInterval(() => {
      if (document.visibilityState !== "hidden") refreshRelease(statusReleaseId);
    }, 5000);
    return () => clearInterval(t);
  }, [statusReleaseId, statusReleaseState, refreshRelease]);
  const statusLiveStatus = releaseDetail.peek(statusReleaseId).data?.live_status;

  return { baselineReleaseId, deployedSnapshot, currentReleaseDetail, statusReleaseId, statusLiveStatus };
}

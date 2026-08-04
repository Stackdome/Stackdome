import { useEffect } from "react";
import type { Stack } from "@/api/stack-types";
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
 * — converged: stack.converged_release — what's actually serving traffic; source for
 *   live per-resource status (public ingress endpoints in the header).
 * — status: the active non-terminal release when a deploy is under way (so the
 *   canvas/drawer reflect the in-flight rollout, not stale converged-release data),
 *   else the live converged_release. Distinct from converged, which stays pinned to
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
  const baselineReleaseId = activeRelease?.id ?? stack?.converged_release?.id;
  useEffect(() => {
    if (baselineReleaseId) releaseDetail.ensure(baselineReleaseId);
  }, [baselineReleaseId, releaseDetail]);
  const deployedSnapshot = releaseDetail.peek(baselineReleaseId).data?.snapshot;

  const convergedReleaseId = stack?.converged_release?.id;
  useEffect(() => {
    if (convergedReleaseId) releaseDetail.ensure(convergedReleaseId);
  }, [convergedReleaseId, releaseDetail]);
  const convergedReleaseDetail = releaseDetail.peek(convergedReleaseId).data;

  const nonTerminalRelease = releases.find((r) => !isTerminal(r.state));
  const statusReleaseId = nonTerminalRelease?.id ?? stack?.converged_release?.id;
  const statusReleaseState = nonTerminalRelease?.state ?? stack?.converged_release?.state;
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

  // TLS certs finish issuing after the release converges, so ingress URLs land on
  // the release detail after the terminal refetch. Keep polling the converged
  // release while a deployed public port still has no ingress URL.
  // Always terminates: a failed issuance still populates ingress with http URLs.
  const ingressPending = (() => {
    if (!convergedReleaseDetail) return false;
    const live = convergedReleaseDetail.live_status?.resources ?? {};
    return (convergedReleaseDetail.snapshot?.resources ?? []).some(
      (r) =>
        r.name &&
        (r.ports ?? []).some((p) => p.exposed_to_public) &&
        !live[r.name]?.public_ingress?.length,
    );
  })();
  useEffect(() => {
    if (!convergedReleaseId || !ingressPending) return;
    const t = setInterval(() => {
      if (document.visibilityState !== "hidden") refreshRelease(convergedReleaseId);
    }, 5000);
    return () => clearInterval(t);
  }, [convergedReleaseId, ingressPending, refreshRelease]);

  return { baselineReleaseId, deployedSnapshot, convergedReleaseDetail, statusReleaseId, statusLiveStatus };
}

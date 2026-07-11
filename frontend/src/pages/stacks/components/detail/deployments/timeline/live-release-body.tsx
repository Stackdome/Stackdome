import { useState, useEffect } from "react";
import { StageTracker } from "@/components/branded";
import { ReleaseEventScope, type StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import type { EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
import { ValidationBanner } from "@/pages/stacks/components/detail/ValidationBanner";
import { deriveStages, deriveFailingResources, deriveRecovered, resourceSource, replicaLabel } from "../derive";
import { ReleaseState, isTerminal } from "../release-states";
import { diffSnapshots } from "../release-snapshot-diff";
import type { ReleaseDetail } from "../use-release-detail";
import { releaseValidationBannerItems } from "../release-errors";
import { useReleaseEvents } from "../use-release-events";
import { type ResourceRowVM, type LogContext } from "./resource-row";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import { DeployFailedBanner } from "./deploy-failed-banner";
import { ReleaseActivityFeed } from "./release-activity-feed";

export interface LiveReleaseBodyProps {
  release: StackRelease;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  detail?: ReleaseDetail;
  prevReleaseId?: string;
  prevSeq?: number;
  onJumpToResource?: (resourceName: string, tab: EditSessionTab) => void;
  /** Refetches the releases list — bound by the page to the releases-list refetch. */
  refetchReleases?: () => void;
}

/**
 * Detail-card body for the latest deploy (releases[0]). Renders LIVE progress from
 * release.live_status.resources + the derived tracker (vs a historical node's stored outcome).
 */
export function LiveReleaseBody({ release, stack, logContext, onOpenLogs, detail, prevReleaseId, prevSeq, onJumpToResource, refetchReleases }: LiveReleaseBodyProps) {
  const [validationDismissed, setValidationDismissed] = useState(false);
  // Hybrid progress driver: the event stream pushes release-scoped state changes instead
  // of a fast poll — a resource-scoped event only updates the feed, a release-scoped one
  // (e.g. the deploy completing) also refreshes the releases list + this release's detail.
  const { events, status: eventsStatus } = useReleaseEvents({
    orgId: logContext?.orgId ?? "",
    teamName: logContext?.teamName ?? "",
    stackId: logContext?.stackId ?? "",
    releaseId: release.id,
    terminal: isTerminal(release.state),
    onEvent: (e) => {
      // Every event refreshes this release's detail — resource-scoped events are
      // exactly when live_status.resources moves. Release-scoped ones also refresh
      // the releases list (state transitions).
      if (release.id) detail?.refresh(release.id);
      if (e.scope === ReleaseEventScope.Release) refetchReleases?.();
    },
  });
  // Source image/repo from THIS release's snapshot (the frozen spec it ships), not the
  // live stack spec which may have been edited after deploy. Also prefetch prev snapshot for the diff.
  useEffect(() => {
    if (release.id) detail?.ensure(release.id);
    if (prevReleaseId) detail?.ensure(prevReleaseId);
  }, [detail, release.id, prevReleaseId]);
  // List items carry no live_status overlay (detail only) — read the fresher
  // detail-cache copy, which the event driver above keeps refreshed.
  const detailData = detail?.peek(release.id).data;
  const releaseSnapshot = detailData?.snapshot;
  const liveStatus = detailData?.live_status ?? release.live_status;
  const failing = deriveFailingResources(release, liveStatus);
  const canDiff = !!prevReleaseId && !!detail;
  // Suppress empty "no changes" until prev snapshot resolves — else the diff briefly lies.
  const prevLoaded = !prevReleaseId || !!detail?.peek(prevReleaseId).data;
  const diff = diffSnapshots(detail?.peek(prevReleaseId).data?.snapshot, detail?.peek(release.id).data?.snapshot);
  const recovered = deriveRecovered(release, liveStatus);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(release, failing, liveStatus);
  const liveResources = Object.entries(liveStatus?.resources ?? {});
  const sourceByName = new Map((releaseSnapshot?.resources ?? stack.spec?.stack_resources ?? []).map((r) => [r.name, r]));
  const failureMsg = release.state === ReleaseState.Failed && failing.length === 0 ? release.message : undefined;
  const bannerResources = stack.spec?.stack_resources ?? [];
  // validation_errors also live on the detail payload, not list items.
  const validationItems = release.state === ReleaseState.Failed
    ? releaseValidationBannerItems(detailData ?? release, bannerResources)
    : [];
  const rows: ResourceRowVM[] = liveResources.map(([name, s]) => ({
    name,
    phase: s.state ?? "",
    replicas: replicaLabel(s.available_replicas, s.replicas),
    msg: s.message,
    tag: recoveredNames.has(name) ? "RECOVERED" : undefined,
    failure: failingByName.get(name),
    source: resourceSource(sourceByName.get(name)),
  }));

  return (
    <div>
      <StageTracker stages={stages} />

      {failureMsg && <DeployFailedBanner message={failureMsg} />}

      {!validationDismissed && validationItems.length > 0 && (
        <ValidationBanner
          items={validationItems}
          onJump={onJumpToResource && ((idx, tab) => {
            // Round-trip index → name via the SAME list the items were built
            // against; the receiver re-resolves the name at click time.
            const name = bannerResources[idx]?.name;
            if (name) onJumpToResource(name, tab);
          })}
          onDismiss={() => setValidationDismissed(true)}
        />
      )}

      <ResourceOutcomeList rows={rows} logContext={logContext} onOpenLogs={onOpenLogs} />

      <div className="mt-4">
        <ReleaseActivityFeed events={events} streaming={eventsStatus === "streaming"} />
      </div>

      {canDiff && <ConfigChangesToggle diff={diff} prevSeq={prevSeq} loading={!prevLoaded} />}

      {recovered.length > 0 && (
        <div className="mt-4 rounded-md border border-warn-border bg-warn-bg px-3.5 py-2.5 text-[12.5px] text-fg-muted">
          {recovered.map((r) => (
            <div key={r.name}>
              <span className="font-medium text-foreground">{r.name}</span> recovered
              {r.restartCount != null ? ` after ${r.restartCount} ${r.restartCount === 1 ? "restart" : "restarts"}` : ""} — last failure{" "}
              <span className="text-warn">{r.reason}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

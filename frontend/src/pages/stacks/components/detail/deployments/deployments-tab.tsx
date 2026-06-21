import { useState } from "react";
import { Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import type { Stack } from "@/api/stacks";
import { createRelease, rollbackRelease, cancelRelease } from "@/api/releases";
import { useReleases } from "./use-releases";
import { deriveFailingResources } from "./derive";
import { TimelineRail } from "./timeline/timeline-rail";
import { DriftBanner, ReleaseErrorBanner } from "./timeline/banners";

export interface DeploymentsTabProps {
  orgId: string; teamName: string; stackId: string; stack: Stack; canDeploy: boolean;
  onOpenLogs?: (resourceName?: string) => void;
}

export function DeploymentsTab({ orgId, teamName, stackId, stack, canDeploy, onOpenLogs }: DeploymentsTabProps) {
  const { releases, activeRelease, loading, error, refetch } = useReleases({ orgId, teamName, stackId, enabled: true });
  const [busy, setBusy] = useState(false);
  const { toast } = useToast();

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setBusy(true);
    try { await fn(); toast({ title: ok }); refetch(); }
    catch (e) { toast({ title: "Action failed", description: e instanceof Error ? e.message : "", variant: "destructive" }); }
    finally { setBusy(false); }
  };

  const onDeploy = () => run(() => createRelease(orgId, teamName, stackId), "Deploy started");
  const onRollback = (id: string) => run(() => rollbackRelease(orgId, teamName, stackId, id), "Rollback started");
  const onCancel = (id: string) => run(() => cancelRelease(orgId, teamName, stackId, id), "Release cancelled");
  const onCopyId = (id: string) => { void navigator.clipboard?.writeText(id); toast({ title: "Release ID copied" }); };

  // Drift is heuristic: list releases carry no snapshot, so compare the stack's
  // updated_at against the active release's completed_at. Precise drift needs the
  // backend to expose the current snapshot_revision (filed follow-up §12.4).
  const stackUpdated = (stack as unknown as { updated_at?: string }).updated_at;
  const hasDrift = !!activeRelease && !!stackUpdated && !!activeRelease.completed_at
    && new Date(stackUpdated) > new Date(activeRelease.completed_at);

  const failing = deriveFailingResources(stack);
  let banner: React.ReactNode = null;
  if (hasDrift) banner = <DriftBanner onDeploy={onDeploy} busy={busy} />;
  else if (activeRelease && failing.length > 0) {
    const total = stack.status?.resources?.length ?? failing.length;
    banner = <ReleaseErrorBanner lead={`Deploy #${activeRelease.sequence} ${activeRelease.state === "Failed" ? "failed" : "failing"}`} text={`${failing.length} of ${total} resources failing`} />;
  }

  if (error) return <EmptyState title="Could not load deployments" description={error} />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] font-medium">Deployments</h2>
        {canDeploy && (
          <Button onClick={onDeploy} disabled={busy} className="gap-1.5">
            <Rocket className="h-3.5 w-3.5" /> Deploy
          </Button>
        )}
      </div>

      {loading && releases.length === 0 ? (
        <p className="text-[13px] text-fg-muted">Loading deployments…</p>
      ) : (
        <TimelineRail
          releases={releases}
          activeRelease={activeRelease}
          stack={stack}
          logContext={{ orgId, teamName, stackId }}
          onOpenLogs={onOpenLogs ? (name) => onOpenLogs(name) : undefined}
          banner={banner}
          onRollback={onRollback}
          onCancel={onCancel}
          onCopyId={onCopyId}
        />
      )}
    </div>
  );
}

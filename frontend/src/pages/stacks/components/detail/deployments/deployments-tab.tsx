import { useState } from "react";
import { Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import type { Stack } from "@/api/stacks";
import { createRelease, rollbackRelease, cancelRelease } from "@/api/releases";
import { useReleases } from "./use-releases";
import { CurrentDeploymentCard } from "./current-deployment-card";
import { ReleaseHistory } from "./release-history";
import { ReleaseDetailDrawer } from "./release-detail-drawer";

export interface DeploymentsTabProps {
  orgId: string;
  teamName: string;
  stackId: string;
  stack: Stack;
  canDeploy: boolean;
}

export function DeploymentsTab({ orgId, teamName, stackId, stack, canDeploy }: DeploymentsTabProps) {
  const { releases, activeRelease, loading, error, refetch } = useReleases({ orgId, teamName, stackId, enabled: true });
  const [openReleaseId, setOpenReleaseId] = useState<string | null>(null);
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

      {activeRelease && <CurrentDeploymentCard release={activeRelease} stack={stack} logContext={{ orgId, teamName, stackId }} />}

      <ReleaseHistory
        releases={releases}
        onViewDetails={setOpenReleaseId}
        onRollback={onRollback}
        onCancel={onCancel}
      />

      {!loading && releases.length === 0 && !activeRelease && (
        <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
      )}

      {openReleaseId && (() => {
        // releases is newest-first; the "previous" release is the next-older one (index + 1).
        const openIdx = releases.findIndex((r) => r.id === openReleaseId);
        const previousRelease = openIdx >= 0 ? releases[openIdx + 1] : undefined;
        return (
          <ReleaseDetailDrawer
            orgId={orgId} teamName={teamName} stackId={stackId}
            releaseId={openReleaseId}
            previousRelease={previousRelease}
            onClose={() => setOpenReleaseId(null)}
          />
        );
      })()}
    </div>
  );
}

import { Loader2, Rocket, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ConfigDiff } from "@/pages/stacks/components/detail/deployments/timeline/config-diff";
import type { SnapshotDiff, ResourceDiff, ItemDiff } from "@/pages/stacks/components/detail/deployments/release-snapshot-diff";

export interface ViewChangesModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Saved-but-undeployed diff (lifecycle.stagedDiff). Undefined while an edit is
   *  still autosaving — the modal shows a "saving" hint until it resolves. */
  diff?: SnapshotDiff;
  /** Total pending changes (drives the header count). */
  count: number;
  stackName: string;
  /** Revert one resource/volume by name. Removed entries can't be reverted in
   *  isolation (no draft slot to restore into) — use Discard all for those. */
  onDiscardResource: (name: string) => void;
  onDiscardVolume: (name: string) => void;
  onDiscardAll: () => void;
  onDeploy: () => void;
  deployBusy: boolean;
  canWrite: boolean;
}

/** A single diff card + its per-change Discard control. */
function ChangeRow({
  card,
  diff,
  onDiscard,
  discardHint,
}: {
  card: ResourceDiff | ItemDiff;
  diff: SnapshotDiff;
  onDiscard?: () => void;
  discardHint?: string;
}) {
  // "removed" can't be restored in isolation; connections have no session-level
  // revert. Both fall back to Discard all.
  const disabled = !onDiscard || card.change === "removed";
  const button = (
    <Button
      type="button"
      variant="outline"
      size="sm"
      className="h-7 flex-none px-2.5 text-[12px]"
      disabled={disabled}
      onClick={onDiscard}
    >
      <RotateCcw className="size-3" />
      Discard
    </Button>
  );
  return (
    <div className="flex items-start gap-2">
      <div className="min-w-0 flex-1">
        <ConfigDiff diff={diff} hasPrev />
      </div>
      {disabled && discardHint ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="flex-none">{button}</span>
          </TooltipTrigger>
          <TooltipContent side="left">{discardHint}</TooltipContent>
        </Tooltip>
      ) : (
        button
      )}
    </div>
  );
}

/**
 * Review-and-discard surface for undeployed changes. Reuses the Deployments-tab
 * `ConfigDiff` rendering; each change gets a Discard wired to the edit session's
 * per-resource/volume revert, and the footer deploys straight from here.
 */
export function ViewChangesModal({
  open,
  onOpenChange,
  diff,
  count,
  stackName,
  onDiscardResource,
  onDiscardVolume,
  onDiscardAll,
  onDeploy,
  deployBusy,
  canWrite,
}: ViewChangesModalProps) {
  const empty =
    !diff || (diff.resources.length === 0 && diff.volumes.length === 0 && diff.connections.length === 0);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[600px] gap-0 p-0">
        <DialogHeader className="flex-row items-center gap-2.5 space-y-0 border-b border-border px-5 py-4">
          <DialogTitle className="text-[16px]">Undeployed changes</DialogTitle>
          {count > 0 && (
            <span className="rounded-full border border-warn-border bg-warn-bg px-2 py-0.5 font-mono text-[11px] font-bold text-warn">
              {count}
            </span>
          )}
          <span className="flex-1" />
          <DialogDescription className="font-mono text-[11.5px] text-muted-foreground">{stackName}</DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[52vh] flex-col gap-3 overflow-auto px-5 py-4">
          {empty ? (
            <p className="py-6 text-center text-[13px] text-fg-muted">
              {diff ? "No pending changes." : "Saving changes… they'll appear here once saved."}
            </p>
          ) : (
            <>
              {diff!.resources.map((d) => (
                <ChangeRow
                  key={`r-${d.name}`}
                  card={d}
                  diff={{ resources: [d], volumes: [], connections: [] }}
                  onDiscard={() => onDiscardResource(d.name)}
                  discardHint="Removed resources restore via Discard all"
                />
              ))}
              {diff!.volumes.map((v) => (
                <ChangeRow
                  key={`v-${v.name}`}
                  card={v}
                  diff={{ resources: [], volumes: [v], connections: [] }}
                  onDiscard={() => onDiscardVolume(v.name)}
                  discardHint="Removed volumes restore via Discard all"
                />
              ))}
              {diff!.connections.map((c) => (
                <ChangeRow
                  key={`c-${c.name}`}
                  card={c}
                  diff={{ resources: [], volumes: [], connections: [c] }}
                  discardHint="Manage connections on the canvas"
                />
              ))}
            </>
          )}
        </div>

        <DialogFooter className="flex-row items-center border-t border-border px-5 py-3.5 sm:justify-start">
          <Button type="button" variant="outline" size="sm" onClick={onDiscardAll} disabled={empty}>
            Discard all
          </Button>
          <span className="flex-1" />
          <Button type="button" variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          <Button
            type="button"
            variant="default"
            size="sm"
            onClick={() => {
              onDeploy();
              onOpenChange(false);
            }}
            disabled={deployBusy || !canWrite || empty}
          >
            {deployBusy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
            {deployBusy ? "Deploying" : `Deploy${count > 0 ? ` ${count} change${count === 1 ? "" : "s"}` : ""}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

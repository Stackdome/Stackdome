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
import type {
  SnapshotDiff,
  ResourceDiff,
  DiffRow,
} from "@/pages/stacks/components/detail/deployments/release-snapshot-diff";

export interface ViewChangesModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Saved-but-undeployed diff (lifecycle.stagedDiff). Undefined while an edit is
   *  still autosaving — the modal shows a "saving" hint until it resolves. */
  diff?: SnapshotDiff;
  /** Total pending changes (drives the header count). */
  count: number;
  /** Autosave is in a terminal error state — the empty diff means "not saved",
   *  not "nothing pending". */
  errored?: boolean;
  /** Unsaved session dirt exists even though the staged diff (saved-vs-release)
   *  is empty — the edit is still in flight, so the diff can't show it yet. */
  dirty?: boolean;
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

function RowLine({ row }: { row: DiffRow }) {
  return (
    <div className="grid grid-cols-[140px_1fr] items-baseline gap-2.5 px-3 py-0.5 font-mono text-[11px]">
      <span className="text-fg-muted [overflow-wrap:anywhere]">{row.key}</span>
      <span className="flex flex-wrap items-baseline gap-1.5">
        {row.kind === "added" && (
          <>
            <span className="text-fg-muted">—</span>
            <span className="text-fg-muted">→</span>
            <span className="text-success">{row.to}</span>
          </>
        )}
        {row.kind === "removed" && <span className="text-danger">{row.from}</span>}
        {row.kind === "changed" && (
          <>
            <span className="text-danger opacity-80">{row.from}</span>
            <span className="text-fg-muted">→</span>
            <span className="text-success">{row.to}</span>
          </>
        )}
      </span>
    </div>
  );
}

const BADGE_STYLES: Record<string, string> = {
  added: "border-success-border bg-success-bg text-success",
  removed: "border-danger-border bg-danger-bg text-danger",
  modified: "border-brand-border bg-brand-bg text-brand",
  renamed: "border-brand-border bg-brand-bg text-brand",
};

const DOT_STYLES: Record<string, string> = {
  added: "bg-success",
  removed: "bg-danger",
  modified: "bg-brand",
  renamed: "bg-brand",
};

/** Compact flat change card: dot + name + badge + always-visible Discard in the
 *  head, then note and key/value rows. */
function ChangeCard({
  name,
  fromName,
  change,
  note,
  sections,
  onDiscard,
  discardHint,
}: {
  name: string;
  fromName?: string;
  change: ResourceDiff["change"];
  note?: string;
  sections: { label?: string; rows: DiffRow[] }[];
  onDiscard?: () => void;
  discardHint?: string;
}) {
  // "removed" can't be restored in isolation; connections have no session-level
  // revert. Both fall back to Discard all.
  const disabled = !onDiscard || change === "removed";
  const discardButton = (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      className="h-6 flex-none gap-1 px-2 font-mono text-[10.5px] text-fg-muted hover:text-brand"
      disabled={disabled}
      onClick={onDiscard}
    >
      <RotateCcw className="size-3" />
      Discard
    </Button>
  );
  return (
    <div data-change-card className="flex-none overflow-hidden rounded-md border border-border">
      <div className="flex items-center gap-2.5 bg-muted py-1.5 pl-3 pr-1.5">
        <span className={`h-[7px] w-[7px] flex-none rounded-full ${DOT_STYLES[change]}`} />
        <span className="min-w-0 font-mono text-[12px] font-semibold text-foreground [overflow-wrap:anywhere]">
          {change === "renamed" && fromName ? (
            <>
              {fromName} <span className="text-fg-muted">→</span> {name}
            </>
          ) : (
            name
          )}
        </span>
        <span
          className={`flex-none rounded-sm border px-1.5 py-0.5 font-mono text-[9.5px] font-medium uppercase tracking-wider ${BADGE_STYLES[change]}`}
        >
          {change}
        </span>
        <span className="flex-1" />
        {disabled && discardHint ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="flex-none">{discardButton}</span>
            </TooltipTrigger>
            <TooltipContent side="left">{discardHint}</TooltipContent>
          </Tooltip>
        ) : (
          discardButton
        )}
      </div>
      {note && (
        <div className="flex items-start gap-2 border-t border-border px-3 pb-0.5 pt-1.5 text-[11.5px] text-fg-muted">
          <span className="flex-none">−</span>
          <span>{note}</span>
        </div>
      )}
      {sections
        .filter((s) => s.rows.length > 0)
        .map((sec, si) => (
          <div key={si} className="border-t border-border py-1.5">
            {sec.label && (
              <div className="px-3 pb-0.5 font-mono text-[9px] uppercase tracking-widest text-fg-muted">
                {sec.label}
              </div>
            )}
            {sec.rows.map((row, ri) => (
              <RowLine key={ri} row={row} />
            ))}
          </div>
        ))}
    </div>
  );
}

function GroupLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-2 flex-none px-0.5 font-mono text-[10px] font-medium uppercase tracking-[0.09em] text-fg-muted first:mt-0">
      {children}
    </div>
  );
}

/**
 * Review-and-discard surface for undeployed changes. Renders compact grouped
 * change cards (Resources / Volumes / Connections); each card carries its own
 * Discard wired to the edit session's per-resource/volume revert, and the
 * footer deploys straight from here.
 */
export function ViewChangesModal({
  open,
  onOpenChange,
  diff,
  count,
  errored,
  dirty,
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
      <DialogContent className="gap-0 p-0 sm:max-w-[760px]">
        <DialogHeader className="flex-row items-center gap-2.5 space-y-0 border-b border-border py-3.5 pl-5 pr-12">
          <DialogTitle className="text-[15px]">Undeployed changes</DialogTitle>
          {count > 0 && (
            <span className="rounded-full border border-warn-border bg-warn-bg px-2 py-0.5 font-mono text-[11px] font-bold text-warn">
              {count}
            </span>
          )}
          <span className="flex-1" />
          <DialogDescription className="sr-only">Undeployed changes for {stackName}</DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[56vh] flex-col gap-1.5 overflow-auto px-5 pb-4 pt-3">
          {empty ? (
            <p
              className={`py-8 text-center text-[13px] ${errored ? "text-danger" : "text-fg-muted"}`}
            >
              {errored
                ? "Changes couldn't be saved. Fix the highlighted error and try again."
                : dirty
                  ? "Saving changes… they'll appear here once saved."
                  : diff
                    ? "No pending changes."
                    : "Saving changes… they'll appear here once saved."}
            </p>
          ) : (
            <>
              {diff!.resources.length > 0 && <GroupLabel>Resources</GroupLabel>}
              {diff!.resources.map((d) => (
                <ChangeCard
                  key={`r-${d.name}`}
                  name={d.name}
                  fromName={d.fromName}
                  change={d.change}
                  note={d.note}
                  sections={d.sections.map((s) => ({ label: s.kind, rows: s.rows }))}
                  onDiscard={() => onDiscardResource(d.name)}
                  discardHint="Removed resources restore via Discard all"
                />
              ))}
              {diff!.volumes.length > 0 && <GroupLabel>Volumes</GroupLabel>}
              {diff!.volumes.map((v) => (
                <ChangeCard
                  key={`v-${v.name}`}
                  name={v.name}
                  change={v.change}
                  note={v.note}
                  sections={[{ rows: v.rows }]}
                  onDiscard={() => onDiscardVolume(v.name)}
                  discardHint="Deleted volume data can't be restored. Discard all recreates this volume empty."
                />
              ))}
              {diff!.connections.length > 0 && <GroupLabel>Connections</GroupLabel>}
              {diff!.connections.map((c) => (
                <ChangeCard
                  key={`c-${c.name}`}
                  name={c.name}
                  change={c.change}
                  note={c.note}
                  sections={[{ rows: c.rows }]}
                  discardHint="Manage connections on the canvas"
                />
              ))}
            </>
          )}
        </div>

        <DialogFooter className="flex-row items-center border-t border-border px-5 py-3 sm:justify-start">
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="hover:border-danger-border hover:text-danger"
            onClick={onDiscardAll}
            disabled={empty}
          >
            <RotateCcw className="size-3" />
            Discard all
          </Button>
          <span className="flex-1" />
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
            {deployBusy ? "Deploying" : "Deploy"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

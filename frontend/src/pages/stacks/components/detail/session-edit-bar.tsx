import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { Loader2 } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";

interface SessionEditBarProps {
  dirty: StackDiff;
  onDiscardAll: () => void;
  onDeploy: () => void;
  isDeploying?: boolean;
}

/**
 * Sticky editorial session bar — mono caps per the design system,
 * spans the full width of the main content area (breaks out of page p-8).
 */
export default function SessionEditBar({
  dirty,
  onDiscardAll,
  onDeploy,
  isDeploying = false,
}: SessionEditBarProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [slot, setSlot] = useState<HTMLElement | null>(null);

  useEffect(() => {
    setSlot(document.getElementById("page-sticky-bar"));
  }, []);

  const resourceCount = dirty.dirtyResourceIdx.size;
  const volumeCount = dirty.dirtyVolumeIdx.size;
  const dirtyEntities = resourceCount + volumeCount;
  let envCount = 0;
  for (const v of dirty.perResourceDirty.values()) envCount += v.rowsChanged;
  const addonLinks = dirty.addonLinkCount;

  const handleDiscardAllClick = () => {
    if (dirtyEntities > 1) {
      setConfirmOpen(true);
    } else {
      onDiscardAll();
    }
  };

  type Segment = { num: number; label: string };
  const segments: Segment[] = [];
  if (resourceCount > 0) {
    segments.push({ num: resourceCount, label: resourceCount === 1 ? "RESOURCE" : "RESOURCES" });
  }
  if (volumeCount > 0) {
    segments.push({ num: volumeCount, label: volumeCount === 1 ? "VOLUME" : "VOLUMES" });
  }
  if (envCount > 0) {
    segments.push({ num: envCount, label: "ENV" });
  }
  if (addonLinks > 0) {
    segments.push({ num: addonLinks, label: addonLinks === 1 ? "ADDON LINK" : "ADDON LINKS" });
  }

  const bar = (
    <div
      className="flex h-11 items-center gap-3 bg-secondary px-6 text-foreground border-b border-border"
      style={{ boxShadow: "inset 3px 0 0 var(--brand)" }}
    >
        <span className="h-2 w-2 rounded-full bg-brand animate-pulse" aria-hidden />
        <span className="font-mono text-[11px] font-bold uppercase tracking-[1.5px] text-brand">
          Pending changes
        </span>
        {segments.map((s, i) => (
          <span
            key={i}
            className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground"
          >
            <span className="mx-1.5 text-muted-foreground/60">·</span>
            <span className="font-bold text-foreground">{s.num}</span>{" "}
            <span>{s.label}</span>
          </span>
        ))}
        <div className="grow" />
        <button
          type="button"
          onClick={handleDiscardAllClick}
          disabled={isDeploying}
          className="font-mono text-[11px] font-bold uppercase tracking-[1.5px] text-foreground border border-border bg-transparent px-3 py-1.5 rounded-sm hover:border-foreground disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Discard all
        </button>
        <button
          type="button"
          onClick={onDeploy}
          disabled={isDeploying}
          className="inline-flex items-center gap-1.5 font-mono text-[11px] font-bold uppercase tracking-[1.5px] text-primary-foreground bg-brand hover:bg-brand-hover active:bg-brand-press disabled:opacity-60 disabled:cursor-not-allowed px-3 py-1.5 rounded-sm transition-colors"
        >
          {isDeploying ? (
            <>
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              Deploying
            </>
          ) : (
            <>
              Deploy <span aria-hidden>→</span>
            </>
          )}
        </button>
      </div>
  );

  return (
    <>
      {slot ? createPortal(bar, slot) : null}
      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard all changes?</AlertDialogTitle>
            <AlertDialogDescription>
              You have unsaved edits across {dirtyEntities}{" "}
              {dirtyEntities === 1 ? "item" : "items"}. This will revert every
              change in this session.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep editing</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setConfirmOpen(false);
                onDiscardAll();
              }}
            >
              Discard all
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

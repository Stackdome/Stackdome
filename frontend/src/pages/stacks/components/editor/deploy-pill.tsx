import { useEffect } from "react";
import { CornerDownLeft, Command, FileDiff, Loader2, MoreHorizontal, Rocket, Undo2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export interface DeployPillProps {
  isDraft?: boolean;
  /** Draft gate: at least one resource exists on the canvas. */
  hasResources: boolean;
  dirtyTotal: number;
  isStaged: boolean;
  isActive: boolean;
  deployBusy: boolean;
  draftDeploying?: boolean;
  canWrite: boolean;
  onDeploy: () => void;
  onDraftDeploy?: () => void;
  onViewChanges: () => void;
  canDiscardDraft: boolean;
  onDiscardDraft?: () => void;
}

/**
 * Floating action pill centered at the top of the canvas. It exists only when
 * there is something to act on: pending changes (existing stack), a deploy in
 * flight, or a deployable draft. Deploy on a draft creates the stack and starts
 * the first release in one go — there is no separate create action.
 */
export function DeployPill({
  isDraft,
  hasResources,
  dirtyTotal,
  isStaged,
  isActive,
  deployBusy,
  draftDeploying,
  canWrite,
  onDeploy,
  onDraftDeploy,
  onViewChanges,
  canDiscardDraft,
  onDiscardDraft,
}: DeployPillProps) {
  // Same rule the shell rail used: mid-session dirt or a saved-but-undeployed
  // diff; never for drafts (nothing server-side to review).
  const hasChanges = !isDraft && dirtyTotal > 0 && (isStaged || isActive);
  const visible = isDraft ? hasResources : hasChanges || deployBusy;
  const busy = isDraft ? !!draftDeploying : deployBusy;
  const deployDisabled = isDraft
    ? !!draftDeploying
    : deployBusy || !canWrite || !(isStaged || (isActive && dirtyTotal > 0));
  const fireDeploy = isDraft ? onDraftDeploy : onDeploy;

  useEffect(() => {
    if (!visible || deployDisabled || !fireDeploy) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.defaultPrevented) return; // consumed by a nested layer (dialog, drawer…)
      if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        fireDeploy();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [visible, deployDisabled, fireDeploy]);

  if (!visible) return null;

  const isMac = typeof navigator !== "undefined" && navigator.platform.toUpperCase().includes("MAC");

  return (
    <div
      data-testid="deploy-pill"
      className={cn(
        "absolute left-1/2 top-3 z-30 flex -translate-x-1/2 items-center gap-2.5 rounded-md border border-brand-border bg-background py-2 shadow-[var(--edge)] animate-in fade-in slide-in-from-top-2 duration-[260ms]",
        hasChanges ? "pl-4 pr-2" : "px-2",
      )}
    >
      {hasChanges && (
        <span className="whitespace-nowrap text-[13px] font-medium tracking-[-0.01em] text-brand">
          Apply {dirtyTotal} {dirtyTotal === 1 ? "change" : "changes"}
        </span>
      )}
      {hasChanges && (
        <Button type="button" variant="outline" size="sm" className="rounded-md border-border bg-transparent hover:bg-muted" onClick={onViewChanges}>
          <FileDiff className="size-3.5" />
          Details
        </Button>
      )}
      <Button type="button" variant="default" size="sm" className="rounded-md" onClick={fireDeploy} disabled={deployDisabled}>
        {busy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
        {busy ? "Deploying" : "Deploy"}
        {!busy && (
          <kbd
            aria-hidden
            // currentColor so the chip follows the button's foreground —
            // white in light mode, primary-foreground (near-black) in dark.
            // deliberate off-scale: rounded-sm (9px) overwhelms this ~16px key chip
            className="ml-1 flex items-center gap-px rounded-[4px] border border-current/30 bg-current/10 px-1 py-0.5 text-current"
          >
            {isMac ? <Command className="size-2.5" /> : <span className="font-mono text-[9px] font-semibold leading-none">Ctrl</span>}
            <CornerDownLeft className="size-2.5" />
          </kbd>
        )}
      </Button>
      {!isDraft && canDiscardDraft && onDiscardDraft && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button type="button" variant="ghost" size="icon" className="rounded-md" aria-label="Change actions">
              <MoreHorizontal className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[200px]">
            {/* Deferred so the confirm dialog mounts only after the menu has
                closed and released its body pointer-events lock.
                See https://github.com/radix-ui/primitives/issues/1836 */}
            <DropdownMenuItem onSelect={() => setTimeout(() => onDiscardDraft(), 0)}>
              <Undo2 className="size-4" />
              Discard draft changes
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  );
}

import { useEffect } from "react";
import { FileDiff, Loader2, MoreHorizontal, Rocket, Undo2 } from "lucide-react";
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

  return (
    <div
      data-testid="deploy-pill"
      className="absolute left-1/2 top-3 z-30 flex -translate-x-1/2 items-center gap-2 rounded-lg border border-border bg-background/95 py-1.5 pl-3.5 pr-1.5 shadow-lg backdrop-blur animate-in fade-in slide-in-from-top-2 duration-[260ms]"
    >
      {hasChanges && (
        <span className="text-[13px] font-medium text-warn">
          {dirtyTotal} {dirtyTotal === 1 ? "change" : "changes"}
        </span>
      )}
      {hasChanges && (
        <Button type="button" variant="outline" size="sm" onClick={onViewChanges}>
          <FileDiff className="size-3.5" />
          Details
        </Button>
      )}
      <Button type="button" variant="default" size="sm" onClick={fireDeploy} disabled={deployDisabled}>
        {busy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
        {busy ? "Deploying" : "Deploy"}
        {!busy && <kbd className="ml-1 font-mono text-[10px] opacity-70">⌘⏎</kbd>}
      </Button>
      {!isDraft && canDiscardDraft && onDiscardDraft && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button type="button" variant="ghost" size="icon" aria-label="Change actions">
              <MoreHorizontal className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[200px]">
            <DropdownMenuItem onClick={onDiscardDraft}>
              <Undo2 className="size-4" />
              Discard draft changes
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  );
}

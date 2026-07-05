import { useState, useEffect, useCallback, type ReactNode } from "react";
import { Activity, ChevronDown, ChevronRight, FileDiff, History, LayoutGrid, Loader2, MoreHorizontal, Pencil, Rocket, ScrollText, Trash2, Undo2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { StatusPill } from "@/components/branded";
import { statusVariant } from "@/components/branded/status-variant";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AutosaveStatus } from "./AutosaveStatus";
import type { SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";
import { PublicEndpointRow, type PublicEndpoint } from "./PublicEndpointRow";

const COLLAPSE_KEY_PREFIX = "stackdome.editor-header-collapsed.";
const DRAFT_COLLAPSE_ID = "draft";

/** The four editor modes, in display order. Icons per the design bundle. */
const EDITOR_TABS = [
  { id: "configuration", label: "Configuration", Icon: LayoutGrid },
  { id: "deployments", label: "Deployments", Icon: History },
  { id: "logs", label: "Logs", Icon: ScrollText },
  { id: "metrics", label: "Metrics", Icon: Activity },
] as const;

export interface CanvasEditorShellProps {
  stackName: string;
  /** Persistence key for header collapse; falls back to a shared draft key. */
  stackId?: string;
  /** Raw stack status state (mapped to a pill variant), e.g. "Ready". */
  statusState?: string | null;
  /** Human subtitle, e.g. "3 services · 2 volumes". */
  subtitle: string;
  /** Draft (unsaved) stack — primary action is always Create (nothing exists server-side until it runs). */
  isDraft?: boolean;
  /** Render the title as an editable input (draft only). */
  nameEditable: boolean;
  onNameChange?: (name: string) => void;
  /** Validation error message for the stack name — shown when nameEditable and set. */
  nameError?: string;
  activeTab: string;
  onTabChange: (tab: string) => void;

  // ── dirty / action wiring (all from the existing session + deploy lifecycle) ──
  /** An edit session is open. */
  isActive: boolean;
  /** Count of resources with pending changes — drives the Configuration tab badge. */
  dirtyResourceCount: number;
  /** Total dirty entities (resources + volumes + addon links) — drives "View changes (N)". */
  dirtyTotal: number;
  /** A saved-but-undeployed diff exists (lifecycle.phase === "staged"). */
  isStaged: boolean;
  /** Open the review-and-discard modal for undeployed changes. */
  onViewChanges: () => void;
  /** Autosave status for existing stacks (idle/saving/saved/error). */
  syncStatus: SyncStatus;
  deployBusy: boolean;
  canWrite: boolean;
  /** Draft-mode create action. */
  onDraftDeploy?: () => void;
  draftDeploying?: boolean;
  onDeploy: () => void;
  /** Session-scope discard of server-persisted draft changes. */
  onDiscardDraft?: () => void;
  /** Whether the "Discard draft changes" menu item should appear. */
  canDiscardDraft: boolean;
  onDelete: () => void;
  /** Whether Delete is enabled. */
  canDeleteStack: boolean;

  /** Public endpoints to show in the expanded header (one row of chips). */
  publicEndpoints?: PublicEndpoint[];

  // ── mode bodies (rendered by active tab) ──
  configuration: ReactNode;
  deployments: ReactNode;
  logs: ReactNode;
  metrics: ReactNode;
}

/**
 * Full-bleed editor chrome shown when the canvas flag is on. The title row is
 * identity only (name + single status pill + autosave); all actions live on the
 * tab rail: tabs on the left, then the "View changes" review entry, Deploy, and
 * the actions menu on the right.
 *
 * Presentation only: it owns no stack state. The autosave indicator and Deploy
 * button are wired straight to the caller's session + deploy lifecycle.
 */
export function CanvasEditorShell({
  stackName,
  stackId,
  statusState,
  subtitle,
  isDraft,
  nameEditable,
  onNameChange,
  nameError,
  activeTab,
  onTabChange,
  isActive,
  dirtyResourceCount,
  dirtyTotal,
  isStaged,
  onViewChanges,
  syncStatus,
  deployBusy,
  canWrite,
  onDraftDeploy,
  draftDeploying,
  onDeploy,
  onDiscardDraft,
  canDiscardDraft,
  onDelete,
  canDeleteStack,
  publicEndpoints,
  configuration,
  deployments,
  logs,
  metrics,
}: CanvasEditorShellProps) {
  const collapseKey = `${COLLAPSE_KEY_PREFIX}${stackId ?? DRAFT_COLLAPSE_ID}`;
  const [collapsed, setCollapsed] = useState<boolean>(() => {
    try {
      return localStorage.getItem(collapseKey) === "1";
    } catch {
      return false;
    }
  });
  const toggleCollapsed = useCallback(() => {
    setCollapsed((c) => {
      const next = !c;
      try {
        localStorage.setItem(collapseKey, next ? "1" : "0");
      } catch {
        /* storage unavailable — collapse stays session-local */
      }
      return next;
    });
  }, [collapseKey]);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.defaultPrevented) return; // consumed by a nested layer (dialog, drawer…)
      if (e.key === "." && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggleCollapsed();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggleCollapsed]);

  const hasUnsaved = isActive && dirtyTotal > 0;
  // Undeployed changes exist (either mid-session dirt or a saved-but-undeployed
  // diff). Draft stacks have nothing server-side to review, so never for drafts.
  // A staged phase with a zero count means the changes net out — nothing to
  // review, so the entry hides rather than advertise a phantom change.
  const hasChanges = !isDraft && dirtyTotal > 0 && (isStaged || isActive);

  // The canvas (Configuration) stays mounted so its open drawer + node
  // selection survive tab switches; ops views render as an opaque overlay on
  // top when active.
  const opsBody =
    activeTab === "deployments" ? deployments : activeTab === "logs" ? logs : activeTab === "metrics" ? metrics : null;

  // The primary action is always Deploy. On a draft it creates the stack and
  // starts the first release in one go — nothing exists server-side until then.
  const primaryButton = isDraft ? (
    <Button type="button" variant="default" size="sm" onClick={onDraftDeploy} disabled={draftDeploying}>
      {draftDeploying ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
      {draftDeploying ? "Deploying" : "Deploy"}
    </Button>
  ) : (
    <Button
      type="button"
      variant="default"
      size="sm"
      onClick={onDeploy}
      disabled={deployBusy || !canWrite || !(isStaged || hasUnsaved)}
    >
      {deployBusy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
      {deployBusy ? "Deploying" : "Deploy"}
    </Button>
  );

  // "View changes (N)" — warn-toned entry to the review/discard modal. Shown on
  // the action rail only when there are undeployed changes on an existing stack.
  const viewChanges = hasChanges && (
    <button
      type="button"
      onClick={onViewChanges}
      className="flex flex-none items-center gap-1.5 rounded-md border border-warn-border bg-warn-bg px-2.5 py-1.5 text-[13px] font-medium text-warn transition-colors hover:bg-warn-bg/70"
    >
      <FileDiff className="size-3.5" />
      View changes
      <span className="rounded-full bg-warn px-1.5 py-px font-mono text-[10px] font-bold text-background">{dirtyTotal}</span>
    </button>
  );

  const actionsMenu = !isDraft && (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="ghost" size="icon" aria-label="Stack actions">
          <MoreHorizontal className="size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[180px]">
        {canDiscardDraft && onDiscardDraft && (
          <DropdownMenuItem onClick={onDiscardDraft}>
            <Undo2 className="size-4" />
            Discard draft changes
          </DropdownMenuItem>
        )}
        <DropdownMenuItem
          className="text-danger focus:text-danger"
          onClick={onDelete}
          disabled={!canDeleteStack}
        >
          <Trash2 className="size-4 text-danger" />
          Delete stack
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const chevron = (
    <button
      type="button"
      onClick={toggleCollapsed}
      aria-label={collapsed ? "Expand header" : "Collapse header"}
      title={`${collapsed ? "Expand" : "Collapse"} header (⌘.)`}
      className="flex size-6 flex-none items-center justify-center rounded text-fg-muted hover:bg-muted hover:text-foreground"
    >
      {collapsed ? <ChevronRight className="size-4" /> : <ChevronDown className="size-4" />}
    </button>
  );

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background">
      {collapsed && (
        <div className="flex h-11 flex-none items-center gap-3 border-b border-border px-4 animate-in fade-in slide-in-from-top-1 duration-300">
          <span className="truncate text-[14px] font-medium text-foreground">{stackName}</span>
          {statusState && (
            <span
              aria-label={`status ${statusState}`}
              className={cn(
                "size-2 flex-none rounded-full",
                statusVariant("stack", statusState) === "ready"
                  ? "bg-success"
                  : statusVariant("stack", statusState) === "error"
                    ? "bg-danger"
                    : "bg-warn",
              )}
            />
          )}
          <div className="mx-2 flex items-center gap-1">
            {EDITOR_TABS.map(({ id, label, Icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => onTabChange(id)}
                className={cn(
                  "flex items-center gap-1.5 rounded-md border px-2 py-1 text-[12px] font-medium transition-colors",
                  activeTab === id
                    ? "border-brand bg-brand-bg text-brand"
                    : "border-transparent text-muted-foreground hover:text-foreground",
                )}
              >
                <Icon className="size-3.5" />
                {label}
              </button>
            ))}
          </div>
          <div className="flex-1" />
          {viewChanges}
          {primaryButton}
          {chevron}
        </div>
      )}
      {!collapsed && (
        <>
          {/* Stack-title header — identity only (fade/translate on expand per design sd-fade) */}
          <div className="flex-none px-7 pt-6 animate-in fade-in slide-in-from-top-1 duration-300">
            {/* Chevron sits at the row's right so the title stays flush-left with
                the subtitle + endpoints below it (no collapse-toggle indent). */}
            <div className="flex items-center gap-3.5">
              {nameEditable ? (
                // Same type metrics as the post-create h1; the dashed underline +
                // pencil signal that the name is still editable (it freezes at deploy).
                <div className="group flex min-w-0 items-center gap-2.5">
                  <Input
                    aria-label="Stack name"
                    aria-invalid={!!nameError}
                    value={stackName}
                    onChange={(e) => onNameChange?.(e.target.value)}
                    placeholder="name-your-stack"
                    className={cn(
                      "h-auto w-[22ch] rounded-none border-0 border-b border-dashed bg-transparent px-0 text-[29px] font-medium tracking-[-0.02em] shadow-none focus-visible:ring-0 md:text-[29px]",
                      nameError
                        ? "border-danger"
                        : "border-border/60 hover:border-border focus-visible:border-brand",
                    )}
                  />
                  <Pencil className="size-4 flex-none text-muted-foreground/60 transition-opacity group-focus-within:opacity-0" />
                </div>
              ) : (
                <h1 className="truncate text-[29px] font-medium tracking-[-0.02em] text-foreground">{stackName}</h1>
              )}
              {statusState && (
                <StatusPill variant={statusVariant("stack", statusState)} className="flex-none">
                  {statusState}
                </StatusPill>
              )}
              <div className="flex-1" />
              {!isDraft && <AutosaveStatus status={syncStatus} />}
              {chevron}
            </div>
            {nameEditable && nameError && (
              <p className="mt-1 text-[12px] text-danger">{nameError}</p>
            )}
            <p className="mt-[7px] text-[13px] text-muted-foreground">{subtitle}</p>
            <PublicEndpointRow endpoints={publicEndpoints ?? []} statusState={statusState} />
          </div>

          {/* Tab + action rail */}
          <div className="flex-none flex items-center gap-2 border-b border-border px-7 py-[18px] animate-in fade-in slide-in-from-top-1 duration-300">
            {EDITOR_TABS.map(({ id, label, Icon }) => {
              const active = activeTab === id;
              return (
                <button
                  key={id}
                  type="button"
                  onClick={() => onTabChange(id)}
                  className={cn(
                    "flex items-center gap-2 rounded-md border px-[15px] py-2 text-sm font-medium transition-colors",
                    active
                      ? "border-brand bg-brand-bg text-brand"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )}
                >
                  <Icon className="size-[15px]" />
                  {label}
                  {id === "configuration" && dirtyResourceCount > 0 && (
                    <span className="ml-0.5 rounded-full bg-brand-bg px-1.5 py-px font-mono text-[9.5px] font-medium text-brand">
                      {dirtyResourceCount}
                    </span>
                  )}
                </button>
              );
            })}
            <div className="flex-1" />
            {viewChanges}
            {primaryButton}
            {actionsMenu}
          </div>
        </>
      )}

      {/* Mode body. The canvas is always mounted (keeps its drawer/selection);
          ops views overlay it. Ops views own their own max-width + padding. */}
      <div className="relative min-h-0 flex-1 overflow-hidden">
        <div className="absolute inset-0">{configuration}</div>
        {opsBody && <div className="absolute inset-0 overflow-auto bg-background">{opsBody}</div>}
      </div>
    </div>
  );
}

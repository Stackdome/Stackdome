import { useState, useEffect, useCallback, useMemo, type ReactNode } from "react";
import { Activity, ChevronDown, ChevronRight, History, LayoutGrid, MoreHorizontal, Pencil, ScrollText, Trash2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { StatusPill } from "@/components/branded";
import { statusVariant } from "@/components/branded/status-variant";
import type { StackLifecycle } from "@/api/stacks";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AutosaveStatus } from "./AutosaveStatus";
import { DeployPill } from "./DeployPill";
import { DrawerInsetContext } from "@/pages/stacks/lib/canvas/drawer-inset";
import type { SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";
import { PublicEndpointRow, type PublicEndpoint } from "./PublicEndpointRow";

const COLLAPSE_KEY_PREFIX = "stackdome.editor-header-collapsed.";
const DRAFT_COLLAPSE_ID = "draft";

/** The four editor modes, in display order. Icons per the design bundle. */
const EDITOR_TABS = [
  { id: "architecture", label: "Architecture", Icon: LayoutGrid },
  { id: "deployments", label: "Deployments", Icon: History },
  { id: "logs", label: "Logs", Icon: ScrollText },
  { id: "metrics", label: "Metrics", Icon: Activity },
] as const;

export interface CanvasEditorShellProps {
  stackName: string;
  /** Persistence key for header collapse; falls back to a shared draft key. */
  stackId?: string;
  /** Health off the current/latest release (ReleaseHealth: "ok" | "progressing" | "degraded" |
   *  "failed"). Undefined → nothing ever deployed → a neutral "Not deployed" pill. */
  headerHealth?: string;
  /** The latest release failed while a different (healthy) release stays live — the
   *  main pill shows that release's health, so this drives a secondary error hint. */
  latestDeployFailed?: boolean;
  /** Stack entity lifecycle — "deleting" overrides health with a pending "Deleting" pill. */
  lifecycle?: StackLifecycle;
  /** Human subtitle, e.g. "3 services · 2 volumes". */
  subtitle: string;
  /** At least one resource exists on the canvas — gates the draft deploy pill. */
  hasResources: boolean;
  /** Draft (unsaved) stack — Deploy creates the stack and starts the first release in one go. */
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
  architecture: ReactNode;
  deployments: ReactNode;
  logs: ReactNode;
  metrics: ReactNode;
}

/**
 * Full-bleed editor chrome shown when the canvas flag is on. The title row is
 * identity only (name + single status pill); deploy actions now live in the
 * floating canvas deploy pill (see DeployPill) rather than the rail. The rail
 * keeps tabs on the left and, on the right, the autosave indicator, the stack
 * ⋮ actions menu, and the collapse chevron.
 *
 * Presentation only: it owns no stack state. The autosave indicator and
 * deploy pill are wired straight to the caller's session + deploy lifecycle.
 */
export function CanvasEditorShell({
  stackName,
  stackId,
  headerHealth,
  latestDeployFailed,
  lifecycle,
  subtitle,
  hasResources,
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
  architecture,
  deployments,
  logs,
  metrics,
}: CanvasEditorShellProps) {
  // Horizontal space (px from the viewport's right edge) claimed by the
  // floating drawer stack; header rows and the canvas shift left by this much
  // so the drawer pushes content instead of covering it.
  const [drawerInset, setDrawerInset] = useState(0);
  const drawerInsetCtx = useMemo(() => ({ setInset: setDrawerInset }), []);

  // Header status pill: "Deleting" overrides everything, "Not deployed" covers a
  // stack that has never completed a release, otherwise health drives the pill.
  const isDeleting = lifecycle === ("deleting" satisfies StackLifecycle);
  const pillLabel = isDeleting ? "Deleting" : headerHealth ?? "Not deployed";
  const pillVariant = isDeleting ? "pending" : headerHealth ? statusVariant("health", headerHealth) : "neutral";

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

  // The canvas (Configuration) stays mounted so its open drawer + node
  // selection survive tab switches; ops views render as an opaque overlay on
  // top when active.
  const opsBody =
    activeTab === "deployments" ? deployments : activeTab === "logs" ? logs : activeTab === "metrics" ? metrics : null;

  const actionsMenu = !isDraft && (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="ghost" size="icon" aria-label="Stack actions">
          <MoreHorizontal className="size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[180px]">
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
        <div
          className="flex h-11 flex-none items-center gap-3 border-b border-border px-4 transition-[margin] duration-[260ms] animate-in fade-in slide-in-from-top-1"
          style={{ marginRight: drawerInset }}
        >
          <span className="truncate text-[14px] font-medium text-foreground">{stackName}</span>
          {pillLabel && (
            <span
              aria-label={`status ${pillLabel}`}
              className={cn(
                "size-2 flex-none rounded-full",
                pillVariant === "ready"
                  ? "bg-success"
                  : pillVariant === "error"
                    ? "bg-danger"
                    : pillVariant === "neutral"
                      ? "bg-fg-muted"
                      : "bg-warn",
              )}
            />
          )}
          {latestDeployFailed && (
            <span
              aria-label="Latest deploy failed"
              title="Latest deploy failed"
              className="size-2 flex-none rounded-full bg-danger"
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
          {chevron}
        </div>
      )}
      {!collapsed && (
        <>
          {/* Stack-title header — identity only (fade/translate on expand per design sd-fade) */}
          <div
            className="flex-none px-7 pt-6 transition-[margin] duration-[260ms] animate-in fade-in slide-in-from-top-1"
            style={{ marginRight: drawerInset }}
          >
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
              {pillLabel && (
                <StatusPill variant={pillVariant} className="flex-none">
                  {pillLabel}
                </StatusPill>
              )}
              {latestDeployFailed && (
                <button
                  type="button"
                  onClick={() => onTabChange("deployments")}
                  className="flex-none"
                  aria-label="Latest deploy failed — view deployments"
                >
                  <StatusPill variant="error" pulse={false} className="cursor-pointer hover:opacity-80">
                    Deploy failed
                  </StatusPill>
                </button>
              )}
              <div className="flex-1" />
            </div>
            {nameEditable && nameError && (
              <p className="mt-1 text-[12px] text-danger">{nameError}</p>
            )}
            <p className="mt-[7px] text-[13px] text-muted-foreground">{subtitle}</p>
            <PublicEndpointRow endpoints={publicEndpoints ?? []} variant={pillVariant} />
          </div>

          {/* Tab + action rail */}
          <div
            className="flex-none flex items-center gap-2 border-b border-border px-7 py-[18px] transition-[margin] duration-[260ms] animate-in fade-in slide-in-from-top-1"
            style={{ marginRight: drawerInset }}
          >
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
                  {id === "architecture" && dirtyResourceCount > 0 && (
                    <span className="ml-0.5 rounded-full bg-brand-bg px-1.5 py-px font-mono text-[9.5px] font-medium text-brand">
                      {dirtyResourceCount}
                    </span>
                  )}
                </button>
              );
            })}
            <div className="flex-1" />
            {!isDraft && <AutosaveStatus status={syncStatus} />}
            {actionsMenu}
            {chevron}
          </div>
        </>
      )}

      {/* Mode body. The canvas is always mounted (keeps its drawer/selection);
          ops views overlay it. Ops views own their own max-width + padding. */}
      <div className="relative min-h-0 flex-1 overflow-hidden">
        <div className="absolute inset-y-0 left-0 transition-[right] duration-[260ms]" style={{ right: drawerInset }}>
          <DrawerInsetContext.Provider value={drawerInsetCtx}>{architecture}</DrawerInsetContext.Provider>
          {activeTab === "architecture" && (
            <DeployPill
              isDraft={isDraft}
              hasResources={hasResources}
              dirtyTotal={dirtyTotal}
              isStaged={isStaged}
              isActive={isActive}
              deployBusy={deployBusy}
              draftDeploying={draftDeploying}
              canWrite={canWrite}
              onDeploy={onDeploy}
              onDraftDeploy={onDraftDeploy}
              onViewChanges={onViewChanges}
              canDiscardDraft={canDiscardDraft}
              onDiscardDraft={onDiscardDraft}
            />
          )}
        </div>
        {opsBody && <div className="absolute inset-0 overflow-auto bg-background">{opsBody}</div>}
      </div>
    </div>
  );
}

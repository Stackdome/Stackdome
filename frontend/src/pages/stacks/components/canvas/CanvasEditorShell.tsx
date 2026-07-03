import { useState, useEffect, useCallback, type ReactNode } from "react";
import { Activity, ChevronDown, ChevronRight, LayoutGrid, Loader2, MoreHorizontal, Rocket, Save, Terminal, Trash2, Undo2, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { StatusPill, variantFromState } from "@/components/branded";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import { AutosaveStatus } from "./AutosaveStatus";
import type { SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";
import { PublicEndpointRow, type PublicEndpoint } from "./PublicEndpointRow";

const COLLAPSE_KEY_PREFIX = "stackdome.editor-header-collapsed.";
const DRAFT_COLLAPSE_ID = "draft";

/** The four editor modes, in display order. Icons per the design bundle. */
const EDITOR_TABS = [
  { id: "configuration", label: "Configuration", Icon: LayoutGrid },
  { id: "deployments", label: "Deployments", Icon: Rocket },
  { id: "logs", label: "Logs", Icon: Terminal },
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
  /** Render the title as an editable input (draft, or a rename-capable stack). */
  nameEditable: boolean;
  onNameChange?: (name: string) => void;
  /** Validation error message for the stack name — shown when nameEditable and set. */
  nameError?: string;
  labels: { key: string; value: string }[];
  labelsEditable: boolean;
  onAddLabel?: (value: string) => void;
  onRemoveLabel?: (index: number) => void;
  activeTab: string;
  onTabChange: (tab: string) => void;

  // ── dirty / action wiring (all from the existing session + deploy lifecycle) ──
  /** An edit session is open. */
  isActive: boolean;
  /** Count of resources with pending changes — drives the Configuration tab badge. */
  dirtyResourceCount: number;
  /** Total dirty entities (resources + volumes + addon links) — drives "N unsaved changes". */
  dirtyTotal: number;
  /** A saved-but-undeployed diff exists (lifecycle.phase === "staged"). */
  isStaged: boolean;
  /** Autosave status for existing stacks (idle/saving/saved/error). */
  syncStatus: SyncStatus;
  deployBusy: boolean;
  canWrite: boolean;
  /** Draft-mode create action. */
  onCreate?: () => void;
  isCreating?: boolean;
  onDeploy: () => void;
  onDiscardAll: () => void;
  /** Session-scope discard of server-persisted draft changes (wired in Task 6). */
  onDiscardDraft?: () => void;
  /** Whether the "Discard draft changes" menu item should appear. */
  canDiscardDraft: boolean;
  onDelete: () => void;
  /** Whether Delete is enabled (false until Task 7 wires it fully, but shell respects the gate). */
  canDeleteStack: boolean;

  /** Public endpoints to show in the expanded header (one pill per service). */
  publicEndpoints?: PublicEndpoint[];

  // ── mode bodies (rendered by active tab) ──
  configuration: ReactNode;
  deployments: ReactNode;
  logs: ReactNode;
  metrics: ReactNode;
}

/**
 * Full-bleed editor chrome shown when the canvas flag is on. Replaces the
 * standard PageHeader + Radix tabs + sticky action bar with the design's
 * compact top bar and icon tab row, and lets the active mode body fill the
 * viewport edge-to-edge.
 *
 * Presentation only: it owns no stack state. The autosave indicator and
 * Deploy button are wired straight to the caller's session + deploy
 * lifecycle — no save/deploy logic lives here.
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
  labels,
  labelsEditable,
  onAddLabel,
  onRemoveLabel,
  activeTab,
  onTabChange,
  isActive,
  dirtyResourceCount,
  dirtyTotal,
  isStaged,
  syncStatus,
  deployBusy,
  canWrite,
  onCreate,
  isCreating,
  onDeploy,
  onDiscardAll,
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
  const [discardOpen, setDiscardOpen] = useState(false);
  const [labelInput, setLabelInput] = useState("");

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
      if (e.key === "." && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggleCollapsed();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggleCollapsed]);

  const hasUnsaved = isActive && dirtyTotal > 0;

  // The canvas (Configuration) stays mounted so its open drawer + node
  // selection survive tab switches; ops views render as an opaque overlay on
  // top when active.
  const opsBody =
    activeTab === "deployments" ? deployments : activeTab === "logs" ? logs : activeTab === "metrics" ? metrics : null;

  // Draft mode keeps ONE explicit action (nothing exists server-side until it
  // runs); existing stacks autosave, so the primary is always Deploy.
  const primaryButton = isDraft ? (
    <Button type="button" variant="default" size="sm" onClick={onCreate} disabled={isCreating}>
      {isCreating ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
      {isCreating ? "Creating" : "Create stack"}
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
        <div className="flex h-11 flex-none items-center gap-3 border-b border-border px-4">
          {chevron}
          <span className="truncate text-[14px] font-medium text-foreground">{stackName}</span>
          {statusState && (
            <span
              aria-label={`status ${statusState}`}
              className={cn(
                "size-2 flex-none rounded-full",
                variantFromState(statusState) === "ready"
                  ? "bg-success"
                  : variantFromState(statusState) === "error"
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
          {hasUnsaved && (
            <span className="font-mono text-[11px] text-brand">
              {dirtyTotal} unsaved {dirtyTotal === 1 ? "change" : "changes"}
            </span>
          )}
          {!isDraft && <AutosaveStatus status={syncStatus} />}
          {primaryButton}
        </div>
      )}
      {!collapsed && (
        <>
          {/* Stack-title header */}
          <div className="flex-none px-7 pt-6">
            <div className="flex items-center gap-3.5">
              {chevron}
              {nameEditable ? (
                <Input
                  aria-label="Stack name"
                  aria-invalid={!!nameError}
                  value={stackName}
                  onChange={(e) => onNameChange?.(e.target.value)}
                  placeholder="name-your-stack"
                  className={cn(
                    "h-auto w-[22ch] bg-transparent px-0 text-[29px] font-medium tracking-[-0.02em] shadow-none focus-visible:ring-0",
                    nameError ? "border border-danger ring-1 ring-danger" : "border-0",
                  )}
                />
              ) : (
                <h1 className="truncate text-[29px] font-medium tracking-[-0.02em] text-foreground">{stackName}</h1>
              )}
              {statusState && (
                <StatusPill variant={variantFromState(statusState)} className="flex-none">
                  {statusState}
                </StatusPill>
              )}
              {!isDraft && isStaged && (
                <StatusPill variant="info" className="flex-none">DRAFT</StatusPill>
              )}
              <div className="flex-1" />
              {!isDraft && <AutosaveStatus status={syncStatus} />}
              {primaryButton}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button type="button" variant="ghost" size="icon" aria-label="Stack actions">
                    <MoreHorizontal className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-[180px]">
                  {hasUnsaved && (
                    <DropdownMenuItem onClick={() => setDiscardOpen(true)}>
                      <Trash2 className="size-4" />
                  Discard all changes
                    </DropdownMenuItem>
                  )}
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
            </div>
            {nameEditable && nameError && (
              <p className="mt-1 text-[12px] text-danger">{nameError}</p>
            )}
            <p className="mt-[7px] text-[13px] text-muted-foreground">{subtitle}</p>
            {(labelsEditable || labels.length > 0) && (
              <div className="mt-2 flex flex-wrap items-center gap-1.5">
                {labels.map((l, i) => (
                  <span key={`${l.value}-${i}`} className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                    {l.value}
                    {labelsEditable && (
                      <button type="button" aria-label={`Remove label ${l.value}`} onClick={() => onRemoveLabel?.(i)} className="rounded-full hover:text-foreground">
                        <X className="size-3" />
                      </button>
                    )}
                  </span>
                ))}
                {labelsEditable && (
                  <Input
                    value={labelInput}
                    onChange={(e) => setLabelInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && labelInput.trim()) {
                        e.preventDefault();
                        onAddLabel?.(labelInput.trim());
                        setLabelInput("");
                      }
                    }}
                    placeholder="add label…"
                    className="h-6 w-[14ch] border-0 bg-transparent px-0 text-[11px] shadow-none focus-visible:ring-0"
                  />
                )}
              </div>
            )}
            <PublicEndpointRow endpoints={publicEndpoints ?? []} />
          </div>

          {/* Tab row */}
          <div className="flex-none flex items-center gap-2 border-b border-border px-7 py-[18px]">
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
          </div>

        </>
      )}

      {/* Mode body. The canvas is always mounted (keeps its drawer/selection);
          ops views overlay it. Ops views own their own max-width + padding. */}
      <div className="relative min-h-0 flex-1 overflow-hidden">
        <div className="absolute inset-0">{configuration}</div>
        {opsBody && <div className="absolute inset-0 overflow-auto bg-background">{opsBody}</div>}
      </div>

      <AlertDialog open={discardOpen} onOpenChange={setDiscardOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard all changes?</AlertDialogTitle>
            <AlertDialogDescription>
              You have unsaved edits across {dirtyTotal} {dirtyTotal === 1 ? "item" : "items"}. This will revert every
              change in this session.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep editing</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setDiscardOpen(false);
                onDiscardAll();
              }}
            >
              Discard all
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

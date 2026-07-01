import { useState, type ReactNode } from "react";
import { Activity, LayoutGrid, Loader2, MoreHorizontal, Pencil, Rocket, Save, Terminal, Trash2, X } from "lucide-react";
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

/** The four editor modes, in display order. Icons per the design bundle. */
const EDITOR_TABS = [
  { id: "configuration", label: "Configuration", Icon: LayoutGrid },
  { id: "deployments", label: "Deployments", Icon: Rocket },
  { id: "logs", label: "Logs", Icon: Terminal },
  { id: "metrics", label: "Metrics", Icon: Activity },
] as const;

export interface CanvasEditorShellProps {
  stackName: string;
  /** Raw stack status state (mapped to a pill variant), e.g. "Ready". */
  statusState?: string | null;
  /** Human subtitle, e.g. "3 services · 2 volumes". */
  subtitle: string;
  /** Draft (unsaved) stack — primary action is always Save (create). */
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
  isSaving: boolean;
  deployBusy: boolean;
  canWrite: boolean;
  onSave: () => void;
  onDeploy: () => void;
  onDiscardAll: () => void;
  onEdit: () => void;
  onDelete: () => void;

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
 * Presentation only: it owns no stack state. "N unsaved changes" and the
 * Save/Deploy button are wired straight to the caller's session + deploy
 * lifecycle — no save/deploy logic lives here.
 */
export function CanvasEditorShell({
  stackName,
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
  isSaving,
  deployBusy,
  canWrite,
  onSave,
  onDeploy,
  onDiscardAll,
  onEdit,
  onDelete,
  configuration,
  deployments,
  logs,
  metrics,
}: CanvasEditorShellProps) {
  const [discardOpen, setDiscardOpen] = useState(false);
  const [labelInput, setLabelInput] = useState("");

  const hasUnsaved = isActive && dirtyTotal > 0;
  const primaryIsSave = isDraft || hasUnsaved;
  const dirtyLabel = dirtyTotal === 1 ? "1 unsaved change" : `${dirtyTotal} unsaved changes`;

  // The canvas (Configuration) stays mounted so its open drawer + node
  // selection survive tab switches; ops views render as an opaque overlay on
  // top when active.
  const opsBody =
    activeTab === "deployments" ? deployments : activeTab === "logs" ? logs : activeTab === "metrics" ? metrics : null;

  // When there are unsaved edits (or in draft mode) the primary action is Save
  // (the draft must be persisted before it can be deployed — the backend keeps
  // save and deploy separate). Otherwise the primary action is Deploy.
  const primaryButton = primaryIsSave ? (
    <Button type="button" variant="default" size="sm" onClick={onSave} disabled={isSaving}>
      {isSaving ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
      {isSaving ? "Saving" : "Save"}
    </Button>
  ) : (
    <Button type="button" variant="default" size="sm" onClick={onDeploy} disabled={deployBusy || !canWrite}>
      {deployBusy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
      {deployBusy ? "Deploying" : "Deploy"}
    </Button>
  );

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background">
      {/* Stack-title header */}
      <div className="flex-none px-7 pt-6">
        <div className="flex items-center gap-3.5">
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
          {isStaged && !hasUnsaved && (
            <span className="flex-none font-mono text-[11.5px] text-muted-foreground">draft saved · undeployed</span>
          )}
          <div className="flex-1" />
          {hasUnsaved && <span className="flex-none font-mono text-[11.5px] text-brand">{dirtyLabel}</span>}
          {primaryButton}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button type="button" variant="ghost" size="icon" aria-label="Stack actions">
                <MoreHorizontal className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[180px]">
              {canWrite && (
                <DropdownMenuItem onClick={onEdit} disabled={isActive}>
                  <Pencil className="size-4" />
                  {isActive ? "Editing" : "Edit"}
                </DropdownMenuItem>
              )}
              {hasUnsaved && (
                <DropdownMenuItem onClick={() => setDiscardOpen(true)}>
                  <Trash2 className="size-4" />
                  Discard all changes
                </DropdownMenuItem>
              )}
              <DropdownMenuItem className="text-danger focus:text-danger" onClick={onDelete}>
                <Trash2 className="size-4 text-danger" />
                Delete
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

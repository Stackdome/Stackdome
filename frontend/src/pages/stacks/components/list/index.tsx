import { Layers, PlusCircle, Loader2, AlertTriangle, Search, Box, GitBranch, Database, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { getStacksByOrg } from "@/api/stacks";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { PageHeader, EmptyState, StatusPill, variantFromState, type StatusVariant } from "@/components/branded";
import { formatDistanceToNow } from "date-fns";
import { DockerComposeImportDropdown } from "@/pages/stacks/components/shared/import-dropdown";
import DockerComposeImportDialog from "@/pages/stacks/components/shared/docker-compose-import-dialog";
import { useDockerComposeImport } from "@/pages/stacks/hooks/use-docker-compose-import";
import type { Stack } from "@/pages/stacks/types";
import { cn } from "@/lib/utils";

type StatusFilter = "all" | "ready" | "pending" | "error";
type SortKey = "updated" | "created" | "name";

const STATUS_FILTERS: { key: StatusFilter; label: string; variant: StatusVariant | "neutral" }[] = [
  { key: "all", label: "All", variant: "neutral" },
  { key: "ready", label: "Ready", variant: "ready" },
  { key: "pending", label: "Pending", variant: "pending" },
  { key: "error", label: "Failed", variant: "error" },
];

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "updated", label: "Recently updated" },
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

function inferStackIcon(stack: Stack) {
  // Pick an icon based on the first resource's source type
  const first = stack.spec?.stack_resources?.[0];
  if (!first) return Layers;
  if (first.build_spec) return GitBranch;
  if (first.image_spec) return Box;
  return Layers;
}

function shortRevision(rev?: string | null) {
  if (!rev) return null;
  // Show first 7 chars (commit-style) or up to 8 if numeric
  const trimmed = rev.replace(/^v/, "");
  return trimmed.length > 8 ? trimmed.slice(0, 8) : trimmed;
}

function bucketStatus(state?: string | null): StatusFilter {
  const v = variantFromState(state);
  if (v === "ready") return "ready";
  if (v === "pending") return "pending";
  if (v === "error") return "error";
  return "all";
}

export default function StacksPage() {
  const { stacks, setStacks } = useStacks();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("updated");
  const navigate = useNavigate();

  const {
    isLoading: isImportLoading,
    error: importError,
    isDialogOpen: isImportDialogOpen,
    openDialog: openImportDialog,
    closeDialog: closeImportDialog,
    handleImport,
    clearError: clearImportError,
  } = useDockerComposeImport();

  useEffect(() => {
    const currentOrgId = getCurrentOrganizationId();

    if (currentOrgId) {
      const fetchStacks = async () => {
        setIsLoading(true);
        setError(null);
        try {
          const data = await getStacksByOrg(currentOrgId);
          setStacks(data.items || []);
        } catch (err) {
          console.error("Failed to fetch stacks:", err);
          setError(getErrorMessage(err));
        }
        setIsLoading(false);
      };

      fetchStacks();
    } else {
      setError("Organization ID not found. Unable to load stacks.");
      setIsLoading(false);
    }
  }, [setStacks]);

  const handleCreateNewStack = () => {
    navigate("/stacks/create");
  };

  // Aggregate counts by status bucket — used for the subtitle and filter pill counts
  const counts = useMemo(() => {
    const c = { all: stacks.length, ready: 0, pending: 0, error: 0 } as Record<StatusFilter, number>;
    for (const s of stacks) {
      const b = bucketStatus(s.status?.state);
      if (b !== "all") c[b]++;
    }
    return c;
  }, [stacks]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = stacks.filter((s) => {
      if (statusFilter !== "all" && bucketStatus(s.status?.state) !== statusFilter) return false;
      if (q && !(s.name?.toLowerCase().includes(q))) return false;
      return true;
    });

    out = [...out].sort((a, b) => {
      if (sortKey === "name") return (a.name || "").localeCompare(b.name || "");
      const aTime = sortKey === "created"
        ? new Date(a.created_at || 0).getTime()
        : new Date(a.updated_at || a.created_at || 0).getTime();
      const bTime = sortKey === "created"
        ? new Date(b.created_at || 0).getTime()
        : new Date(b.updated_at || b.created_at || 0).getTime();
      return bTime - aTime;
    });

    return out;
  }, [stacks, statusFilter, query, sortKey]);

  if (isLoading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading stacks...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-1 flex-col p-4 pt-0 h-full items-center justify-center text-center">
        <AlertTriangle className="h-10 w-10 text-destructive mb-4" />
        <h2 className="text-2xl font-bold mb-2">Error</h2>
        <p className="text-muted-foreground mb-6">{error}</p>
        <Button onClick={() => window.location.reload()}>Try Again</Button>
      </div>
    );
  }

  const subtitleParts: string[] = [`${stacks.length} deployed`];
  if (counts.ready) subtitleParts.push(`${counts.ready} healthy`);
  if (counts.pending) subtitleParts.push(`${counts.pending} pending`);
  if (counts.error) subtitleParts.push(`${counts.error} failing`);

  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "Sort";

  return (
    <TooltipProvider>
      <div className="flex flex-1 flex-col p-8 space-y-6 h-full">
        <PageHeader
          eyebrow="Platform"
          title="Stacks"
          subtitle={subtitleParts.map((p, i) => (
            <span key={i}>
              {i > 0 && <span className="mx-2 text-muted-foreground/50">·</span>}
              {p}
            </span>
          ))}
          actions={
            <>
              <DockerComposeImportDropdown
                onDockerComposeImport={openImportDialog}
                variant="outline"
              />
              <Button onClick={handleCreateNewStack} className="bg-brand text-white hover:bg-brand-darker">
                <PlusCircle className="h-4 w-4" />
                New Stack
              </Button>
            </>
          }
        />

        {stacks.length === 0 ? (
          <EmptyState
            icon={<Layers className="h-8 w-8" />}
            title="No stacks deployed yet"
            description="Deploy your first stack to get started."
            action={
              <Button onClick={handleCreateNewStack} variant="outline">
                <PlusCircle className="h-4 w-4" />
                Create New Stack
              </Button>
            }
          />
        ) : (
          <>
            {/* Filter / sort toolbar */}
            <div className="flex flex-wrap items-center gap-3">
              <div className="relative flex-1 min-w-[220px]">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Filter stacks…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="pl-9 h-9"
                />
              </div>
              <div className="flex items-center gap-1.5">
                {STATUS_FILTERS.map((f) => {
                  const active = statusFilter === f.key;
                  const count = counts[f.key];
                  return (
                    <button
                      key={f.key}
                      type="button"
                      onClick={() => setStatusFilter(f.key)}
                      className={cn(
                        "inline-flex items-center gap-1.5 rounded-md border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.2px] transition-colors",
                        active
                          ? "border-brand-border bg-brand-bg text-brand"
                          : "border-border text-muted-foreground hover:bg-muted/50"
                      )}
                    >
                      <span>{f.label}</span>
                      <span className="tabular-nums opacity-80">{count}</span>
                    </button>
                  );
                })}
              </div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.2px] text-muted-foreground hover:bg-muted/50"
                  >
                    Sort: <span className="text-foreground">{sortLabel}</span>
                    <ChevronDown className="h-3 w-3" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {SORT_OPTIONS.map((o) => (
                    <DropdownMenuItem
                      key={o.key}
                      onClick={() => setSortKey(o.key)}
                      className={cn(sortKey === o.key && "text-brand")}
                    >
                      {o.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {filtered.length === 0 ? (
              <EmptyState
                icon={<Search className="h-8 w-8" />}
                title="No stacks match"
                description="Try a different search or status filter."
              />
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {filtered.map((stack) => {
                  const Icon = inferStackIcon(stack);
                  const variant = variantFromState(stack.status?.state);
                  const resourceCount = stack.spec?.stack_resources?.length || 0;
                  const volumeCount = stack.spec?.volumes?.length || 0;
                  const rev = shortRevision(stack.revision);
                  const observed = shortRevision(stack.status?.observed_revision);
                  const driftedRevision = rev && observed && rev !== observed;
                  const updatedAt = stack.updated_at || stack.created_at;
                  const message = stack.status?.message;

                  return (
                    <Link
                      key={stack.id || stack.name}
                      to={`/stacks/${stack.id}`}
                      className="block group"
                    >
                      <Card className="flex flex-col w-full h-full hover:border-brand-border hover:bg-muted/20 transition-colors duration-150 p-4 gap-3">
                        {/* Header: icon + name + namespace + status pill */}
                        <div className="flex items-start gap-3">
                          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-brand-border bg-brand-bg text-brand">
                            <Icon className="h-4 w-4" />
                          </span>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-2">
                              <span
                                className="truncate font-medium text-base group-hover:text-brand transition-colors"
                                title={stack.name}
                              >
                                {stack.name}
                              </span>
                              {stack.status?.state && (
                                <StatusPill variant={variant} className="shrink-0">
                                  {stack.status.state}
                                </StatusPill>
                              )}
                            </div>
                            {stack.namespace && (
                              <p className="font-mono text-[11px] text-muted-foreground truncate mt-0.5 tracking-[0.3px]">
                                {stack.namespace}
                              </p>
                            )}
                          </div>
                        </div>

                        {/* Metric row: resources / volumes / revision */}
                        <div className="grid grid-cols-3 gap-2 pt-1 border-t border-border/60">
                          <div className="flex flex-col gap-0.5 pt-2">
                            <span className="font-mono tabular-nums text-base font-medium leading-none">
                              {resourceCount}
                            </span>
                            <span className="font-mono text-[10px] uppercase tracking-[1.2px] text-muted-foreground">
                              Resources
                            </span>
                          </div>
                          <div className="flex flex-col gap-0.5 pt-2">
                            <span className="font-mono tabular-nums text-base font-medium leading-none">
                              {volumeCount}
                            </span>
                            <span className="font-mono text-[10px] uppercase tracking-[1.2px] text-muted-foreground">
                              Volumes
                            </span>
                          </div>
                          <div className="flex flex-col gap-0.5 pt-2 min-w-0">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="font-mono tabular-nums text-base font-medium leading-none truncate">
                                  {rev ? `#${rev}` : "—"}
                                </span>
                              </TooltipTrigger>
                              {(rev || observed) && (
                                <TooltipContent>
                                  <div className="font-mono text-xs space-y-1">
                                    {rev && <div>Spec: {rev}</div>}
                                    {observed && <div>Observed: {observed}</div>}
                                  </div>
                                </TooltipContent>
                              )}
                            </Tooltip>
                            <span className="font-mono text-[10px] uppercase tracking-[1.2px] text-muted-foreground">
                              Revision
                            </span>
                          </div>
                        </div>

                        {/* Footer: optional drift hint + relative time */}
                        <div className="flex items-center justify-between gap-2 mt-auto font-mono text-[10px] uppercase tracking-[1.2px] text-muted-foreground">
                          {driftedRevision ? (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="inline-flex items-center gap-1 text-warn">
                                  <Database className="h-3 w-3" />
                                  Rolling out
                                </span>
                              </TooltipTrigger>
                              <TooltipContent>
                                <div className="text-xs max-w-xs">
                                  Spec revision <span className="font-mono">{rev}</span> hasn't been observed yet
                                  {message ? <>: {message}</> : "."}
                                </div>
                              </TooltipContent>
                            </Tooltip>
                          ) : message ? (
                            <span className="truncate normal-case tracking-normal text-muted-foreground/80">
                              {message}
                            </span>
                          ) : (
                            <span />
                          )}
                          <span className="shrink-0 text-right">
                            {updatedAt
                              ? formatDistanceToNow(new Date(updatedAt), { addSuffix: true }).replace(/^about\s/, "")
                              : "—"}
                          </span>
                        </div>
                      </Card>
                    </Link>
                  );
                })}
              </div>
            )}
          </>
        )}

        {/* Docker Compose Import Dialog */}
        <DockerComposeImportDialog
          open={isImportDialogOpen}
          onOpenChange={closeImportDialog}
          onImport={handleImport}
          isLoading={isImportLoading}
          error={importError}
          onClearError={clearImportError}
        />
      </div>
    </TooltipProvider>
  );
}

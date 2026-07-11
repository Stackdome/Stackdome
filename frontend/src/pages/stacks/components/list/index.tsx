import { Layers, PlusCircle, Loader2, AlertTriangle, Search, Box, GitBranch, GitPullRequest, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link, useSearchParams } from "react-router-dom";
import { useCallback, useEffect, useMemo, useState } from "react";
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
import { PageHeader, EmptyState, type StatusVariant } from "@/components/branded";
import { statusVariant } from "@/components/branded/status-variant";
import { formatDistanceToNow } from "date-fns";
import { StackCreateWizard } from "@/pages/stacks/components/wizard/stack-create-wizard";
import { EnableRepoWizard } from "@/pages/previews/components/enable-repo-wizard/enable-repo-wizard";
import { PreviewEnvCard } from "./preview-env-card";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { listAllPreviewConfigs, type StackPreviewConfig } from "@/api/preview-configs";
import type { PreviewPhase } from "@/api/preview-envs";
import type { Stack } from "@/pages/stacks/types";
import { useCurrentUser } from "@/hooks/use-current-user";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { cn } from "@/lib/utils";

type StatusFilter = "all" | "ready" | "pending" | "error";
type SortKey = "updated" | "created" | "name";
type ViewMode = "deployed" | "previews";

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

const VIEW_MODES: { key: ViewMode; label: string }[] = [
  { key: "deployed", label: "Deployed" },
  { key: "previews", label: "Previews" },
];

function inferStackIcon(stack: Stack) {
  // Pick an icon based on the first resource's source type
  const first = stack.spec?.stack_resources?.[0];
  if (!first) return Layers;
  if (first.source?.git) return GitBranch;
  if (first.source?.image) return Box;
  return Layers;
}

function bucketStatus(state?: string | null): StatusFilter {
  const v = statusVariant("stack", state);
  if (v === "ready") return "ready";
  if (v === "pending") return "pending";
  if (v === "error") return "error";
  return "all";
}

function bucketPhase(phase?: PreviewPhase | null): StatusFilter {
  if (phase === "Ready") return "ready";
  if (phase === "Failed") return "error";
  // Provisioning / Deploying / Deleting / not yet reported → in flight
  return "pending";
}

export default function StacksPage() {
  const { stacks, setStacks } = useStacks();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("updated");
  const [wizardOpen, setWizardOpen] = useState(false);
  const [enableRepoOpen, setEnableRepoOpen] = useState(false);
  const [configs, setConfigs] = useState<StackPreviewConfig[]>([]);
  const { canWriteAnyTeam } = useCurrentUser();
  const { defaultTeamName } = useResourceTeams();
  const [searchParams, setSearchParams] = useSearchParams();

  const view: ViewMode = searchParams.get("view") === "previews" ? "previews" : "deployed";
  const repoFilter = searchParams.get("repo");

  const { envs, loading: envsLoading, refresh: refreshEnvs } = usePreviewEnvs();

  const setView = (next: ViewMode) => {
    const params = new URLSearchParams(searchParams);
    if (next === "previews") params.set("view", "previews");
    else {
      params.delete("view");
      params.delete("repo");
    }
    setSearchParams(params, { replace: true });
    setStatusFilter("all");
    setQuery("");
  };

  const setRepoFilter = (configId: string | null) => {
    const params = new URLSearchParams(searchParams);
    if (configId) params.set("repo", configId);
    else params.delete("repo");
    setSearchParams(params, { replace: true });
  };

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

  const refreshConfigs = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName) return;
    try {
      setConfigs(await listAllPreviewConfigs(orgId, defaultTeamName));
    } catch {
      // Non-fatal: previews view degrades to env names without repo labels.
    }
  }, [defaultTeamName]);

  useEffect(() => {
    void refreshConfigs();
  }, [refreshConfigs]);

  const configNameById = useMemo(() => {
    const m = new Map<string, string>();
    for (const c of configs) if (c.id) m.set(c.id, c.name ?? c.id);
    return m;
  }, [configs]);

  // Stacks created by preview environments are shown in the Previews view only.
  const previewStackIds = useMemo(() => {
    const s = new Set<string>();
    for (const e of envs) if (e.stack_id) s.add(e.stack_id);
    return s;
  }, [envs]);

  const deployedStacks = useMemo(
    () => stacks.filter((s) => !s.id || !previewStackIds.has(s.id)),
    [stacks, previewStackIds],
  );

  // Aggregate counts by status bucket — used for the filter pill counts
  const counts = useMemo(() => {
    const c = { all: 0, ready: 0, pending: 0, error: 0 } as Record<StatusFilter, number>;
    if (view === "deployed") {
      c.all = deployedStacks.length;
      for (const s of deployedStacks) {
        const b = bucketStatus(s.status?.state);
        if (b !== "all") c[b]++;
      }
    } else {
      const scoped = repoFilter ? envs.filter((e) => e.config_id === repoFilter) : envs;
      c.all = scoped.length;
      for (const e of scoped) c[bucketPhase(e.status?.phase)]++;
    }
    return c;
  }, [view, deployedStacks, envs, repoFilter]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = deployedStacks.filter((s) => {
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
  }, [deployedStacks, statusFilter, query, sortKey]);

  const filteredEnvs = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = envs.filter((e) => {
      if (repoFilter && e.config_id !== repoFilter) return false;
      if (statusFilter !== "all" && bucketPhase(e.status?.phase) !== statusFilter) return false;
      if (q) {
        const configName = (e.config_id && configNameById.get(e.config_id)) || "";
        const haystack = `pr #${e.pr_number ?? ""} ${e.branch ?? ""} ${configName} ${e.name ?? ""}`.toLowerCase();
        if (!haystack.includes(q)) return false;
      }
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
  }, [envs, repoFilter, statusFilter, query, sortKey, configNameById]);

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

  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "Sort";
  const repoLabel = repoFilter ? configNameById.get(repoFilter) ?? "Repo" : "All repos";
  const showToolbar = view === "deployed" ? deployedStacks.length > 0 : envs.length > 0;

  const viewToggle = (
    <div className="inline-flex items-center rounded-md border border-border p-0.5">
      {VIEW_MODES.map((m) => {
        const active = view === m.key;
        return (
          <button
            key={m.key}
            type="button"
            onClick={() => setView(m.key)}
            className={cn(
              "inline-flex items-center rounded-[5px] px-3 h-7 font-mono text-[11px] uppercase tracking-[1.5px] transition-colors",
              active ? "bg-brand text-primary-foreground" : "text-muted-foreground hover:bg-muted/50",
            )}
          >
            {m.label}
          </button>
        );
      })}
    </div>
  );

  return (
    <TooltipProvider>
      <div className="flex flex-1 flex-col p-8 space-y-6 h-full">
        <PageHeader
          eyebrow="Platform"
          title="Stacks"
          subtitle={
            view === "deployed"
              ? "Provision and manage your application stacks"
              : "Preview environments deployed from pull requests"
          }
          actions={
            view === "deployed" && canWriteAnyTeam ? (
              <Button onClick={() => setWizardOpen(true)}>
                <PlusCircle className="h-4 w-4" />
                New Stack
              </Button>
            ) : undefined
          }
        />

        {/* Filter / sort toolbar */}
        <div className="flex flex-wrap items-center gap-3">
          {viewToggle}
          {showToolbar && (
            <>
              <div className="relative flex-1 min-w-[220px]">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder={view === "deployed" ? "Filter stacks…" : "Filter previews…"}
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
                        "inline-flex items-center gap-1.5 rounded-md border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] transition-colors",
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
              {view === "previews" && configs.length > 0 && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                    >
                      Repo: <span className="normal-case tracking-normal text-foreground">{repoLabel}</span>
                      <ChevronDown className="h-3 w-3" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent
                    align="end"
                    className="min-w-[200px]"
                    onCloseAutoFocus={(e) => e.preventDefault()}
                  >
                    <DropdownMenuItem
                      onClick={() => setRepoFilter(null)}
                      className={cn(
                        "font-mono text-[11px] uppercase tracking-[1.5px]",
                        !repoFilter && "text-brand"
                      )}
                    >
                      All repos
                    </DropdownMenuItem>
                    {configs.map((c) => (
                      <DropdownMenuItem
                        key={c.id}
                        onClick={() => setRepoFilter(c.id ?? null)}
                        className={cn(
                          "font-mono text-[11px]",
                          repoFilter === c.id && "text-brand"
                        )}
                      >
                        {c.name}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                  >
                      Sort: <span className="text-foreground">{sortLabel}</span>
                    <ChevronDown className="h-3 w-3" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="min-w-[200px]"
                  onCloseAutoFocus={(e) => e.preventDefault()}
                >
                  {SORT_OPTIONS.map((o) => (
                    <DropdownMenuItem
                      key={o.key}
                      onClick={() => setSortKey(o.key)}
                      className={cn(
                        "font-mono text-[11px] uppercase tracking-[1.5px]",
                        sortKey === o.key && "text-brand"
                      )}
                    >
                      {o.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </>
          )}
        </div>

        {view === "deployed" ? (
          deployedStacks.length === 0 ? (
            <EmptyState
              icon={<Layers className="h-8 w-8" />}
              title="No stacks deployed yet"
              description="Deploy your first stack to get started."
              action={
                canWriteAnyTeam ? (
                  <Button onClick={() => setWizardOpen(true)}>
                    <PlusCircle className="h-4 w-4" />
                    Create New Stack
                  </Button>
                ) : undefined
              }
            />
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={<Search className="h-8 w-8" />}
              title="No stacks match"
              description="Try a different search or status filter."
            />
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
              {filtered.map((stack) => {
                const Icon = inferStackIcon(stack);
                const variant = statusVariant("stack", stack.status?.state);
                const resourceCount = stack.spec?.stack_resources?.length || 0;
                const volumeCount = stack.spec?.volumes?.length || 0;
                const updatedAt = stack.updated_at || stack.created_at;

                return (
                  <Link
                    key={stack.id || stack.name}
                    to={`/stacks/${stack.id}`}
                    className="block group"
                  >
                    <Card className="flex flex-col w-full h-full min-h-[180px] hover:border-brand-border hover:bg-muted/20 transition-colors duration-150 p-5 gap-5">
                      {/* Header: icon + name + namespace + status pill */}
                      <div className="flex items-start gap-3">
                        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-brand-border bg-brand-bg text-brand">
                          <Icon className="h-[18px] w-[18px]" />
                        </span>
                        <div className="flex-1 min-w-0 flex items-start justify-between gap-2">
                          <span
                            className="font-medium text-base leading-tight group-hover:text-brand transition-colors break-words line-clamp-2 [overflow-wrap:anywhere]"
                            title={stack.name}
                          >
                            {stack.name}
                          </span>
                          {stack.status?.state && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span
                                  className={cn(
                                    "mt-1.5 inline-block h-2 w-2 shrink-0 rounded-full",
                                    variant === "ready" && "bg-success",
                                    variant === "pending" && "bg-warn",
                                    variant === "error" && "bg-danger",
                                    variant === "info" && "bg-info",
                                    variant === "neutral" && "bg-fg-muted",
                                  )}
                                  aria-label={stack.status.state}
                                />
                              </TooltipTrigger>
                              <TooltipContent side="top">
                                <span className="font-mono text-[11px] uppercase tracking-[1.5px]">
                                  {stack.status.state}
                                </span>
                              </TooltipContent>
                            </Tooltip>
                          )}
                        </div>
                      </div>

                      {/* Footer: counts on left, relative time on right */}
                      <div className="flex items-end justify-between gap-2 mt-auto pt-4 border-t border-border/60 font-mono text-[11px] text-muted-foreground whitespace-nowrap">
                        <div className="flex flex-col gap-1 tabular-nums">
                          <span>{resourceCount} {resourceCount === 1 ? "resource" : "resources"}</span>
                          <span>{volumeCount} {volumeCount === 1 ? "volume" : "volumes"}</span>
                        </div>
                        <span className="uppercase tracking-[0.5px] text-right">
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
          )
        ) : envsLoading ? (
          <div className="flex flex-1 items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : configs.length === 0 ? (
          <EmptyState
            icon={<GitPullRequest className="h-8 w-8" />}
            title="Preview every pull request"
            description="Connect GitHub and enable a repository — each pull request can get its own temporary environment with a shareable URL."
            action={
              canWriteAnyTeam ? (
                <Button onClick={() => setEnableRepoOpen(true)}>
                  <PlusCircle className="h-4 w-4" />
                  Enable repository
                </Button>
              ) : undefined
            }
          />
        ) : filteredEnvs.length === 0 ? (
          <EmptyState
            icon={<GitPullRequest className="h-8 w-8" />}
            title="No preview environments"
            description={
              envs.length === 0
                ? "Open a pull request, or create one from a repository configuration."
                : "Try a different search, status, or repo filter."
            }
          />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
            {filteredEnvs.map((env) => (
              <PreviewEnvCard
                key={env.id}
                env={env}
                configName={env.config_id ? configNameById.get(env.config_id) : undefined}
              />
            ))}
          </div>
        )}

        <StackCreateWizard open={wizardOpen} onOpenChange={setWizardOpen} />
        <EnableRepoWizard
          open={enableRepoOpen}
          onOpenChange={setEnableRepoOpen}
          onCreated={() => {
            void refreshConfigs();
            void refreshEnvs();
          }}
        />
      </div>
    </TooltipProvider>
  );
}

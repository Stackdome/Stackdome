import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  AlertTriangle, ArrowDown, ChevronDown, GitPullRequest, Loader2, Plus, Search, Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useToast } from "@/components/ui/use-toast";
import { PageHeader, EmptyState, type StatusVariant } from "@/components/branded";
import { previewStatusVariant, statusVariantLabel } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";
import { getPreviewConfig, type StackPreviewConfig } from "@/api/preview-configs";
import { deletePreviewEnv, type PreviewStack } from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { useCurrentUser } from "@/hooks/use-current-user";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { PreviewEnvCard } from "./components/preview-env-card";
import { ConfigSettingsModal } from "./components/config-settings-modal";
import { NewPreviewEnvModal } from "./components/new-preview-env-modal";
import { SyncEnvDialog } from "./components/sync-env-dialog";

type StatusFilter = "all" | "ready" | "pending" | "error";
type SortKey = "updated" | "created" | "name";

const STATUS_FILTERS: { key: StatusFilter; label: string; variant: StatusVariant | "neutral" }[] = [
  { key: "all", label: "All", variant: "neutral" },
  { key: "ready", label: statusVariantLabel.ready, variant: "ready" },
  { key: "pending", label: statusVariantLabel.pending, variant: "pending" },
  { key: "error", label: statusVariantLabel.error, variant: "error" },
];

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "updated", label: "Recently updated" },
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

function bucketStatus(phase?: string | null): StatusFilter {
  const v = previewStatusVariant(phase);
  if (v === "ready") return "ready";
  if (v === "pending") return "pending";
  if (v === "error") return "error";
  return "all";
}

export default function PreviewConfigDetailPage() {
  const { configId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { defaultTeamName } = useResourceTeams();
  const { canWriteAnyTeam } = useCurrentUser();

  const [config, setConfig] = useState<StackPreviewConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [configError, setConfigError] = useState<string | null>(null);
  const [fetchNonce, setFetchNonce] = useState(0);

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("updated");

  const { envs, loading: envsLoading, error: envsError, refresh } = usePreviewEnvs(configId);
  const [syncing, setSyncing] = useState<PreviewStack | null>(null);
  const [deleting, setDeleting] = useState<PreviewStack | null>(null);

  // Breadcrumb shows the config's name instead of its UUID path segment.
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  useEffect(() => {
    if (!configId) return;
    const path = `/previews/${configId}`;
    setCustomLabel(path, config?.name ?? "Preview configuration");
    setPathLoading(path, configLoading);
  }, [configId, config?.name, configLoading, setCustomLabel, setPathLoading]);

  useEffect(() => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !configId) return;
    let cancelled = false;
    setConfigLoading(true);
    setConfigError(null);
    getPreviewConfig(orgId, defaultTeamName, configId)
      .then((cfg) => {
        if (cancelled) return;
        setConfig(cfg);
      })
      .catch((e) => {
        if (!cancelled) setConfigError(getErrorMessage(e));
      })
      .finally(() => {
        if (!cancelled) setConfigLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [configId, defaultTeamName, fetchNonce]);

  const confirmDeleteEnv = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !deleting?.id) return;
    try {
      await deletePreviewEnv(orgId, defaultTeamName, deleting.id);
      toast({ title: `Deleting PR #${deleting.pr_number} environment` });
      await refresh();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setDeleting(null);
    }
  };

  // Aggregate counts by status bucket — used for the filter pill counts.
  const counts = useMemo(() => {
    const c = { all: 0, ready: 0, pending: 0, error: 0 } as Record<StatusFilter, number>;
    c.all = envs.length;
    for (const e of envs) {
      const b = bucketStatus(e.status?.phase);
      if (b !== "all") c[b]++;
    }
    return c;
  }, [envs]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = envs.filter((e) => {
      if (statusFilter !== "all" && bucketStatus(e.status?.phase) !== statusFilter) return false;
      if (q) {
        const prMatch = (e.pr_number ?? "").toLowerCase().includes(q);
        const branchMatch = (e.branch ?? "").toLowerCase().includes(q);
        const commitMatch = (e.commit ?? "").toLowerCase().startsWith(q);
        if (!prMatch && !branchMatch && !commitMatch) return false;
      }
      return true;
    });

    out = [...out].sort((a, b) => {
      if (sortKey === "name") return (a.branch || "").localeCompare(b.branch || "");
      const aTime = sortKey === "created"
        ? new Date(a.created_at || 0).getTime()
        : new Date(a.updated_at || a.created_at || 0).getTime();
      const bTime = sortKey === "created"
        ? new Date(b.created_at || 0).getTime()
        : new Date(b.updated_at || b.created_at || 0).getTime();
      return bTime - aTime;
    });

    return out;
  }, [envs, statusFilter, query, sortKey]);

  if (configLoading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    );
  }

  if (configError || !config) {
    return (
      <div className="flex flex-1 flex-col p-8 h-full items-center justify-center">
        <EmptyState
          icon={<AlertTriangle className="h-8 w-8" />}
          title="Couldn't load this configuration"
          description={configError ?? undefined}
          action={<Button onClick={() => setFetchNonce((n) => n + 1)}>Retry</Button>}
        />
      </div>
    );
  }

  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "Sort";
  const showToolbar = envs.length > 0;

  return (
    <div className="flex flex-1 flex-col p-8 space-y-6 h-full">
      <PageHeader
        eyebrow={<Link to="/previews">← Previews</Link>}
        title={config.name}
        subtitle={config.git_repository?.repo_url}
        actions={
          canWriteAnyTeam ? (
            <>
              <Button variant="outline" onClick={() => setSettingsOpen(true)}>
                <Settings className="h-4 w-4" />
                Settings
              </Button>
              <Button onClick={() => setCreateOpen(true)}>
                <Plus className="h-4 w-4" />
                New preview environment
              </Button>
            </>
          ) : undefined
        }
      />

      {/* Filter / sort toolbar */}
      <div className="flex flex-wrap items-center gap-3">
        {showToolbar && (
          <>
            <div className="relative w-[300px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search PR #, branch, commit…"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="pl-9 h-9"
              />
            </div>
            <div className="flex items-stretch overflow-hidden rounded-md border border-border">
              {STATUS_FILTERS.map((f, i) => {
                const active = statusFilter === f.key;
                const count = counts[f.key];
                return (
                  <button
                    key={f.key}
                    type="button"
                    onClick={() => setStatusFilter(f.key)}
                    className={cn(
                      "inline-flex items-center gap-1.5 px-3.5 h-9 font-mono text-[11px] uppercase tracking-[1.5px] transition-colors",
                      i > 0 && "border-l border-border",
                      active
                        ? "bg-brand-bg text-brand"
                        : "text-muted-foreground hover:bg-muted/50"
                    )}
                  >
                    <span>{f.label}</span>
                    <span className="tabular-nums opacity-80">{count}</span>
                  </button>
                );
              })}
            </div>
            <div className="flex-1" />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="inline-flex items-center gap-1.5 rounded-md border border-border px-3.5 h-9 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground transition-colors hover:border-brand-border hover:text-foreground"
                >
                  <ArrowDown className="h-3 w-3" />
                  {sortLabel}
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

      {envsError && <p className="text-sm text-destructive">{envsError}</p>}

      {envsLoading ? (
        <p className="text-sm text-muted-foreground">Loading environments…</p>
      ) : envs.length === 0 ? (
        <EmptyState
          icon={<GitPullRequest className="h-8 w-8" />}
          title="No preview environments yet"
          description="Create a preview environment from a pull request branch."
          action={
            canWriteAnyTeam ? (
              <Button onClick={() => setCreateOpen(true)}>
                <Plus className="h-4 w-4" />
                New preview environment
              </Button>
            ) : undefined
          }
        />
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={<Search className="h-8 w-8" />}
          title="No environments match"
          description="Try a different search or status filter."
        />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
          {filtered.map((env) => (
            <PreviewEnvCard
              key={env.id}
              env={env}
              configName={config.name}
              onSync={setSyncing}
              onDelete={setDeleting}
            />
          ))}
        </div>
      )}

      <ConfigSettingsModal
        open={settingsOpen}
        onOpenChange={setSettingsOpen}
        config={config}
        onSaved={setConfig}
        onDeleted={() => navigate("/previews")}
      />
      <NewPreviewEnvModal
        open={createOpen}
        onOpenChange={setCreateOpen}
        config={config}
        onCreated={() => void refresh()}
      />
      <SyncEnvDialog
        env={syncing}
        onOpenChange={(o) => !o && setSyncing(null)}
        onSynced={() => void refresh()}
      />
      <AlertDialog open={deleting != null} onOpenChange={(o) => !o && setDeleting(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete PR #{deleting?.pr_number} environment?</AlertDialogTitle>
            <AlertDialogDescription>
              The environment&apos;s stack and resources are torn down. This cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => void confirmDeleteEnv()}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

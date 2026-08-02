import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  AlertTriangle, ChevronDown, GitPullRequest, Loader2, Plus, Search, Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useConfirm } from "@/components/branded/confirm";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useToast } from "@/components/ui/use-toast";
import { PageHeader, EmptyState } from "@/components/branded";
import { previewStatusVariant, statusVariantLabel } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";
import { getPreviewConfig, type StackPreviewConfig } from "@/api/preview-configs";
import { deletePreviewEnv, type PreviewStack } from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useCurrentUser } from "@/hooks/use-current-user";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { PreviewEnvCard } from "./components/preview-env-card";
import { ConfigSettingsModal } from "./components/config-settings-modal";
import { NewPreviewEnvModal } from "./components/new-preview-env-modal";
import { SyncEnvDialog } from "./components/sync-env-dialog";

type SortKey = "updated" | "created" | "name";

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "updated", label: "Recently updated" },
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

const ALL_STATUSES = "all";

/** Display order for the status filter — the exact words the cards render,
 *  healthiest first. Unknown words sort last, alphabetically. */
const STATUS_WORD_ORDER = [statusVariantLabel.ready, statusVariantLabel.pending, statusVariantLabel.error];

function statusWordRank(word: string): number {
  const i = STATUS_WORD_ORDER.indexOf(word);
  return i === -1 ? STATUS_WORD_ORDER.length : i;
}

/** The status word a card shows for this env phase — filter and cards must agree. */
function cardStatusWord(phase?: string | null): string {
  return statusVariantLabel[previewStatusVariant(phase)];
}

export default function PreviewConfigDetailPage() {
  const { configId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { defaultProjectName } = useResourceProjects();
  const { canWriteAnyProject } = useCurrentUser();

  const [config, setConfig] = useState<StackPreviewConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [configError, setConfigError] = useState<string | null>(null);
  const [fetchNonce, setFetchNonce] = useState(0);

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>(ALL_STATUSES);
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("updated");

  const { envs, loading: envsLoading, error: envsError, refresh } = usePreviewEnvs(configId);
  const [syncing, setSyncing] = useState<PreviewStack | null>(null);
  const confirm = useConfirm();

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
    if (!orgId || !defaultProjectName || !configId) return;
    let cancelled = false;
    setConfigLoading(true);
    setConfigError(null);
    getPreviewConfig(orgId, defaultProjectName, configId)
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
  }, [configId, defaultProjectName, fetchNonce]);

  const requestDeleteEnv = async (env: PreviewStack) => {
    const ok = await confirm({
      title: `Delete PR #${env.pr_number} environment?`,
      description: "The environment's stack and resources are torn down. This cannot be undone.",
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultProjectName || !env.id) return;
    try {
      await deletePreviewEnv(orgId, defaultProjectName, env.id);
      toast({ title: `Deleting PR #${env.pr_number} environment`, variant: "info" });
      await refresh();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  // Every card status word with its count (0 when absent), healthiest first —
  // drives the STATUS filter dropdown. Unknown words from the data still
  // surface, appended after the known set.
  const statusOptions = useMemo(() => {
    const counts = new Map<string, number>();
    for (const e of envs) {
      const word = cardStatusWord(e.status?.phase);
      counts.set(word, (counts.get(word) ?? 0) + 1);
    }
    const words = [...new Set([...STATUS_WORD_ORDER, ...counts.keys()])];
    return words
      .map((word) => ({ word, count: counts.get(word) ?? 0 }))
      .sort((a, b) => statusWordRank(a.word) - statusWordRank(b.word) || a.word.localeCompare(b.word));
  }, [envs]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = envs.filter((e) => {
      if (statusFilter !== ALL_STATUSES && cardStatusWord(e.status?.phase) !== statusFilter) return false;
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
          canWriteAnyProject ? (
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
            <div className="relative w-full min-w-[220px] max-w-[340px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search PR #, branch, commit…"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="pl-9 h-9"
              />
            </div>
            <div className="ml-auto flex items-center gap-2">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                  >
                    Status: <span className="text-foreground">{statusFilter === ALL_STATUSES ? "All" : statusFilter}</span>
                    <ChevronDown className="h-3 w-3 flex-none" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="min-w-[200px]"
                  onCloseAutoFocus={(e) => e.preventDefault()}
                >
                  <DropdownMenuItem
                    onSelect={() => setStatusFilter(ALL_STATUSES)}
                    className={cn(
                      "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                      statusFilter === ALL_STATUSES && "text-brand"
                    )}
                  >
                    <span>All</span>
                    <span className="tabular-nums opacity-80">{envs.length}</span>
                  </DropdownMenuItem>
                  {statusOptions.map((o) => (
                    <DropdownMenuItem
                      key={o.word}
                      onSelect={() => setStatusFilter(o.word)}
                      className={cn(
                        "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                        statusFilter === o.word && "text-brand"
                      )}
                    >
                      <span>{o.word}</span>
                      <span className="tabular-nums opacity-80">{o.count}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                  >
                    Sort: <span className="text-foreground">{sortLabel}</span>
                    <ChevronDown className="h-3 w-3 flex-none" />
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
                      onSelect={() => setSortKey(o.key)}
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
            </div>
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
            canWriteAnyProject ? (
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
        <div
          // auto-fill sizes off the container (not viewport breakpoints), and
          // the 300px cap keeps cards from stretching on in-between widths.
          className="grid gap-4"
          style={{ gridTemplateColumns: "repeat(auto-fill, minmax(260px, 300px))" }}
        >
          {filtered.map((env) => (
            <PreviewEnvCard
              key={env.id}
              env={env}
              onSync={setSyncing}
              onDelete={(env) => void requestDeleteEnv(env)}
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
    </div>
  );
}

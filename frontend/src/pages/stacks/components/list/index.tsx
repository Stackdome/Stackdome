import { Layers, PlusCircle, Loader2, AlertTriangle, Search, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Navigate, useSearchParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { getStacksByOrg, deleteStack } from "@/api/stacks";
import { useToast } from "@/components/ui/use-toast";
import { useConfirm } from "@/components/branded/confirm";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { PageHeader, EmptyState } from "@/components/branded";
import { StackCreateWizard } from "@/pages/stacks/components/wizard/stack-create-wizard";
import type { Stack } from "@/pages/stacks/types";
import { DeployStackCard, headerStatus } from "./stack-card";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { useCurrentUser } from "@/hooks/use-current-user";
import { cn } from "@/lib/utils";

type SortKey = "updated" | "created" | "name";

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "updated", label: "Recently updated" },
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

const ALL_STATUSES = "all";

/** Display order for the status filter — the exact labels headerStatus renders
 *  on cards, healthiest first. Unknown labels sort last, alphabetically. */
const STATUS_LABEL_ORDER = ["ok", "progressing", "degraded", "unavailable", "failed", "Not deployed", "Deleting"];

function statusLabelRank(label: string): number {
  const i = STATUS_LABEL_ORDER.indexOf(label);
  return i === -1 ? STATUS_LABEL_ORDER.length : i;
}

export default function StacksPage() {
  const { stacks, setStacks } = useStacks();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>(ALL_STATUSES);
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("updated");
  const [wizardOpen, setWizardOpen] = useState(false);
  const { canWriteAnyProject } = useCurrentUser();
  const { projectNameById } = useResourceProjects();
  const { toast } = useToast();
  const confirm = useConfirm();
  const [searchParams] = useSearchParams();

  const { envs, loading: envsLoading } = usePreviewEnvs();

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

  // Stacks created by preview environments are shown on the Previews page only.
  const previewStackIds = useMemo(() => {
    const s = new Set<string>();
    for (const e of envs) if (e.stack_id) s.add(e.stack_id);
    return s;
  }, [envs]);

  const deployedStacks = useMemo(
    () => stacks.filter((s) => !s.id || !previewStackIds.has(s.id)),
    [stacks, previewStackIds],
  );

  // Every card status label with its count (0 when absent), healthiest first —
  // drives the STATUS filter dropdown. Unknown labels from the data still
  // surface, appended after the known set.
  const statusOptions = useMemo(() => {
    const counts = new Map<string, number>();
    for (const s of deployedStacks) {
      const label = headerStatus(s).label;
      counts.set(label, (counts.get(label) ?? 0) + 1);
    }
    const labels = [...new Set([...STATUS_LABEL_ORDER, ...counts.keys()])];
    return labels
      .map((label) => ({ label, count: counts.get(label) ?? 0 }))
      .sort((a, b) => statusLabelRank(a.label) - statusLabelRank(b.label) || a.label.localeCompare(b.label));
  }, [deployedStacks]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    let out = deployedStacks.filter((s) => {
      if (statusFilter !== ALL_STATUSES && headerStatus(s).label !== statusFilter) return false;
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

  const requestDelete = async (stack: Stack) => {
    const ok = await confirm({
      title: "Delete stack?",
      description: `This permanently deletes “${stack.name}” and all of its deployed resources. This action cannot be undone.`,
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    const projectName = projectNameById(stack.project_id);
    if (!orgId || !projectName || !stack.id) {
      toast({ title: "Delete failed", description: "The stack's project could not be resolved.", variant: "destructive" });
      return;
    }
    try {
      await deleteStack(orgId, projectName, stack.id);
      setStacks((prev) => prev.filter((s) => s.id !== stack.id));
      toast({ title: "Stack deleted", description: `"${stack.name}" was deleted.`, variant: "success" });
    } catch {
      toast({ title: "Delete failed", description: "The stack could not be deleted.", variant: "destructive" });
    }
  };

  // Old previews-tab links redirect to the dedicated /previews page.
  if (searchParams.get("view") === "previews") {
    return <Navigate to="/previews" replace />;
  }

  // Wait for the preview-env list too: rendering before the exclusion set
  // arrives flashes preview-created stacks in the deployed grid.
  if (isLoading || envsLoading) {
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

  return (
    <div className="flex flex-1 flex-col p-8 space-y-6 h-full">
      <PageHeader
        eyebrow="Platform"
        title="Stacks"
        subtitle="Provision and manage your application stacks"
        actions={
          canWriteAnyProject ? (
            <Button onClick={() => setWizardOpen(true)}>
              <PlusCircle className="h-4 w-4" />
                New Stack
            </Button>
          ) : undefined
        }
      />

      {deployedStacks.length === 0 ? (
        <EmptyState
          icon={<Layers className="h-8 w-8" />}
          title="No stacks deployed yet"
          description="Deploy your first stack to get started."
          action={
            canWriteAnyProject ? (
              <Button onClick={() => setWizardOpen(true)}>
                <PlusCircle className="h-4 w-4" />
                  Create New Stack
              </Button>
            ) : undefined
          }
        />
      ) : (
        <>
          {/* Filter / sort toolbar: capped search left, hug-content controls
              right — triggers size to their value like standard list-filter
              buttons, no reserved dead space. */}
          <div className="flex flex-wrap items-center gap-3">
            <div className="relative w-full min-w-[220px] max-w-[340px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Filter stacks…"
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
                    onClick={() => setStatusFilter(ALL_STATUSES)}
                    className={cn(
                      "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                      statusFilter === ALL_STATUSES && "text-brand"
                    )}
                  >
                    <span>All</span>
                    <span className="tabular-nums opacity-80">{deployedStacks.length}</span>
                  </DropdownMenuItem>
                  {statusOptions.map((o) => (
                    <DropdownMenuItem
                      key={o.label}
                      onClick={() => setStatusFilter(o.label)}
                      className={cn(
                        "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                        statusFilter === o.label && "text-brand"
                      )}
                    >
                      <span>{o.label}</span>
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
            </div>
          </div>

          {filtered.length === 0 ? (
            <EmptyState
              icon={<Search className="h-8 w-8" />}
              title="No stacks match"
              description="Try a different search or status filter."
            />
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
              {filtered.map((stack) => (
                <DeployStackCard
                  key={stack.id || stack.name}
                  stack={stack}
                  onDelete={canWriteAnyProject ? (s) => void requestDelete(s) : undefined}
                />
              ))}
            </div>
          )}
        </>
      )}

      <StackCreateWizard open={wizardOpen} onOpenChange={setWizardOpen} />
    </div>
  );
}

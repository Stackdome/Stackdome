import { Plus, AlertTriangle, Search, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Navigate, useNavigate, useSearchParams } from "react-router-dom";
import { useEffect, useMemo, useRef, useState } from "react";
import { getStacksByOrg, deleteStack } from "@/api/stacks";
import { getOrganization } from "@/api/organizations";
import { buildHelloStackSeed } from "@/pages/stacks/lib/onboarding/hello-stack-seed";
import { startCanvasStage, isTourDone, markTourDone } from "@/pages/stacks/lib/onboarding/tour";
import { WelcomeDialog } from "@/pages/stacks/components/onboarding/welcome-dialog";
import { useToast } from "@/components/ui/use-toast";
import { useConfirm } from "@/components/branded/confirm";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { PageHeader, EmptyState } from "@/components/branded";
import { SearchGlyph, StackArchitectureGlyph } from "@/components/branded/empty-state";
import { ViewToggle, useViewMode } from "@/components/branded/view-toggle";
import type { Stack } from "@/api/stack-types";
import { DeployStackCard, StackCardSkeleton, STACK_CARD_GRID } from "./stack-card";
import { DeployStackRow, StackRowHeader, StackRowSkeleton } from "./stack-row";
import { needsAttention, stackRollupState } from "./status";
import { usePreviewEnvs } from "@/hooks/use-preview-envs";
import { useCurrentUser } from "@/hooks/use-current-user";
import { cn } from "@/lib/utils";
import { NEW_STACK_PATH, STACK_DRAFT_PATH } from "@/pages/stacks/lib/routes";

type SortKey = "attention" | "updated" | "created" | "name";

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "attention", label: "Needs attention" },
  { key: "updated", label: "Recently updated" },
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

const ALL_STATUSES = "all";

/** Display order for the status filter — the exact words StatusText renders,
 *  healthiest first. Unknown words from the data sort last, alphabetically. */
const STATUS_ORDER = ["Healthy", "Deploying", "Degraded", "Unavailable", "Failed", "NotDeployed", "Deleting"];

function statusRank(state: string): number {
  const i = STATUS_ORDER.indexOf(state);
  return i === -1 ? STATUS_ORDER.length : i;
}

/** The rollup states, in the same words the rows show. `NotDeployed` is one
 *  token on the wire and two words on screen; the filter menu shows the latter
 *  and matches on the former. */
function statusLabel(state: string): string {
  return state.replace(/([a-z])([A-Z])/g, "$1 $2");
}

export default function StacksPage() {
  const { stacks, setStacks } = useStacks();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>(ALL_STATUSES);
  const [query, setQuery] = useState("");
  // Failures first by default: on the page this replaces, the one row that
  // needed a human was fourth.
  const [sortKey, setSortKey] = useState<SortKey>("attention");
  const [view, setView] = useViewMode("stacks");
  const { canWriteAnyProject, canWrite } = useCurrentUser();
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

  // The demo exposes a public port, so an org without a domain cannot finish
  // the tour.
  const navigate = useNavigate();
  const tourOffered = useRef(false);
  const [welcomeOpen, setWelcomeOpen] = useState(false);
  useEffect(() => {
    if (isLoading || error || stacks.length > 0 || tourOffered.current) return;
    if (!canWriteAnyProject || isTourDone()) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    tourOffered.current = true;
    getOrganization(orgId)
      .then((org) => {
        if ((org.domains ?? []).length === 0) return;
        setWelcomeOpen(true);
      })
      .catch(() => {
        // No tour on a failed lookup — the normal empty state still shows.
      });
  }, [isLoading, error, stacks, canWriteAnyProject]);

  const acceptTour = () => {
    setWelcomeOpen(false);
    startCanvasStage();
    navigate(STACK_DRAFT_PATH, { state: { seed: buildHelloStackSeed() } });
  };

  const closeTour = () => setWelcomeOpen(false);

  const optOutTour = () => {
    markTourDone();
    setWelcomeOpen(false);
  };

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

  const attentionCount = useMemo(
    () => deployedStacks.filter(needsAttention).length,
    [deployedStacks],
  );

  // Every rollup state with its count (0 when absent), healthiest first — drives
  // the Status filter. Unknown states from the data still surface, appended
  // after the known set.
  const statusOptions = useMemo(() => {
    const counts = new Map<string, number>();
    for (const s of deployedStacks) {
      const state = stackRollupState(s);
      counts.set(state, (counts.get(state) ?? 0) + 1);
    }
    const states = [...new Set([...STATUS_ORDER, ...counts.keys()])];
    return states
      .map((state) => ({ state, count: counts.get(state) ?? 0 }))
      .sort((a, b) => statusRank(a.state) - statusRank(b.state) || a.state.localeCompare(b.state));
  }, [deployedStacks]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const out = deployedStacks.filter((s) => {
      if (statusFilter !== ALL_STATUSES && stackRollupState(s) !== statusFilter) return false;
      if (q && !s.name?.toLowerCase().includes(q)) return false;
      return true;
    });

    const recency = (s: Stack, key: SortKey) =>
      key === "created"
        ? new Date(s.created_at || 0).getTime()
        : new Date(s.updated_at || s.created_at || 0).getTime();

    return [...out].sort((a, b) => {
      if (sortKey === "name") return (a.name || "").localeCompare(b.name || "");
      // Failures first, then most recently touched within each group — a broken
      // stack nobody has poked in a week still outranks a healthy deploy from
      // this morning.
      if (sortKey === "attention") {
        const rank = Number(needsAttention(b)) - Number(needsAttention(a));
        if (rank !== 0) return rank;
        return recency(b, "updated") - recency(a, "updated");
      }
      return recency(b, sortKey) - recency(a, sortKey);
    });
  }, [deployedStacks, statusFilter, query, sortKey]);

  const requestDelete = async (stack: Stack) => {
    // §6a level 3 — a stack has dependents and data, so the gate is a retype.
    // The body says what BREAKS, in plain words, not that it cannot be undone.
    const ok = await confirm({
      title: `Delete ${stack.name}?`,
      description:
        "Every service in this stack stops and its containers, volumes and routes are torn down. Any traffic still pointed at it starts failing immediately.",
      confirmLabel: "Delete stack",
      variant: "destructive",
      gate: { kind: "retype", name: stack.name },
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      toast({ title: "Delete failed", description: "No organization selected.", variant: "destructive" });
      return;
    }
    const projectName = projectNameById(stack.project_id);
    if (!projectName || !stack.id) {
      toast({ title: "Delete failed", description: "The stack's project could not be resolved.", variant: "destructive" });
      return;
    }
    try {
      await deleteStack(orgId, projectName, stack.id);
      setStacks((prev) => prev.filter((s) => s.id !== stack.id));
      toast({ title: "Stack deleted", description: `"${stack.name}" was deleted.`, variant: "success" });
    } catch (err) {
      toast({ title: "Delete failed", description: getErrorMessage(err), variant: "destructive" });
    }
  };

  // Old previews-tab links redirect to the dedicated /previews page.
  if (searchParams.get("view") === "previews") {
    return <Navigate to="/previews" replace />;
  }

  // Wait for the preview-env list too: rendering before the exclusion set
  // arrives flashes preview-created stacks in the deployed grid.
  const loading = isLoading || envsLoading;
  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "Sort";

  // §12a — the section's tools live in the header's second row, not in the page
  // body. Rendered in every state including loading: they do not depend on the
  // data, and the old centred spinner threw the whole layout away and then
  // threw it back.
  const toolbar = error ? undefined : (
    <>
      <div className="relative w-[300px]">
        <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-fg-muted" />
        <Input
          placeholder="Filter stacks…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-[30px]"
          aria-label="Filter stacks"
        />
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          {/* Filters are working controls: `flat`, never a pill (§9). */}
          <Button variant="outline" shape="flat">
            <span className="text-fg-2">Status:</span>{" "}
            <span>{statusFilter === ALL_STATUSES ? "All" : statusLabel(statusFilter)}</span>
            <ChevronDown className="h-3.5 w-3.5 flex-none text-fg-2" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="min-w-[200px]" onCloseAutoFocus={(e) => e.preventDefault()}>
          <DropdownMenuItem
            onSelect={() => setStatusFilter(ALL_STATUSES)}
            className={cn("justify-between text-body", statusFilter === ALL_STATUSES && "font-semibold text-foreground")}
          >
            <span>All</span>
            <span className="tabular-nums text-fg-2">{deployedStacks.length}</span>
          </DropdownMenuItem>
          {statusOptions.map((o) => (
            <DropdownMenuItem
              key={o.state}
              onSelect={() => setStatusFilter(o.state)}
              className={cn("justify-between text-body", statusFilter === o.state && "font-semibold text-foreground")}
            >
              <span>{statusLabel(o.state)}</span>
              <span className="tabular-nums text-fg-2">{o.count}</span>
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" shape="flat">
            <span className="text-fg-2">Sort:</span> <span>{sortLabel}</span>
            <ChevronDown className="h-3.5 w-3.5 flex-none text-fg-2" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="min-w-[200px]" onCloseAutoFocus={(e) => e.preventDefault()}>
          {SORT_OPTIONS.map((o) => (
            <DropdownMenuItem
              key={o.key}
              onSelect={() => setSortKey(o.key)}
              className={cn("text-body", sortKey === o.key && "font-semibold text-foreground")}
            >
              {o.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* §11 — content toolbar, right side, last: after the filters, before
          nothing. */}
      <div className="ml-auto">
        <ViewToggle value={view} onValueChange={setView} />
      </div>
    </>
  );

  return (
    <div className="flex flex-1 flex-col h-full">
      <PageHeader
        // §12a's one fact. It counts what is on screen — the old bar said
        // "8 stacks" while six rendered, because it counted before the
        // preview-created stacks were excluded. Attention is appended because
        // the number that decides whether you keep reading is the second one.
        status={
          !loading && deployedStacks.length > 0 ? (
            <span className="text-name tabular-nums text-fg-muted">
              {deployedStacks.length} {deployedStacks.length === 1 ? "stack" : "stacks"}
              {attentionCount > 0 && ` · ${attentionCount} need${attentionCount === 1 ? "s" : ""} attention`}
            </span>
          ) : undefined
        }
        actions={
          canWriteAnyProject ? (
            <Button onClick={() => navigate(NEW_STACK_PATH)}>
              <Plus />
              New stack
            </Button>
          ) : undefined
        }
        toolbar={toolbar}
      />

      {error ? (
        <div className="flex flex-1 flex-col items-center justify-center py-16 text-center">
          <AlertTriangle className="h-8 w-8 text-danger mb-4" />
          <h2 className="text-head font-semibold mb-2">Stacks could not be loaded</h2>
          <p className="text-fg-2 mb-6">{error}</p>
          <Button onClick={() => window.location.reload()}>Try again</Button>
        </div>
      ) : loading ? (
        /* §15 loading — the header and toolbar are already on screen because
           they do not depend on the data. Six rows at the real pitch means
           nothing moves when it lands. */
        view === "list" ? (
          <>
            <StackRowHeader />
            <StackRowSkeleton />
          </>
        ) : (
          <StackCardSkeleton />
        )
      ) : deployedStacks.length === 0 ? (
        /* First run. This is the only screen in the product that gets to
           define its core noun, so it does — and it earns the decorated
           glyph because it is the first thing a new user ever sees. `gap-6`
           because the bigger illustration wants more air than a 34px lens. */
        <EmptyState
          className="flex-1 gap-6"
          icon={<StackArchitectureGlyph />}
          title="No stacks yet"
          description="A stack is your app and everything it needs to run — services, databases and domains, deployed together from a Git branch."
          action={
            canWriteAnyProject ? (
              <Button onClick={() => navigate(NEW_STACK_PATH)}>
                <Plus />
                New stack
              </Button>
            ) : undefined
          }
        />
      ) : (
        <>
          {filtered.length === 0 ? (
            /* A filter that matched nothing is small and recoverable, so it
               gets a small mark and a way back — never the first-run glyph. */
            <EmptyState
              className="flex-1"
              icon={<SearchGlyph />}
              title="No stacks match"
              description="Try a different search, or clear the filters."
              action={
                /* `outline`, not `secondary`. The board's secondary TONE is a
                   control fill plus a hairline; code's `secondary` variant
                   dropped the hairline on purpose, so it renders as an edgeless
                   blob on the sheet. `outline` is the variant that matches. */
                <Button
                  variant="outline"
                  onClick={() => {
                    setQuery("");
                    setStatusFilter(ALL_STATUSES);
                  }}
                >
                  Clear filters
                </Button>
              }
            />
          ) : view === "list" ? (
            /* Hairlines, not cards: separation is a 1px rule, and the list is
               not boxed either — the rows and the sheet edge are the only
               boundaries there are (§7). */
            <div>
              <StackRowHeader />
              {filtered.map((stack) => (
                <DeployStackRow
                  key={stack.id || stack.name}
                  stack={stack}
                  projectName={projectNameById(stack.project_id)}
                  onDelete={canWrite(stack.project_id ?? "") ? (s) => void requestDelete(s) : undefined}
                />
              ))}
            </div>
          ) : (
            /* Same rows, same filters, same sort — a card view that quietly
               drops a fact is a different page wearing the same name (§11). */
            <div className={STACK_CARD_GRID}>
              {filtered.map((stack) => (
                <DeployStackCard
                  key={stack.id || stack.name}
                  stack={stack}
                  projectName={projectNameById(stack.project_id)}
                  onDelete={canWrite(stack.project_id ?? "") ? (s) => void requestDelete(s) : undefined}
                />
              ))}
            </div>
          )}
        </>
      )}

      <WelcomeDialog
        open={welcomeOpen}
        onTakeTour={acceptTour}
        onClose={closeTour}
        onOptOut={optOutTour}
      />
    </div>
  );
}

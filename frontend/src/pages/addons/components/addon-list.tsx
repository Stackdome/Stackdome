import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Search, ChevronDown, ChevronRight } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EmptyState } from "@/components/branded";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import type { PostgresAddon } from "@/api/addons";
import { StatusPill } from "./status-pill";
import { AddonTypeIcon } from "./addon-type-icon";
import {
  filterAndSortAddons,
  countByBucket,
  type AddonStatusFilter,
  type AddonSortKey,
} from "../lib/addon-list-filter";

const STATUS_FILTERS: { key: AddonStatusFilter; label: string }[] = [
  { key: "all", label: "All" },
  { key: "ready", label: "Ready" },
  { key: "pending", label: "Pending" },
  { key: "error", label: "Failed" },
];
const SORT_OPTIONS: { key: AddonSortKey; label: string }[] = [
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

export function AddonList({
  addons,
  canWrite,
}: {
  addons: PostgresAddon[];
  // Gate per-row mutating actions by the row's project. Show by default when
  // undefined (caller hasn't opted into role-based gating).
  canWrite?: (projectId?: string) => boolean;
}) {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<AddonStatusFilter>("all");
  const [sortKey, setSortKey] = useState<AddonSortKey>("created");

  const counts = useMemo(() => countByBucket(addons), [addons]);
  const rows = useMemo(
    () => filterAndSortAddons(addons, query, status, sortKey),
    [addons, query, status, sortKey],
  );
  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "Sort";

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 min-w-[220px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Filter addons…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex items-center gap-1.5">
          {STATUS_FILTERS.map((f) => {
            const active = status === f.key;
            return (
              <Button
                key={f.key}
                type="button"
                variant={active ? "outline" : "ghost"}
                size="sm"
                onClick={() => setStatus(f.key)}
                className={cn(
                  "gap-1.5",
                  active && "border-border-strong bg-foreground/[0.06] font-medium text-foreground",
                )}
              >
                <span>{f.label}</span>
                <span className="tabular-nums text-fg-2">{counts[f.key]}</span>
              </Button>
            );
          })}
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              <span className="text-fg-2">Sort:</span> <span>{sortLabel}</span>
              <ChevronDown className="h-3.5 w-3.5 flex-none text-fg-2" />
            </Button>
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
                className={cn("text-[13px]", sortKey === o.key && "font-semibold text-foreground")}
              >
                {o.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {rows.length === 0 ? (
        <EmptyState
          icon={<Search className="h-8 w-8" />}
          title="No addons match"
          description="Try a different search or status filter."
        />
      ) : (
        <div className="rounded-md border border-border overflow-hidden divide-y divide-border">
          {rows.map((a) => (
            <Link
              key={a.id || a.name}
              to={`/addons/postgres/${a.id}`}
              // Per-row write capability; reserved for gating any per-row
              // mutating affordances. Defaults to writable when the caller
              // hasn't opted into role-based gating.
              data-can-write={canWrite ? canWrite(a.project_id) : true}
              className="flex items-center justify-between px-4 py-3 hover:bg-muted/40 transition-colors group"
            >
              <div className="flex items-center gap-3 min-w-0">
                <AddonTypeIcon type="postgres" size={20} className="shrink-0" />
                <span className="font-medium break-words">
                  {a.name}
                </span>
                <span className="text-xs text-muted-foreground">postgres</span>
              </div>
              <div className="flex items-center gap-3 text-xs text-muted-foreground whitespace-nowrap">
                <StatusPill state={a.status?.state} />
                <span className="font-mono">PG {a.spec.version.major}</span>
                <span className="font-mono">{a.spec.storage.size ?? "—"}</span>
                <span>
                  {a.spec?.backup ? "backups on" : "backups off"}
                </span>
                <span>
                  {a.created_at
                    ? formatDistanceToNow(new Date(a.created_at), {
                      addSuffix: true,
                    }).replace(/^about\s/, "")
                    : "—"}
                </span>
                <ChevronRight className="h-4 w-4" />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

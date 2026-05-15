import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Search, ChevronDown, ChevronRight } from "lucide-react";
import { Input } from "@/components/ui/input";
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

export function AddonList({ addons }: { addons: PostgresAddon[] }) {
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
            className="pl-9 h-9"
          />
        </div>
        <div className="flex items-center gap-1.5">
          {STATUS_FILTERS.map((f) => {
            const active = status === f.key;
            return (
              <button
                key={f.key}
                type="button"
                onClick={() => setStatus(f.key)}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-md border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] transition-colors",
                  active
                    ? "border-brand-border bg-brand-bg text-brand"
                    : "border-border text-muted-foreground hover:bg-muted/50",
                )}
              >
                <span>{f.label}</span>
                <span className="tabular-nums opacity-80">{counts[f.key]}</span>
              </button>
            );
          })}
        </div>
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
                  sortKey === o.key && "text-brand",
                )}
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
              className="flex items-center justify-between px-4 py-3 hover:bg-muted/40 transition-colors group"
            >
              <div className="flex items-center gap-2 min-w-0">
                <span className="font-medium group-hover:text-brand transition-colors break-words">
                  {a.name}
                </span>
                <span className="text-xs text-muted-foreground">postgres</span>
              </div>
              <div className="flex items-center gap-3 text-xs text-muted-foreground whitespace-nowrap">
                <StatusPill state={a.status?.state} />
                <span className="font-mono">PG {a.spec.version.major}</span>
                <span className="font-mono">{a.spec.storage.size ?? "—"}</span>
                <span>
                  {a.spec?.backup?.enabled ? "backups on" : "backups off"}
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

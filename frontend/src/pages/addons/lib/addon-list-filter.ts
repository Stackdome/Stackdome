import type { PostgresAddon } from "@/api/addons";
import { variantFromState } from "@/components/branded";

export type AddonStatusFilter = "all" | "ready" | "pending" | "error";
export type AddonSortKey = "created" | "name";

export function bucketAddonStatus(state?: string | null): AddonStatusFilter {
  const v = variantFromState(state);
  if (v === "ready") return "ready";
  if (v === "pending") return "pending";
  if (v === "error") return "error";
  return "all";
}

export function filterAndSortAddons(
  addons: PostgresAddon[],
  query: string,
  status: AddonStatusFilter,
  sortKey: AddonSortKey,
): PostgresAddon[] {
  const q = query.trim().toLowerCase();
  const out = addons.filter((a) => {
    if (status !== "all" && bucketAddonStatus(a.status?.state) !== status) return false;
    if (q && !a.name?.toLowerCase().includes(q)) return false;
    return true;
  });
  return [...out].sort((a, b) => {
    if (sortKey === "name") return (a.name || "").localeCompare(b.name || "");
    return new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime();
  });
}

export function countByBucket(
  addons: PostgresAddon[],
): Record<AddonStatusFilter, number> {
  const c: Record<AddonStatusFilter, number> = {
    all: addons.length,
    ready: 0,
    pending: 0,
    error: 0,
  };
  for (const a of addons) {
    const b = bucketAddonStatus(a.status?.state);
    if (b !== "all") c[b]++;
  }
  return c;
}

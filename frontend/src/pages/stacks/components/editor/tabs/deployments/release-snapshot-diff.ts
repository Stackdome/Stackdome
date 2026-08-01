import type { components } from "@/api/types/openapi";
import type { FormEnvRow, FormMountRow } from "@/pages/stacks/lib/connection-mapping";
import { canonicalFromSnapshot } from "@/pages/stacks/lib/stack-model/from-api";
import { diffStacks, type EntityDiff, type FieldChange } from "@/pages/stacks/lib/stack-model/diff";
import { labelForField, type FieldSection } from "@/pages/stacks/lib/stack-model/policy";
import { isStructurallyEmpty } from "@/pages/stacks/lib/stack-model/equal";

export type Snap = components["schemas"]["StackReleaseSnapshot"];

export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
export interface DiffSection { kind: FieldSection; rows: DiffRow[]; }
export interface ResourceDiff { name: string; change: "added" | "removed" | "modified" | "renamed"; sections: DiffSection[]; note?: string; fromName?: string; }
export interface ItemDiff { name: string; change: "added" | "removed" | "modified"; rows: DiffRow[]; note?: string; }
export interface SnapshotDiff { resources: ResourceDiff[]; volumes: ItemDiff[]; }

const REMOVED_RESOURCE_NOTE =
  "Resource removed from this release — workload and config deleted from the stack.";
const REMOVED_VOLUME_NOTE = "Volume removed from this release.";

/** Where an env row's value comes from, in the words the drawer uses. */
function envRowLabel(row: FormEnvRow): string {
  switch (row.from) {
    case "stack": return row.value;
    case "self": return `this resource · ${row.selfOutput}`;
    case "secret": return `secret · ${row.secretKey}`;
    case "addon": return `addon · ${row.credField ?? ""}`;
    case "resource": return `${row.resourceName} · ${row.output}`;
    case "resourceTemplate": return `${row.resourceName} · ${row.template}`;
  }
}

function mountLabel(m: FormMountRow): string {
  const sub = m.source_sub_path ? `/${m.source_sub_path}` : "";
  return `${m.source_volume_name ?? ""}${sub}${m.read_only ? " (read only)" : ""}`;
}

function formatValue(path: string, value: unknown): string | undefined {
  // An absent value has many spellings; none of them is worth a row.
  if (isStructurallyEmpty(value)) return undefined;
  if (path.startsWith("env.")) return envRowLabel(value as FormEnvRow);
  if (path.startsWith("mounts.")) return mountLabel(value as FormMountRow);
  if (path === "ports") {
    const ports = (value as { number?: number }[]).map((p) => p.number).filter((n) => n != null);
    return ports.length ? ports.join(", ") : undefined;
  }
  if (Array.isArray(value)) return value.length ? value.map(String).join(" ") : undefined;
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function toRow(change: FieldChange): DiffRow {
  return {
    key: labelForField(change.path),
    from: formatValue(change.path, change.from),
    to: formatValue(change.path, change.to),
    kind: change.kind,
  };
}

/** A row whose value formats away to nothing on both sides says nothing to a
 *  reader — an empty array replacing an absent one, say. */
function isLegible(row: DiffRow): boolean {
  return row.from !== undefined || row.to !== undefined;
}

function toSections(fields: FieldChange[]): DiffSection[] {
  const bySection = new Map<FieldSection, DiffRow[]>();
  for (const field of fields) {
    const row = toRow(field);
    if (!isLegible(row)) continue;
    const rows = bySection.get(field.section) ?? [];
    rows.push(row);
    bySection.set(field.section, rows);
  }
  const order: FieldSection[] = ["configuration", "deployment", "environment"];
  return order
    .filter((kind) => bySection.get(kind)?.length)
    .map((kind) => ({ kind, rows: bySection.get(kind)! }));
}

function toResourceDiff(entry: EntityDiff): ResourceDiff {
  return {
    name: entry.name,
    fromName: entry.fromName,
    change: entry.change,
    sections: toSections(entry.fields),
    ...(entry.change === "removed" ? { note: REMOVED_RESOURCE_NOTE } : {}),
  };
}

function toItemDiff(entry: EntityDiff): ItemDiff {
  return {
    name: entry.name,
    // Volumes have no rename affordance; a renamed volume is a new one.
    change: entry.change === "renamed" ? "added" : entry.change,
    rows: toSections(entry.fields).flatMap((s) => s.rows),
    ...(entry.change === "removed" ? { note: REMOVED_VOLUME_NOTE } : {}),
  };
}

/**
 * Presentation for the canonical diff: labels, grouping, and value formatting.
 * What counts as a change is decided in the model — this module only decides
 * how to say it.
 */
export function diffSnapshots(prev?: Snap, cur?: Snap): SnapshotDiff {
  // No predecessor — the caller distinguishes "initial release" from "no change".
  if (prev == null) return { resources: [], volumes: [] };
  const diff = diffStacks(canonicalFromSnapshot(prev), canonicalFromSnapshot(cur));
  return {
    resources: diff.resources.map(toResourceDiff),
    volumes: diff.volumes.map(toItemDiff),
  };
}

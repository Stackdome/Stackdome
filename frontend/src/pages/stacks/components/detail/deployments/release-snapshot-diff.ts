import type { components } from "@/api/types/openapi";

export type Snap = components["schemas"]["StackReleaseSnapshot"];
type SnapResource = components["schemas"]["StackResource"];
type SnapVolume = NonNullable<Snap["volumes"]>[number];

export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
export interface DiffSection { kind: "configuration" | "environment"; rows: DiffRow[]; }
export interface ResourceDiff { name: string; change: "added" | "removed" | "modified"; sections: DiffSection[]; note?: string; }
export interface ItemDiff { name: string; change: "added" | "removed" | "modified"; rows: DiffRow[]; note?: string; }
export interface SnapshotDiff { resources: ResourceDiff[]; volumes: ItemDiff[]; connections: ItemDiff[]; }

function resourcesOf(snap: unknown): SnapResource[] {
  const r = (snap as { resources?: SnapResource[] } | null | undefined)?.resources;
  return Array.isArray(r) ? r : [];
}

function configScalars(r: SnapResource): Record<string, string | undefined> {
  return {
    "image": r.image_spec?.image,
    "ports": (r.ports ?? []).map((p) => p.number).join(", ") || undefined,
    "command": (r.execution_config?.command ?? []).join(" ") || undefined,
    "args": (r.execution_config?.args ?? []).join(" ") || undefined,
  };
}

function envMap(r: SnapResource): Record<string, string> {
  const out: Record<string, string> = {};
  for (const e of r.execution_config?.environment_variables ?? []) {
    if (e?.name) out[e.name] = e.value ?? (e.self_output ? "(output)" : "");
  }
  return out;
}

function configRows(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffRow[] {
  const p = prev ? configScalars(prev) : {};
  const c = cur ? configScalars(cur) : {};
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key];
    const to = c[key];
    if (from === to) continue;
    if (from == null) rows.push({ key, to, kind: "added" });
    else if (to == null) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

function envRows(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffRow[] {
  const p = prev ? envMap(prev) : {};
  const c = cur ? envMap(cur) : {};
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key];
    const to = c[key];
    if (from === to) continue;
    if (from === undefined) rows.push({ key, to, kind: "added" });
    else if (to === undefined) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

function sectionsFor(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffSection[] {
  const sections: DiffSection[] = [];
  const cfg = configRows(prev, cur);
  if (cfg.length) sections.push({ kind: "configuration", rows: cfg });
  const env = envRows(prev, cur);
  if (env.length) sections.push({ kind: "environment", rows: env });
  return sections;
}

function diffResources(prev: unknown, cur: unknown): ResourceDiff[] {
  const prevByName = new Map(resourcesOf(prev).map((r) => [r.name ?? "", r]));
  const curByName = new Map(resourcesOf(cur).map((r) => [r.name ?? "", r]));
  const out: ResourceDiff[] = [];
  for (const name of new Set([...prevByName.keys(), ...curByName.keys()])) {
    const p = prevByName.get(name);
    const c = curByName.get(name);
    if (p && !c) { out.push({ name, change: "removed", sections: [], note: "Resource removed from this release — workload and config deleted from the stack." }); continue; }
    if (!p && c) { out.push({ name, change: "added", sections: sectionsFor(undefined, c) }); continue; }
    const sections = sectionsFor(p, c);
    if (sections.length) out.push({ name, change: "modified", sections });
  }
  return out;
}

/** Generic scalar diff used by volumes/connections (resources have their own sectioned diff). */
function scalarRows(p: Record<string, string | undefined>, c: Record<string, string | undefined>): DiffRow[] {
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key];
    const to = c[key];
    if (from === to) continue;
    if (from == null) rows.push({ key, to, kind: "added" });
    else if (to == null) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

/** Diff two name-keyed lists into added/removed/modified ItemDiffs by their scalar fields. */
function diffNamed<T>(prev: T[], cur: T[], nameOf: (t: T) => string, scalars: (t: T) => Record<string, string | undefined>, removedNote: string): ItemDiff[] {
  const p = new Map(prev.map((t) => [nameOf(t), t]));
  const c = new Map(cur.map((t) => [nameOf(t), t]));
  const out: ItemDiff[] = [];
  for (const name of new Set([...p.keys(), ...c.keys()])) {
    const a = p.get(name);
    const b = c.get(name);
    if (a && !b) { out.push({ name, change: "removed", rows: [], note: removedNote }); continue; }
    if (!a && b) { out.push({ name, change: "added", rows: scalarRows({}, scalars(b)) }); continue; }
    const rows = scalarRows(scalars(a as T), scalars(b as T));
    if (rows.length) out.push({ name, change: "modified", rows });
  }
  return out;
}

function volumeScalars(v: SnapVolume): Record<string, string | undefined> {
  return { size: v.spec?.size, storage_class: v.spec?.storage_class, access_mode: v.spec?.access_mode };
}

export function diffSnapshots(prev?: Snap, cur?: Snap): SnapshotDiff {
  if (prev == null) return { resources: [], volumes: [], connections: [] }; // no predecessor — caller distinguishes "initial"
  const resources = diffResources(prev, cur);
  const volumes = diffNamed(prev.volumes ?? [], cur?.volumes ?? [], (v) => v.name ?? "", volumeScalars, "Volume removed from this release.");
  return { resources, volumes, connections: [] };
}

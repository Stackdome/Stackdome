import type { components } from "@/api/types/openapi";
import { deepEqual, pairByFingerprint } from "@/pages/stacks/lib/stack-diff";

export type Snap = components["schemas"]["StackReleaseSnapshot"];
type SnapResource = components["schemas"]["StackResource"];
type SnapVolume = NonNullable<Snap["volumes"]>[number];

export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
export interface DiffSection { kind: "configuration" | "environment"; rows: DiffRow[]; }
export interface ResourceDiff { name: string; change: "added" | "removed" | "modified" | "renamed"; sections: DiffSection[]; note?: string; fromName?: string; }
export interface ItemDiff { name: string; change: "added" | "removed" | "modified"; rows: DiffRow[]; note?: string; }
export interface SnapshotDiff { resources: ResourceDiff[]; volumes: ItemDiff[]; connections: ItemDiff[]; }

function resourcesOf(snap: unknown): SnapResource[] {
  const r = (snap as { resources?: SnapResource[] } | null | undefined)?.resources;
  return Array.isArray(r) ? r : [];
}

function configScalars(r: SnapResource): Record<string, string | undefined> {
  const git = r.source?.git;
  return {
    "image": r.source?.image?.ref,
    "repo": git?.repo_url,
    "branch": git?.branch,
    "tag": git?.tag,
    "commit": git?.commit,
    "dockerfile": git?.dockerfile_path,
    "build context": git?.build_context,
    "ports": (r.ports ?? []).map((p) => p.number).join(", ") || undefined,
    "command": (r.execution_config?.command ?? []).join(" ") || undefined,
    "args": (r.execution_config?.args ?? []).join(" ") || undefined,
  };
}

/** Revision keys the deploy-time pin resolver writes into snapshots. */
const REVISION_KEYS = ["branch", "tag", "commit"] as const;
type RevisionKey = (typeof REVISION_KEYS)[number];

/** Revision keys the resource's git source leaves unpinned. The resolver fills
 *  each unpinned key into the snapshot at deploy time, so a value present only
 *  on the baseline for such a key is a resolver fact, not drift — a spec
 *  tracking a branch must not read as dirty from the resolved commit. */
export function unpinnedRevisionKeys(r: SnapResource | undefined): RevisionKey[] {
  const git = r?.source?.git;
  if (!git) return [];
  return REVISION_KEYS.filter((k) => !git[k]);
}

/** Server-written fields — never user intent, excluded from the catch-all. */
const RESOURCE_SERVER_FIELDS = ["id", "stack_id", "revision", "outputs"] as const;
const VOLUME_SERVER_FIELDS = ["id", "project_id", "status"] as const;

const GENERIC_ROW: DiffRow = { key: "other configuration", kind: "changed" };

function canonicalResource(r: SnapResource, dropRevisions: RevisionKey[]): unknown {
  const out = { ...r } as Record<string, unknown>;
  for (const k of RESOURCE_SERVER_FIELDS) delete out[k];
  if (dropRevisions.length && r.source?.git) {
    const git = { ...r.source.git } as Record<string, unknown>;
    for (const k of dropRevisions) delete git[k];
    out.source = { ...r.source, git };
  }
  const env = r.execution_config?.environment_variables;
  if (env) {
    out.execution_config = {
      ...r.execution_config,
      environment_variables: [...env].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
    };
  }
  return out;
}

/** Real config drift outside the projected scalar set. Both sides are
 *  canonicalized the same way, so resolver-written revisions on the baseline
 *  of keys the spec leaves unpinned never count. */
function resourceResidual(prev: SnapResource, cur: SnapResource): boolean {
  const drop = unpinnedRevisionKeys(cur);
  return !deepEqual(canonicalResource(prev, drop), canonicalResource(cur, drop));
}

function volumeResidual(prev: SnapVolume, cur: SnapVolume): boolean {
  const canon = (v: SnapVolume) => {
    const out = { ...v } as Record<string, unknown>;
    for (const k of VOLUME_SERVER_FIELDS) delete out[k];
    return out;
  };
  return !deepEqual(canon(prev), canon(cur));
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
  // `cur` is the user's spec side. Each revision key it leaves unpinned is a
  // deploy-time resolver fact on the baseline, not drift — drop it.
  for (const k of unpinnedRevisionKeys(cur)) delete p[k];
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

/** Identity of a resource by its config + env, ignoring the name. A removed and an
 *  added resource with the same fingerprint are the same resource renamed.
 *  Revisions are excluded: the snapshot side carries resolver-written
 *  branch/commit an unpinned spec side lacks, which would break the pairing. */
function resourceFingerprint(r: SnapResource): string {
  const cfg = configScalars(r);
  for (const k of REVISION_KEYS) delete cfg[k];
  return JSON.stringify({ cfg, env: envMap(r) });
}

function diffResources(prev: unknown, cur: unknown): ResourceDiff[] {
  const prevByName = new Map(resourcesOf(prev).map((r) => [r.name ?? "", r]));
  const curByName = new Map(resourcesOf(cur).map((r) => [r.name ?? "", r]));
  const out: ResourceDiff[] = [];
  const removed: SnapResource[] = [];
  const added: SnapResource[] = [];
  for (const name of new Set([...prevByName.keys(), ...curByName.keys()])) {
    const p = prevByName.get(name);
    const c = curByName.get(name);
    if (p && !c) { removed.push(p); continue; }
    if (!p && c) { added.push(c); continue; }
    const sections = sectionsFor(p, c);
    // Catch-all: a change outside the projected scalars must still surface —
    // an invisible real diff is how the phantom-pill class of bugs starts.
    if (!sections.length && resourceResidual(p!, c!)) {
      sections.push({ kind: "configuration", rows: [GENERIC_ROW] });
    }
    if (sections.length) out.push({ name, change: "modified", sections });
  }
  // Collapse a removed + added pair with identical config into a single rename
  // (the backend reconciles by name, so a rename is a delete + create).
  const pairs = pairByFingerprint(removed, added, resourceFingerprint, resourceFingerprint);
  const renamedRemoved = new Set(pairs.map(([r]) => r));
  const renamedAdded = new Set(pairs.map(([, a]) => a));
  for (const [r, a] of pairs) {
    out.push({ name: a.name ?? "", fromName: r.name ?? "", change: "renamed", sections: [] });
  }
  for (const r of removed) {
    if (renamedRemoved.has(r)) continue;
    out.push({ name: r.name ?? "", change: "removed", sections: sectionsFor(r, undefined), note: "Resource removed from this release — workload and config deleted from the stack." });
  }
  for (const a of added) {
    if (renamedAdded.has(a)) continue;
    out.push({ name: a.name ?? "", change: "added", sections: sectionsFor(undefined, a) });
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

/** Diff two name-keyed lists into added/removed/modified ItemDiffs by their scalar
 *  fields. `residual` (optional) detects real drift outside the scalar set and
 *  degrades it to one generic row instead of silence. */
function diffNamed<T>(prev: T[], cur: T[], nameOf: (t: T) => string, scalars: (t: T) => Record<string, string | undefined>, removedNote: string, residual?: (a: T, b: T) => boolean): ItemDiff[] {
  const p = new Map(prev.map((t) => [nameOf(t), t]));
  const c = new Map(cur.map((t) => [nameOf(t), t]));
  const out: ItemDiff[] = [];
  for (const name of new Set([...p.keys(), ...c.keys()])) {
    const a = p.get(name);
    const b = c.get(name);
    if (a && !b) { out.push({ name, change: "removed", rows: scalarRows(scalars(a as T), {}), note: removedNote }); continue; }
    if (!a && b) { out.push({ name, change: "added", rows: scalarRows({}, scalars(b)) }); continue; }
    const rows = scalarRows(scalars(a as T), scalars(b as T));
    if (!rows.length && residual?.(a as T, b as T)) rows.push(GENERIC_ROW);
    if (rows.length) out.push({ name, change: "modified", rows });
  }
  return out;
}

function volumeScalars(v: SnapVolume): Record<string, string | undefined> {
  return { size: v.spec?.size, storage_class: v.spec?.storage_class, access_mode: v.spec?.access_mode };
}

type SnapConn = NonNullable<Snap["connections"]>[number];

function nodeLabel(n?: { type?: string; name?: string; id?: string }): string {
  return n?.name ?? n?.id ?? n?.type ?? "?";
}

/** Identity label for a connection: kind + endpoints (used as the diff key). */
function connName(c: SnapConn): string {
  return `${c.kind ?? "?"} · ${nodeLabel(c.from)} → ${nodeLabel(c.to)}`;
}

/** Snapshots store value references (output accessors / templates), never resolved secrets. */
function valueLabel(v?: { output?: string; template?: string }): string {
  return v?.output ?? v?.template ?? "(reference)";
}

function connScalars(c: SnapConn): Record<string, string | undefined> {
  const out: Record<string, string | undefined> = {};
  for (const m of c.mappings ?? []) {
    const key = m.target?.name ?? m.target?.path ?? "(target)";
    out[key] = valueLabel(m.value);
  }
  return out;
}

/** Rewrite resource endpoints of the previous snapshot's connections through the
 *  rename map, so a connection whose only "change" is a renamed owner keys the
 *  same as its current counterpart instead of diffing as a phantom remove + add. */
function remapConnResources(conns: SnapConn[], renames: Map<string, string>): SnapConn[] {
  if (renames.size === 0) return conns;
  const remap = (n: SnapConn["from"]): SnapConn["from"] =>
    n?.type === "stack_resource" && n.name && renames.has(n.name) ? { ...n, name: renames.get(n.name) } : n;
  return conns.map((c) => ({ ...c, from: remap(c.from), to: remap(c.to) }));
}

export function diffSnapshots(prev?: Snap, cur?: Snap): SnapshotDiff {
  if (prev == null) return { resources: [], volumes: [], connections: [] }; // no predecessor — caller distinguishes "initial"
  const resources = diffResources(prev, cur);
  const renames = new Map(
    resources.filter((r) => r.change === "renamed" && r.fromName).map((r) => [r.fromName!, r.name]),
  );
  const volumes = diffNamed(prev.volumes ?? [], cur?.volumes ?? [], (v) => v.name ?? "", volumeScalars, "Volume removed from this release.", volumeResidual);
  const connections = diffNamed(
    remapConnResources(prev.connections ?? [], renames),
    cur?.connections ?? [],
    connName,
    connScalars,
    "Connection removed from this release.",
  );
  return { resources, volumes, connections };
}

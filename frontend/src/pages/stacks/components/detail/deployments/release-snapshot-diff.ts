import type { components } from "@/api/types/openapi";

type SnapResource = components["schemas"]["StackResource"];

export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
export interface DiffSection { kind: "configuration" | "environment"; rows: DiffRow[]; }
export interface ResourceDiff { name: string; change: "added" | "removed" | "modified"; sections: DiffSection[]; note?: string; }

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

export function diffSnapshots(prev: unknown, cur: unknown): ResourceDiff[] {
  if (prev == null) return []; // initial release — nothing to compare
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

import type { CanonicalResource, CanonicalStack, CanonicalVolume } from "./canonical";
import { deepEqual, pairByFingerprint } from "./equal";
import { REVISION_KEYS, sectionForField, type FieldSection } from "./policy";

export type ChangeKind = "added" | "removed" | "changed";
export type EntityChange = "added" | "removed" | "modified" | "renamed";

export interface FieldChange {
  /** Dot-path in canonical shape, e.g. `source.git.branch` or `env.DATABASE_URL`. */
  path: string;
  from?: unknown;
  to?: unknown;
  kind: ChangeKind;
  section: FieldSection;
}

export interface EntityDiff {
  name: string;
  fromName?: string;
  change: EntityChange;
  fields: FieldChange[];
}

export interface StackDiff {
  resources: EntityDiff[];
  volumes: EntityDiff[];
}

export interface DiffOptions {
  /** The baseline is a deployed release, so revisions the current spec leaves
   *  unpinned were written there by the pin resolver rather than by a user. */
  baselineIsRelease?: boolean;
}

export function isEmptyDiff(d: StackDiff): boolean {
  return d.resources.length === 0 && d.volumes.length === 0;
}

/**
 * A resource as comparable leaves. Nested subtrees worth naming individually
 * (the git source, each env var, each mount) get their own path; everything
 * else compares whole, including fields this module has never heard of — a
 * field added to the API surfaces under its own name rather than vanishing.
 */
function flattenResource(r: CanonicalResource): Map<string, unknown> {
  const out = new Map<string, unknown>();
  const { name, env, mounts, source, execution_config, ...rest } = r as CanonicalResource &
    Record<string, unknown>;
  void name;

  for (const [key, value] of Object.entries(rest)) out.set(key, value);

  if (source?.git) {
    for (const [k, v] of Object.entries(source.git)) out.set(`source.git.${k}`, v);
  } else if (source?.image) {
    for (const [k, v] of Object.entries(source.image)) out.set(`source.image.${k}`, v);
  } else if (source) {
    out.set("source", source);
  }

  out.set("execution_config.command", execution_config?.command ?? []);
  out.set("execution_config.args", execution_config?.args ?? []);

  for (const row of env ?? []) {
    const { name: envName, ...value } = row as { name: string } & Record<string, unknown>;
    out.set(`env.${envName}`, value);
  }
  for (const m of mounts ?? []) out.set(`mounts.${m.target_path ?? ""}`, m);

  return out;
}

function flattenVolume(v: CanonicalVolume): Map<string, unknown> {
  const out = new Map<string, unknown>();
  const { name, spec, ...rest } = v as CanonicalVolume & Record<string, unknown>;
  void name;
  for (const [key, value] of Object.entries(rest)) out.set(key, value);
  for (const [k, val] of Object.entries(spec ?? {})) out.set(`spec.${k}`, val);
  return out;
}

/**
 * The pin resolver writes the revision it resolved into a release snapshot, so
 * a key the current spec leaves unpinned carries a deploy-time fact there, not
 * drift: a branch-tracking resource must not read as changed against whatever
 * commit the branch happened to point at.
 *
 * ONLY sound when the baseline is a release snapshot. Against a baseline that
 * holds user intent — the saved spec, the session baseline — the same rule
 * erases the user's own act of clearing a pin, and the edit is never written.
 */
function dropResolvedRevisions(prev: Map<string, unknown>, cur: Map<string, unknown>): void {
  for (const key of REVISION_KEYS) {
    const path = `source.git.${key}`;
    const pinned = cur.get(path);
    if (pinned === undefined || pinned === "") prev.delete(path);
  }
}

function fieldChanges(prev: Map<string, unknown>, cur: Map<string, unknown>): FieldChange[] {
  const changes: FieldChange[] = [];
  for (const path of new Set([...prev.keys(), ...cur.keys()])) {
    const from = prev.get(path);
    const to = cur.get(path);
    if (deepEqual(from, to)) continue;
    const kind: ChangeKind =
      from === undefined ? "added" : to === undefined ? "removed" : "changed";
    changes.push({ path, from, to, kind, section: sectionForField(path) });
  }
  return changes.sort((a, b) => a.path.localeCompare(b.path));
}

function resourceFields(
  prev: CanonicalResource,
  cur: CanonicalResource,
  opts: DiffOptions,
): FieldChange[] {
  const p = flattenResource(prev);
  const c = flattenResource(cur);
  if (opts.baselineIsRelease) dropResolvedRevisions(p, c);
  return fieldChanges(p, c);
}

/** Content identity ignoring the name and any resolver-written revision, so a
 *  rename pairs across the delete + create the backend reconciles it as. */
function resourceFingerprint(r: CanonicalResource): string {
  const flat = flattenResource(r);
  for (const key of REVISION_KEYS) flat.delete(`source.git.${key}`);
  return JSON.stringify([...flat.entries()].sort(([a], [b]) => a.localeCompare(b)));
}

function volumeFingerprint(v: CanonicalVolume): string {
  return JSON.stringify([...flattenVolume(v).entries()].sort(([a], [b]) => a.localeCompare(b)));
}

function diffEntities<T extends { name: string }>(
  prev: T[],
  cur: T[],
  fieldsOf: (a: T, b: T) => FieldChange[],
  addedFieldsOf: (t: T) => FieldChange[],
  removedFieldsOf: (t: T) => FieldChange[],
  fingerprint: (t: T) => string,
): EntityDiff[] {
  const prevByName = new Map(prev.map((t) => [t.name, t]));
  const curByName = new Map(cur.map((t) => [t.name, t]));
  const out: EntityDiff[] = [];
  const removed: T[] = [];
  const added: T[] = [];

  for (const name of new Set([...prevByName.keys(), ...curByName.keys()])) {
    const p = prevByName.get(name);
    const c = curByName.get(name);
    if (p && !c) { removed.push(p); continue; }
    if (!p && c) { added.push(c); continue; }
    const fields = fieldsOf(p!, c!);
    if (fields.length) out.push({ name, change: "modified", fields });
  }

  const pairs = pairByFingerprint(removed, added, fingerprint, fingerprint);
  const renamedFrom = new Set(pairs.map(([r]) => r));
  const renamedTo = new Set(pairs.map(([, a]) => a));
  for (const [from, to] of pairs) {
    out.push({ name: to.name, fromName: from.name, change: "renamed", fields: [] });
  }
  for (const r of removed) {
    if (renamedFrom.has(r)) continue;
    out.push({ name: r.name, change: "removed", fields: removedFieldsOf(r) });
  }
  for (const a of added) {
    if (renamedTo.has(a)) continue;
    out.push({ name: a.name, change: "added", fields: addedFieldsOf(a) });
  }
  return out;
}

function presentFields(flat: Map<string, unknown>, side: "from" | "to"): FieldChange[] {
  const changes: FieldChange[] = [];
  for (const [path, value] of flat) {
    if (value === undefined) continue;
    changes.push({
      path,
      [side]: value,
      kind: side === "to" ? "added" : "removed",
      section: sectionForField(path),
    } as FieldChange);
  }
  return changes.sort((a, b) => a.path.localeCompare(b.path));
}

/**
 * The one diff: autosave, the deploy surfaces, the drawer tints and the release
 * timeline all read their answer from this.
 */
export function diffStacks(
  prev: CanonicalStack,
  cur: CanonicalStack,
  opts: DiffOptions = {},
): StackDiff {
  return {
    resources: diffEntities(
      prev.resources,
      cur.resources,
      (a, b) => resourceFields(a, b, opts),
      (r) => presentFields(flattenResource(r), "to"),
      (r) => presentFields(flattenResource(r), "from"),
      resourceFingerprint,
    ),
    volumes: diffEntities(
      prev.volumes,
      cur.volumes,
      (a, b) => fieldChanges(flattenVolume(a), flattenVolume(b)),
      (v) => presentFields(flattenVolume(v), "to"),
      (v) => presentFields(flattenVolume(v), "from"),
      volumeFingerprint,
    ),
  };
}

/** Fields of one resource, for the drawer's per-field tints. */
export function resourceFieldChanges(
  prev: CanonicalResource | undefined,
  cur: CanonicalResource | undefined,
  opts: DiffOptions = {},
): FieldChange[] {
  if (!prev || !cur) return [];
  return resourceFields(prev, cur, opts);
}

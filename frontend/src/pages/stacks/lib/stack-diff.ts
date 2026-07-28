import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

/**
 * Pure diff helpers for the stack edit session.
 *
 * Compares a working `draft` against an immutable `baseline` snapshot via a
 * small recursive deep-equality helper (no external dependency). Results are
 * intended for memoization in `useStackEditSession`.
 */

export type ResourceArr = Partial<FormStackResourceData>[];
export type VolumeArr = Partial<FormVolumeExtendedData>[];

export interface PerResourceDirty {
  rowsChanged: number;
  fieldsChanged: number;
}

export interface PerVolumeDirty {
  fieldsChanged: number;
}

export interface StackDiff {
  dirtyResourceIdx: Set<number>;
  dirtyVolumeIdx: Set<number>;
  perResourceDirty: Map<number, PerResourceDirty>;
  perVolumeDirty: Map<number, PerVolumeDirty>;
}

/** Recursive structural equality that treats `undefined` like a missing key. */
/** True when `v` carries no semantic content — undefined/null, empty string,
 *  empty array, or an object whose values are all themselves structurally empty.
 *  Used by deepEqual so that `{cmd: []}` vs `undefined` (a common form/baseline
 *  mismatch produced by clearing a comma-separated field) reads as equal. */
function isStructurallyEmpty(v: unknown): boolean {
  if (v === null || v === undefined) return true;
  if (typeof v === "string") return v === "";
  if (Array.isArray(v)) return v.every(isStructurallyEmpty);
  if (typeof v === "object") {
    return Object.values(v as Record<string, unknown>).every(isStructurallyEmpty);
  }
  return false;
}

export function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (isStructurallyEmpty(a) && isStructurallyEmpty(b)) return true;
  if (a === null || b === null) return a === b;
  if (typeof a !== typeof b) return false;
  if (typeof a !== "object") return false;

  if (Array.isArray(a) || Array.isArray(b)) {
    if (!Array.isArray(a) || !Array.isArray(b)) return false;
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) {
      if (!deepEqual(a[i], b[i])) return false;
    }
    return true;
  }

  const ao = a as Record<string, unknown>;
  const bo = b as Record<string, unknown>;
  // Exclude structurally-empty values (undefined, [], {}) so that server fields
  // like `depends_on: []` compare equal to missing form fields.
  const aKeys = Object.keys(ao).filter((k) => !isStructurallyEmpty(ao[k]));
  const bKeys = Object.keys(bo).filter((k) => !isStructurallyEmpty(bo[k]));
  if (aKeys.length !== bKeys.length) return false;
  for (const k of aKeys) {
    if (!deepEqual(ao[k], bo[k])) return false;
  }
  return true;
}

/** Count top-level field changes between two objects (shallow on keys, deep on values). */
function countChangedFields(
  a: Record<string, unknown> | undefined | null,
  b: Record<string, unknown> | undefined | null,
): number {
  const ao = (a ?? {}) as Record<string, unknown>;
  const bo = (b ?? {}) as Record<string, unknown>;
  const keys = new Set<string>([...Object.keys(ao), ...Object.keys(bo)]);
  let n = 0;
  for (const k of keys) {
    if (k === "status") continue; // server telemetry, never user dirt
    if (!deepEqual(ao[k], bo[k])) n++;
  }
  return n;
}

/**
 * Drop server-only telemetry before any dirt comparison. `status` is written by
 * the cluster, not the user: the baseline may come from a release snapshot whose
 * status was captured at deploy time while the draft carries the live status —
 * that drift must never read as an undeployed change.
 */
/** Server-computed fields carried on form data for display only. `status` is
 *  live telemetry; `outputs` are derived from the spec by the server (e.g. a
 *  port edit renames its outputs), and the session never rewrites the draft's
 *  copy after a save — comparing either as user intent manufactures phantom
 *  dirt that survives a deploy until a full page refresh. */
function omitServerComputed<T>(x: T): T {
  if (!x || typeof x !== "object" || Array.isArray(x)) return x;
  const { status, outputs, ...rest } = x as Record<string, unknown>;
  void status;
  void outputs;
  return rest as T;
}

export function isResourceDirty(
  draftResource: Partial<FormStackResourceData> | undefined,
  baselineResource: Partial<FormStackResourceData> | undefined,
): boolean {
  return !deepEqual(omitServerComputed(draftResource), omitServerComputed(baselineResource));
}

export function isVolumeDirty(
  draftVolume: Partial<FormVolumeExtendedData> | undefined,
  baselineVolume: Partial<FormVolumeExtendedData> | undefined,
): boolean {
  return !deepEqual(omitServerComputed(draftVolume), omitServerComputed(baselineVolume));
}

function getEnvVars(
  r: Partial<FormStackResourceData> | undefined,
): Array<Record<string, unknown>> {
  return ((r?.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >);
}

function countEnvRowsChanged(
  draft: Partial<FormStackResourceData> | undefined,
  base: Partial<FormStackResourceData> | undefined,
): number {
  const d = getEnvVars(draft);
  const b = getEnvVars(base);
  const max = Math.max(d.length, b.length);
  let n = 0;
  for (let i = 0; i < max; i++) {
    if (!deepEqual(d[i], b[i])) n++;
  }
  return n;
}

/**
 * Per-resource and per-volume diff caches, keyed by draft ref. Most
 * resources are reference-stable across keystrokes (only the resource
 * being edited gets a fresh ref), so we can reuse the prior diff result
 * for the others and skip O(R) deepEqual walks per render. The cache is
 * a WeakMap so stale entries are GC'd with the resource clone.
 *
 * Each entry stores both the dirty flag and the per-X stats; on a cache
 * hit we return both without walking.
 */
// Each entry also records the baseline reference it was computed against.
// When the baseline changes (e.g. after a rebase), the cache key is the
// same draft object but the baseline ref differs, so we recompute.
type ResourceDiffEntry =
  | { dirty: false; baseline: unknown }
  | { dirty: true; baseline: unknown; stats: PerResourceDirty };
type VolumeDiffEntry =
  | { dirty: false; baseline: unknown }
  | { dirty: true; baseline: unknown; stats: PerVolumeDirty };
const resourceDiffCache = new WeakMap<object, ResourceDiffEntry>();
const volumeDiffCache = new WeakMap<object, VolumeDiffEntry>();

function diffOneResource(
  d: Partial<FormStackResourceData> | undefined,
  b: Partial<FormStackResourceData> | undefined,
): ResourceDiffEntry {
  if (d && typeof d === "object") {
    const cached = resourceDiffCache.get(d);
    if (cached !== undefined && cached.baseline === b) return cached;
  }
  const dirty = isResourceDirty(d, b);
  const entry: ResourceDiffEntry = dirty
    ? {
      dirty: true,
      baseline: b,
      stats: {
        rowsChanged: countEnvRowsChanged(d, b),
        fieldsChanged: countChangedFields(
            d as Record<string, unknown> | undefined,
            b as Record<string, unknown> | undefined,
        ),
      },
    }
    : { dirty: false, baseline: b };
  if (d && typeof d === "object") resourceDiffCache.set(d, entry);
  return entry;
}

function diffOneVolume(
  d: Partial<FormVolumeExtendedData> | undefined,
  b: Partial<FormVolumeExtendedData> | undefined,
): VolumeDiffEntry {
  if (d && typeof d === "object") {
    const cached = volumeDiffCache.get(d);
    if (cached !== undefined && cached.baseline === b) return cached;
  }
  const dirty = isVolumeDirty(d, b);
  const entry: VolumeDiffEntry = dirty
    ? {
      dirty: true,
      baseline: b,
      stats: {
        fieldsChanged: countChangedFields(
            d as Record<string, unknown> | undefined,
            b as Record<string, unknown> | undefined,
        ),
      },
    }
    : { dirty: false, baseline: b };
  if (d && typeof d === "object") volumeDiffCache.set(d, entry);
  return entry;
}

export function diffStack(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
): StackDiff {
  const dirtyResourceIdx = new Set<number>();
  const perResourceDirty = new Map<number, PerResourceDirty>();
  const maxR = Math.max(draft.resources.length, baseline.resources.length);
  for (let i = 0; i < maxR; i++) {
    const entry = diffOneResource(draft.resources[i], baseline.resources[i]);
    if (!entry.dirty) continue;
    dirtyResourceIdx.add(i);
    perResourceDirty.set(i, entry.stats);
  }

  const dirtyVolumeIdx = new Set<number>();
  const perVolumeDirty = new Map<number, PerVolumeDirty>();
  const maxV = Math.max(draft.volumes.length, baseline.volumes.length);
  for (let i = 0; i < maxV; i++) {
    const entry = diffOneVolume(draft.volumes[i], baseline.volumes[i]);
    if (!entry.dirty) continue;
    dirtyVolumeIdx.add(i);
    perVolumeDirty.set(i, entry.stats);
  }

  return {
    dirtyResourceIdx,
    dirtyVolumeIdx,
    perResourceDirty,
    perVolumeDirty,
  };
}

/** Deep clone via JSON round-trip. Form data is plain JSON so this is safe.
 *  Passes undefined through (JSON.parse(JSON.stringify(undefined)) throws). */
export function cloneJson<T>(v: T): T {
  if (v === undefined) return v;
  return JSON.parse(JSON.stringify(v)) as T;
}

export function revertResource(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, resources: draft.resources.slice() };
  // A name-aligned baseline can carry nullish holes for draft-only resources
  // (see alignBaselineToDraft) — treat those the same as "past the end".
  const baselineEntry = idx < baseline.resources.length ? baseline.resources[idx] : undefined;
  if (baselineEntry != null) {
    // Keep the draft's live status: the baseline's was captured at deploy time
    // and restoring it would show stale telemetry until the next refresh.
    const liveStatus = (draft.resources[idx] as { status?: unknown } | undefined)?.status;
    const restored = {
      ...cloneJson(baselineEntry),
      ...(liveStatus !== undefined ? { status: liveStatus } : {}),
    } as (typeof next.resources)[number];
    // A restored mount may reference a volume the draft has since deleted —
    // reattaching it would create a dangling mount, so drop those rows.
    if (Array.isArray(restored.volume_mounts)) {
      const draftVolumeNames = new Set(draft.volumes.map((v) => v?.name).filter(Boolean));
      restored.volume_mounts = restored.volume_mounts.filter(
        (m) => m?.source_volume_name && draftVolumeNames.has(m.source_volume_name),
      );
    }
    next.resources[idx] = restored;
  } else {
    // The resource only exists in the draft — drop it.
    next.resources.splice(idx, 1);
  }
  return next;
}

/**
 * Reorder a baseline array so it aligns positionally with the draft, matching
 * entries by `name`. All downstream diffing (diffStack, per-drawer baselines)
 * is positional, but the server returns stack_resources in unstable order and
 * a release snapshot's order need not match the live stack's — so the baseline
 * must be re-keyed onto the draft's order before any index-wise comparison.
 *
 * Draft entries with no baseline match get an `undefined` hole (they read as
 * "added"); baseline entries missing from the draft are appended at the end so
 * deletions still register as dirt.
 */
export function alignBaselineToDraft<T extends { name?: string }>(
  baseline: T[],
  draft: Array<{ name?: string }>,
  fingerprint?: (entry: unknown) => string,
): (T | undefined)[] {
  const byName = new Map<string, T>();
  for (const b of baseline) {
    // First occurrence wins on (malformed) duplicate names.
    if (b?.name && !byName.has(b.name)) byName.set(b.name, b);
  }
  const used = new Set<string>();
  const aligned: (T | undefined)[] = draft.map((d) => {
    if (!d?.name) return undefined;
    const match = byName.get(d.name);
    if (match) used.add(d.name);
    return match;
  });
  // Rename pass: a draft entry with no name match paired with a leftover
  // baseline entry of identical content is the same entity renamed. Slot the
  // baseline into the draft's position so positional diffing reads one changed
  // field (the name) instead of an addition plus a deletion.
  if (fingerprint) {
    const holes = aligned.flatMap((a, i) => (a === undefined && draft[i]?.name ? [i] : []));
    const leftovers = baseline.filter((b) => b?.name && !used.has(b.name));
    const pairs = pairByFingerprint(holes, leftovers, (i) => fingerprint(draft[i]), fingerprint);
    for (const [i, b] of pairs) {
      aligned[i] = b;
      used.add(b.name as string);
    }
  }
  for (const b of baseline) {
    if (!b?.name || !used.has(b.name)) aligned.push(b);
  }
  return aligned;
}

/**
 * Greedily pair entries from two lists whose fingerprints match; each entry is
 * used at most once. Rename detection: a removed entry and an added entry with
 * the same content fingerprint are one renamed entity, not two changes.
 */
export function pairByFingerprint<A, B>(
  as: A[],
  bs: B[],
  fpA: (a: A) => string,
  fpB: (b: B) => string,
): Array<[A, B]> {
  const pairs: Array<[A, B]> = [];
  const usedB = new Set<number>();
  for (const a of as) {
    const fp = fpA(a);
    const idx = bs.findIndex((b, i) => !usedB.has(i) && fpB(b) === fp);
    if (idx >= 0) {
      usedB.add(idx);
      pairs.push([a, bs[idx]]);
    }
  }
  return pairs;
}

/**
 * Content identity for rename detection: everything except the entry's `name`
 * and its live `status` telemetry (status drifts between the deploy-time
 * baseline and the live draft and is never dirt — see diffOneResource).
 */
export function renameFingerprint(entry: unknown): string {
  if (entry == null || typeof entry !== "object") return JSON.stringify(entry ?? null);
  const { name: _name, status: _status, ...rest } = entry as Record<string, unknown>;
  return JSON.stringify(rest);
}

/**
 * Per-row status of env vars in a resource's draft, relative to baseline.
 * Rows are matched by `name` (the env KEY) for stable identity across edits;
 * unnamed rows fall back to positional matching. Baseline rows that don't
 * appear in the draft don't surface here — they live in baseline only.
 */
export type EnvRowStatus = "unchanged" | "modified" | "added";

export function envRowsDiff(
  draftRows: Array<Record<string, unknown>>,
  baselineRows: Array<Record<string, unknown>>,
): EnvRowStatus[] {
  const baselineByName = new Map<string, Record<string, unknown>>();
  baselineRows.forEach((r) => {
    const name = typeof r?.name === "string" ? r.name : "";
    if (name) baselineByName.set(name, r);
  });

  return draftRows.map((draft, i) => {
    const draftName = typeof draft?.name === "string" ? draft.name : "";
    if (draftName) {
      const base = baselineByName.get(draftName);
      if (!base) return "added";
      return deepEqual(draft, base) ? "unchanged" : "modified";
    }
    // Unnamed row — fall back to positional match.
    const base = baselineRows[i];
    if (!base) return "added";
    return deepEqual(draft, base) ? "unchanged" : "modified";
  });
}

/**
 * Revert a single env row in `draft.resources[resourceIdx]` to its baseline
 * value, matched by name (or positional fallback for unnamed rows). If the
 * row is "added" with no baseline counterpart, it's removed entirely.
 */
export function revertEnvRow(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  resourceIdx: number,
  envIdx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const draftResource = draft.resources[resourceIdx];
  if (!draftResource) return draft;
  const draftRows = (draftResource.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >;
  const draftRow = draftRows[envIdx];
  if (!draftRow) return draft;

  const baselineResource = baseline.resources[resourceIdx];
  const baselineRows = (baselineResource?.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >;

  const draftName = typeof draftRow.name === "string" ? draftRow.name : "";
  let baselineRow: Record<string, unknown> | undefined;
  if (draftName) {
    baselineRow = baselineRows.find((r) => (typeof r?.name === "string" ? r.name : "") === draftName);
  } else {
    baselineRow = baselineRows[envIdx];
  }

  const nextRows = draftRows.slice();
  if (baselineRow) {
    nextRows[envIdx] = cloneJson(baselineRow);
  } else {
    // No baseline counterpart — the row is newly added; drop it.
    nextRows.splice(envIdx, 1);
  }

  const nextResources = draft.resources.slice();
  nextResources[resourceIdx] = {
    ...draftResource,
    execution_config: {
      ...draftResource.execution_config,
      environment_variables: nextRows as never,
    },
  };
  return { ...draft, resources: nextResources };
}

/**
 * Tabs in the resource accordion. Used by callers to decide which tab
 * triggers should render a dirty dot.
 */
export type ResourceTab = "configuration" | "deployment" | "environment";

/**
 * Buckets indicating which tabs of a resource contain dirty fields. The
 * env-vars list is its own tab (Environment); everything else lives in
 * Configuration or Deployment per the existing tab structure.
 */
export interface ResourceDirtyTabs {
  configuration: boolean;
  deployment: boolean;
  environment: boolean;
}

const DEPLOYMENT_KEYS = new Set([
  "init_spec",
]);

export function dirtyTabsForResource(
  draft: Partial<FormStackResourceData> | undefined,
  baseline: Partial<FormStackResourceData> | undefined,
): ResourceDirtyTabs {
  const out: ResourceDirtyTabs = {
    configuration: false,
    deployment: false,
    environment: false,
  };
  if (deepEqual(draft, baseline)) return out;

  const dKeys = new Set(Object.keys((draft ?? {}) as Record<string, unknown>));
  const bKeys = new Set(Object.keys((baseline ?? {}) as Record<string, unknown>));
  const keys = new Set<string>([...dKeys, ...bKeys]);

  for (const k of keys) {
    if (k === "status") continue; // server telemetry, never user dirt
    const dv = (draft as Record<string, unknown> | undefined)?.[k];
    const bv = (baseline as Record<string, unknown> | undefined)?.[k];
    if (deepEqual(dv, bv)) continue;

    if (k === "execution_config") {
      // Split env-vars vs command/args. env-vars → environment; rest → deployment? No — command/args live with image, so put under configuration.
      const dEC = (dv ?? {}) as Record<string, unknown>;
      const bEC = (bv ?? {}) as Record<string, unknown>;
      if (!deepEqual(dEC.environment_variables, bEC.environment_variables)) {
        out.environment = true;
      }
      if (!deepEqual(dEC.command, bEC.command) || !deepEqual(dEC.args, bEC.args)) {
        out.configuration = true;
      }
    } else if (DEPLOYMENT_KEYS.has(k)) {
      out.deployment = true;
    } else {
      out.configuration = true;
    }
  }

  return out;
}

// --- Generic dot-path helpers, used by per-field dirty/reset infra. ---

/** Read a nested value via dot-path. Returns undefined for missing segments. */
export function getAtPath(obj: unknown, path: string): unknown {
  if (!path) return obj;
  const parts = path.split(".");
  let cur: unknown = obj;
  for (const p of parts) {
    if (cur == null || typeof cur !== "object") return undefined;
    cur = (cur as Record<string, unknown>)[p];
  }
  return cur;
}

/** Immutably set a nested value via dot-path; returns a clone of `obj` with `path` set.
 * Preserves array-ness: `{...arr}` would coerce arrays into plain objects with
 * numeric keys, so we shallow-clone arrays with `arr.slice()` instead. */
export function setAtPath<T>(obj: T, path: string, value: unknown): T {
  if (!path) return value as T;
  const parts = path.split(".");
  const cloneNode = (node: unknown): Record<string, unknown> | unknown[] => {
    if (Array.isArray(node)) return node.slice() as unknown[];
    if (node && typeof node === "object") return { ...(node as Record<string, unknown>) };
    return {};
  };
  const root = cloneNode(obj);
  let cur: Record<string, unknown> | unknown[] = root;
  for (let i = 0; i < parts.length - 1; i++) {
    const k = parts[i];
    const idx = Array.isArray(cur) ? Number(k) : k;
    const next = (cur as Record<string | number, unknown>)[idx as string];
    const cloned = cloneNode(next);
    (cur as Record<string | number, unknown>)[idx as string] = cloned;
    cur = cloned;
  }
  const lastKey = parts[parts.length - 1];
  const lastIdx = Array.isArray(cur) ? Number(lastKey) : lastKey;
  (cur as Record<string | number, unknown>)[lastIdx as string] = value;
  return root as T;
}

/** Is a single dot-path different between draft and baseline (deep)? */
export function isPathDirty(
  draft: unknown,
  baseline: unknown,
  path: string,
): boolean {
  return !deepEqual(getAtPath(draft, path), getAtPath(baseline, path));
}

/** Revert a single dot-path field on a resource to its baseline value. */
export function revertResourceField(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  resourceIdx: number,
  path: string,
): { resources: ResourceArr; volumes: VolumeArr } {
  const draftResource = draft.resources[resourceIdx];
  if (!draftResource) return draft;
  const baselineResource = baseline.resources[resourceIdx];
  const baselineValue = getAtPath(baselineResource, path);
  const nextResource = setAtPath(draftResource, path, cloneJson(baselineValue));
  const nextResources = draft.resources.slice();
  nextResources[resourceIdx] = nextResource;
  return { ...draft, resources: nextResources };
}

export function revertVolume(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, volumes: draft.volumes.slice() };
  // Nullish holes from a name-aligned baseline mean "draft-only" — drop the row.
  const baselineEntry = idx < baseline.volumes.length ? baseline.volumes[idx] : undefined;
  if (baselineEntry != null) {
    const liveStatus = (draft.volumes[idx] as { status?: unknown } | undefined)?.status;
    next.volumes[idx] = {
      ...cloneJson(baselineEntry),
      ...(liveStatus !== undefined ? { status: liveStatus } : {}),
    } as (typeof next.volumes)[number];
  } else {
    next.volumes.splice(idx, 1);
  }
  return next;
}

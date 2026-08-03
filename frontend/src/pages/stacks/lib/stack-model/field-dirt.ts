import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";
import { getAtPath } from "./path";
import { canonicalResourceFromForm } from "./from-form";
import { resourceFieldChanges } from "./diff";
import { deepEqual } from "./equal";
import type { CanonicalResource } from "./canonical";
import type { FieldSection } from "./policy";

/**
 * Draft objects are reference-stable except for the resource being edited, so
 * one entry per object holds this to one canonicalization per keystroke rather
 * than one per field the drawer renders.
 */
const cache = new WeakMap<object, CanonicalResource>();

function canonical(resource: Partial<FormStackResourceData> | undefined): CanonicalResource | undefined {
  if (!resource || typeof resource !== "object" || !resource.name?.trim()) return undefined;
  const hit = cache.get(resource);
  if (hit) return hit;
  const built = canonicalResourceFromForm(resource as FormStackResourceData);
  cache.set(resource, built);
  return built;
}

/**
 * The drawer's UI-only fields, in canonical terms. `gitRevisionValue` and
 * `gitRevisionType` both name two paths: which of branch/tag carries the value
 * is itself what a revision-type switch changes.
 */
const FORM_PATH_TO_CANONICAL: Record<string, string[]> = {
  gitCommitPin: ["source.git.commit"],
  gitRevisionValue: ["source.git.branch", "source.git.tag"],
  gitRevisionType: ["source.git.branch", "source.git.tag"],
};

/** The Build-from toggle edits which kind of source there is, not what is in
 *  it — it must not light up because a field inside the source changed. */
function sourceKind(r: CanonicalResource): string {
  if (r.source?.git) return "git";
  if (r.source?.image) return "image";
  return "";
}

/** A mount row addressed by its position in the drawer's list. */
const MOUNT_ROW_PATH = /^volume_mounts\.(\d+)$/;

function canonicalPaths(formPath: string): string[] {
  return FORM_PATH_TO_CANONICAL[formPath] ?? [formPath];
}

/** Changed paths for one resource, memoized per draft object and baseline pair. */
const changedPathsCache = new WeakMap<object, { baseline: unknown; paths: Set<string> }>();

function changedPaths(d: CanonicalResource, b: CanonicalResource, key: object): Set<string> {
  const hit = changedPathsCache.get(key);
  if (hit && hit.baseline === b) return hit.paths;
  const paths = new Set(resourceFieldChanges(b, d).map((f) => f.path));
  changedPathsCache.set(key, { baseline: b, paths });
  return paths;
}

/** Is one drawer field edited relative to the baseline? */
export function isFieldDirty(
  draft: Partial<FormStackResourceData> | undefined,
  baseline: Partial<FormStackResourceData> | undefined,
  formPath: string,
): boolean {
  const d = canonical(draft);
  const b = canonical(baseline);
  if (!d || !b) return !deepEqual(getAtPath(draft, formPath), getAtPath(baseline, formPath));
  if (formPath === "sourceType") return sourceKind(d) !== sourceKind(b);
  const dirty = changedPaths(d, b, draft as object);
  const mountRow = MOUNT_ROW_PATH.exec(formPath);
  if (mountRow) {
    // Canonical keys a mount by its target path, so the row index has to be
    // resolved against the draft. Exact match only: a prefix test would let
    // `/data` light `/data.bak`, and a bare `mounts` would light every row.
    const target = draft?.volume_mounts?.[Number(mountRow[1])]?.target_path;
    return !!target && dirty.has(`mounts.${target}`);
  }
  return canonicalPaths(formPath).some(
    (p) => dirty.has(p) || [...dirty].some((path) => path.startsWith(`${p}.`)),
  );
}

export interface ResourceDirtyTabs {
  configuration: boolean;
  deployment: boolean;
  environment: boolean;
}

const NO_TABS: ResourceDirtyTabs = { configuration: false, deployment: false, environment: false };

/** Which drawer tabs hold changes, for the tab dots. */
export function dirtyTabsForResource(
  draft: Partial<FormStackResourceData> | undefined,
  baseline: Partial<FormStackResourceData> | undefined,
): ResourceDirtyTabs {
  const fields = resourceFieldChanges(canonical(baseline), canonical(draft));
  if (!fields.length) return NO_TABS;
  const tabs: ResourceDirtyTabs = { ...NO_TABS };
  for (const f of fields) tabs[f.section as FieldSection] = true;
  return tabs;
}

export function isResourceDirty(
  draft: Partial<FormStackResourceData> | undefined,
  baseline: Partial<FormStackResourceData> | undefined,
): boolean {
  const d = canonical(draft);
  const b = canonical(baseline);
  if (!d || !b) return !deepEqual(draft, baseline);
  return resourceFieldChanges(b, d).length > 0;
}

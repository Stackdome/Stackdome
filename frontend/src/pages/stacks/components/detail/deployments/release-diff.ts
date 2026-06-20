import { deepEqual } from "@/pages/stacks/lib/stack-diff";

export interface SnapshotChange {
  path: string;
  before: unknown;
  after: unknown;
  kind: "added" | "removed" | "changed";
}

function isPlainObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

/** Recursive JSON diff. Arrays and scalars are leaves; objects are descended. */
export function diffSnapshots(before: unknown, after: unknown, prefix = ""): SnapshotChange[] {
  if (deepEqual(before, after)) return [];

  if (isPlainObject(before) && isPlainObject(after)) {
    const keys = new Set([...Object.keys(before), ...Object.keys(after)]);
    const out: SnapshotChange[] = [];
    for (const key of keys) {
      const path = prefix ? `${prefix}.${key}` : key;
      const hasB = key in before, hasA = key in after;
      if (hasB && !hasA) out.push({ path, before: before[key], after: undefined, kind: "removed" });
      else if (!hasB && hasA) out.push({ path, before: undefined, after: after[key], kind: "added" });
      else out.push(...diffSnapshots(before[key], after[key], path));
    }
    return out;
  }

  return [{ path: prefix, before, after, kind: "changed" }];
}

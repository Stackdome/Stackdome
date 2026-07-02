import type { Stack, StackUpdateRequest, StackResource, Volume } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";
import { cleanServerResource } from "./server-state";

function cleanVolume(v: Volume) {
  const { id, status, ...rest } = v as Volume & { status?: unknown };
  void id; void status;
  return rest;
}

/**
 * Turn a release snapshot back into a whole-stack PUT body. Replace-all
 * semantics are exactly right for a revert; connection ids are kept so the
 * backend upserts instead of churning rows.
 */
export function snapshotToUpdateRequest(
  snap: StackReleaseSnapshot,
  current: { name: string; labels?: Stack["labels"] },
): StackUpdateRequest {
  const connections = snap.connections ?? [];
  return {
    name: current.name,
    labels: current.labels,
    spec: {
      stack_resources: (snap.resources ?? []).map((r) => cleanServerResource(r as StackResource)),
      volumes: (snap.volumes ?? []).length > 0 ? (snap.volumes ?? []).map((v) => cleanVolume(v as Volume)) : undefined,
      ...(connections.length > 0 ? { connections } : {}),
    },
  } as StackUpdateRequest;
}

/** Volumes on the stack that the deployed snapshot doesn't know — draft
 *  artifacts to remove after the PUT (which never deletes volumes). */
export function volumesToDelete(stack: Stack, snap: StackReleaseSnapshot): { id: string; name: string }[] {
  const snapNames = new Set((snap.volumes ?? []).map((v) => v.name).filter(Boolean));
  return (stack.spec?.volumes ?? [])
    .filter((v) => v.id && v.name && !snapNames.has(v.name))
    .map((v) => ({ id: v.id!, name: v.name! }));
}

import type { Stack, StackResource, StackUpdateRequest, StackResourceUpdateRequest, Volume } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";

/**
 * Strip the fields the server owns and the PUT schema rejects. Deliberately NOT
 * the canonical adapter: that one is built for comparing, so it drops nameless
 * resources, sorts env vars and materializes defaults. This body is replace-all
 * — anything it fails to carry is deleted from the stack.
 */
function cleanResource(r: StackResource): StackResourceUpdateRequest {
  const { id, stack_id, revision, outputs, status, ...rest } = r as StackResource & {
    outputs?: unknown;
    status?: unknown;
  };
  void id; void stack_id; void revision; void outputs; void status;
  return rest as StackResourceUpdateRequest;
}

function cleanVolume(v: Volume) {
  // Strip every readOnly Volume field (id, project_id, status) — the whole-stack
  // PUT is schema-validated server-side and rejects readOnly properties.
  const { id, project_id, status, ...rest } = v as Volume & { status?: unknown; project_id?: string };
  void id; void project_id; void status;
  return rest;
}

/**
 * gorm's Updates() skips nil fields, so a revert must send explicit empty
 * collections to CLEAR authored values the deployment never had.
 *
 * All ARRAY-valued resource fields are normalised here so a "Discard draft
 * changes" revert cannot leave a post-deploy canvas edit (e.g. an added port)
 * stranded in the DB.
 *
 * KNOWN LIMITATION: struct-pointer fields (source, lifecycle_config) cannot be
 * cleared (set to nil) from the frontend — JSON null still unmarshals to a nil
 * pointer that gorm's Updates() skips. Clearing an optional nested field (e.g.
 * dropping source.git.push) is therefore not fully reverted by a discard.
 * Fixing this requires a backend change (dedicated PATCH endpoint or explicit
 * NULL handling in the store). Tracked as a follow-up.
 */
function withExplicitEmptyCollections(r: StackResourceUpdateRequest): StackResourceUpdateRequest {
  return {
    ...r,
    depends_on: r.depends_on ?? [],
    volume_mounts: r.volume_mounts ?? [],
    ports: r.ports ?? [],
    labels: r.labels ?? [],
    annotations: r.annotations ?? [],
    execution_config: {
      ...(r.execution_config ?? {}),
      command: r.execution_config?.command ?? [],
      args: r.execution_config?.args ?? [],
      environment_variables: r.execution_config?.environment_variables ?? [],
    },
  };
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
      stack_resources: (snap.resources ?? []).map((r) =>
        withExplicitEmptyCollections(cleanResource(r as StackResource)),
      ),
      volumes: (snap.volumes ?? []).length > 0 ? (snap.volumes ?? []).map((v) => cleanVolume(v as Volume)) : undefined,
      ...(connections.length > 0 ? { connections } : {}),
    },
  } as StackUpdateRequest;
}

/**
 * Whole-stack PUT body from the LIVE server stack with only the labels swapped.
 * PUT replace-all semantics require the full resource/volume/connection set —
 * an incomplete body silently wipes bindings.
 */
export function stackToUpdateRequest(stack: Stack, labels: Stack["labels"]): StackUpdateRequest {
  const connections = stack.spec?.connections ?? [];
  return {
    name: stack.name,
    labels,
    spec: {
      stack_resources: (stack.spec?.stack_resources ?? []).map((r) =>
        withExplicitEmptyCollections(cleanResource(r as StackResource)),
      ),
      volumes:
        (stack.spec?.volumes ?? []).length > 0
          ? (stack.spec?.volumes ?? []).map((v) => cleanVolume(v as Volume))
          : undefined,
      ...(connections.length > 0 ? { connections } : {}),
    },
  };
}

/** Volumes on the stack that the deployed snapshot doesn't know — draft
 *  artifacts to remove after the PUT (which never deletes volumes). */
export function volumesToDelete(stack: Stack, snap: StackReleaseSnapshot): { id: string; name: string }[] {
  const snapNames = new Set((snap.volumes ?? []).map((v) => v.name).filter(Boolean));
  return (stack.spec?.volumes ?? [])
    .filter((v) => v.id && v.name && !snapNames.has(v.name))
    .map((v) => ({ id: v.id!, name: v.name! }));
}

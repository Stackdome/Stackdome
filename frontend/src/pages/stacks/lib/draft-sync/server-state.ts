import type { Stack, StackResource, StackResourceUpdateRequest, Volume, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";

/**
 * The engine's mirror of what the server currently holds for a stack, indexed
 * for diffing. Server ids live here and ONLY here — form state stays id-free.
 */
export interface ServerConnectionEntry {
  id?: string;
  conn: StackConnection;
}

export interface ServerStackState {
  resourcesByName: Map<string, StackResourceUpdateRequest>;
  volumeIdByName: Map<string, string>;
  volumesByName: Map<string, VolumeUpdateRequest>;
  connections: Map<string, ServerConnectionEntry>;
}

type NodeRef = { type?: string; id?: string; name?: string } | undefined;

function nodeKey(n: NodeRef): string {
  if (!n) return "";
  return `${n.type ?? ""}:${n.id ?? n.name ?? ""}`;
}

/**
 * Content identity of a connection — everything except its mappings. Mirrors
 * the backend's uniqueness check (kind + from + to + config discriminator);
 * mapping changes are updates to the same connection, identity changes are a
 * different connection.
 */
export function connectionIdentityKey(c: StackConnection): string {
  const cfg = c.config as { database?: string; superuser?: boolean } | undefined;
  const cfgKey = cfg ? (cfg.superuser ? "superuser" : `db:${cfg.database ?? ""}`) : "";
  return [c.kind ?? "", nodeKey(c.from), nodeKey(c.to), cfgKey].join("|");
}

/** Strip server-computed fields so a server resource compares against form-derived ones. */
export function cleanServerResource(r: StackResource): StackResourceUpdateRequest {
  const { id, stack_id, revision, status, outputs, ...rest } = r as StackResource & { outputs?: unknown };
  void id; void stack_id; void revision; void status; void outputs;
  // volume_mounts are stored as volume_mount connections since the volume_mounts
  // table was dropped. The server always returns volume_mounts: [] on resources,
  // while the desired state strips them too — both sides must be undefined so
  // deepEqual sees no phantom diff on every autosave cycle.
  // workload_type is zod-defaulted to "Service" on the form side; mirror that
  // default here so a server resource without it doesn't read as dirty.
  return {
    ...rest,
    volume_mounts: undefined,
    workload_type: rest.workload_type ?? "Service",
  } as StackResourceUpdateRequest;
}

function cleanServerVolume(v: Volume): VolumeUpdateRequest {
  const { id, status, ...rest } = v as Volume & { status?: unknown };
  void id; void status;
  return rest as VolumeUpdateRequest;
}

export function serverStateFromStack(stack: Stack): ServerStackState {
  const resourcesByName = new Map<string, StackResourceUpdateRequest>();
  for (const r of stack.spec?.stack_resources ?? []) {
    if (r.name) resourcesByName.set(r.name, cleanServerResource(r));
  }
  const volumeIdByName = new Map<string, string>();
  const volumesByName = new Map<string, VolumeUpdateRequest>();
  for (const v of stack.spec?.volumes ?? []) {
    if (!v.name) continue;
    if (v.id) volumeIdByName.set(v.name, v.id);
    volumesByName.set(v.name, cleanServerVolume(v));
  }
  const connections = new Map<string, ServerConnectionEntry>();
  for (const c of stack.spec?.connections ?? []) {
    connections.set(connectionIdentityKey(c), { id: c.id, conn: c });
  }
  return { resourcesByName, volumeIdByName, volumesByName, connections };
}

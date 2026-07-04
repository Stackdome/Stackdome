import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import { deepEqual } from "@/pages/stacks/lib/stack-diff";
import type { ServerStackState } from "./server-state";
import type { DesiredStackState } from "./desired-state";

/**
 * One thin-endpoint mutation. Op order is a correctness invariant:
 * volumes exist before resources mount them; a resource exists before its
 * connections; connections die before their resource (the backend does not
 * cascade). Volume update/delete are intentionally absent from autosave — no
 * thin endpoint for size edits, and deletion is handled outside this engine
 * entirely by the confirm-gated `use-volume-delete` flow (immediate,
 * synchronous `DELETE /volumes/{id}`, not a staged sync op) plus revert
 * (wholesale removal of draft-only volumes).
 */
export type SyncOp =
  | { kind: "createVolume"; volume: VolumeUpdateRequest }
  | { kind: "createResource"; resource: StackResourceUpdateRequest }
  | { kind: "updateResource"; name: string; resource: StackResourceUpdateRequest }
  | { kind: "deleteConnection"; id: string; identityKey: string }
  | { kind: "updateConnection"; id: string; identityKey: string; conn: StackConnection }
  | { kind: "createConnection"; identityKey: string; conn: StackConnection }
  | { kind: "deleteResource"; name: string };

export function computeSyncOps(server: ServerStackState, desired: DesiredStackState): SyncOp[] {
  const createVolumes: SyncOp[] = [];
  for (const [name, volume] of desired.volumes) {
    if (!server.volumesByName.has(name)) createVolumes.push({ kind: "createVolume", volume });
  }

  const createResources: SyncOp[] = [];
  const updateResources: SyncOp[] = [];
  for (const [name, resource] of desired.resources) {
    const existing = server.resourcesByName.get(name);
    if (!existing) createResources.push({ kind: "createResource", resource });
    else if (!deepEqual(existing, resource)) updateResources.push({ kind: "updateResource", name, resource });
  }

  const deleteResources: SyncOp[] = [];
  for (const name of server.resourcesByName.keys()) {
    if (!desired.resources.has(name) && !desired.held.has(name)) {
      deleteResources.push({ kind: "deleteResource", name });
    }
  }
  const deletedResourceNames = new Set(deleteResources.map((op) => (op as { name: string }).name));

  const deleteConnections: SyncOp[] = [];
  const updateConnections: SyncOp[] = [];
  const createConnections: SyncOp[] = [];
  const connTouchesHeld = (conn: StackConnection): boolean => {
    const toName = conn.to?.name;
    const fromName = conn.from?.type === "stack_resource" ? conn.from?.name : undefined;
    return (!!toName && desired.held.has(toName)) || (!!fromName && desired.held.has(fromName));
  };

  for (const [key, entry] of server.connections) {
    const want = desired.connections.get(key);
    if (want) {
      // Compare both mappings and config — volume_mount connections carry their
      // mount_path / sub_path in config with no mappings, so a config-only
      // diff (e.g. mount_path change) must still produce an updateConnection.
      const serverContent = { m: entry.conn.mappings ?? [], c: entry.conn.config ?? {} };
      const wantContent = { m: want.mappings ?? [], c: want.config ?? {} };
      if (entry.id && !deepEqual(serverContent, wantContent)) {
        updateConnections.push({ kind: "updateConnection", id: entry.id, identityKey: key, conn: want });
      }
      continue;
    }
    // Connections tied to a resource being deleted MUST go; otherwise spare
    // held resources' connections. Id-less entries are skipped and heal after
    // the next refetch rebuilds the mirror.
    const toName = entry.conn.to?.name;
    const forcedByDelete = !!toName && deletedResourceNames.has(toName);
    if (!forcedByDelete && connTouchesHeld(entry.conn)) continue;
    if (entry.id) deleteConnections.push({ kind: "deleteConnection", id: entry.id, identityKey: key });
  }
  for (const [key, conn] of desired.connections) {
    if (!server.connections.has(key) && !connTouchesHeld(conn)) {
      createConnections.push({ kind: "createConnection", identityKey: key, conn });
    }
  }

  return [
    ...createVolumes,
    ...createResources,
    ...updateResources,
    ...deleteConnections,
    ...updateConnections,
    ...createConnections,
    ...deleteResources,
  ];
}

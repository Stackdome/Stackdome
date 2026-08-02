import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import { deepEqual } from "@/pages/stacks/lib/stack-model/equal";
import { resourcesByName, volumesByName, type CanonicalStack } from "@/pages/stacks/lib/stack-model/canonical";
import { diffStacks } from "@/pages/stacks/lib/stack-model/diff";
import { connectionsOf, resourceToApi, volumeToApi } from "@/pages/stacks/lib/stack-model/to-api";
import type { CanonicalDraft } from "@/pages/stacks/lib/stack-model/from-form";
import { connectionIdentityKey, type ServerConnectionIndex } from "./server-state";

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

/**
 * The writes the server needs, derived from the stack diff.
 *
 * The backend reconciles by name, so a rename — one entry in the diff — expands
 * into a create plus a delete.
 */
export function computeSyncOps(
  server: CanonicalStack,
  draft: CanonicalDraft,
  serverConnections: ServerConnectionIndex,
): SyncOp[] {
  const diff = diffStacks(server, draft);
  const draftResources = resourcesByName(draft);
  const serverResources = resourcesByName(server);
  const serverVolumes = volumesByName(server);

  const createVolumes: SyncOp[] = draft.volumes
    .filter((v) => !serverVolumes.has(v.name))
    .map((volume) => ({ kind: "createVolume", volume: volumeToApi(volume) }));

  const createResources: SyncOp[] = [];
  const updateResources: SyncOp[] = [];
  const deleteResources: SyncOp[] = [];
  for (const entry of diff.resources) {
    const desired = draftResources.get(entry.name);
    switch (entry.change) {
      case "added":
        if (desired) createResources.push({ kind: "createResource", resource: resourceToApi(desired) });
        break;
      case "modified": {
        // "modified" also covers env references and mounts, which are written
        // as connections rather than through the resource endpoint — so the
        // resource is only written when its own payload differs.
        const held = serverResources.get(entry.name);
        if (desired && !deepEqual(held && resourceToApi(held), resourceToApi(desired))) {
          updateResources.push({ kind: "updateResource", name: entry.name, resource: resourceToApi(desired) });
        }
        break;
      }
      case "renamed":
        if (desired) createResources.push({ kind: "createResource", resource: resourceToApi(desired) });
        if (entry.fromName && !draft.held.has(entry.fromName)) {
          deleteResources.push({ kind: "deleteResource", name: entry.fromName });
        }
        break;
      case "removed":
        // A half-typed resource reads as absent from the draft, so held names
        // are never deleted from the server.
        if (!draft.held.has(entry.name)) deleteResources.push({ kind: "deleteResource", name: entry.name });
        break;
    }
  }
  const deletedResourceNames = new Set(
    deleteResources.map((op) => (op as { name: string }).name),
  );

  const desiredConnections = new Map(
    connectionsOf(draft).map((c) => [connectionIdentityKey(c), c] as const),
  );

  const touchesHeld = (conn: StackConnection): boolean => {
    const toName = conn.to?.name;
    const fromName = conn.from?.type === "stack_resource" ? conn.from?.name : undefined;
    return (!!toName && draft.held.has(toName)) || (!!fromName && draft.held.has(fromName));
  };

  const deleteConnections: SyncOp[] = [];
  const updateConnections: SyncOp[] = [];
  const createConnections: SyncOp[] = [];

  for (const [key, entry] of serverConnections) {
    const want = desiredConnections.get(key);
    if (want) {
      // volume_mount connections carry their mount_path / sub_path in config with
      // no mappings, so a config-only change must still produce an update.
      const held = { m: entry.conn.mappings ?? [], c: entry.conn.config ?? {} };
      const wanted = { m: want.mappings ?? [], c: want.config ?? {} };
      if (entry.id && !deepEqual(held, wanted)) {
        updateConnections.push({ kind: "updateConnection", id: entry.id, identityKey: key, conn: want });
      }
      continue;
    }
    // Connections tied to a resource being deleted MUST go; otherwise spare
    // held resources' connections. Id-less entries are skipped and heal after
    // the next refetch rebuilds the mirror.
    const toName = entry.conn.to?.name;
    const forcedByDelete = !!toName && deletedResourceNames.has(toName);
    if (!forcedByDelete && touchesHeld(entry.conn)) continue;
    if (entry.id) deleteConnections.push({ kind: "deleteConnection", id: entry.id, identityKey: key });
  }
  for (const [key, conn] of desiredConnections) {
    if (!serverConnections.has(key) && !touchesHeld(conn)) {
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

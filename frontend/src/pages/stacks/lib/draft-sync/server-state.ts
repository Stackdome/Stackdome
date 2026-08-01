import type { Stack } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";

/**
 * Server ids, indexed for the sync engine — and nothing else. What the server
 * *holds* is answered by the canonical model (`canonicalFromStack`); this file
 * exists only because connections and volumes are addressed by id when written,
 * and form state stays id-free.
 */
export interface ServerConnectionEntry {
  id?: string;
  conn: StackConnection;
}

export type ServerConnectionIndex = Map<string, ServerConnectionEntry>;

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

export function serverConnectionIndex(stack: Stack): ServerConnectionIndex {
  const connections: ServerConnectionIndex = new Map();
  for (const c of stack.spec?.connections ?? []) {
    connections.set(connectionIdentityKey(c), { id: c.id, conn: c });
  }
  return connections;
}

export function volumeIdsByName(stack: Stack): Map<string, string> {
  const ids = new Map<string, string>();
  for (const v of stack.spec?.volumes ?? []) {
    if (v.name && v.id) ids.set(v.name, v.id);
  }
  return ids;
}

import type { StackConnection, ConnectionMapping } from "@/api/connections";
import type { components } from "@/api/types/openapi";
import { ADDON_OUTPUT_FIELDS, type AddonOutputField } from "@/pages/stacks/lib/addon-presets";

export type EnvVar = components["schemas"]["EnvVar"];

// The form-side env-row union. Re-exported from form-schema as FormEnvVarData.
export type FormEnvRow =
  | { from: "stack"; name: string; value: string }
  | { from: "secret"; name: string; secretId: string; secretKey: string }
  | {
      from: "addon";
      name: string;
      addonId: string;
      database?: string;
      superuser: boolean;
      credField?: AddonOutputField;
    }
  | { from: "resource"; name: string; resourceName: string; output: string }
  | {
      from: "resourceTemplate";
      name: string;
      resourceName: string;
      template: string;
      /** Template key → output name on the source resource. */
      values: Record<string, string>;
    }
  | { from: "self"; name: string; selfOutput: string };

const envMapping = (name: string, output: string): ConnectionMapping => ({
  target: { type: "env", name },
  value: { output },
});

// Split a resource's form env-rows into the two persistence channels:
//  - envVars: literal + self rows (ride the stack PUT)
//  - connections: secret/addon/resource rows, grouped one connection per source.
export function splitEnvRows(
  resourceName: string,
  rows: FormEnvRow[],
): { envVars: EnvVar[]; connections: StackConnection[] } {
  const envVars: EnvVar[] = [];
  // Insertion-ordered groups keyed by source identity.
  type GroupedConn = StackConnection & { mappings: ConnectionMapping[] };
  const groups = new Map<string, GroupedConn>();

  const ensureGroup = (key: string, conn: () => StackConnection): GroupedConn => {
    let g = groups.get(key);
    if (!g) {
      g = { ...conn(), mappings: [] };
      groups.set(key, g);
    }
    return g;
  };

  for (const row of rows) {
    switch (row.from) {
      case "stack":
        envVars.push({ name: row.name, value: row.value });
        break;
      case "self":
        envVars.push({ name: row.name, self_output: row.selfOutput });
        break;
      case "secret": {
        if (!row.secretId || !row.secretKey) break; // skip in-progress rows
        const key = `secret::${row.secretId}`;
        const g = ensureGroup(key, () => ({
          kind: "env",
          from: { type: "secret", id: row.secretId },
          to: { type: "stack_resource", name: resourceName },
        }));
        g.mappings.push(envMapping(row.name, row.secretKey));
        break;
      }
      case "addon": {
        if (!row.addonId || !row.credField) break; // skip in-progress rows
        const db = row.superuser ? "" : row.database ?? "";
        const key = `addon::${row.addonId}::${db}::${row.superuser}`;
        const g = ensureGroup(key, () => ({
          kind: "env",
          from: { type: "addon/postgres", id: row.addonId },
          to: { type: "stack_resource", name: resourceName },
          config: row.superuser
            ? { superuser: true }
            : { database: row.database, superuser: false },
        }));
        g.mappings.push(envMapping(row.name, row.credField));
        break;
      }
      case "resource": {
        if (!row.resourceName || !row.output) break; // skip in-progress rows
        const key = `resource::${row.resourceName}`;
        const g = ensureGroup(key, () => ({
          kind: "env",
          from: { type: "stack_resource", name: row.resourceName },
          to: { type: "stack_resource", name: resourceName },
        }));
        g.mappings.push(envMapping(row.name, row.output));
        break;
      }
      case "resourceTemplate": {
        if (!row.resourceName || !row.template) break; // skip in-progress rows
        const key = `resource::${row.resourceName}`;
        const g = ensureGroup(key, () => ({
          kind: "env",
          from: { type: "stack_resource", name: row.resourceName },
          to: { type: "stack_resource", name: resourceName },
        }));
        g.mappings.push({
          target: { type: "env", name: row.name },
          value: {
            template: row.template,
            values: Object.fromEntries(
              Object.entries(row.values).map(([key, output]) => [key, { output }]),
            ),
          },
        });
        break;
      }
    }
  }

  const connections = [...groups.values()].filter((c) => (c.mappings?.length ?? 0) > 0);
  return { envVars, connections };
}

const ADDON_FIELD_SET = new Set<string>(ADDON_OUTPUT_FIELDS);

type AddonConfig = { database?: string; superuser?: boolean };

// Expand the stack's connections into form rows for one resource (the rows whose
// `to` is this resource). Literal/self rows come from env vars, added separately.
export function connectionsToEnvRows(
  resourceName: string,
  connections: StackConnection[],
): FormEnvRow[] {
  const rows: FormEnvRow[] = [];
  for (const c of connections) {
    if (c.kind !== "env") continue;
    if (c.to?.type !== "stack_resource" || c.to?.name !== resourceName) continue;
    const cfg = c.config as AddonConfig | undefined;
    for (const m of c.mappings ?? []) {
      const envName = m.target?.name ?? "";
      const output = m.value?.output ?? "";
      // unknown source types are silently skipped (consistent with splitEnvRows)
      if (c.from?.type === "secret" && c.from?.id) {
        rows.push({ from: "secret", name: envName, secretId: c.from.id, secretKey: output });
      } else if (c.from?.type === "addon/postgres" && c.from?.id) {
        if (!ADDON_FIELD_SET.has(output)) continue;
        rows.push({
          from: "addon",
          name: envName,
          addonId: c.from.id,
          database: cfg?.superuser ? undefined : cfg?.database,
          superuser: cfg?.superuser ?? false,
          credField: output as AddonOutputField,
        });
      } else if (c.from?.type === "stack_resource" && c.from?.name) {
        if (m.value?.template) {
          rows.push({
            from: "resourceTemplate",
            name: envName,
            resourceName: c.from.name,
            template: m.value.template,
            values: Object.fromEntries(
              Object.entries(m.value.values ?? {}).map(([key, ref]) => [key, ref.output ?? ""]),
            ),
          });
        } else {
          rows.push({ from: "resource", name: envName, resourceName: c.from.name, output });
        }
      }
    }
  }
  return rows;
}

// Build the full desired connection set from every resource's rows.
export function buildDesiredConnections(
  resources: { name: string; rows: FormEnvRow[] }[],
): StackConnection[] {
  return resources.flatMap((r) => splitEnvRows(r.name, r.rows).connections);
}

// ── Volume mount ↔ connection translation ────────────────────────────────────

/**
 * A loose form-side mount row. Required fields are optional here so that
 * in-progress rows (user is still typing) are representable without casting.
 */
export type FormMountRow = {
  source_volume_name?: string;
  source_sub_path?: string;
  target_path?: string;
  read_only?: boolean;
};

type VolumeMountCfg = { mount_path?: string; sub_path?: string; read_only?: boolean };

/**
 * Convert form mount rows to volume_mount StackConnections.
 * Skips in-progress rows — those missing source_volume_name OR target_path.
 */
export function mountsToConnections(resourceName: string, mounts: FormMountRow[]): StackConnection[] {
  const conns: StackConnection[] = [];
  for (const m of mounts) {
    if (!m.source_volume_name || !m.target_path) continue; // skip in-progress rows
    conns.push({
      kind: "volume_mount",
      from: { type: "volume", name: m.source_volume_name },
      to: { type: "stack_resource", name: resourceName },
      config: {
        mount_path: m.target_path,
        ...(m.source_sub_path ? { sub_path: m.source_sub_path } : {}),
        ...(m.read_only !== undefined ? { read_only: m.read_only } : {}),
      },
    });
  }
  return conns;
}

/**
 * Expand volume_mount connections for a resource back into form mount rows.
 * Ignores non-volume_mount connections and connections to other resources.
 */
export function connectionsToMounts(resourceName: string, connections: StackConnection[]): FormMountRow[] {
  const rows: FormMountRow[] = [];
  for (const c of connections) {
    if (c.kind !== "volume_mount") continue;
    if (c.to?.type !== "stack_resource" || c.to?.name !== resourceName) continue;
    const vmcfg = c.config as VolumeMountCfg | undefined;
    if (!c.from?.name || !vmcfg?.mount_path) continue; // malformed — skip
    rows.push({
      source_volume_name: c.from.name,
      source_sub_path: vmcfg.sub_path ?? "",
      target_path: vmcfg.mount_path,
      ...(vmcfg.read_only !== undefined ? { read_only: vmcfg.read_only } : {}),
    });
  }
  return rows;
}

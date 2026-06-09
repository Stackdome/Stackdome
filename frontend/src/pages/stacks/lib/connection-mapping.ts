import type { StackConnection, ConnectionMapping } from "@/api/connections";
import type { components } from "@/api/types/openapi";
import { ADDON_OUTPUT_FIELDS, type AddonOutputField } from "@/pages/stacks/lib/addon-presets";

const SIMPLE_KEY = /^[A-Za-z0-9_]+$/;

// Mirror of pkg/models/output_descriptor.go secretOutputAccessor: simple keys
// get a dot accessor; anything else is bracket-quoted with ' and \ escaped.
export function secretOutputAccessor(key: string): string {
  if (SIMPLE_KEY.test(key)) return `key.${key}`;
  const escaped = key.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
  return `key['${escaped}']`;
}

// Reverse secretOutputAccessor. Returns the key, or null if the accessor is not
// a secret-key accessor.
export function parseSecretOutput(output: string): string | null {
  if (output.startsWith("key.") && output.length > 4) return output.slice(4);
  const m = output.match(/^key\['(.*)'\]$/s);
  if (!m) return null;
  return m[1].replace(/\\(['\\])/g, "$1");
}

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
        g.mappings.push(envMapping(row.name, secretOutputAccessor(row.secretKey)));
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
        const key = parseSecretOutput(output);
        if (key === null) continue;
        rows.push({ from: "secret", name: envName, secretId: c.from.id, secretKey: key });
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
        rows.push({ from: "resource", name: envName, resourceName: c.from.name, output });
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

import type { StackConnection, ConnectionMapping } from "@/api/connections";
import type { components } from "@/api/types/openapi";
import type { AddonOutputField } from "@/pages/stacks/lib/addon-presets";

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

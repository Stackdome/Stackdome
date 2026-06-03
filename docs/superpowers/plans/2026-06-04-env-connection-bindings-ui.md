# Env Connection Bindings UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the per-env-row "From" source selector (Literal / Secret / Addon / Resource / Self) on a stack resource's Environment tab, persisting Secret/Addon/Resource bindings through the stack-connections API instead of the deleted inline `execution_config` fields.

**Architecture:** A pure mapping module (`lib/connection-mapping.ts`) converts between form env-rows and `StackConnection[]` + `EnvVar[]`, with no React or network dependencies. The form schema regains its discriminated `FormEnvVar` union (with two new `resource`/`self` arms) and serializes literal+self rows into resource env vars while delegating secret/addon/resource rows to the mapping module. On save, the stack PUT carries env vars; a connection diff (`buildDesiredConnections` vs the stack's loaded `spec.connections`) fires granular `POST`/`PUT`/`DELETE` against a new `api/connections.ts` client. On create, connections ride in `spec.connections`. UI components (`env-row`, `env-addon-group`) are recovered from git `8891fee~1` and adapted.

**Tech Stack:** React 19, Vite, Zod, Vitest, TypeScript, generated OpenAPI types (`frontend/src/api/types/openapi.d.ts`).

**Standing constraints (from the user, must hold throughout):**
1. Frontend-only — no backend changes. Use only existing, spec-defined APIs.
2. No teams/roles surfaced; everything scopes to the hidden **default team** (`useResourceTeams().defaultTeamName`).
3. `env` connection `kind` only. Volume/build-artifact connections are out of scope.

---

## File Structure

| File | Responsibility | Action |
|------|----------------|--------|
| `frontend/src/api/connections.ts` | Connections CRUD client (team-scoped). Thin typed wrapper. | Create |
| `frontend/src/api/tests/connections.test.ts` | URL/verb contract test for the client. | Create |
| `frontend/src/pages/stacks/lib/addon-presets.ts` | Postgres addon output field list (`ADDON_OUTPUT_FIELDS`) + cluster-wide set. | Create (was deleted in `8891fee`) |
| `frontend/src/pages/stacks/lib/connection-mapping.ts` | Pure rows↔connections+envVars, output accessors, connection diff. | Create |
| `frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts` | Unit tests for the mapping module. | Create |
| `frontend/src/pages/stacks/schemas/form-schema.ts` | `FormEnvVar` union (5 arms); serialize literal/self → env vars; deserialize env vars → literal/self rows. | Modify |
| `frontend/src/pages/stacks/schemas/tests/form-schema.test.ts` | Round-trip coverage for the union + serializer. | Modify (recover + adapt) |
| `frontend/src/pages/stacks/components/shared/env-row.tsx` | Per-row From selector + sub-pickers (secret/addon/resource/self). | Replace (recover + adapt) |
| `frontend/src/pages/stacks/components/shared/env-addon-group.tsx` | Grouped addon-binding card + edit/detach state machine. | Create (recover) |
| `frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx` | EnvRow source-switch + validation tests. | Replace (recover + adapt) |
| `frontend/src/pages/stacks/components/shared/stack-resource-environment-tab.tsx` | Renders rows; "Add variable" menu offers all source types; wires pickers. | Modify |
| `frontend/src/pages/stacks/components/detail/index.tsx` | Merge connection-derived rows into baseline on load; run connection diff after stack PUT on save. | Modify |

---

## Reference: recovered source

The pre-revert components live at git ref `8891fee~1`. Recover exact blobs with:

```bash
git show 8891fee~1:frontend/src/pages/stacks/components/shared/env-row.tsx
git show 8891fee~1:frontend/src/pages/stacks/components/shared/env-addon-group.tsx
git show 8891fee~1:frontend/src/pages/stacks/lib/addon-presets.ts
git show 8891fee~1:frontend/src/pages/stacks/schemas/tests/form-schema.test.ts
git show 8891fee~1:frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx
```

The recovered code targets the **deleted** inline fields (`environment_variables_from_secret`, `env_from_addons`) and the old `CRED_FIELDS` labels (`connectionString`, `caCertificate`). Every task below states exactly how to adapt those to the connections model and backend output names.

**Backend output names (the accessor vocabulary — must match exactly):**
- Postgres addon: `host`, `port`, `database`, `username`, `password`, `sslmode`, `ca_certificate`, `url`
- Secret: `key.<KEY>` (simple key) or `key['<KEY>']` (special chars) — mirrors `pkg/models/output_descriptor.go` `secretOutputAccessor`
- Resource: `host`, `url.<port>`, `public.<port>.url`, `public.<port>.host`
- Self: same resource-output vocabulary, but persisted as an `EnvVar.self_output`, **not** a connection.

**Run commands from `frontend/`.** Tests: `pnpm test:run <path>`. Type-check: `pnpm exec tsc -b`.

---

## Task 1: Connections API client

**Files:**
- Create: `frontend/src/api/connections.ts`
- Test: `frontend/src/api/tests/connections.test.ts`

- [ ] **Step 1: Write the failing test**

```typescript
// frontend/src/api/tests/connections.test.ts
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("../client", () => ({
  default: {
    get: vi.fn(() => Promise.resolve({ data: { items: [] } })),
    post: vi.fn(() => Promise.resolve({ data: { id: "c1" } })),
    put: vi.fn(() => Promise.resolve({ data: { id: "c1" } })),
    delete: vi.fn(() => Promise.resolve({ data: undefined })),
  },
}));

import api from "../client";
import {
  listStackConnections,
  createStackConnection,
  updateStackConnection,
  deleteStackConnection,
} from "../connections";

const ORG = "org1";
const TEAM = "default";
const STACK = "stack1";
const base = `/organizations/${ORG}/teams/${TEAM}/stacks/${STACK}/connections`;

describe("connections api client", () => {
  beforeEach(() => vi.clearAllMocks());

  it("lists connections at the team-scoped stack path", async () => {
    await listStackConnections(ORG, TEAM, STACK);
    expect(api.get).toHaveBeenCalledWith(base);
  });

  it("creates a connection with POST to the collection path", async () => {
    const conn = { kind: "env" as const, from: { type: "secret" as const, id: "s1" }, to: { type: "stack_resource" as const, name: "web" } };
    await createStackConnection(ORG, TEAM, STACK, conn);
    expect(api.post).toHaveBeenCalledWith(base, conn);
  });

  it("updates a connection with PUT to the item path", async () => {
    const conn = { id: "c1", kind: "env" as const, from: { type: "secret" as const, id: "s1" }, to: { type: "stack_resource" as const, name: "web" } };
    await updateStackConnection(ORG, TEAM, STACK, "c1", conn);
    expect(api.put).toHaveBeenCalledWith(`${base}/c1`, conn);
  });

  it("deletes a connection with DELETE to the item path", async () => {
    await deleteStackConnection(ORG, TEAM, STACK, "c1");
    expect(api.delete).toHaveBeenCalledWith(`${base}/c1`);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test:run src/api/tests/connections.test.ts`
Expected: FAIL — `Failed to resolve import "../connections"`.

- [ ] **Step 3: Write minimal implementation**

```typescript
// frontend/src/api/connections.ts
import api from "./client";
import type { components } from "./types/openapi";

export type StackConnection = components["schemas"]["StackConnection"];
export type StackConnectionList = components["schemas"]["StackConnectionList"];
export type ConnectionMapping = components["schemas"]["ConnectionMapping"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];
export type StackConnectionConfig = components["schemas"]["StackConnectionConfig"];

// Connections are team-scoped, nested under a stack. The UI scopes to the default team.
const base = (orgId: string, teamName: string, stackId: string) =>
  `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/connections`;

export async function listStackConnections(
  orgId: string,
  teamName: string,
  stackId: string,
): Promise<StackConnectionList> {
  const res = await api.get(base(orgId, teamName, stackId));
  return res.data as StackConnectionList;
}

export async function createStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connection: StackConnection,
): Promise<StackConnection> {
  const res = await api.post(base(orgId, teamName, stackId), connection);
  return res.data as StackConnection;
}

export async function updateStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connectionId: string,
  connection: StackConnection,
): Promise<StackConnection> {
  const res = await api.put(`${base(orgId, teamName, stackId)}/${connectionId}`, connection);
  return res.data as StackConnection;
}

export async function deleteStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connectionId: string,
): Promise<void> {
  await api.delete(`${base(orgId, teamName, stackId)}/${connectionId}`);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm test:run src/api/tests/connections.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/connections.ts frontend/src/api/tests/connections.test.ts
git commit -m "feat(stacks): add team-scoped stack connections API client"
```

---

## Task 2: Addon output field presets

**Files:**
- Create: `frontend/src/pages/stacks/lib/addon-presets.ts`

This file was deleted in `8891fee`. Recreate it with **backend output names** (not the old `connectionString`/`caCertificate` labels).

- [ ] **Step 1: Write the implementation directly** (a constant table — no behavior to TDD; it is exercised by Task 3's tests)

```typescript
// frontend/src/pages/stacks/lib/addon-presets.ts

// Postgres addon output accessors, exactly as the backend exposes them
// (pkg/models/output_descriptor.go). These are the `value.output` strings
// written into an addon env connection's mappings.
export const ADDON_OUTPUT_FIELDS = [
  "host",
  "port",
  "database",
  "username",
  "password",
  "sslmode",
  "ca_certificate",
  "url",
] as const;

export type AddonOutputField = (typeof ADDON_OUTPUT_FIELDS)[number];

// Fields that are identical across every database in the addon (cluster-wide),
// shown with a "cluster" hint in pickers.
export const CLUSTER_WIDE_FIELDS: ReadonlySet<AddonOutputField> = new Set<AddonOutputField>([
  "host",
  "port",
  "sslmode",
  "ca_certificate",
]);
```

- [ ] **Step 2: Type-check**

Run: `pnpm exec tsc -b`
Expected: no new errors referencing `addon-presets.ts`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/stacks/lib/addon-presets.ts
git commit -m "feat(stacks): add postgres addon output field presets"
```

---

## Task 3: Connection-mapping module — output accessors

**Files:**
- Create: `frontend/src/pages/stacks/lib/connection-mapping.ts`
- Test: `frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts`

Start with the secret output accessor (the trickiest piece — it must byte-match the backend).

- [ ] **Step 1: Write the failing test**

```typescript
// frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts
import { describe, it, expect } from "vitest";
import { secretOutputAccessor, parseSecretOutput } from "../connection-mapping";

describe("secretOutputAccessor", () => {
  it("uses dot form for simple keys", () => {
    expect(secretOutputAccessor("LOCKBOX_MASTER_KEY")).toBe("key.LOCKBOX_MASTER_KEY");
    expect(secretOutputAccessor("a1_B2")).toBe("key.a1_B2");
  });

  it("uses bracket form for keys with special characters", () => {
    expect(secretOutputAccessor("my-key")).toBe("key['my-key']");
    expect(secretOutputAccessor("has space")).toBe("key['has space']");
    expect(secretOutputAccessor("1leading")).toBe("key['1leading']");
  });

  it("escapes single quotes and backslashes in bracket form", () => {
    expect(secretOutputAccessor("a'b")).toBe("key['a\\'b']");
    expect(secretOutputAccessor("a\\b")).toBe("key['a\\\\b']");
  });
});

describe("parseSecretOutput", () => {
  it("reverses the dot form", () => {
    expect(parseSecretOutput("key.LOCKBOX_MASTER_KEY")).toBe("LOCKBOX_MASTER_KEY");
  });
  it("reverses the bracket form, unescaping", () => {
    expect(parseSecretOutput("key['my-key']")).toBe("my-key");
    expect(parseSecretOutput("key['a\\'b']")).toBe("a'b");
    expect(parseSecretOutput("key['a\\\\b']")).toBe("a\\b");
  });
  it("returns null for unrecognized accessors", () => {
    expect(parseSecretOutput("host")).toBeNull();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: FAIL — `Failed to resolve import "../connection-mapping"`.

- [ ] **Step 3: Write minimal implementation**

```typescript
// frontend/src/pages/stacks/lib/connection-mapping.ts

const SIMPLE_KEY = /^[A-Za-z_][A-Za-z0-9_]*$/;

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
  if (output.startsWith("key.")) return output.slice(4);
  const m = output.match(/^key\['(.*)'\]$/s);
  if (!m) return null;
  return m[1].replace(/\\(['\\])/g, "$1");
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/connection-mapping.ts frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts
git commit -m "feat(stacks): add secret output accessor encode/decode"
```

---

## Task 4: Connection-mapping module — rows → {envVars, connections}

**Files:**
- Modify: `frontend/src/pages/stacks/lib/connection-mapping.ts`
- Test: `frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts`

This task introduces the `FormEnvRow` shape used by the mapping module. It is the **same** discriminated union the form schema will adopt in Task 6; define it here as the canonical type and re-export it from the schema later.

- [ ] **Step 1: Write the failing test** (append to the existing test file)

```typescript
import {
  splitEnvRows,
  type FormEnvRow,
} from "../connection-mapping";

describe("splitEnvRows", () => {
  it("emits literal rows as env vars with value", () => {
    const rows: FormEnvRow[] = [{ from: "stack", name: "NODE_ENV", value: "production" }];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([{ name: "NODE_ENV", value: "production" }]);
    expect(connections).toEqual([]);
  });

  it("emits self rows as env vars with self_output", () => {
    const rows: FormEnvRow[] = [{ from: "self", name: "TOOLJET_HOST", selfOutput: "public.http.url" }];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([{ name: "TOOLJET_HOST", self_output: "public.http.url" }]);
    expect(connections).toEqual([]);
  });

  it("groups all secret rows for one secret into a single connection", () => {
    const rows: FormEnvRow[] = [
      { from: "secret", name: "LOCKBOX_MASTER_KEY", secretId: "s1", secretKey: "LOCKBOX_MASTER_KEY" },
      { from: "secret", name: "SECRET_KEY_BASE", secretId: "s1", secretKey: "SECRET_KEY_BASE" },
    ];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([]);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "web" },
    });
    expect(connections[0].config).toBeUndefined();
    expect(connections[0].mappings).toEqual([
      { target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "key.LOCKBOX_MASTER_KEY" } },
      { target: { type: "env", name: "SECRET_KEY_BASE" }, value: { output: "key.SECRET_KEY_BASE" } },
    ]);
  });

  it("groups 5 addon fields for one (addon, database) into a single connection with config", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
      { from: "addon", name: "PG_PORT", addonId: "a1", database: "tooljet", superuser: false, credField: "port" },
      { from: "addon", name: "PG_USER", addonId: "a1", database: "tooljet", superuser: false, credField: "username" },
      { from: "addon", name: "PG_PASS", addonId: "a1", database: "tooljet", superuser: false, credField: "password" },
      { from: "addon", name: "PG_DB", addonId: "a1", database: "tooljet", superuser: false, credField: "database" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "tooljet", superuser: false },
    });
    expect(connections[0].mappings).toHaveLength(5);
    expect(connections[0].mappings![0]).toEqual({
      target: { type: "env", name: "PG_HOST" },
      value: { output: "host" },
    });
  });

  it("splits addon rows that differ by database into separate connections", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "A_HOST", addonId: "a1", database: "db1", superuser: false, credField: "host" },
      { from: "addon", name: "B_HOST", addonId: "a1", database: "db2", superuser: false, credField: "host" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(2);
  });

  it("groups resource rows for one source resource into a single connection", () => {
    const rows: FormEnvRow[] = [
      { from: "resource", name: "SMTP_DOMAIN", resourceName: "mailhog", output: "host" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "stack_resource", name: "mailhog" },
      to: { type: "stack_resource", name: "web" },
    });
    expect(connections[0].config).toBeUndefined();
    expect(connections[0].mappings).toEqual([
      { target: { type: "env", name: "SMTP_DOMAIN" }, value: { output: "host" } },
    ]);
  });

  it("uses superuser config and omits database when superuser is true", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "PG_URL", addonId: "a1", superuser: true, credField: "url" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections[0].config).toEqual({ superuser: true });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: FAIL — `splitEnvRows` not exported.

- [ ] **Step 3: Write minimal implementation** (append to `connection-mapping.ts`)

```typescript
import type { StackConnection, ConnectionMapping } from "@/api/connections";
import type { components } from "@/api/types/openapi";
import type { AddonOutputField } from "@/pages/stacks/lib/addon-presets";

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
  const groups = new Map<string, StackConnection>();

  const ensureGroup = (key: string, conn: () => StackConnection): StackConnection => {
    let g = groups.get(key);
    if (!g) {
      g = conn();
      g.mappings = [];
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
        g.mappings!.push(envMapping(row.name, secretOutputAccessor(row.secretKey)));
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
        g.mappings!.push(envMapping(row.name, row.credField));
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
        g.mappings!.push(envMapping(row.name, row.output));
        break;
      }
    }
  }

  const connections = [...groups.values()].filter((c) => (c.mappings?.length ?? 0) > 0);
  return { envVars, connections };
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: PASS (all groups).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/connection-mapping.ts frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts
git commit -m "feat(stacks): map env rows to grouped stack connections"
```

---

## Task 5: Connection-mapping module — connections → rows, and diff

**Files:**
- Modify: `frontend/src/pages/stacks/lib/connection-mapping.ts`
- Test: `frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts`

- [ ] **Step 1: Write the failing test** (append)

```typescript
import {
  connectionsToEnvRows,
  buildDesiredConnections,
  diffConnections,
} from "../connection-mapping";
import type { StackConnection } from "@/api/connections";

describe("connectionsToEnvRows", () => {
  it("expands a secret connection into secret rows", () => {
    const conns: StackConnection[] = [{
      id: "c1", kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "key.LOCKBOX_MASTER_KEY" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "secret", name: "LOCKBOX_MASTER_KEY", secretId: "s1", secretKey: "LOCKBOX_MASTER_KEY" },
    ]);
  });

  it("expands an addon connection into addon rows using config database", () => {
    const conns: StackConnection[] = [{
      id: "c2", kind: "env",
      from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "tooljet", superuser: false },
      mappings: [{ target: { type: "env", name: "PG_HOST" }, value: { output: "host" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
    ]);
  });

  it("expands a resource connection into resource rows", () => {
    const conns: StackConnection[] = [{
      id: "c3", kind: "env",
      from: { type: "stack_resource", name: "mailhog" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "env", name: "SMTP_DOMAIN" }, value: { output: "host" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "resource", name: "SMTP_DOMAIN", resourceName: "mailhog", output: "host" },
    ]);
  });

  it("ignores connections whose `to` is a different resource", () => {
    const conns: StackConnection[] = [{
      id: "c4", kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "other" },
      mappings: [{ target: { type: "env", name: "X" }, value: { output: "key.X" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([]);
  });
});

describe("buildDesiredConnections", () => {
  it("collects connections across every resource", () => {
    const resources = [
      { name: "web", rows: [{ from: "secret", name: "X", secretId: "s1", secretKey: "X" }] as FormEnvRow[] },
      { name: "api", rows: [{ from: "resource", name: "H", resourceName: "web", output: "host" }] as FormEnvRow[] },
    ];
    expect(buildDesiredConnections(resources)).toHaveLength(2);
  });
});

describe("diffConnections", () => {
  const secretConn = (id: string | undefined, names: string[]): StackConnection => ({
    id, kind: "env",
    from: { type: "secret", id: "s1" },
    to: { type: "stack_resource", name: "web" },
    mappings: names.map((n) => ({ target: { type: "env", name: n }, value: { output: `key.${n}` } })),
  });

  it("creates a desired connection with no loaded match", () => {
    const { creates, updates, deletes } = diffConnections([], [secretConn(undefined, ["A"])]);
    expect(creates).toHaveLength(1);
    expect(updates).toEqual([]);
    expect(deletes).toEqual([]);
  });

  it("deletes a loaded connection with no desired match", () => {
    const { creates, updates, deletes } = diffConnections([secretConn("c1", ["A"])], []);
    expect(deletes).toEqual(["c1"]);
    expect(creates).toEqual([]);
    expect(updates).toEqual([]);
  });

  it("updates when mappings change within the same source group, carrying the loaded id", () => {
    const { creates, updates, deletes } = diffConnections(
      [secretConn("c1", ["A"])],
      [secretConn(undefined, ["A", "B"])],
    );
    expect(creates).toEqual([]);
    expect(deletes).toEqual([]);
    expect(updates).toHaveLength(1);
    expect(updates[0].id).toBe("c1");
    expect(updates[0].mappings).toHaveLength(2);
  });

  it("emits nothing when desired equals loaded", () => {
    const { creates, updates, deletes } = diffConnections([secretConn("c1", ["A"])], [secretConn(undefined, ["A"])]);
    expect(creates).toEqual([]);
    expect(updates).toEqual([]);
    expect(deletes).toEqual([]);
  });

  it("treats different addon databases as distinct groups", () => {
    const a = (db: string): StackConnection => ({
      kind: "env", from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" }, config: { database: db, superuser: false },
      mappings: [{ target: { type: "env", name: "H" }, value: { output: "host" } }],
    });
    const { creates, deletes } = diffConnections([{ ...a("db1"), id: "c1" }], [a("db2")]);
    expect(creates).toHaveLength(1);
    expect(deletes).toEqual(["c1"]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: FAIL — `connectionsToEnvRows`/`buildDesiredConnections`/`diffConnections` not exported.

- [ ] **Step 3: Write minimal implementation** (append)

```typescript
import { ADDON_OUTPUT_FIELDS } from "@/pages/stacks/lib/addon-presets";

const ADDON_FIELD_SET = new Set<string>(ADDON_OUTPUT_FIELDS);

// Identity of a connection source-group: same identity => same edge, diffed by
// mappings. Mirrors the backend discriminator (from + to + config db/superuser).
function connectionIdentity(c: StackConnection): string {
  const from = c.from?.type === "stack_resource"
    ? `stack_resource:${c.from?.name ?? ""}`
    : `${c.from?.type}:${c.from?.id ?? ""}`;
  const to = `${c.to?.type}:${c.to?.name ?? ""}`;
  const cfg = c.config as { database?: string; superuser?: boolean } | undefined;
  const config = cfg ? `${cfg.database ?? ""}:${cfg.superuser ?? false}` : "";
  return `${from}->${to}|${config}`;
}

// Stable signature of a connection's mappings, for change detection.
function mappingsSignature(c: StackConnection): string {
  return (c.mappings ?? [])
    .map((m) => `${m.target?.name ?? ""}=${m.value?.output ?? ""}`)
    .sort()
    .join("|");
}

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
    const cfg = c.config as { database?: string; superuser?: boolean } | undefined;
    for (const m of c.mappings ?? []) {
      const envName = m.target?.name ?? "";
      const output = m.value?.output ?? "";
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
          credField: output as FormEnvRow extends { credField?: infer T } ? T : never,
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

// Diff desired vs loaded connections at source-group granularity.
export function diffConnections(
  loaded: StackConnection[],
  desired: StackConnection[],
): { creates: StackConnection[]; updates: StackConnection[]; deletes: string[] } {
  const loadedByIdentity = new Map<string, StackConnection>();
  for (const c of loaded) loadedByIdentity.set(connectionIdentity(c), c);

  const creates: StackConnection[] = [];
  const updates: StackConnection[] = [];
  const seen = new Set<string>();

  for (const d of desired) {
    const identity = connectionIdentity(d);
    seen.add(identity);
    const match = loadedByIdentity.get(identity);
    if (!match) {
      creates.push(d);
    } else if (mappingsSignature(match) !== mappingsSignature(d)) {
      updates.push({ ...d, id: match.id });
    }
  }

  const deletes: string[] = [];
  for (const c of loaded) {
    if (!seen.has(connectionIdentity(c)) && c.id) deletes.push(c.id);
  }

  return { creates, updates, deletes };
}
```

> Note on the `credField` cast: the `as FormEnvRow extends ...` trick keeps `tsc` honest without `any`. If it proves awkward under the repo's TS config, replace it with `output as import("@/pages/stacks/lib/addon-presets").AddonOutputField` (the membership is already guaranteed by the `ADDON_FIELD_SET.has(output)` guard above).

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm test:run src/pages/stacks/lib/tests/connection-mapping.test.ts`
Expected: PASS (all mapping tests).

- [ ] **Step 5: Type-check, then commit**

Run: `pnpm exec tsc -b`
Expected: no new errors in `connection-mapping.ts`.

```bash
git add frontend/src/pages/stacks/lib/connection-mapping.ts frontend/src/pages/stacks/lib/tests/connection-mapping.test.ts
git commit -m "feat(stacks): expand connections to rows and diff desired vs loaded"
```

---

## Task 6: Form schema — restore the 5-arm union + literal/self serialization

**Files:**
- Modify: `frontend/src/pages/stacks/schemas/form-schema.ts`
- Test: `frontend/src/pages/stacks/schemas/tests/form-schema.test.ts`

The form schema keeps owning literal/self env-var serialization. Secret/addon/resource rows are validated in the union but are **dropped** from the resource API payload (they persist via connections — Task 8). Deserialization builds literal/self rows from env vars; connection-derived rows are merged in by the caller (Task 8).

- [ ] **Step 1: Replace `FormEnvVarSchema`** (lines 20-26 of current `form-schema.ts`) with the 5-arm union:

```typescript
import { ADDON_OUTPUT_FIELDS } from "@/pages/stacks/lib/addon-presets";
// (add alongside the existing imports at the top of the file)

const FormEnvVarSchema = z.union([
  z.object({
    from: z.literal("stack"),
    name: z.string().min(1, "Required"),
    value: z.string(),
  }),
  z.object({
    from: z.literal("secret"),
    name: z.string().min(1, "Required"),
    secretId: z.string().min(1, "Pick a secret"),
    secretKey: z.string().min(1, "Pick a key"),
  }),
  z
    .object({
      from: z.literal("addon"),
      name: z.string().min(1, "Required"),
      addonId: z.string().min(1, "Pick an addon"),
      database: z.string().optional(),
      superuser: z.boolean().default(false),
      credField: z.enum(ADDON_OUTPUT_FIELDS).optional(),
    })
    .refine((d) => d.superuser || (typeof d.database === "string" && d.database.length > 0), {
      message: "Pick a database",
      path: ["database"],
    })
    .refine((d) => typeof d.credField === "string" && d.credField.length > 0, {
      message: "Pick a field",
      path: ["credField"],
    }),
  z.object({
    from: z.literal("resource"),
    name: z.string().min(1, "Required"),
    resourceName: z.string().min(1, "Pick a resource"),
    output: z.string().min(1, "Pick an output"),
  }),
  z.object({
    from: z.literal("self"),
    name: z.string().min(1, "Required"),
    selfOutput: z.string().min(1, "Pick an output"),
  }),
]);
```

- [ ] **Step 2: Replace the serializer env-var block in `convertFormResourceToApiResource`** (current lines 214-224) so only literal + self rows become env vars:

```typescript
  // Only literal (stack) and self rows persist as env vars. Secret/addon/resource
  // rows persist as stack connections, handled in the save orchestration.
  const envVars = (rest.execution_config?.environment_variables ?? []) as FormEnvVarData[];

  const apiEnvVars = envVars.flatMap((r) => {
    if (r.from === "stack") return [{ name: r.name, value: r.value }];
    if (r.from === "self") return [{ name: r.name, self_output: r.selfOutput }];
    return [];
  });

  const processedExecutionConfig = rest.execution_config
    ? {
      ...rest.execution_config,
      environment_variables: apiEnvVars,
    }
    : undefined;
```

- [ ] **Step 3: Replace the deserializer env-var block in `convertApiResourceToFormResource`** (current lines 290-297) so env vars fan out into literal/self rows:

```typescript
  // Env vars deserialize into literal + self rows. Secret/addon/resource rows
  // are merged in by the caller from the stack's connections.
  const processedEnvVars: FormEnvVarData[] = (
    resource.execution_config?.environment_variables ?? []
  ).map((v) =>
    v.self_output
      ? { from: "self" as const, name: v.name, selfOutput: v.self_output }
      : { from: "stack" as const, name: v.name, value: v.value ?? "" },
  );
```

- [ ] **Step 4: Recover and adapt the form-schema tests**

Recover the pre-revert test file, then adapt it: secret/addon assertions that checked `environment_variables_from_secret` / `env_from_addons` no longer apply (those rows are dropped from the resource payload now). Keep the literal-row tests, add self-row tests.

```bash
git show 8891fee~1:frontend/src/pages/stacks/schemas/tests/form-schema.test.ts > frontend/src/pages/stacks/schemas/tests/form-schema.test.ts
```

Then edit the recovered file:
- Delete any `describe`/`it` blocks asserting `environment_variables_from_secret` or `env_from_addons` on the converted resource (search for those identifiers).
- Add this self-row round-trip test:

```typescript
it("serializes a self row into a self_output env var and back", () => {
  const form = {
    name: "web",
    sourceType: "image" as const,
    image_spec: { image: "nginx" },
    execution_config: {
      environment_variables: [{ from: "self" as const, name: "URL", selfOutput: "public.http.url" }],
    },
  };
  const api = convertFormResourceToApiResource(form as never);
  expect(api.execution_config?.environment_variables).toEqual([
    { name: "URL", self_output: "public.http.url" },
  ]);
  const back = convertApiResourceToFormResource(api as never);
  expect(back.execution_config?.environment_variables).toEqual([
    { from: "self", name: "URL", selfOutput: "public.http.url" },
  ]);
});

it("drops secret/addon/resource rows from the resource payload", () => {
  const form = {
    name: "web",
    sourceType: "image" as const,
    image_spec: { image: "nginx" },
    execution_config: {
      environment_variables: [
        { from: "stack" as const, name: "A", value: "1" },
        { from: "secret" as const, name: "B", secretId: "s1", secretKey: "B" },
        { from: "addon" as const, name: "C", addonId: "a1", database: "d", superuser: false, credField: "host" as const },
        { from: "resource" as const, name: "D", resourceName: "other", output: "host" },
      ],
    },
  };
  const api = convertFormResourceToApiResource(form as never);
  expect(api.execution_config?.environment_variables).toEqual([{ name: "A", value: "1" }]);
});
```

- [ ] **Step 5: Run tests to verify they fail, then pass**

Run: `pnpm test:run src/pages/stacks/schemas/tests/form-schema.test.ts`
Expected: after Steps 1-3 are in place, the new tests PASS and no deleted-field assertions remain.

- [ ] **Step 6: Type-check, then commit**

Run: `pnpm exec tsc -b`
Expected: no new errors in `form-schema.ts`.

```bash
git add frontend/src/pages/stacks/schemas/form-schema.ts frontend/src/pages/stacks/schemas/tests/form-schema.test.ts
git commit -m "feat(stacks): restore env source union; serialize literal+self rows"
```

---

## Task 7: EnvRow component — source selector + sub-pickers

**Files:**
- Replace: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Create: `frontend/src/pages/stacks/components/shared/env-addon-group.tsx`
- Replace: `frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx`

- [ ] **Step 1: Recover the pre-revert components**

```bash
git show 8891fee~1:frontend/src/pages/stacks/components/shared/env-row.tsx > frontend/src/pages/stacks/components/shared/env-row.tsx
git show 8891fee~1:frontend/src/pages/stacks/components/shared/env-addon-group.tsx > frontend/src/pages/stacks/components/shared/env-addon-group.tsx
```

- [ ] **Step 2: Adapt `env-row.tsx` imports** — the old file imports `CRED_FIELDS`, `CLUSTER_WIDE_FIELDS`, `type CredField` from `addon-presets`. Rename to the new presets:

Replace:
```typescript
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  type CredField,
} from "@/pages/stacks/lib/addon-presets";
```
With:
```typescript
import {
  ADDON_OUTPUT_FIELDS,
  CLUSTER_WIDE_FIELDS,
  type AddonOutputField,
} from "@/pages/stacks/lib/addon-presets";
```

Then, throughout `env-row.tsx`, replace every `CRED_FIELDS` identifier with `ADDON_OUTPUT_FIELDS` and every `CredField` type with `AddonOutputField` (the `AddonCredFieldPicker` component and the `AddonBindingPatch` type both reference them).

- [ ] **Step 3: Add the Resource and Self arms to the From selector**

In `env-row.tsx`, the `<SelectContent>` for the From selector currently lists Stack / Secret / Addon. Add Resource and Self:

```tsx
            <SelectContent>
              <SelectItem value="stack">Stack</SelectItem>
              <SelectItem value="secret">Secret</SelectItem>
              <SelectItem value="addon">Addon</SelectItem>
              <SelectItem value="resource">Resource</SelectItem>
              <SelectItem value="self">Self</SelectItem>
            </SelectContent>
```

Extend `EnvRowProps` with the data needed for the new pickers and add render arms in the Value cell. Add these props to the `EnvRowProps` interface:

```typescript
  /** Sibling resources (excluding this one) for the Resource source picker. */
  resourceOptions?: { name: string; outputs: string[] }[];
  /** This resource's own declared outputs for the Self source picker. */
  selfOutputs?: string[];
  onChangeResource?: (resourceName: string, output: string) => void;
  onChangeSelf?: (selfOutput: string) => void;
```

In the Value cell (`col-span-6`), after the existing `row.from === "addon"` block, add:

```tsx
          {row.from === "resource" && (
            <ResourceOutputCell
              resourceOptions={resourceOptions ?? []}
              resourceName={row.resourceName}
              output={row.output}
              onChange={(rn, out) => onChangeResource?.(rn, out)}
            />
          )}
          {row.from === "self" && (
            <SelfOutputCell
              outputs={selfOutputs ?? []}
              selfOutput={row.selfOutput}
              onChange={(out) => onChangeSelf?.(out)}
            />
          )}
```

And add the two small cell components at the bottom of the file (next to `SecretValueCell`):

```tsx
function ResourceOutputCell({
  resourceOptions,
  resourceName,
  output,
  onChange,
}: {
  resourceOptions: { name: string; outputs: string[] }[];
  resourceName: string;
  output: string;
  onChange: (resourceName: string, output: string) => void;
}) {
  const selected = resourceOptions.find((r) => r.name === resourceName);
  const outputs = selected?.outputs ?? [];
  return (
    <div className="space-y-2">
      <Select
        value={resourceName || ""}
        onValueChange={(v) => onChange(v, "")}
        disabled={resourceOptions.length === 0}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={resourceOptions.length === 0 ? "No other resources" : "select resource..."} />
        </SelectTrigger>
        <SelectContent>
          {resourceOptions.map((r) => (
            <SelectItem key={r.name} value={r.name}>{r.name}</SelectItem>
          ))}
        </SelectContent>
      </Select>
      {resourceName && (
        <Select value={output || ""} onValueChange={(v) => onChange(resourceName, v)} disabled={outputs.length === 0}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder={outputs.length === 0 ? "No outputs" : "select output..."} />
          </SelectTrigger>
          <SelectContent>
            {outputs.map((o) => (<SelectItem key={o} value={o}>{o}</SelectItem>))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}

function SelfOutputCell({
  outputs,
  selfOutput,
  onChange,
}: {
  outputs: string[];
  selfOutput: string;
  onChange: (selfOutput: string) => void;
}) {
  return (
    <Select value={selfOutput || ""} onValueChange={onChange} disabled={outputs.length === 0}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder={outputs.length === 0 ? "No outputs declared" : "select output..."} />
      </SelectTrigger>
      <SelectContent>
        {outputs.map((o) => (<SelectItem key={o} value={o}>{o}</SelectItem>))}
      </SelectContent>
    </Select>
  );
}
```

- [ ] **Step 4: Adapt `env-addon-group.tsx`** — same identifier rename: `CRED_FIELDS` → `ADDON_OUTPUT_FIELDS`, `CredField` → `AddonOutputField`. No behavioral change.

- [ ] **Step 5: Recover and adapt the EnvRow test**

```bash
git show 8891fee~1:frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx > frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx
```

In the recovered test, replace `CRED_FIELDS` references with `ADDON_OUTPUT_FIELDS` and update any assertion that selected `connectionString`/`caCertificate` to use `url`/`ca_certificate`. Add one test for the Self arm:

```tsx
it("renders self outputs and reports the chosen one", () => {
  const onChangeSelf = vi.fn();
  render(
    <EnvRow
      row={{ from: "self", name: "URL", selfOutput: "" }}
      index={0}
      resourceIndex={0}
      selfOutputs={["public.http.url", "host"]}
      onChangeName={() => {}}
      onChangeValue={() => {}}
      onChangeFrom={() => {}}
      onChangeSecret={() => {}}
      onChangeAddon={() => {}}
      onChangeSelf={onChangeSelf}
      onRemove={() => {}}
    />,
  );
  // open the self-output select and pick the first option
  // (use the same Radix pointer-event helper the recovered test already imports)
});
```

> If the recovered test uses a Radix `Select` interaction helper, reuse it verbatim for the self picker. If it instead asserts on rendered option text, assert that `public.http.url` appears.

- [ ] **Step 6: Run the EnvRow tests**

Run: `pnpm test:run src/pages/stacks/components/shared/tests/env-row.test.tsx`
Expected: PASS.

- [ ] **Step 7: Type-check, then commit**

Run: `pnpm exec tsc -b`
Expected: `env-row.tsx` / `env-addon-group.tsx` errors only where Task 8 still needs to pass the new props (acceptable mid-stack; the env-tab in Task 8 supplies them). If `tsc` blocks, proceed to Task 8 and type-check at its end.

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/src/pages/stacks/components/shared/env-addon-group.tsx frontend/src/pages/stacks/components/shared/tests/env-row.test.tsx
git commit -m "feat(stacks): restore env-row source selector; add resource+self arms"
```

---

## Task 8: Wire env-tab + save flow (deserialize merge + connection diff)

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-environment-tab.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx`

### 8a — Environment tab: offer all source types + pass picker data

- [ ] **Step 1: Extend the env-tab props** to receive what the pickers need. In `StackResourceEnvironmentTabProps` add:

```typescript
  /** Sibling resources (name + declared outputs) for the Resource source picker. */
  resourceOptions: { name: string; outputs: string[] }[];
  /** This resource's own declared outputs for the Self source picker. */
  selfOutputs: string[];
  secrets: import("@/api/secrets").Secret[];
  secretsLoading: boolean;
  addonNameById: Map<string, string>;
```

- [ ] **Step 2: Replace the single "Add variable" button** (current lines 294-299) with a small menu that can add any source type. Use the existing dropdown primitive already used elsewhere in the file's imports if present; otherwise add buttons inline:

```tsx
      <div className="mt-2 flex flex-wrap gap-2">
        <Button variant="ghost" size="sm" onClick={() => addEnvVar({ from: "stack", name: "", value: "" })}>
          <PlusCircle className="h-4 w-4 mr-2" />Add variable
        </Button>
        <Button variant="ghost" size="sm" onClick={() => addEnvVar({ from: "secret", name: "", secretId: "", secretKey: "" })}>
          From secret
        </Button>
        <Button variant="ghost" size="sm" onClick={() => addEnvVar({ from: "addon", name: "", addonId: "", superuser: false })}>
          From addon
        </Button>
        <Button variant="ghost" size="sm" onClick={() => addEnvVar({ from: "resource", name: "", resourceName: "", output: "" })}>
          From resource
        </Button>
        <Button variant="ghost" size="sm" onClick={() => addEnvVar({ from: "self", name: "", selfOutput: "" })}>
          From self output
        </Button>
      </div>
```

- [ ] **Step 3: Pass the new props through to `<EnvRow>`** in the map (current lines 273-291). Add the secret/addon/resource/self handlers:

```tsx
            <EnvRow
              key={envIdx}
              row={env}
              index={envIdx}
              resourceIndex={index}
              secrets={secrets}
              secretsLoading={secretsLoading}
              addonNameById={addonNameById}
              resourceOptions={resourceOptions}
              selfOutputs={selfOutputs}
              rowErrors={rowErrorsForIndex(envIdx)}
              status={envRowStatuses[envIdx] ?? "unchanged"}
              onReset={onDiscardEnvRow ? () => onDiscardEnvRow(envIdx) : undefined}
              onChangeName={(name) => replaceEnvVar(envIdx, { ...env, name })}
              onChangeValue={(value) => replaceEnvVar(envIdx, { ...env, value } as FormEnvVarData)}
              onChangeFrom={(from) => replaceEnvVar(envIdx, freshRowForSource(from, env.name))}
              onChangeSecret={(secretId, secretKey) =>
                replaceEnvVar(envIdx, { from: "secret", name: env.name, secretId, secretKey })}
              onChangeAddon={(patch) =>
                replaceEnvVar(envIdx, applyAddonPatch(env, patch))}
              onChangeResource={(resourceName, output) =>
                replaceEnvVar(envIdx, { from: "resource", name: env.name, resourceName, output })}
              onChangeSelf={(selfOutput) =>
                replaceEnvVar(envIdx, { from: "self", name: env.name, selfOutput })}
              onBlur={() => markEnvRowDirty(envIdx)}
              onRemove={() => removeEnvVar(envIdx)}
            />
```

Add the two small helpers above the `return` in the impl:

```tsx
  const freshRowForSource = (from: FormEnvVarData["from"], name: string): FormEnvVarData => {
    switch (from) {
      case "secret": return { from: "secret", name, secretId: "", secretKey: "" };
      case "addon": return { from: "addon", name, addonId: "", superuser: false };
      case "resource": return { from: "resource", name, resourceName: "", output: "" };
      case "self": return { from: "self", name, selfOutput: "" };
      default: return { from: "stack", name, value: "" };
    }
  };

  const applyAddonPatch = (
    env: FormEnvVarData,
    patch: import("./env-row").AddonBindingPatch,
  ): FormEnvVarData => {
    const base = env.from === "addon" ? env : { from: "addon" as const, name: env.name, addonId: "", superuser: false };
    return {
      ...base,
      addonId: patch.addonId ?? base.addonId,
      database: patch.database === null ? undefined : patch.database ?? base.database,
      superuser: patch.superuser ?? base.superuser,
      credField: patch.credField ?? base.credField,
    };
  };
```

> The addon **database** picker: the recovered `env-row` only renders the cred-field picker; database/addon selection comes through `onChangeAddon` patches driven by the addon panel/group. For this plan's scope, the database is supplied when a row is seeded from an addon connection (deserialize) or chosen via the addon group. A bare "From addon" row with no database fails validation ("Pick a database") until bound — acceptable, matches the old UX. A follow-up may add an inline database dropdown sourced from `usePostgresAddons` spec databases.

### 8b — Detail page: merge connection rows on load

- [ ] **Step 4: Merge connection-derived rows into each baseline resource.** In `detail/index.tsx`, locate where `baselineResources` is built (the `useMemo` around line 148 mapping `stackToShow.spec.stack_resources`). Wrap each mapped resource so connection rows are appended to its env vars:

```typescript
import { connectionsToEnvRows } from "@/pages/stacks/lib/connection-mapping";
// (add to imports)

  const baselineResources = useMemo<FormStackResourceData[]>(() => {
    const connections = stackToShow?.spec?.connections ?? [];
    return (stackToShow?.spec?.stack_resources || []).map((r) => {
      const form = convertApiResourceToFormResource(r);
      const connRows = connectionsToEnvRows(form.name, connections);
      if (connRows.length === 0) return form;
      return {
        ...form,
        execution_config: {
          ...(form.execution_config ?? {}),
          environment_variables: [
            ...((form.execution_config?.environment_variables ?? []) as FormEnvVarData[]),
            ...connRows,
          ],
        },
      };
    });
  }, [stackToShow]);
```

> Match the exact current shape of the `baselineResources` memo when editing — the surrounding mapping helper name (`convertApiResourceToFormResource` vs a local `mapResourceToFormData`) must be whatever the file already uses. Keep that, only add the `connRows` append.

### 8c — Detail page: connection diff after stack PUT

- [ ] **Step 5: Run the connection diff in `performSave`.** In `detail/index.tsx`, after the existing `const updatedStack = await updateStack(orgId, teamName, id, apiData);` (current line 309), add the connection reconciliation:

```typescript
import {
  buildDesiredConnections,
  diffConnections,
} from "@/pages/stacks/lib/connection-mapping";
import {
  createStackConnection,
  updateStackConnection,
  deleteStackConnection,
} from "@/api/connections";
// (add to imports)
```

```typescript
      // Reconcile connections (secret/addon/resource env bindings) after the
      // stack PUT. StackUpdateRequest carries no connections, so they are diffed
      // and applied via the dedicated connections API.
      const desired = buildDesiredConnections(
        resources.map((r) => ({
          name: r.name ?? "",
          rows: (r.execution_config?.environment_variables ?? []) as FormEnvRow[],
        })),
      );
      const loadedConnections = stackToShow.spec?.connections ?? [];
      const { creates, updates, deletes } = diffConnections(loadedConnections, desired);

      const connectionResults = await Promise.allSettled([
        ...creates.map((c) => createStackConnection(orgId, teamName, id, c)),
        ...updates.map((c) => updateStackConnection(orgId, teamName, id, c.id!, c)),
        ...deletes.map((cid) => deleteStackConnection(orgId, teamName, id, cid)),
      ]);
      const connectionFailures = connectionResults.filter((r) => r.status === "rejected").length;

      // Re-fetch so the panel and baseline reflect the new connection set.
      const refreshed = await getStackById(orgId, teamName, id);
      setFetchedStack(refreshed);

      if (connectionFailures > 0) {
        toast({
          title: "Stack saved, but some bindings failed",
          description: `${connectionFailures} connection change(s) did not apply. Re-open the resource to retry.`,
          variant: "destructive",
        });
      }
```

Then replace the existing `setFetchedStack(updatedStack);` line that immediately followed the PUT with nothing (the re-fetch above supersedes it). Import `FormEnvRow` as a type from `@/pages/stacks/lib/connection-mapping`.

- [ ] **Step 6: Pass the env-tab its new props** wherever `<StackResourceEnvironmentTab>` is rendered (in `stack-resource-detail.tsx` or `detail/index.tsx` — follow the existing render site). Supply:
  - `secrets` / `secretsLoading` from `useSecrets()`
  - `addonNameById` from `usePostgresAddons()` (build `new Map(addons.map(a => [a.id!, a.name!]))`)
  - `resourceOptions` = the other draft resources mapped to `{ name, outputs: (r.outputs ?? []).map(o => o.name) }`
  - `selfOutputs` = this resource's `(outputs ?? []).map(o => o.name)`

> `StackResource.outputs` is a readonly `OutputDescriptor[]` returned by the API; use `resource.outputs?.map(o => o.name) ?? []`. For brand-new unsaved resources with no computed outputs yet, the Self/Resource pickers show "No outputs declared" until the stack is saved once — acceptable.

- [ ] **Step 7: Type-check and run the full stacks test suite**

Run: `pnpm exec tsc -b`
Expected: zero new errors in the touched files.

Run: `pnpm test:run src/pages/stacks`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/shared/stack-resource-environment-tab.tsx frontend/src/pages/stacks/components/detail/index.tsx frontend/src/pages/stacks/components/shared/stack-resource-detail.tsx
git commit -m "feat(stacks): wire env source bindings through the connections API"
```

---

## Task 9: Live verification against the ToolJet stack

**Goal:** prove the round-trip end-to-end against the running app — no code, a manual/Playwright check. The local env and the interconnected ToolJet stack (addon + secret + mailhog + self) already exist from the earlier session.

- [ ] **Step 1: Ensure the env is up**

```bash
mage run            # API on :8000 (separate terminal / background)
pnpm --prefix frontend dev   # Vite on :5173
```

- [ ] **Step 2: Load + read-back (deserialize)** — drive `http://localhost:5173` with the Playwright MCP. Open the `tooljet` stack → `tooljet` resource → Environment tab. Confirm rows render with their sources: addon rows (`PG_HOST` ← host, …), secret rows (`LOCKBOX_MASTER_KEY` ← key), the mailhog resource row (`SMTP_DOMAIN` ← host), and the self row (`TOOLJET_HOST` ← public output). No literal duplication of bound vars.

- [ ] **Step 3: Edit + save (diff)** — change one secret row's key, add one new addon field row, remove one resource row. Save Changes. Capture the network panel: assert the expected `POST`/`PUT`/`DELETE …/connections` calls fire with `2xx`/`201`, and the stack `PUT` returns `200`.

- [ ] **Step 4: Reload (persistence)** — refresh; confirm the edited bindings persist and match what was saved. Verify against the DB:

```bash
psql "postgres://postgres:foobar-bizz-buzz@localhost:5432/stackdome_dev" \
  -c "select id, kind, (config->>'database') db, jsonb_array_length(mappings) n from stack_connections order by id;"
```

Expected: connection rows reflect the edit (changed mapping, new field, removed group).

- [ ] **Step 5: Record findings** — note pass/fail per step inline. Any failure → a finding, fix via TDD in the relevant module (re-enter the matching task), not an ad-hoc patch.

---

## Self-Review

**Spec coverage:**
- Restore env-centric From selector with all 5 sources → Tasks 6 (union), 7 (UI). ✓
- Secret/Addon/Resource → connections; Literal/Self → env vars → Tasks 4, 6. ✓
- Output accessors (addon field names, secret `key.<K>`/`key['<K>']`, resource outputs) → Tasks 2, 3, 4. ✓
- Persistence: create via `spec.connections`, update via diff CRUD → create path already rides `spec.connections` (unchanged); update path Task 8c. ✓
- New `api/connections.ts` + `lib/connection-mapping.ts` (pure, unit-tested) → Tasks 1, 3-5. ✓
- Diff at connection/source-group granularity, identity = from+to+config → Task 5 `connectionIdentity`. ✓
- Best-effort-atomic save + partial-failure toast → Task 8c `Promise.allSettled` + toast. ✓
- Out of scope (volume/build connections, topology view, backend) → respected. ✓
- Testing: round-trip, grouping, accessor edge cases, diff, EnvRow switch, live ToolJet → Tasks 3-7, 9. ✓

**Note on the create path:** New-stack creation already serializes connections into `spec.connections` via the existing create flow; this plan does not alter stack creation. If the new-stack form must also let users author secret/addon/resource rows pre-save, the same env-tab (Task 8a) is reused there and `buildDesiredConnections` feeds `spec.connections` on the create payload — wire it at the create form's submit if that surface is in scope for the milestone. Verify the create form's submit handler builds `spec.connections` from `buildDesiredConnections`; if it doesn't yet, add a follow-up task mirroring Task 8c's `buildDesiredConnections` call into the create payload.

**Type consistency:** `FormEnvRow` (mapping module) and `FormEnvVarData` (schema) describe the same union; Task 6 re-exports the schema-inferred type and Task 4/8 use `FormEnvRow` — keep them structurally identical (same arm field names: `secretId`/`secretKey`, `addonId`/`database`/`superuser`/`credField`, `resourceName`/`output`, `selfOutput`). `ADDON_OUTPUT_FIELDS`/`AddonOutputField`/`CLUSTER_WIDE_FIELDS` are the single naming used in Tasks 2, 6, 7. Client fn names `listStackConnections`/`createStackConnection`/`updateStackConnection`/`deleteStackConnection` are consistent across Tasks 1 and 8.

**Placeholder scan:** No TBD/TODO. Every code step shows code; every command shows expected output. The two "follow-up may add" notes (inline addon database dropdown; create-form wiring) are explicitly scoped as optional follow-ups, not gaps in the core round-trip.

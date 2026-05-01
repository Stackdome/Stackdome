# Stack ↔ Addon Env Linkage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users link a Postgres addon's credentials into a stack resource as environment variables, with lossless round-trip through the form, a count badge on the collapsed resource header, and clear handling of orphaned addon references.

**Architecture:** Replace the existing per-row `useSecret` boolean with a `from` discriminator (`stack | secret | addon`) on a discriminated-union form schema. Rewrite the form ↔ API converter so addon rows fan out from / regroup back to `env_from_addons[]`. Add an `Add from addon` mass-action dialog that handles the 1:N expansion (one addon source produces many env-var rows). Read addons via the existing `usePostgresAddons` hook.

**Tech Stack:** React 19, TypeScript, Zod (discriminated union), shadcn UI primitives (Dialog, Select), Vitest + @testing-library/react for tests.

---

## File Structure

**Create:**
- `frontend/src/pages/stacks/lib/addon-presets.ts` — credential field list, default env names, preset definitions. Pure data + small helper.
- `frontend/src/pages/stacks/components/shared/env-row.tsx` — renders one row in the env table; switches on `row.from`. Replaces ~120 lines of inline JSX in `stack-resource-item.tsx`.
- `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx` — the mass-action dialog.
- `frontend/__tests__/stacks-env-roundtrip.test.ts` — vitest covering the converter both directions, fan-out, regroup, orphan passthrough.
- `frontend/__tests__/add-from-addon-dialog.test.tsx` — component test covering presets, validation, superuser hiding, single-database auto-select.
- `frontend/__tests__/addon-presets.test.ts` — small unit tests for the preset library.

**Modify:**
- `frontend/src/pages/stacks/schemas/form-schema.ts` — replace `FormEnvVarSchema` with a `z.discriminatedUnion("from", […])`. Rewrite both converters. Drop the boolean `useSecret`/`selectedSecretId`/`selectedSecretKey` fields.
- `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` — delegate row rendering to `<EnvRow>`, add the `Add from addon` toolbar button, add the addon-count badge to the collapsed accordion trigger header, wire up the dialog state.

**Untouched:**
- `frontend/src/api/stacks.ts` — no changes; existing PUT/GET cover the request shape.
- Backend — no changes.

---

## Task 1: Addon presets library

**Files:**
- Create: `frontend/src/pages/stacks/lib/addon-presets.ts`
- Test: `frontend/__tests__/addon-presets.test.ts`

- [ ] **Step 1: Write the failing test**

Create `frontend/__tests__/addon-presets.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  DEFAULT_ENV_NAMES,
  applyPreset,
} from "../src/pages/stacks/lib/addon-presets";

describe("CRED_FIELDS", () => {
  it("lists all 8 fields the backend supports", () => {
    expect(CRED_FIELDS).toEqual([
      "host",
      "port",
      "username",
      "password",
      "database",
      "sslmode",
      "connectionString",
      "caCertificate",
    ]);
  });
});

describe("CLUSTER_WIDE_FIELDS", () => {
  it("contains the four cluster-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("host")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("port")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("sslmode")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("caCertificate")).toBe(true);
  });

  it("does not contain database-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("username")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("password")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("database")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("connectionString")).toBe(false);
  });
});

describe("DEFAULT_ENV_NAMES", () => {
  it("provides defaults for every credential field", () => {
    expect(DEFAULT_ENV_NAMES.host).toBe("PG_HOST");
    expect(DEFAULT_ENV_NAMES.port).toBe("PG_PORT");
    expect(DEFAULT_ENV_NAMES.username).toBe("PG_USER");
    expect(DEFAULT_ENV_NAMES.password).toBe("PG_PASS");
    expect(DEFAULT_ENV_NAMES.database).toBe("PG_DB");
    expect(DEFAULT_ENV_NAMES.sslmode).toBe("PG_SSLMODE");
    expect(DEFAULT_ENV_NAMES.connectionString).toBe("DATABASE_URL");
    expect(DEFAULT_ENV_NAMES.caCertificate).toBe("PG_CA_CERT");
  });
});

describe("applyPreset", () => {
  it("postgres-conventions selects host/port/username/password/database with default names", () => {
    const result = applyPreset("postgres-conventions");
    expect([...result.selected].sort()).toEqual(
      ["database", "host", "password", "port", "username"].sort(),
    );
    expect(result.envNames.host).toBe("PG_HOST");
    expect(result.envNames.password).toBe("PG_PASS");
  });

  it("connection-string selects only connectionString as DATABASE_URL", () => {
    const result = applyPreset("connection-string");
    expect([...result.selected]).toEqual(["connectionString"]);
    expect(result.envNames.connectionString).toBe("DATABASE_URL");
  });

  it("clear returns empty selection", () => {
    const result = applyPreset("clear");
    expect(result.selected.size).toBe(0);
    expect(Object.keys(result.envNames)).toEqual([]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && pnpm test:run __tests__/addon-presets.test.ts
```

Expected: FAIL with module-not-found error for `../src/pages/stacks/lib/addon-presets`.

- [ ] **Step 3: Implement the library**

Create `frontend/src/pages/stacks/lib/addon-presets.ts`:

```ts
export const CRED_FIELDS = [
  "host",
  "port",
  "username",
  "password",
  "database",
  "sslmode",
  "connectionString",
  "caCertificate",
] as const;

export type CredField = (typeof CRED_FIELDS)[number];

export const CLUSTER_WIDE_FIELDS: ReadonlySet<CredField> = new Set<CredField>([
  "host",
  "port",
  "sslmode",
  "caCertificate",
]);

export const DEFAULT_ENV_NAMES: Record<CredField, string> = {
  host: "PG_HOST",
  port: "PG_PORT",
  username: "PG_USER",
  password: "PG_PASS",
  database: "PG_DB",
  sslmode: "PG_SSLMODE",
  connectionString: "DATABASE_URL",
  caCertificate: "PG_CA_CERT",
};

export type Preset = "postgres-conventions" | "connection-string" | "clear";

export interface PresetResult {
  selected: Set<CredField>;
  envNames: Partial<Record<CredField, string>>;
}

export function applyPreset(preset: Preset): PresetResult {
  switch (preset) {
    case "postgres-conventions":
      return {
        selected: new Set<CredField>([
          "host",
          "port",
          "username",
          "password",
          "database",
        ]),
        envNames: {
          host: "PG_HOST",
          port: "PG_PORT",
          username: "PG_USER",
          password: "PG_PASS",
          database: "PG_DB",
        },
      };
    case "connection-string":
      return {
        selected: new Set<CredField>(["connectionString"]),
        envNames: { connectionString: "DATABASE_URL" },
      };
    case "clear":
      return { selected: new Set<CredField>(), envNames: {} };
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd frontend && pnpm test:run __tests__/addon-presets.test.ts
```

Expected: PASS, all tests green.

- [ ] **Step 5: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/lib/addon-presets.ts frontend/__tests__/addon-presets.test.ts
git commit -m "feat(stacks): add addon credential preset library

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Form schema discriminated union + converter rewrite

**Files:**
- Modify: `frontend/src/pages/stacks/schemas/form-schema.ts`
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` (minimal compile-fix)
- Test: `frontend/__tests__/stacks-env-roundtrip.test.ts`

This is the round-trip-fix task. Changes the row schema, rewrites both converters, and minimally adapts the existing UI to the new field names so the tree still compiles. Full UI overhaul (dialog, badge, env-row component) lands in later tasks.

- [ ] **Step 1: Write the failing test**

Create `frontend/__tests__/stacks-env-roundtrip.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import {
  convertApiResourceToFormResource,
  convertFormStackToApiStack,
} from "../src/pages/stacks/schemas/form-schema";
import type { StackResource } from "@/api/stacks";

const TOOLJET_ADDON_ID = "57fa98c8-27ca-47a8-9761-15504d60d349";

const baseResource = (extras: Partial<StackResource["execution_config"]> = {}) => ({
  id: "r-1",
  stack_id: "s-1",
  name: "tooljet",
  image_spec: { image: "tooljet/tooljet-ce:latest" },
  execution_config: {
    args: ["npm", "run", "start:prod"],
    environment_variables: [],
    environment_variables_from_secret: [],
    env_from_addons: [],
    ...extras,
  },
});

describe("env round-trip", () => {
  it("loads a stack literal env var as a 'stack' row", () => {
    const r = baseResource({ environment_variables: [{ name: "PORT", value: "80" }] });
    const form = convertApiResourceToFormResource(r as any);
    expect(form.execution_config?.environment_variables).toHaveLength(1);
    const row = form.execution_config!.environment_variables![0];
    expect(row.from).toBe("stack");
    expect(row.name).toBe("PORT");
    expect((row as any).value).toBe("80");
  });

  it("loads a secret-backed env var as a 'secret' row", () => {
    const r = baseResource({
      environment_variables_from_secret: [
        { name: "STRIPE", secret_ref: { secret_id: "sec-1" }, key: "live" },
      ],
    });
    const form = convertApiResourceToFormResource(r as any);
    const row = form.execution_config!.environment_variables![0];
    expect(row.from).toBe("secret");
    expect((row as any).secretId).toBe("sec-1");
    expect((row as any).secretKey).toBe("live");
  });

  it("fans out one env_from_addons entry into one row per credField", () => {
    const r = baseResource({
      env_from_addons: [
        {
          postgres: {
            addon_id: TOOLJET_ADDON_ID,
            database: "tooljet",
            superuser: false,
            env_mapping: {
              host: "PG_HOST",
              port: "PG_PORT",
              username: "PG_USER",
            },
          },
        },
      ],
    });
    const form = convertApiResourceToFormResource(r as any);
    const addonRows = form.execution_config!.environment_variables!.filter(
      (r) => r.from === "addon",
    );
    expect(addonRows).toHaveLength(3);
    expect(addonRows.map((r) => r.name).sort()).toEqual(
      ["PG_HOST", "PG_PORT", "PG_USER"].sort(),
    );
    addonRows.forEach((r) => {
      expect((r as any).addonId).toBe(TOOLJET_ADDON_ID);
      expect((r as any).database).toBe("tooljet");
      expect((r as any).superuser).toBe(false);
      expect((r as any).addonType).toBe("postgres");
    });
  });

  it("regroups addon rows back into a single env_from_addons entry on save", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            ...baseResource(),
            execution_config: {
              environment_variables: [
                { from: "addon", name: "PG_HOST", addonType: "postgres", addonId: TOOLJET_ADDON_ID, database: "tooljet", superuser: false, credField: "host" },
                { from: "addon", name: "PG_PORT", addonType: "postgres", addonId: TOOLJET_ADDON_ID, database: "tooljet", superuser: false, credField: "port" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as any);
    const ec = api.spec.stack_resources[0].execution_config!;
    expect(ec.env_from_addons).toHaveLength(1);
    expect(ec.env_from_addons![0].postgres!.addon_id).toBe(TOOLJET_ADDON_ID);
    expect(ec.env_from_addons![0].postgres!.database).toBe("tooljet");
    expect(ec.env_from_addons![0].postgres!.env_mapping).toEqual({
      host: "PG_HOST",
      port: "PG_PORT",
    });
    expect(ec.environment_variables).toEqual([]);
  });

  it("emits two env_from_addons entries when same addon has rows for two databases", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            ...baseResource(),
            execution_config: {
              environment_variables: [
                { from: "addon", name: "PG_HOST",         addonType: "postgres", addonId: TOOLJET_ADDON_ID, database: "tooljet",    superuser: false, credField: "host" },
                { from: "addon", name: "TOOLJET_DB_HOST", addonType: "postgres", addonId: TOOLJET_ADDON_ID, database: "tooljet-db", superuser: false, credField: "host" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as any);
    const entries = api.spec.stack_resources[0].execution_config!.env_from_addons!;
    expect(entries).toHaveLength(2);
    const dbs = entries.map((e) => e.postgres!.database).sort();
    expect(dbs).toEqual(["tooljet", "tooljet-db"]);
  });

  it("omits database from the API entry when superuser=true", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            ...baseResource(),
            execution_config: {
              environment_variables: [
                { from: "addon", name: "PG_HOST", addonType: "postgres", addonId: TOOLJET_ADDON_ID, superuser: true, credField: "host" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as any);
    const pg = api.spec.stack_resources[0].execution_config!.env_from_addons![0].postgres!;
    expect(pg.superuser).toBe(true);
    expect(pg.database).toBeUndefined();
  });

  it("preserves all three lists when mixed", () => {
    const r = baseResource({
      environment_variables: [{ name: "NODE_ENV", value: "production" }],
      environment_variables_from_secret: [
        { name: "STRIPE", secret_ref: { secret_id: "sec-1" }, key: "live" },
      ],
      env_from_addons: [
        {
          postgres: {
            addon_id: TOOLJET_ADDON_ID,
            database: "tooljet",
            superuser: false,
            env_mapping: { host: "PG_HOST" },
          },
        },
      ],
    });
    const form = convertApiResourceToFormResource(r as any);
    const rows = form.execution_config!.environment_variables!;
    expect(rows.map((r) => r.from).sort()).toEqual(["addon", "secret", "stack"]);
  });

  it("drops addon groups whose mapping is empty on save", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            ...baseResource(),
            execution_config: {
              environment_variables: [
                { from: "stack", name: "X", value: "y" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as any);
    const ec = api.spec.stack_resources[0].execution_config!;
    expect(ec.env_from_addons ?? []).toHaveLength(0);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && pnpm test:run __tests__/stacks-env-roundtrip.test.ts
```

Expected: FAIL — most assertions fail (current schema lacks `from`; converters drop `env_from_addons`).

- [ ] **Step 3: Replace `FormEnvVarSchema` and rewrite converters**

Open `frontend/src/pages/stacks/schemas/form-schema.ts`. Replace lines 20-27 (the current `FormEnvVarSchema`) with the discriminated union:

```ts
import { CRED_FIELDS } from "@/pages/stacks/lib/addon-presets";

const FormEnvVarSchema = z.discriminatedUnion("from", [
  z.object({
    from: z.literal("stack"),
    name: z.string().min(1, "Environment variable name is required"),
    value: z.string(),
  }),
  z.object({
    from: z.literal("secret"),
    name: z.string().min(1, "Environment variable name is required"),
    secretId: z.string().min(1),
    secretKey: z.string().min(1),
  }),
  z.object({
    from: z.literal("addon"),
    name: z.string().min(1, "Environment variable name is required"),
    addonType: z.literal("postgres"),
    addonId: z.string().min(1),
    database: z.string().optional(),
    superuser: z.boolean().default(false),
    credField: z.enum(CRED_FIELDS),
  }),
]);

type FormEnvVarData = z.infer<typeof FormEnvVarSchema>;
```

Now rewrite the body of `convertFormResourceToApiResource` (around lines 196-216). Replace the `processedExecutionConfig` block with:

```ts
  const envVars = (rest.execution_config?.environment_variables ?? []) as FormEnvVarData[];

  const literalEnvs = envVars
    .filter((r): r is Extract<FormEnvVarData, { from: "stack" }> => r.from === "stack")
    .map((r) => ({ name: r.name, value: r.value }));

  const secretEnvs = envVars
    .filter((r): r is Extract<FormEnvVarData, { from: "secret" }> => r.from === "secret")
    .map((r) => ({
      name: r.name,
      secret_ref: { secret_id: r.secretId },
      key: r.secretKey,
    }));

  const groups = new Map<
    string,
    {
      addonId: string;
      database?: string;
      superuser: boolean;
      mapping: Record<string, string>;
    }
  >();
  for (const r of envVars) {
    if (r.from !== "addon") continue;
    const key = `${r.addonId}::${r.database ?? ""}::${r.superuser}`;
    let g = groups.get(key);
    if (!g) {
      g = {
        addonId: r.addonId,
        database: r.database,
        superuser: r.superuser,
        mapping: {},
      };
      groups.set(key, g);
    }
    g.mapping[r.credField] = r.name;
  }

  const env_from_addons = [...groups.values()]
    .filter((g) => Object.keys(g.mapping).length > 0)
    .sort((a, b) => {
      if (a.addonId !== b.addonId) return a.addonId.localeCompare(b.addonId);
      return (a.database ?? "").localeCompare(b.database ?? "");
    })
    .map((g) => ({
      postgres: {
        addon_id: g.addonId,
        ...(g.superuser ? {} : { database: g.database }),
        superuser: g.superuser,
        env_mapping: g.mapping,
      },
    }));

  const processedExecutionConfig = rest.execution_config
    ? {
        ...rest.execution_config,
        environment_variables: literalEnvs,
        environment_variables_from_secret: secretEnvs,
        env_from_addons,
      }
    : undefined;
```

Then rewrite the `processedEnvVars` block in `convertApiResourceToFormResource` (around lines 277-294). Replace it with:

```ts
  const literalRows: FormEnvVarData[] = (
    resource.execution_config?.environment_variables ?? []
  ).map((v) => ({
    from: "stack" as const,
    name: v.name,
    value: v.value,
  }));

  const secretRows: FormEnvVarData[] = (
    resource.execution_config?.environment_variables_from_secret ?? []
  ).map((v) => ({
    from: "secret" as const,
    name: v.name,
    secretId: v.secret_ref.secret_id,
    secretKey: v.key,
  }));

  const credOrderIndex = (f: string) =>
    CRED_FIELDS.indexOf(f as (typeof CRED_FIELDS)[number]);

  const addonRows: FormEnvVarData[] = (
    resource.execution_config?.env_from_addons ?? []
  ).flatMap((entry) => {
    const pg = entry.postgres;
    if (!pg) return [];
    return Object.entries(pg.env_mapping ?? {})
      .sort(([a], [b]) => credOrderIndex(a) - credOrderIndex(b))
      .map(([credField, envName]) => ({
        from: "addon" as const,
        name: envName,
        addonType: "postgres" as const,
        addonId: pg.addon_id,
        database: pg.database,
        superuser: pg.superuser ?? false,
        credField: credField as (typeof CRED_FIELDS)[number],
      }));
  });

  const processedEnvVars = [...literalRows, ...secretRows, ...addonRows];
```

Also remove the now-dead `FormEnvVarSchema`-bound fields from the resource conversion. The block in `convertApiResourceToFormResource` that returns the resource (around line 304) keeps `useImageSecret`, `selectedImageSecretId`, `useGitSecret`, `selectedGitSecretId` — leave those alone. Only the `processedEnvVars` shape changes.

Update the type export (around line 416) to also export `FormEnvVarData`:

```ts
export type {
  FormStackData,
  FormStackResourceData,
  FormVolumeData,
  FormVolumeExtendedData,
  FormEnvVarData,
};
```

- [ ] **Step 4: Patch `stack-resource-item.tsx` to compile against the new shape**

Open `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`. Three blocks reference the old fields. Update them as below — this is a minimal compile fix; full UI rework lands in Task 5.

Around line 116-127 (the `addEnvVar` helper that adds an empty row): change the pushed row to `{ from: "stack", name: "", value: "" }`.

Around lines 156-189 (the `addMultipleEnvVars` helper): the new rows produced from the paste dialog should be `{ from: "stack" as const, name, value }` (drop `useSecret`/`selectedSecretId`/`selectedSecretKey`).

Around lines 1130-1290 (where each row is rendered): for each row, branch on `row.from`. For `from === "stack"`, render the existing Key + Value text inputs. For `from === "secret"`, render the existing secret picker + key picker. For `from === "addon"`, render a temporary placeholder for now:

```tsx
<div className="text-xs text-muted-foreground italic px-3 py-2">
  ⚙ {row.addonId.slice(0, 8)}… · {row.database ?? "(superuser)"} · {row.credField}
</div>
```

Likewise update any `setUseSecret` / `setSelectedSecretId` / `setSelectedSecretKey` callbacks to write to `row.from`, `row.secretId`, `row.secretKey`. The toggle that flips between literal and secret becomes a flip between `{ from: "stack", name, value: "" }` and `{ from: "secret", name, secretId: "", secretKey: "" }`.

(Concrete edit hint: search for `useSecret` in the file. There are about a dozen occurrences. Each needs to be migrated to the discriminator. The `Use secret` toggle stays visually as today; only the underlying data shape changes.)

- [ ] **Step 5: Run test to verify it passes**

```bash
cd frontend && pnpm test:run __tests__/stacks-env-roundtrip.test.ts && pnpm tsc --noEmit
```

Expected: tests PASS, type-check clean.

- [ ] **Step 6: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/schemas/form-schema.ts frontend/src/pages/stacks/components/shared/stack-resource-item.tsx frontend/__tests__/stacks-env-roundtrip.test.ts
git commit -m "feat(stacks): round-trip env_from_addons through form schema

Replace per-row useSecret boolean with a 'from' discriminator
(stack | secret | addon) on a Zod discriminated union. Rewrite both
converters: load fans out env_from_addons entries into one form row
per credField; save groups them back by (addonId, database, superuser).

Closes the silent-drop bug where the form converter never read or
wrote env_from_addons, causing any stack consuming a Postgres addon
to lose its linkage on save.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Extract `<EnvRow>` component

Pulls the per-row JSX out of `stack-resource-item.tsx` so the file stops growing and the row variants are easier to reason about. No behavior change.

**Files:**
- Create: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/pages/stacks/components/shared/env-row.tsx`:

```tsx
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { Secret } from "@/api/secrets";

export type EnvFrom = FormEnvVarData["from"];

interface EnvRowProps {
  row: FormEnvVarData;
  index: number;
  resourceIndex: number;
  secrets: Secret[];
  secretsLoading: boolean;
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onChangeFrom: (from: EnvFrom) => void;
  onChangeSecret: (secretId: string, secretKey: string) => void;
  onRemove: () => void;
}

export function EnvRow({
  row,
  index,
  resourceIndex,
  secrets,
  secretsLoading,
  onChangeName,
  onChangeValue,
  onChangeFrom,
  onChangeSecret,
  onRemove,
}: EnvRowProps) {
  return (
    <div className="flex items-start gap-2 py-2 border-b">
      <Input
        id={`env-name-${resourceIndex}-${index}`}
        className="flex-1"
        placeholder="KEY"
        value={row.name}
        onChange={(e) => onChangeName(e.target.value)}
      />
      <div className="flex-[2]">
        {row.from === "stack" && (
          <Input
            placeholder="value"
            value={row.value}
            onChange={(e) => onChangeValue(e.target.value)}
          />
        )}
        {row.from === "secret" && (
          <SecretValueCell
            secrets={secrets}
            loading={secretsLoading}
            secretId={row.secretId}
            secretKey={row.secretKey}
            onChange={onChangeSecret}
          />
        )}
        {row.from === "addon" && (
          <AddonValueCell
            addonId={row.addonId}
            database={row.database}
            credField={row.credField}
            superuser={row.superuser}
          />
        )}
      </div>
      <Select
        value={row.from}
        onValueChange={(v) => onChangeFrom(v as EnvFrom)}
        disabled={row.from === "addon"}
      >
        <SelectTrigger className="w-[110px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="stack">Stack</SelectItem>
          <SelectItem value="secret">Secret</SelectItem>
          <SelectItem value="addon" disabled>
            Addon
          </SelectItem>
        </SelectContent>
      </Select>
      <Button variant="ghost" size="icon" onClick={onRemove} aria-label="Remove env var">
        <X className="h-4 w-4" />
      </Button>
    </div>
  );
}

function SecretValueCell({
  secrets,
  loading,
  secretId,
  secretKey,
  onChange,
}: {
  secrets: Secret[];
  loading: boolean;
  secretId: string;
  secretKey: string;
  onChange: (secretId: string, secretKey: string) => void;
}) {
  const selected = secrets.find((s) => s.id === secretId);
  const keys = selected?.data ? Object.keys(selected.data) : [];
  return (
    <div className="flex gap-2">
      <Select
        value={secretId}
        onValueChange={(id) => onChange(id, "")}
        disabled={loading}
      >
        <SelectTrigger className="flex-1">
          <SelectValue placeholder="Select secret" />
        </SelectTrigger>
        <SelectContent>
          {secrets.map((s) => (
            <SelectItem key={s.id} value={s.id!}>
              {s.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={secretKey}
        onValueChange={(k) => onChange(secretId, k)}
        disabled={!secretId || keys.length === 0}
      >
        <SelectTrigger className="flex-1">
          <SelectValue placeholder="Key" />
        </SelectTrigger>
        <SelectContent>
          {keys.map((k) => (
            <SelectItem key={k} value={k}>
              {k}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

function AddonValueCell({
  addonId,
  database,
  credField,
  superuser,
}: {
  addonId: string;
  database?: string;
  credField: string;
  superuser: boolean;
}) {
  const dbLabel = superuser ? "(superuser)" : database ?? "—";
  const shortId = addonId.slice(0, 8);
  return (
    <div className="text-sm text-muted-foreground italic px-2 py-1.5">
      ⚙ {shortId}… · {dbLabel} · {credField}
    </div>
  );
}
```

(Note: `Secret` type comes from `@/api/secrets` — confirm import name; it's whatever's currently used in `stack-resource-item.tsx`.)

- [ ] **Step 2: Replace inline row JSX in `stack-resource-item.tsx` with `<EnvRow>`**

Find the row-rendering block in the env tab (around lines 1130-1290). Replace the entire `.map(...)` body with:

```tsx
{(resource.execution_config?.environment_variables || []).map((row, envIdx) => (
  <EnvRow
    key={envIdx}
    row={row}
    index={envIdx}
    resourceIndex={index}
    secrets={secrets.userSecrets}
    secretsLoading={secrets.isLoading}
    onChangeName={(name) => updateEnvVar(envIdx, { ...row, name })}
    onChangeValue={(value) =>
      row.from === "stack" && updateEnvVar(envIdx, { ...row, value })
    }
    onChangeFrom={(from) => switchRowFrom(envIdx, from)}
    onChangeSecret={(secretId, secretKey) =>
      updateEnvVar(envIdx, {
        from: "secret",
        name: row.name,
        secretId,
        secretKey,
      })
    }
    onRemove={() => removeEnvVar(envIdx)}
  />
))}
```

Add the `updateEnvVar`, `switchRowFrom`, and `removeEnvVar` helpers at the top of the component if they aren't already present:

```ts
const updateEnvVar = (envIdx: number, next: FormEnvVarData) => {
  update({
    execution_config: {
      ...resource.execution_config,
      environment_variables: (resource.execution_config?.environment_variables || []).map(
        (r, i) => (i === envIdx ? next : r),
      ),
    },
  });
};

const switchRowFrom = (envIdx: number, from: EnvFrom) => {
  const current = resource.execution_config?.environment_variables?.[envIdx];
  if (!current) return;
  if (from === "stack") {
    updateEnvVar(envIdx, { from: "stack", name: current.name, value: "" });
  } else if (from === "secret") {
    updateEnvVar(envIdx, { from: "secret", name: current.name, secretId: "", secretKey: "" });
  }
  // 'addon' from here is a no-op; addon rows are added via the dialog only.
};

const removeEnvVar = (envIdx: number) => {
  update({
    execution_config: {
      ...resource.execution_config,
      environment_variables: (resource.execution_config?.environment_variables || []).filter(
        (_, i) => i !== envIdx,
      ),
    },
  });
};
```

Add the import at the top:

```ts
import { EnvRow, type EnvFrom } from "./env-row";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
```

- [ ] **Step 3: Verify type-check and tests pass**

```bash
cd frontend && pnpm tsc --noEmit && pnpm test:run __tests__/stacks-env-roundtrip.test.ts
```

Expected: clean.

- [ ] **Step 4: Manual smoke check**

```bash
cd frontend && pnpm dev
```

Open the running tooljet-addon stack edit page in the browser, open the Environment Variables tab on the `tooljet` resource, and verify:
- Literal rows render normally with editable name and value.
- Secret rows render with the two pickers.
- Addon rows render the placeholder caption (`⚙ <id>… · <db> · <field>`) and From column reads "Addon" and is disabled.

Stop the dev server.

- [ ] **Step 5: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/src/pages/stacks/components/shared/stack-resource-item.tsx
git commit -m "refactor(stacks): extract EnvRow component for env table

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Add-from-Addon dialog

Builds the mass-action dialog. Receives the addon list, the existing env-var names in the resource (for collision validation), and an `onAdd` callback that delivers the inserted rows.

**Files:**
- Create: `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx`
- Test: `frontend/__tests__/add-from-addon-dialog.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/__tests__/add-from-addon-dialog.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AddFromAddonDialog } from "../src/pages/stacks/components/shared/add-from-addon-dialog";
import type { PostgresAddon } from "@/api/addons";

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => ({
  id: "addon-1",
  name: "tooljet-db",
  status: { state: "Ready" },
  spec: {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "app" }],
    configuration: { enable_superuser_access: false },
  } as any,
  ...(over as any),
});

describe("AddFromAddonDialog", () => {
  it("auto-selects the only database when addon has one", async () => {
    const single = mkAddon({ spec: { ...mkAddon().spec, databases: [{ name: "only" }] } as any });
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[single]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    await screen.findByText(/only/i);
    expect(screen.getByText(/only/i)).toBeInTheDocument();
  });

  it("hides Superuser toggle when addon does not enable_superuser_access", () => {
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    expect(screen.queryByLabelText(/superuser/i)).not.toBeInTheDocument();
  });

  it("shows Superuser toggle when addon enables it", () => {
    const su = mkAddon({
      spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
    });
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[su]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    expect(screen.getByLabelText(/superuser/i)).toBeInTheDocument();
  });

  it("Postgres conventions preset ticks 5 fields with default names", async () => {
    const user = userEvent.setup();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    expect(screen.getByDisplayValue("PG_HOST")).toBeInTheDocument();
    expect(screen.getByDisplayValue("PG_PASS")).toBeInTheDocument();
    expect(screen.getByDisplayValue("PG_DB")).toBeInTheDocument();
  });

  it("blocks Add when an env name collides with existing env", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set(["PG_HOST"])}
        onAdd={onAdd}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    const addBtn = screen.getByRole("button", { name: /^add$/i });
    expect(addBtn).toBeDisabled();
    expect(screen.getByText(/already exists/i)).toBeInTheDocument();
  });

  it("calls onAdd with one row per ticked field on confirm", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={onAdd}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    await user.click(screen.getByRole("button", { name: /^add$/i }));
    expect(onAdd).toHaveBeenCalledTimes(1);
    const rows = onAdd.mock.calls[0][0];
    expect(rows).toHaveLength(5);
    rows.forEach((r: any) => {
      expect(r.from).toBe("addon");
      expect(r.addonType).toBe("postgres");
      expect(r.addonId).toBe("addon-1");
    });
    expect(rows.map((r: any) => r.credField).sort()).toEqual(
      ["database", "host", "password", "port", "username"].sort(),
    );
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && pnpm test:run __tests__/add-from-addon-dialog.test.tsx
```

Expected: FAIL — module not found.

- [ ] **Step 3: Implement the dialog**

Create `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx`:

```tsx
import { useEffect, useMemo, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  DEFAULT_ENV_NAMES,
  applyPreset,
  type CredField,
  type Preset,
} from "@/pages/stacks/lib/addon-presets";
import type { PostgresAddon } from "@/api/addons";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

interface AddFromAddonDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  addons: PostgresAddon[];
  existingEnvNames: Set<string>;
  onAdd: (rows: FormEnvVarData[]) => void;
}

export function AddFromAddonDialog({
  open,
  onOpenChange,
  addons,
  existingEnvNames,
  onAdd,
}: AddFromAddonDialogProps) {
  const [addonId, setAddonId] = useState<string>(addons[0]?.id ?? "");
  const [database, setDatabase] = useState<string>("");
  const [superuser, setSuperuser] = useState(false);
  const [selected, setSelected] = useState<Set<CredField>>(new Set());
  const [envNames, setEnvNames] = useState<Partial<Record<CredField, string>>>({});

  const addon = addons.find((a) => a.id === addonId);
  const databases = addon?.spec?.databases ?? [];
  const supportsSuperuser =
    addon?.spec?.configuration?.enable_superuser_access === true;

  // Auto-select database when there's only one
  useEffect(() => {
    if (databases.length === 1 && !database) {
      setDatabase(databases[0]?.name ?? "");
    }
  }, [databases, database]);

  // Reset state when addon switches
  useEffect(() => {
    setDatabase("");
    setSuperuser(false);
    setSelected(new Set());
    setEnvNames({});
  }, [addonId]);

  const toggleField = (field: CredField, on: boolean) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (on) next.add(field);
      else next.delete(field);
      return next;
    });
    setEnvNames((prev) => {
      if (on && !prev[field]) {
        return { ...prev, [field]: DEFAULT_ENV_NAMES[field] };
      }
      if (!on) {
        const next = { ...prev };
        delete next[field];
        return next;
      }
      return prev;
    });
  };

  const onPreset = (preset: Preset) => {
    const result = applyPreset(preset);
    setSelected(result.selected);
    setEnvNames(result.envNames);
  };

  const collisions = useMemo(() => {
    const out: string[] = [];
    selected.forEach((f) => {
      const name = envNames[f] ?? "";
      if (name && existingEnvNames.has(name)) out.push(name);
    });
    return out;
  }, [selected, envNames, existingEnvNames]);

  const validationError = useMemo(() => {
    if (!addon) return "Pick an addon.";
    if (!superuser && !database) return "Pick a database.";
    if (selected.size === 0) return "Tick at least one credential field.";
    for (const f of selected) {
      const name = envNames[f] ?? "";
      if (!name.trim()) return `Env name for "${f}" is empty.`;
    }
    if (collisions.length) {
      return `Env name "${collisions[0]}" already exists in this resource.`;
    }
    const names = [...selected].map((f) => envNames[f]);
    if (new Set(names).size !== names.length) {
      return "Two ticked fields share the same env name.";
    }
    return null;
  }, [addon, superuser, database, selected, envNames, collisions]);

  const onConfirm = () => {
    if (validationError) return;
    const rows: FormEnvVarData[] = [...selected].map((credField) => ({
      from: "addon",
      name: envNames[credField] ?? DEFAULT_ENV_NAMES[credField],
      addonType: "postgres",
      addonId,
      database: superuser ? undefined : database,
      superuser,
      credField,
    }));
    onAdd(rows);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>Add from Addon</DialogTitle>
          <DialogDescription>
            Inject credentials from an addon as environment variables.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label className="text-sm">Addon</Label>
            {addons.length === 0 ? (
              <div className="rounded-md border bg-muted/40 px-3 py-3 text-sm">
                <p className="text-muted-foreground mb-2">
                  No addons yet. Create a Postgres addon to inject its credentials here.
                </p>
                <a
                  href="/addons/create/postgres"
                  target="_blank"
                  rel="noreferrer"
                  className="text-primary underline"
                >
                  + Create Postgres addon
                </a>
              </div>
            ) : (
              <Select value={addonId} onValueChange={setAddonId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select an addon" />
                </SelectTrigger>
                <SelectContent>
                  {addons.map((a) => (
                    <SelectItem key={a.id} value={a.id!}>
                      {a.name} (Postgres · {a.status?.state ?? "Unknown"})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {!superuser && (
            <div>
              <Label className="text-sm">Database</Label>
              <Select value={database} onValueChange={setDatabase}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a database" />
                </SelectTrigger>
                <SelectContent>
                  {databases.map((d) => (
                    <SelectItem key={d.name} value={d.name!}>
                      {d.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {supportsSuperuser && (
            <div className="flex items-center gap-2">
              <Switch
                id="superuser-toggle"
                checked={superuser}
                onCheckedChange={setSuperuser}
              />
              <Label htmlFor="superuser-toggle" className="text-sm">
                Superuser
              </Label>
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <Label className="text-sm">Inject credentials</Label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="sm">
                    Apply preset ▾
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onClick={() => onPreset("postgres-conventions")}>
                    Postgres conventions
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onPreset("connection-string")}>
                    Connection string only
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onPreset("clear")}>
                    Clear
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div className="border rounded-md divide-y">
              {CRED_FIELDS.map((field) => (
                <div key={field} className="flex items-center gap-2 px-3 py-2">
                  <input
                    type="checkbox"
                    id={`cred-${field}`}
                    checked={selected.has(field)}
                    onChange={(e) => toggleField(field, e.target.checked)}
                  />
                  <Label htmlFor={`cred-${field}`} className="w-44 text-sm font-mono">
                    {field}
                  </Label>
                  <span className="text-muted-foreground">→</span>
                  <Input
                    value={envNames[field] ?? ""}
                    placeholder={DEFAULT_ENV_NAMES[field]}
                    disabled={!selected.has(field)}
                    onChange={(e) =>
                      setEnvNames((p) => ({ ...p, [field]: e.target.value }))
                    }
                    className="flex-1"
                  />
                  {CLUSTER_WIDE_FIELDS.has(field) && (
                    <span className="text-xs text-muted-foreground">cluster</span>
                  )}
                </div>
              ))}
            </div>

            {validationError ? (
              <p className="text-xs text-destructive mt-2">{validationError}</p>
            ) : (
              <p className="text-xs text-muted-foreground mt-2">
                Adds {selected.size} environment variable{selected.size === 1 ? "" : "s"}.
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={validationError !== null}>
            Add
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd frontend && pnpm test:run __tests__/add-from-addon-dialog.test.tsx
```

Expected: PASS, all 6 cases green. If a test fails because `Switch` or `DropdownMenu` aren't reachable via the queries used (these are Radix primitives), adjust the query (e.g., use `getByText` for menu items) but keep the assertion intent intact.

- [ ] **Step 5: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx frontend/__tests__/add-from-addon-dialog.test.tsx
git commit -m "feat(stacks): add 'Add from Addon' env-var dialog

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Wire dialog into the env tab + addon row rendering

Hooks the dialog up to the env tab, fetches addons via the existing hook, replaces the `EnvRow` addon placeholder with the proper pill caption, and adds the toolbar button.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`

- [ ] **Step 1: Replace the addon placeholder caption with the proper pill in `env-row.tsx`**

Open `frontend/src/pages/stacks/components/shared/env-row.tsx`. The component already takes `addons: PostgresAddon[]` — actually, it doesn't yet. Add a new optional prop `addonNameById: Map<string, string>` to render the human name instead of the truncated UUID:

```tsx
// Add to EnvRowProps:
addonNameById?: Map<string, string>;
```

And update `AddonValueCell` to take and render the name:

```tsx
function AddonValueCell({
  addonId,
  database,
  credField,
  superuser,
  addonName,
  isOrphan,
}: {
  addonId: string;
  database?: string;
  credField: string;
  superuser: boolean;
  addonName?: string;
  isOrphan?: boolean;
}) {
  const dbLabel = superuser ? "(superuser)" : database ?? "—";
  const display = isOrphan ? "<missing addon>" : addonName ?? `${addonId.slice(0, 8)}…`;
  return (
    <div
      className={`text-sm italic px-2 py-1.5 ${
        isOrphan ? "text-yellow-600" : "text-muted-foreground"
      }`}
    >
      ⚙ {display} · {dbLabel} · {credField}
    </div>
  );
}
```

Wire `addonNameById` in the parent invocation:

```tsx
{row.from === "addon" && (
  <AddonValueCell
    addonId={row.addonId}
    database={row.database}
    credField={row.credField}
    superuser={row.superuser}
    addonName={addonNameById?.get(row.addonId)}
    isOrphan={addonNameById !== undefined && !addonNameById.has(row.addonId)}
  />
)}
```

- [ ] **Step 2: Wire the dialog into `stack-resource-item.tsx`**

At the top of the component, fetch the addon list:

```ts
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import { AddFromAddonDialog } from "./add-from-addon-dialog";

const { addons } = usePostgresAddons();
const addonNameById = useMemo(
  () => new Map(addons.filter((a) => a.id).map((a) => [a.id!, a.name])),
  [addons],
);
const [addonDialogOpen, setAddonDialogOpen] = useState(false);
```

Find the env-tab toolbar (around line 1062-1086, the row with `Clear All` and `Paste Variables`). Add a third button between them and after `Paste Variables`:

```tsx
<Button
  variant="ghost"
  size="sm"
  className="gap-2"
  onClick={() => setAddonDialogOpen(true)}
>
  <Cog className="h-4 w-4" />
  <span>Add from addon</span>
</Button>
```

(Import `Cog` from `lucide-react` at the top.)

Then render the dialog at the end of the env tab body:

```tsx
<AddFromAddonDialog
  open={addonDialogOpen}
  onOpenChange={setAddonDialogOpen}
  addons={addons}
  existingEnvNames={new Set(
    (resource.execution_config?.environment_variables || []).map((r) => r.name),
  )}
  onAdd={(rows) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [
          ...(resource.execution_config?.environment_variables || []),
          ...rows,
        ],
      },
    });
  }}
/>
```

Pass `addonNameById` into the `EnvRow` invocations:

```tsx
<EnvRow
  /* existing props */
  addonNameById={addonNameById}
/>
```

- [ ] **Step 3: Verify type-check, tests, and manual smoke**

```bash
cd frontend && pnpm tsc --noEmit && pnpm test:run
```

Expected: clean. Then:

```bash
cd frontend && pnpm dev
```

Open the running tooljet-addon stack edit page. In the tooljet resource's Environment Variables tab:
- Confirm the addon-source rows now show `⚙ tooljet-db · tooljet · host` etc. (real addon name).
- Click `Add from addon`. Pick `tooljet-db`. Pick database `app`. Apply Postgres conventions. Click Add.
- Verify 5 new rows appear in the env list with `From: Addon` and addon-source captions.
- Save the stack. In DevTools, confirm the PUT body's `env_from_addons` contains both the original two entries (tooljet, tooljet-db) and the new entry for `app`.
- Stop the dev server.

- [ ] **Step 4: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/components/shared/stack-resource-item.tsx frontend/src/pages/stacks/components/shared/env-row.tsx
git commit -m "feat(stacks): wire Add from Addon dialog into env-vars tab

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Addon-count badge on the collapsed accordion header

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`

- [ ] **Step 1: Compute the unique-addon count**

At the top of the component (after the existing memo for `addonNameById`):

```ts
const addonCount = useMemo(() => {
  const ids = new Set<string>();
  (resource.execution_config?.environment_variables || []).forEach((r) => {
    if (r.from === "addon") ids.add(r.addonId);
  });
  return ids.size;
}, [resource.execution_config?.environment_variables]);
```

- [ ] **Step 2: Render the badge in the AccordionTrigger header**

Find the trigger header (the row that shows the resource name + image). Add the badge before the chevron, right-aligned:

```tsx
{addonCount > 0 && (
  <span className="ml-auto mr-2 inline-flex items-center gap-1 rounded-full border bg-muted/60 px-2 py-0.5 text-xs text-muted-foreground">
    <Cog className="h-3 w-3" />
    {addonCount} {addonCount === 1 ? "addon" : "addons"}
  </span>
)}
```

(Re-uses the `Cog` import from Task 5. If the existing trigger row layout is `flex items-center justify-between`, the `ml-auto` will push the badge to the right. Otherwise wrap as needed.)

- [ ] **Step 3: Manual smoke**

```bash
cd frontend && pnpm dev
```

On the running tooljet-addon stack, expand Stack Resources:
- The `tooljet` row shows `⚙ 1 addon` (it uses two databases of the same addon, but that's still one addon).
- `mailhog` and `redis` show no badge.
- Expand `tooljet`, click Add from addon, link a second distinct addon (create a temp Postgres if needed). Collapse — the badge updates to `⚙ 2 addons`.

Stop the dev server.

- [ ] **Step 4: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/components/shared/stack-resource-item.tsx
git commit -m "feat(stacks): show addon-count badge on collapsed resource header

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Orphan addon detection + warning UI

When a stack references an `addon_id` that no longer exists in the addon list, the row renders read-only with a warning. Save still emits the row as-is so the user can decide.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/stacks-env-roundtrip.test.ts` (extend with orphan-passthrough test)

- [ ] **Step 1: Extend the round-trip test for orphan passthrough**

Add to `frontend/__tests__/stacks-env-roundtrip.test.ts` inside the existing `describe("env round-trip", …)`:

```ts
  it("preserves an orphaned addon row through save", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            ...baseResource(),
            execution_config: {
              environment_variables: [
                { from: "addon", name: "PG_HOST", addonType: "postgres", addonId: "deleted-addon-id", database: "tooljet", superuser: false, credField: "host" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as any);
    const entries = api.spec.stack_resources[0].execution_config!.env_from_addons!;
    expect(entries).toHaveLength(1);
    expect(entries[0].postgres!.addon_id).toBe("deleted-addon-id");
    expect(entries[0].postgres!.env_mapping).toEqual({ host: "PG_HOST" });
  });
```

```bash
cd frontend && pnpm test:run __tests__/stacks-env-roundtrip.test.ts
```

Expected: PASS (the converter already handles this; this test is a regression guard).

- [ ] **Step 2: Render orphan rows with warning styling**

Update the `EnvRow` rendering to treat orphan addon rows as read-only:

```tsx
// In EnvRow, before the value cell:
const isOrphanAddon =
  row.from === "addon" &&
  addonNameById !== undefined &&
  !addonNameById.has(row.addonId);

return (
  <div
    className={`flex items-start gap-2 py-2 border-b ${
      isOrphanAddon ? "bg-yellow-500/5 border-l-4 border-l-yellow-500" : ""
    }`}
  >
    <Input
      /* existing key input */
      readOnly={isOrphanAddon}
    />
    <div className="flex-[2]">
      {/* existing value cells; AddonValueCell already shows the orphan style */}
    </div>
    <Select
      /* existing source picker */
      disabled={row.from === "addon" || isOrphanAddon}
    >
      {/* … */}
    </Select>
    <Button
      variant="ghost"
      size="icon"
      onClick={onRemove}
      aria-label="Remove env var"
    >
      <X className="h-4 w-4" />
    </Button>
    {isOrphanAddon && (
      <span className="absolute -bottom-4 left-0 text-xs text-yellow-700">
        Addon was deleted. This variable won't resolve. Remove to clean up.
      </span>
    )}
  </div>
);
```

(If positioning the warning beneath the row needs a wrapping `<div className="relative">…<warning>` structure, adjust accordingly. The visual goal is a small yellow caption directly under the row.)

- [ ] **Step 3: Manual orphan smoke**

```bash
cd frontend && pnpm dev
```

In the addons list, create a throwaway Postgres addon `temp-pg` (any sane defaults). Wait for Ready. Open a stack and add a `From: Addon` row referencing `temp-pg / postgres / host → PG_HOST`. Save. Then go back to addons and delete `temp-pg`. Reload the stack edit page.

Verify:
- The PG_HOST row renders with a yellow strip and the warning caption.
- Inputs are read-only; only the `×` works.
- Click `×`, save. Row goes away; collapsed badge updates.

Stop the dev server.

- [ ] **Step 4: Commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/stacks-env-roundtrip.test.ts
git commit -m "feat(stacks): warn on orphaned addon refs in env list

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Final verification

End-to-end manual run-through against the running platform.

- [ ] **Step 1: Type-check + full test run**

```bash
cd frontend && pnpm tsc --noEmit && pnpm lint && pnpm test:run
```

Expected: all green. Fix any drift.

- [ ] **Step 2: End-to-end manual scenario**

Start the platform if not already running, then exercise:

1. Open `http://localhost:5173/`, log in.
2. Navigate to the running `tooljet-addon` stack edit page.
3. Confirm the `tooljet` resource shows `⚙ 1 addon` badge on the collapsed header.
4. Expand `tooljet` → Environment Variables tab. Confirm the existing addon rows render with the addon name (`tooljet-db`), correct database, correct credField.
5. Save the stack with no changes. In DevTools Network tab, inspect the PUT body and verify `env_from_addons` contains the original two entries verbatim.
6. Click `Add from addon` → pick `tooljet-db` → pick database `app` → Apply Postgres conventions → Add.
7. Save. Verify the PUT body now has three `env_from_addons` entries (tooljet, tooljet-db, app).
8. Trigger an orphan: delete `tooljet-db` (won't succeed if stack still references it — surface the 409 from the addon delete page; that's expected). Stop here for orphan testing if you don't want to break the stack.

- [ ] **Step 3: Update the spec status**

Edit `docs/superpowers/specs/2026-05-01-stack-addon-env-linkage-design.md`:

```markdown
**Status:** Implemented (2026-05-01)
```

- [ ] **Step 4: Final commit**

```bash
cd /Users/akshaysasidharan/code/stackdome
git add docs/superpowers/specs/2026-05-01-stack-addon-env-linkage-design.md
git commit -m "docs: mark stack-addon env linkage spec as implemented

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Known Limitations Carried Forward

- **Save-time cross-row name uniqueness** (spec § Edge cases > "Same env-var name appears twice"): the dialog blocks insertion when an addon row collides with the existing list, but the form does not currently flag a collision created by typing two `Stack` rows with the same name (or by editing a literal row to match an addon row). Backend will accept the duplicate and one of the two will overwrite the other in the pod env. Tracked as follow-up; not blocking this plan.

## Self-Review Checklist (for the implementer)

Before marking this plan complete, run through:

- [ ] All 8 tasks committed.
- [ ] `pnpm tsc --noEmit` clean across the repo.
- [ ] `pnpm lint` clean.
- [ ] `pnpm test:run` all green; new test files appear.
- [ ] Manual: existing tooljet-addon stack opens, saves without losing addon links, badge shows correctly.
- [ ] Manual: Add-from-Addon dialog inserts rows; Postgres conventions preset works; collisions block save.
- [ ] Manual: orphan addon row renders with warning; save preserves the row.

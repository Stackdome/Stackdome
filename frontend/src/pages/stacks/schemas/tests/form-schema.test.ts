import { describe, it, expect } from "vitest";
import {
  convertApiResourceToFormResource,
  convertFormStackToApiStack,
} from "../form-schema";
import type { FormStackData } from "../form-schema";
import type { StackResource } from "@/api/stacks";

type ApiResourceArg = Parameters<typeof convertApiResourceToFormResource>[0];
type Loose = Record<string, unknown>;

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
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    expect(form.execution_config?.environment_variables).toHaveLength(1);
    const row = form.execution_config!.environment_variables![0];
    expect(row.from).toBe("stack");
    expect(row.name).toBe("PORT");
    expect((row as Loose).value).toBe("80");
  });

  it("loads a secret-backed env var as a 'secret' row", () => {
    const r = baseResource({
      environment_variables_from_secret: [
        { name: "STRIPE", secret_ref: { secret_id: "sec-1" }, key: "live" },
      ],
    });
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    const row = form.execution_config!.environment_variables![0];
    expect(row.from).toBe("secret");
    expect((row as Loose).secretId).toBe("sec-1");
    expect((row as Loose).secretKey).toBe("live");
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
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    const addonRows = form.execution_config!.environment_variables!.filter(
      (r) => r.from === "addon",
    );
    expect(addonRows).toHaveLength(3);
    expect(addonRows.map((r) => r.name).sort()).toEqual(
      ["PG_HOST", "PG_PORT", "PG_USER"].sort(),
    );
    addonRows.forEach((r) => {
      expect((r as Loose).addonId).toBe(TOOLJET_ADDON_ID);
      expect((r as Loose).database).toBe("tooljet");
      expect((r as Loose).superuser).toBe(false);
      expect((r as Loose).addonType).toBe("postgres");
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
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
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
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
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
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
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
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    const rows = form.execution_config!.environment_variables!;
    expect(rows.map((r) => r.from).sort()).toEqual(["addon", "secret", "stack"]);
  });

  it("preserves an orphaned addon row through save", () => {
    const formStack = {
      name: "s",
      labels: [],
      annotations: [],
      spec: {
        stack_resources: [
          {
            id: "r-1",
            stack_id: "s-1",
            name: "tooljet",
            image_spec: { image: "tooljet/tooljet-ce:latest" },
            execution_config: {
              environment_variables: [
                {
                  from: "addon",
                  name: "PG_HOST",
                  addonType: "postgres",
                  addonId: "deleted-addon-id",
                  database: "tooljet",
                  superuser: false,
                  credField: "host",
                },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
    const entries = api.spec.stack_resources[0].execution_config!.env_from_addons!;
    expect(entries).toHaveLength(1);
    expect(entries[0].postgres!.addon_id).toBe("deleted-addon-id");
    expect(entries[0].postgres!.env_mapping).toEqual({ host: "PG_HOST" });
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
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
    const ec = api.spec.stack_resources[0].execution_config!;
    expect(ec.env_from_addons ?? []).toHaveLength(0);
  });
});

describe("FormEnvVarSchema (addon variant) — refines", () => {
  it("requires database when superuser is false", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: undefined,
      superuser: false,
      credField: "host",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes("database"))).toBe(true);
    }
  });

  it("allows missing database when superuser is true", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: undefined,
      superuser: true,
      credField: "host",
    });
    expect(result.success).toBe(true);
  });

  it("requires credField on addon rows", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: "tooljet",
      superuser: false,
      credField: undefined,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes("credField"))).toBe(true);
    }
  });

  it("uses 'Pick an addon' message on empty addonId", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "",
      database: "tooljet",
      superuser: false,
      credField: "host",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(
        result.error.issues.find((i) => i.path.includes("addonId"))?.message,
      ).toMatch(/pick an addon/i);
    }
  });
});

describe("FormStackSchema — depends_on cross-validation", () => {
  /** Minimal valid stack resource skeleton — fields irrelevant to depends_on. */
  const stackResource = (over: { name: string; depends_on?: string[] }) => ({
    name: over.name,
    sourceType: "image" as const,
    useImageSecret: false,
    useGitSecret: false,
    image_spec: { image: "nginx:latest" },
    depends_on: over.depends_on,
  });

  const stackOf = (...resources: ReturnType<typeof stackResource>[]) => ({
    name: "demo",
    labels: [],
    spec: { stack_resources: resources, volumes: [] },
  });

  it("accepts depends_on that points to an existing resource", async () => {
    const { FormStackSchema } = await import("../form-schema");
    const result = FormStackSchema.safeParse(
      stackOf(
        stackResource({ name: "redis" }),
        stackResource({ name: "api", depends_on: ["redis"] }),
      ),
    );
    expect(result.success).toBe(true);
  });

  it("accepts an absent or empty depends_on (it is optional)", async () => {
    const { FormStackSchema } = await import("../form-schema");
    expect(
      FormStackSchema.safeParse(stackOf(stackResource({ name: "solo" }))).success,
    ).toBe(true);
    expect(
      FormStackSchema.safeParse(
        stackOf(stackResource({ name: "solo", depends_on: [] })),
      ).success,
    ).toBe(true);
  });

  it("rejects depends_on referencing an unknown resource", async () => {
    const { FormStackSchema } = await import("../form-schema");
    const result = FormStackSchema.safeParse(
      stackOf(
        stackResource({ name: "redis" }),
        stackResource({ name: "api", depends_on: ["ghost"] }),
      ),
    );
    expect(result.success).toBe(false);
    if (!result.success) {
      const issue = result.error.issues.find(
        (i) =>
          i.path[0] === "spec" &&
          i.path[1] === "stack_resources" &&
          i.path[2] === 1 &&
          i.path[3] === "depends_on" &&
          i.path[4] === 0,
      );
      expect(issue?.message).toMatch(/unknown resource/i);
    }
  });

  it("flags a dangling reference once its target is renamed", async () => {
    // Reproduces the bug the user hit on the edit page: rename a depended-upon
    // resource and the dependent's depends_on entry now points nowhere.
    const { FormStackSchema } = await import("../form-schema");
    const result = FormStackSchema.safeParse(
      stackOf(
        stackResource({ name: "redis-renamed" }),
        stackResource({ name: "api", depends_on: ["redis"] }),
      ),
    );
    expect(result.success).toBe(false);
  });

  it("emits one issue per dangling entry, not just the first", async () => {
    const { FormStackSchema } = await import("../form-schema");
    const result = FormStackSchema.safeParse(
      stackOf(
        stackResource({ name: "api", depends_on: ["ghost1", "ghost2"] }),
      ),
    );
    expect(result.success).toBe(false);
    if (!result.success) {
      const dangling = result.error.issues.filter(
        (i) =>
          i.path[0] === "spec" &&
          i.path[1] === "stack_resources" &&
          i.path[3] === "depends_on",
      );
      expect(dangling).toHaveLength(2);
    }
  });
});

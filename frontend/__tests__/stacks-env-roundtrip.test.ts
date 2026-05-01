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

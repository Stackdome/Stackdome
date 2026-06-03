import { describe, it, expect } from "vitest";
import {
  convertApiResourceToFormResource,
  convertFormStackToApiStack,
} from "../form-schema";
import type { FormStackData } from "../form-schema";
import type { StackResource } from "@/api/stacks";

type ApiResourceArg = Parameters<typeof convertApiResourceToFormResource>[0];
type Loose = Record<string, unknown>;

const baseResource = (extras: Partial<StackResource["execution_config"]> = {}) => ({
  id: "r-1",
  stack_id: "s-1",
  name: "tooljet",
  image_spec: { image: "tooljet/tooljet-ce:latest" },
  execution_config: {
    args: ["npm", "run", "start:prod"],
    environment_variables: [],
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

  it("loads multiple literal env vars preserving order", () => {
    const r = baseResource({
      environment_variables: [
        { name: "NODE_ENV", value: "production" },
        { name: "PORT", value: "80" },
      ],
    });
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    const rows = form.execution_config!.environment_variables!;
    expect(rows.map((r) => r.from)).toEqual(["stack", "stack"]);
    expect(rows.map((r) => r.name)).toEqual(["NODE_ENV", "PORT"]);
  });

  it("defaults a missing value to an empty string on load", () => {
    const r = baseResource({ environment_variables: [{ name: "FLAG" }] as never });
    const form = convertApiResourceToFormResource(r as unknown as ApiResourceArg);
    const row = form.execution_config!.environment_variables![0];
    expect((row as Loose).value).toBe("");
  });

  it("emits only literal environment_variables on save", () => {
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
                { from: "stack", name: "NODE_ENV", value: "production" },
                { from: "stack", name: "PORT", value: "80" },
              ],
            },
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(formStack as unknown as FormStackData);
    const ec = api.spec.stack_resources[0].execution_config! as Loose;
    expect(ec.environment_variables).toEqual([
      { name: "NODE_ENV", value: "production" },
      { name: "PORT", value: "80" },
    ]);
    // Removed feature — no secret/addon arrays are emitted anymore.
    expect(ec.environment_variables_from_secret).toBeUndefined();
    expect(ec.env_from_addons).toBeUndefined();
  });
});

describe("FormEnvVarSchema — literal rows", () => {
  it("accepts a literal stack row", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({ from: "stack", name: "PORT", value: "80" });
    expect(result.success).toBe(true);
  });

  it("requires a non-empty name", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({ from: "stack", name: "", value: "80" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes("name"))).toBe(true);
    }
  });

  it("rejects the removed secret variant", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "secret",
      name: "STRIPE",
      secretId: "sec-1",
      secretKey: "live",
    });
    expect(result.success).toBe(false);
  });

  it("rejects the removed addon variant", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: "tooljet",
      superuser: false,
      credField: "host",
    });
    expect(result.success).toBe(false);
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

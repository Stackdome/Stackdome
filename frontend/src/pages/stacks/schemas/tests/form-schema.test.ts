import { describe, it, expect } from "vitest";
import {
  convertApiResourceToFormResource,
  convertFormResourceToApiResource,
  convertFormStackToApiStack,
} from "../form-schema";

type ApiResourceArg = Parameters<typeof convertApiResourceToFormResource>[0];
type Loose = Record<string, unknown>;

const baseResource = (extras: Record<string, unknown> = {}) => ({
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

  it("defaults a missing value to an empty string on load", () => {
    const api = { name: "web", execution_config: { environment_variables: [{ name: "EMPTY" }] } };
    const form = convertApiResourceToFormResource(api as never);
    expect(form.execution_config?.environment_variables).toEqual([
      { from: "stack", name: "EMPTY", value: "" },
    ]);
  });

  it("round-trips a literal env var through serialize + deserialize", () => {
    const form = {
      name: "web",
      sourceType: "image" as const,
      image_spec: { image: "nginx" },
      execution_config: {
        environment_variables: [{ from: "stack" as const, name: "PORT", value: "80" }],
      },
    };
    const api = convertFormResourceToApiResource(form as never);
    expect(api.execution_config?.environment_variables).toEqual([
      { name: "PORT", value: "80" },
    ]);
    const back = convertApiResourceToFormResource(api as never);
    expect(back.execution_config?.environment_variables).toEqual([
      { from: "stack", name: "PORT", value: "80" },
    ]);
  });

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
});

describe("convertFormStackToApiStack — spec.connections", () => {
  it("emits spec.connections built from secret/addon/resource rows on convertFormStackToApiStack", () => {
    const form = {
      name: "tooljet",
      labels: [],
      spec: {
        stack_resources: [
          {
            name: "web",
            sourceType: "image" as const,
            image_spec: { image: "nginx" },
            execution_config: {
              environment_variables: [
                { from: "stack" as const, name: "A", value: "1" },
                { from: "secret" as const, name: "B", secretId: "s1", secretKey: "B" },
                { from: "resource" as const, name: "C", resourceName: "mailhog", output: "host" },
              ],
            },
          },
        ],
      },
    };
    const api = convertFormStackToApiStack(form as never);
    // literal stays an env var:
    expect(api.spec.stack_resources[0].execution_config?.environment_variables).toEqual([{ name: "A", value: "1" }]);
    // secret + resource become connections (2 groups, to web):
    expect(api.spec.connections).toBeDefined();
    expect(api.spec.connections!.map((c) => c.from)).toEqual(
      expect.arrayContaining([{ type: "secret", id: "s1" }, { type: "stack_resource", name: "mailhog" }]),
    );
    api.spec.connections!.forEach((c) => expect(c.to).toEqual({ type: "stack_resource", name: "web" }));
  });

  it("emits volume_mount connections from resource volume_mounts and strips the dead field", () => {
    const form = {
      name: "tooljet",
      labels: [],
      spec: {
        stack_resources: [
          {
            name: "db",
            sourceType: "image" as const,
            image_spec: { image: "postgres:16" },
            volume_mounts: [{ source_volume_name: "pgdata", target_path: "/var/lib/postgresql/data" }],
          },
        ],
        volumes: [{ name: "pgdata", spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } }],
      },
    };
    const api = convertFormStackToApiStack(form as never);
    expect(api.spec.connections).toEqual([
      {
        kind: "volume_mount",
        from: { type: "volume", name: "pgdata" },
        to: { type: "stack_resource", name: "db" },
        config: { mount_path: "/var/lib/postgresql/data" },
      },
    ]);
    // Mounts persist only as connections — the resource field is server-ignored.
    expect(api.spec.stack_resources[0].volume_mounts).toBeUndefined();
  });

  it("skips volume_mount connections whose volume is missing from the spec", () => {
    const form = {
      name: "tooljet",
      labels: [],
      spec: {
        stack_resources: [
          {
            name: "db",
            sourceType: "image" as const,
            image_spec: { image: "postgres:16" },
            volume_mounts: [{ source_volume_name: "ghost", target_path: "/data" }],
          },
        ],
        volumes: [],
      },
    };
    const api = convertFormStackToApiStack(form as never);
    expect(api.spec.connections).toBeUndefined();
  });

  it("omits spec.connections when there are no secret/addon/resource rows", () => {
    const form = {
      name: "s", labels: [],
      spec: { stack_resources: [{ name: "web", sourceType: "image" as const, image_spec: { image: "nginx" },
        execution_config: { environment_variables: [{ from: "stack" as const, name: "A", value: "1" }] } }] },
    };
    const api = convertFormStackToApiStack(form as never);
    expect(api.spec.connections).toBeUndefined();
  });
});

describe("FormEnvVarSchema (addon variant) — refines", () => {
  it("requires database when superuser is false", async () => {
    const { FormEnvVarSchema } = await import("../form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
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

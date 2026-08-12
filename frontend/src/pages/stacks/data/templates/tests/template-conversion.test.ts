import { describe, it, expect } from "vitest";
import { templates, getTemplateById } from "@/pages/stacks/data/templates/registry";
import { templateToFormData } from "@/pages/stacks/data/templates/template-to-form";

describe("template conversion round-trip", () => {
  it.each(templates.map((t) => [t.id, t] as const))(
    "template %s converts without errors",
    (_id, template) => {
      const { data } = templateToFormData(template);
      expect(data.spec.stack_resources.length).toBeGreaterThan(0);
    },
  );

  it("keeps every bundled PostgreSQL resource below the mounted filesystem root", () => {
    const postgresResources = templates.flatMap((template) => {
      const { data } = templateToFormData(template);
      return data.spec.stack_resources.filter(
        (resource) =>
          /(^|\/)postgres(?=:|$)/.test(resource.source?.image?.ref ?? "") &&
          resource.volume_mounts?.some(
            (mount) => mount.target_path === "/var/lib/postgresql/data",
          ),
      );
    });

    expect(postgresResources.length).toBeGreaterThan(0);
    for (const postgres of postgresResources) {
      expect(envByName(postgres).PGDATA).toEqual({
        from: "stack",
        name: "PGDATA",
        value: "/var/lib/postgresql/data/pgdata",
      });
    }
  });

  it("tooljet template produces the full 6-service stack", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    const names = data.spec.stack_resources.map((r) => r.name).sort();
    expect(names).toEqual(
      ["otel-stack", "postgresql", "postgrest", "redis", "tooljet", "tooljet-worker"].sort(),
    );

    const volumeNames = (data.spec.volumes ?? []).map((v) => v.name).sort();
    expect(volumeNames).toEqual(["otel-stack-data", "postgres-data"].sort());

    const app = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    const env = envByName(app);
    expect(env.ENABLE_OTEL).toEqual({ from: "stack", name: "ENABLE_OTEL", value: "true" });
    expect(env.OTEL_EXPORTER_OTLP_TRACES).toEqual({
      from: "resourceTemplate",
      name: "OTEL_EXPORTER_OTLP_TRACES",
      resourceName: "otel-stack",
      template: "http://{{host}}:4318/v1/traces",
      values: { host: "host" },
    });
    expect(env.WORKER).toBeUndefined();

    const worker = data.spec.stack_resources.find((r) => r.name === "tooljet-worker")!;
    const workerEnv = envByName(worker);
    expect(workerEnv.WORKER).toEqual({ from: "stack", name: "WORKER", value: "true" });
  });

  it("tooljet template resolves referential env vars at deploy time", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    const app = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    const env = envByName(app);
    expect(env.TOOLJET_HOST).toEqual({ from: "self", name: "TOOLJET_HOST", selfOutput: "public_url" });
    expect(env.PG_HOST).toEqual({ from: "resource", name: "PG_HOST", resourceName: "postgresql", output: "host" });
    expect(env.TOOLJET_DB_HOST).toEqual({ from: "resource", name: "TOOLJET_DB_HOST", resourceName: "postgresql", output: "host" });
    expect(env.PGRST_HOST).toEqual({ from: "resource", name: "PGRST_HOST", resourceName: "postgrest", output: "host" });
    expect(env.REDIS_HOST).toEqual({ from: "resource", name: "REDIS_HOST", resourceName: "redis", output: "host" });

    const worker = data.spec.stack_resources.find((r) => r.name === "tooljet-worker")!;
    const workerEnv = envByName(worker);
    expect(workerEnv.TOOLJET_HOST).toEqual({ from: "resource", name: "TOOLJET_HOST", resourceName: "tooljet", output: "public_url" });
    expect(workerEnv.REDIS_HOST).toEqual({ from: "resource", name: "REDIS_HOST", resourceName: "redis", output: "host" });
  });

  it("tooljet template runs the image entrypoint and exposes the OTLP port", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    // Only the server runs the migration entrypoint. The worker must start
    // with worker:prod — two containers running db:setup:prod concurrently
    // block each other's migrations until statement_timeout kills one
    // (pg error 57014).
    const server = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    expect(server.execution_config?.command).toBe(
      "./server/ee-entrypoint.sh npm run start:prod",
    );
    const worker = data.spec.stack_resources.find((r) => r.name === "tooljet-worker")!;
    expect(worker.execution_config?.command).toBe(
      "npm --prefix server run worker:prod",
    );
    expect(worker.depends_on).toContain("tooljet");

    const otelStack = data.spec.stack_resources.find((r) => r.name === "otel-stack")!;
    const ports = (otelStack.ports ?? [])
      .map((p) => ({ number: p.number, public: p.exposed_to_public }))
      .sort((a, b) => (a.number ?? 0) - (b.number ?? 0));
    expect(ports).toEqual([
      { number: 3000, public: true },
      { number: 4318, public: false },
    ]);
  });
});

type ConvertedResource = ReturnType<typeof templateToFormData>["data"]["spec"]["stack_resources"][number];

function envByName(resource: ConvertedResource) {
  return Object.fromEntries(
    (resource.execution_config?.environment_variables ?? []).map((e) => [e.name, e]),
  );
}

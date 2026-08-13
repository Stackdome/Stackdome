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

  it("tooljet template produces the full 4-service stack", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    const names = data.spec.stack_resources.map((r) => r.name).sort();
    expect(names).toEqual(["postgresql", "postgrest", "redis", "tooljet"].sort());

    const volumeNames = (data.spec.volumes ?? []).map((v) => v.name).sort();
    expect(volumeNames).toEqual(["postgres-data"]);

    // The server runs the workflow processors itself — there is no separate
    // worker resource.
    const app = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    const env = envByName(app);
    expect(env.WORKER).toEqual({ from: "stack", name: "WORKER", value: "true" });
    expect(env.ENABLE_OTEL).toBeUndefined();
    expect(env.OTEL_EXPORTER_OTLP_TRACES).toBeUndefined();
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
  });

  it("tooljet template runs the migration entrypoint and exposes only the app port", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    const server = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    expect(server.execution_config?.command).toBe(
      "./server/ee-entrypoint.sh npm run start:prod",
    );

    const publicPorts = data.spec.stack_resources.flatMap((r) =>
      (r.ports ?? []).filter((p) => p.exposed_to_public).map((p) => ({ name: r.name, number: p.number })),
    );
    expect(publicPorts).toEqual([{ name: "tooljet", number: 80 }]);
  });
});

type ConvertedResource = ReturnType<typeof templateToFormData>["data"]["spec"]["stack_resources"][number];

function envByName(resource: ConvertedResource) {
  return Object.fromEntries(
    (resource.execution_config?.environment_variables ?? []).map((e) => [e.name, e]),
  );
}

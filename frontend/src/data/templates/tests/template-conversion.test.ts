import { describe, it, expect } from "vitest";
import { templates, getTemplateById } from "@/data/templates/registry";
import { templateToFormData } from "@/data/templates/template-to-form";

describe("template conversion round-trip", () => {
  it.each(templates.map((t) => [t.id, t] as const))(
    "template %s converts without errors",
    (_id, template) => {
      const { data } = templateToFormData(template);
      expect(data.spec.stack_resources.length).toBeGreaterThan(0);
    },
  );

  it("tooljet template produces the full 6-service stack", () => {
    const tooljet = getTemplateById("tooljet")!;
    const { data } = templateToFormData(tooljet);

    const names = data.spec.stack_resources.map((r) => r.name).sort();
    expect(names).toEqual(
      ["lgtm", "postgresql", "postgrest", "redis", "tooljet", "tooljet-worker"].sort(),
    );

    const volumeNames = (data.spec.volumes ?? []).map((v) => v.name).sort();
    expect(volumeNames).toEqual(["grafana-data", "postgres-data"].sort());

    const app = data.spec.stack_resources.find((r) => r.name === "tooljet")!;
    const env = Object.fromEntries(
      (app.environment_variables ?? []).map((e: { name: string; value: string }) => [e.name, e.value]),
    );
    expect(env.ENABLE_OTEL).toBe("true");
    expect(env.OTEL_EXPORTER_OTLP_TRACES).toBe("http://lgtm:4318/v1/traces");
    expect(env.WORKER).toBeUndefined();

    const worker = data.spec.stack_resources.find((r) => r.name === "tooljet-worker")!;
    const workerEnv = Object.fromEntries(
      (worker.environment_variables ?? []).map((e: { name: string; value: string }) => [e.name, e.value]),
    );
    expect(workerEnv.WORKER).toBe("true");
  });
});

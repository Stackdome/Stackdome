import { describe, it, expect } from "vitest";
import { templates, getTemplateById } from "../registry";
import { templateToFormData } from "../template-to-form";
import { FormStackSchema } from "@/pages/stacks/schemas/form-schema";

const REQUIRED_FIELDS = [
  "initials",
  "icon",
  "category",
  "shortDescription",
  "longDescription",
  "website",
  "docs",
  "version",
  "stackYaml",
] as const;

describe("templates registry", () => {
  // Every registered template must be display-complete and produce a valid
  // create-form stack — adding a template that breaks either fails here.
  it.each(templates.map((t) => [t.id, t] as const))(
    "%s: has all display fields and converts to a valid create-form stack",
    (_id, template) => {
      expect(template.name).toBeTruthy();
      for (const field of REQUIRED_FIELDS) {
        expect(
          template[field],
          `expected non-empty ${field} on ${template.id}`,
        ).toBeTruthy();
      }

      const { data } = templateToFormData(template);
      const parsed = FormStackSchema.safeParse(data);
      expect(parsed.success, JSON.stringify(parsed.error?.issues, null, 2)).toBe(
        true,
      );
      expect(data.name).toBe(template.id);
    },
  );

  it("ships the ToolJet preset as app + postgresql with a volume", () => {
    const { data } = templateToFormData(getTemplateById("tooljet")!);
    const resources = data.spec!.stack_resources!;
    const names = resources.map((r) => r.name);
    expect(names).toContain("tooljet");
    expect(names).toContain("postgresql");
    expect(
      resources.find((r) => r.name === "tooljet")!.depends_on,
    ).toContain("postgresql");
    expect(data.spec!.volumes!.length).toBeGreaterThan(0);
  });

  it("ships the n8n preset as app + postgres with a volume", () => {
    const { data } = templateToFormData(getTemplateById("n8n")!);
    const resources = data.spec!.stack_resources!;
    const names = resources.map((r) => r.name);
    expect(names).toContain("n8n");
    expect(names).toContain("postgres");
    expect(resources.find((r) => r.name === "n8n")!.depends_on).toContain(
      "postgres",
    );
    expect(data.spec!.volumes!.length).toBeGreaterThan(0);
  });

  it("ships the OpenClaw preset as a single gateway service with a volume", () => {
    const { data } = templateToFormData(getTemplateById("openclaw")!);
    const resources = data.spec!.stack_resources!;
    const names = resources.map((r) => r.name);
    expect(names).toEqual(["openclaw"]);
    expect(
      resources.find((r) => r.name === "openclaw")!.depends_on ?? [],
    ).toHaveLength(0);
    expect(data.spec!.volumes!.length).toBeGreaterThan(0);
  });
});

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
  it("exposes the ToolJet template with all display fields", () => {
    const t = getTemplateById("tooljet");
    expect(t).toBeDefined();
    expect(t!.name).toBe("ToolJet");
    for (const field of REQUIRED_FIELDS) {
      expect(t![field], `expected non-empty ${field}`).toBeTruthy();
    }
    expect(templates).toContain(t);
  });

  it("converts the ToolJet preset into a valid create-form stack", () => {
    const t = getTemplateById("tooljet")!;
    const { data } = templateToFormData(t);

    const parsed = FormStackSchema.safeParse(data);
    expect(parsed.success, JSON.stringify(parsed.error?.issues, null, 2)).toBe(true);

    const resources = data.spec!.stack_resources!;
    const names = resources.map((r) => r.name);
    expect(names).toContain("tooljet");
    expect(names).toContain("postgresql");

    const app = resources.find((r) => r.name === "tooljet")!;
    expect(app.depends_on).toContain("postgresql");

    expect(data.spec!.volumes!.length).toBeGreaterThan(0);
  });

  it("names the prefilled stack after the template", () => {
    const t = getTemplateById("tooljet")!;
    expect(templateToFormData(t).data.name).toBe("tooljet");
  });
});

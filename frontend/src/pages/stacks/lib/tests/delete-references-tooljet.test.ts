import { describe, it, expect } from "vitest";
import { templateToFormData } from "@/pages/stacks/data/templates/template-to-form";
import { getTemplateById } from "@/pages/stacks/data/templates/registry";
import { FormStackSchema } from "@/pages/stacks/schemas/form-schema";
import { deleteResourceAndReferences, findResourceDependents } from "../delete-references";

/**
 * The ToolJet template carries every reference shape at once: `depends_on`,
 * structured env refs, and a mounted volume.
 *
 * A stranded reference surfaces as a form-schema "Unknown resource" issue, so a
 * clean parse after the deletes is what proves the pruning.
 */
describe("deleting from the ToolJet template", () => {
  const draft = () => templateToFormData(getTemplateById("tooljet")!).data;

  const unknownResourceIssues = (data: unknown) => {
    const parsed = FormStackSchema.safeParse(data);
    if (parsed.success) return [];
    return parsed.error.issues.filter((i) => /Unknown resource/i.test(i.message));
  };

  it("names the resources that depend on postgrest before deleting it", () => {
    const data = draft();
    const found = findResourceDependents(data.spec.stack_resources, data.spec.volumes ?? [], "postgrest");
    expect(found.dependsOn).toEqual(["tooljet"]);
    expect(found.envRefs.map((r) => `${r.resource}.${r.envName}`)).toEqual(["tooljet.PGRST_HOST"]);
  });

  it("leaves no unknown-resource issues after deleting postgrest and redis", () => {
    const data = draft();
    expect(unknownResourceIssues(data)).toEqual([]);

    let resources = deleteResourceAndReferences(data.spec.stack_resources, "postgrest");
    resources = deleteResourceAndReferences(resources, "redis");

    const after = { ...data, spec: { ...data.spec, stack_resources: resources } };
    expect(unknownResourceIssues(after)).toEqual([]);
    expect(resources.map((r) => r?.name)).not.toContain("postgrest");
    expect(resources.map((r) => r?.name)).not.toContain("redis");
  });

  /** Keeps the assertion above honest: a plain filter must still produce the
   *  issues, or a clean parse would prove nothing. */
  it("still breaks when the resource is filtered out without pruning references", () => {
    const data = draft();
    const naive = data.spec.stack_resources.filter(
      (r) => r.name !== "postgrest" && r.name !== "redis",
    );
    const after = { ...data, spec: { ...data.spec, stack_resources: naive } };
    expect(unknownResourceIssues(after).length).toBeGreaterThan(0);
  });

  /** Asked before the delete, the way the dialog asks it. Asking afterwards
   *  passes on the "nobody mounts it" path even with the lookup removed. */
  it("reports postgres-data as unattached when postgresql goes", () => {
    const data = draft();
    const found = findResourceDependents(data.spec.stack_resources, data.spec.volumes ?? [], "postgresql");
    expect(found.orphanedVolumes).toEqual(["postgres-data"]);
  });

  it("does not blame an unrelated delete for the volumes it never mounted", () => {
    const data = draft();
    const found = findResourceDependents(data.spec.stack_resources, data.spec.volumes ?? [], "postgrest");
    expect(found.orphanedVolumes).toEqual([]);
  });
});

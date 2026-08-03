import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser";
import {
  convertDockerComposeToStackData,
  type ConversionResult,
} from "@/pages/stacks/lib/docker-compose-converter";
import type { DockerComposeFile } from "@/types/docker-compose";
import type { FormStackData } from "@/pages/stacks/schemas/form-schema";
import type { Template } from "@/pages/stacks/data/templates/types";

export interface TemplateFormResult {
  data: FormStackData;
  warnings: NonNullable<ConversionResult["warnings"]>;
}

/**
 * Convert a template's docker-compose preset into create-form data, reusing the
 * same parse + convert pipeline as the manual Docker Compose import. The stack is
 * named after the template so the create form opens with a sensible default.
 */
export function templateToFormData(template: Template): TemplateFormResult {
  const parsed = parseAndValidateDockerCompose(template.stackYaml);
  const result = convertDockerComposeToStackData(parsed as DockerComposeFile);

  if (!result.success || !result.data) {
    const reason =
      result.errors?.map((e) => e.message).join(", ") ||
      "Failed to convert template preset";
    throw new Error(`Template "${template.id}" is not convertible: ${reason}`);
  }

  return {
    data: { ...result.data, name: template.id },
    warnings: result.warnings ?? [],
  };
}

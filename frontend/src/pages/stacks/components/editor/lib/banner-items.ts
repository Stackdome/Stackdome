import { mapFieldErrors } from "@/pages/stacks/lib/map-field-errors";
import type { ParsedFieldError } from "@/api/errors";
import type { ValidationBannerItem } from "@/pages/stacks/components/editor/validation-banner";

/** Summary-banner rows for the last draft-deploy failure. Fat dialect: resource
 * errors carry a jump index; stack-level errors (name/settings/connections) do not. */
export function buildBannerItems(
  deployFieldErrors: ParsedFieldError[],
  resources: ReadonlyArray<{ name?: string }>,
): ValidationBannerItem[] {
  if (deployFieldErrors.length === 0) return [];
  const mapped = mapFieldErrors(deployFieldErrors, { dialect: "fat" });
  const items: ValidationBannerItem[] = [];
  if (mapped.stackName) items.push({ label: "Stack name", message: mapped.stackName });
  for (const [idxStr, fields] of Object.entries(mapped.resources)) {
    const idx = Number(idxStr);
    const label = resources[idx]?.name?.trim() || `Resource ${idx + 1}`;
    for (const [fieldKey, message] of Object.entries(fields)) {
      // Env errors live on the Environment tab; everything else renders on
      // Configuration. Jump opens the tab holding the offending field.
      const tab = fieldKey.startsWith("execution_config.environment_variables")
        ? "environment"
        : "configuration";
      items.push({ label, message, resourceIndex: idx, tab });
    }
  }
  for (const m of mapped.settings) items.push({ label: "Stack settings", message: m });
  for (const m of mapped.connections) items.push({ label: "Connection", message: m });
  for (const u of mapped.unmapped) items.push({ label: u.field, message: u.message });
  return items;
}

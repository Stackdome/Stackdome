import type { StackRelease } from "@/api/releases";
import type { StackResource } from "@/api/stacks";
import { fieldTab } from "@/pages/stacks/lib/map-field-errors";
import type { ValidationBannerItem } from "@/pages/stacks/components/editor/validation-banner";

// Async release-time validation (image/git/registry checks etc, run after a
// release is created) surfaces on the failed release node through the same
// banner design as the sync create-release errors.
export function releaseValidationBannerItems(
  release: StackRelease,
  resources: StackResource[],
): ValidationBannerItem[] {
  return (release.validation_errors ?? []).map((e) => ({
    label: e.resource_name || "Stack",
    message: `${e.code} — ${e.message}`,
    resourceIndex: jumpTargetIndex(e.resource_name ?? "", resources),
    tab: fieldTab(e.field ?? ""),
  }));
}

// Click-time jump resolution. The canvas drawer indexes into the CURRENT
// resource list (the session draft while editing), which can drift from the
// list the banner was rendered against — resolve by name at click time, never
// bake a render-time index across that boundary.
export function jumpTargetIndex(resourceName: string, resources: { name?: string }[]): number | undefined {
  if (!resourceName) return undefined;
  const idx = resources.findIndex((r) => r.name === resourceName);
  return idx >= 0 ? idx : undefined;
}

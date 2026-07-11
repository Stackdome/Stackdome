import type { StackRelease } from "@/api/releases";
import type { StackResource } from "@/api/stacks";
import { fieldTab } from "@/pages/stacks/lib/map-field-errors";
import type { ValidationBannerItem } from "@/pages/stacks/components/detail/ValidationBanner";

// Async release-time validation (image/git/registry checks etc, run after a
// release is created) surfaces on the failed release node through the same
// banner design as the sync create-release errors.
export function releaseValidationBannerItems(
  release: StackRelease,
  resources: StackResource[],
): ValidationBannerItem[] {
  return (release.validation_errors ?? []).map((e) => {
    const idx = resources.findIndex((r) => r.name === e.resource_name);
    return {
      label: e.resource_name || "Stack",
      message: `${e.code} — ${e.message}`,
      resourceIndex: idx >= 0 ? idx : undefined,
      tab: fieldTab(e.field ?? ""),
    };
  });
}

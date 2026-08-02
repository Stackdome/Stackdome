import type { StackReleaseSnapshot } from "@/api/releases";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import { canonicalFromDraft } from "@/pages/stacks/lib/stack-model/from-form";
import { connectionsOf, resourceToApi, volumeToApi } from "@/pages/stacks/lib/stack-model/to-api";

/**
 * Adapt the live edit-session draft into release-snapshot shape so the staged
 * diff can compare it against a release baseline WITHOUT waiting for the
 * autosave round-trip. Reuses the exact desired-state derivation autosave
 * persists, so what the diff shows is what a save will produce (invalid/held
 * resources excluded on both paths).
 */
export function draftToSnapshot(draft: EditSessionDraft): StackReleaseSnapshot {
  const canonical = canonicalFromDraft(draft);
  return {
    resources: canonical.resources.map(resourceToApi) as StackReleaseSnapshot["resources"],
    volumes: canonical.volumes.map(volumeToApi) as StackReleaseSnapshot["volumes"],
    connections: connectionsOf(canonical),
  };
}

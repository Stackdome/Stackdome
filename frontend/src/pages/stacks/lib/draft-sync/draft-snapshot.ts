import type { StackReleaseSnapshot } from "@/api/releases";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import { buildDesiredState } from "./desired-state";

/**
 * Adapt the live edit-session draft into release-snapshot shape so the staged
 * diff can compare it against a release baseline WITHOUT waiting for the
 * autosave round-trip. Reuses the exact desired-state derivation autosave
 * persists, so what the diff shows is what a save will produce (invalid/held
 * resources excluded on both paths).
 */
export function draftToSnapshot(draft: EditSessionDraft): StackReleaseSnapshot {
  const desired = buildDesiredState(draft);
  return {
    resources: [...desired.resources.values()] as StackReleaseSnapshot["resources"],
    volumes: [...desired.volumes.values()] as StackReleaseSnapshot["volumes"],
    connections: [...desired.connections.values()],
  };
}

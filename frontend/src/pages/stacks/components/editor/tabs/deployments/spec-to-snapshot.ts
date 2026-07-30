import type { Stack } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";

/**
 * Adapt the saved spec into snapshot shape for diffSnapshots().
 * Spec stores resources under `stack_resources`; snapshots under `resources`.
 */
export function specToSnapshot(stack: Stack): StackReleaseSnapshot {
  return {
    resources: stack.spec?.stack_resources ?? [],
    volumes: stack.spec?.volumes ?? [],
    connections: stack.spec?.connections ?? [],
  };
}

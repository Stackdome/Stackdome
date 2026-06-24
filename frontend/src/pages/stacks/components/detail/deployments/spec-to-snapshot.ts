import type { Stack } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";

/**
 * Adapt the stack's current saved spec into the release-snapshot shape so it can
 * be diffed against a real release snapshot via diffSnapshots(). The spec stores
 * resources under `stack_resources`; snapshots store them under `resources`.
 */
export function specToSnapshot(stack: Stack): StackReleaseSnapshot {
  return {
    resources: stack.spec?.stack_resources ?? [],
    volumes: stack.spec?.volumes ?? [],
    connections: stack.spec?.connections ?? [],
  };
}

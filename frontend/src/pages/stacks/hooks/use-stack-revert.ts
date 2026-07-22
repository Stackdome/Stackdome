import { useCallback, useState } from "react";
import type { Stack } from "@/api/stacks";
import { applyStack } from "@/api/stacks";
import { deleteVolume } from "@/api/volumes";
import type { StackReleaseSnapshot } from "@/api/releases";
import { snapshotToUpdateRequest, volumesToDelete } from "@/pages/stacks/lib/draft-sync/snapshot-to-update";
import { parseApiError } from "@/api/errors";

export interface UseStackRevertArgs {
  ids: { orgId: string; projectName: string; stackId: string } | null;
  stack: Stack | undefined;
  liveSnapshot: StackReleaseSnapshot | undefined;
  /** Page-provided, ticket-gated stack refetch (applies the payload to page
   *  state itself and returns the newest applied stack) — keeps this hook's
   *  refetch inside the page's response-ordering domain. */
  fetchStack: () => Promise<Stack>;
  /** session.discard — the page's session auto-start effect re-seeds from the refreshed stack. */
  onReverted: (fresh: Stack) => void;
  /** Called with the backend reason when the revert throws. */
  onError?: (message: string) => void;
}

/** Restore the authored stack to the last deployed snapshot. */
export function useStackRevert({ ids, stack, liveSnapshot, fetchStack, onReverted, onError }: UseStackRevertArgs) {
  const [reverting, setReverting] = useState(false);

  const revert = useCallback(async (): Promise<boolean> => {
    if (!ids || !stack || !liveSnapshot) return false;
    setReverting(true);
    try {
      const req = snapshotToUpdateRequest(liveSnapshot, { name: stack.name, labels: stack.labels });
      await applyStack(ids.orgId, ids.projectName, ids.stackId, req);
      // Draft-only volumes are unmounted after the PUT; remove them (destroys
      // the cluster volume — the confirm dialog carries that warning).
      for (const v of volumesToDelete(stack, liveSnapshot)) {
        await deleteVolume(ids.orgId, ids.projectName, v.id);
      }
      const fresh = await fetchStack();
      onReverted(fresh);
      return true;
    } catch (err) {
      onError?.(parseApiError(err).topLevel);
      return false;
    } finally {
      setReverting(false);
    }
  }, [ids, stack, liveSnapshot, fetchStack, onReverted, onError]);

  return { reverting, revert };
}

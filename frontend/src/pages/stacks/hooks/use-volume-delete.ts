import { useCallback, useState } from "react";
import axios from "axios";
import type { Stack, Volume } from "@/api/stacks";
import { deleteVolume } from "@/api/volumes";

export type VolumeDeleteToast = (t: { title: string; description?: string; variant?: "destructive" | "success" }) => void;

export interface UseVolumeDeleteArgs {
  ids: { orgId: string; projectName: string; stackId: string } | null;
  draftSync: { flush(): Promise<boolean>; notifyExternalUpdate(stack: Stack): void };
  /** Page-provided, ticket-gated stack refetch — applies the payload to page
   *  state itself and returns the newest applied stack, so this hook's own
   *  GETs can't clobber fresher state with a late response. */
  fetchStack: () => Promise<Stack>;
  /** Failure path: re-add the volume to the session draft (it reappears floating). */
  onRestoreVolume: (fresh: Volume) => void;
  toast: VolumeDeleteToast;
}

export interface UseVolumeDelete {
  deleting: boolean;
  deleteVolume: (name: string) => Promise<boolean>;
}

/**
 * Confirm-gated, immediate server-side volume deletion. The caller has already
 * applied the local draft edit (dropping the volume + its mounts) before
 * invoking this; on any failure path the volume is restored into the draft so
 * the UI never lies about what's actually on the server.
 */
export function useVolumeDelete({
  ids,
  draftSync,
  fetchStack,
  onRestoreVolume,
  toast,
}: UseVolumeDeleteArgs): UseVolumeDelete {
  const [deleting, setDeleting] = useState(false);

  /** Refetch the stack, keep the autosave mirror truthful, and — if the volume
   *  still exists server-side — restore it into the local draft (it never
   *  actually disappeared, so the local edit must not lie about that). */
  const refetchAndRestore = useCallback(
    async (name: string): Promise<void> => {
      if (!ids) return;
      try {
        const fresh = await fetchStack();
        draftSync.notifyExternalUpdate(fresh);
        const vol = (fresh.spec?.volumes ?? []).find((v) => v.name === name);
        if (vol) onRestoreVolume(vol as Volume);
      } catch {
        toast({
          title: "Couldn't refresh stack",
          description: "Reload the page to see the current state.",
          variant: "destructive",
        });
      }
    },
    [ids, fetchStack, draftSync, onRestoreVolume, toast],
  );

  const runDelete = useCallback(
    async (name: string): Promise<boolean> => {
      if (!ids) return true;
      setDeleting(true);
      try {
        // Let React commit the caller's draft edit first — flush()'s startCycle
        // reads the current session draft, and buildDesiredState already drops
        // mounts of missing volumes, so this deletes the volume_mount connections.
        await new Promise((resolve) => setTimeout(resolve, 0));

        const ok = await draftSync.flush();
        if (!ok) {
          await refetchAndRestore(name);
          toast({
            title: "Delete blocked",
            description: `Draft changes couldn't be saved, so "${name}" was not deleted.`,
            variant: "destructive",
          });
          return false;
        }

        const fresh = await fetchStack();
        const id = fresh.spec?.volumes?.find((v) => v.name === name)?.id;
        if (id == null) {
          // Never persisted server-side (or already gone) — the local delete is
          // already truthful; just keep the mirror in sync.
          draftSync.notifyExternalUpdate(fresh);
          return true;
        }

        await deleteVolume(ids.orgId, ids.projectName, id);

        // fetchStack applies the payload to page state itself (ticket-gated).
        const fresh2 = await fetchStack();
        draftSync.notifyExternalUpdate(fresh2);
        toast({ title: "Volume deleted", description: `"${name}" and its data were deleted.`, variant: "success" });
        return true;
      } catch (err) {
        if (axios.isAxiosError(err) && err.response?.status === 409) {
          await refetchAndRestore(name);
          toast({
            title: "Volume is in use",
            description: `"${name}" is mounted by a running deployment and can't be deleted. The volume was not deleted.`,
            variant: "destructive",
          });
          return false;
        }
        await refetchAndRestore(name);
        toast({
          title: "Couldn't delete volume",
          description: `"${name}" was not deleted. Check your connection and try again.`,
          variant: "destructive",
        });
        return false;
      } finally {
        setDeleting(false);
      }
    },
    [ids, fetchStack, draftSync, refetchAndRestore, toast],
  );

  return { deleting, deleteVolume: runDelete };
}

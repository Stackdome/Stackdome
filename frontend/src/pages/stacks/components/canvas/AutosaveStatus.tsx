import { Check, Loader2, TriangleAlert } from "lucide-react";
import { SYNC_STATUS, type SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";

/** Compact autosave state readout for the canvas header. Purely presentational. */
export function AutosaveStatus({ status }: { status: SyncStatus }) {
  if (status === SYNC_STATUS.idle) return null;
  if (status === SYNC_STATUS.saving) {
    return (
      <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-muted-foreground">
        <Loader2 className="size-3 animate-spin" aria-hidden /> Saving…
      </span>
    );
  }
  if (status === SYNC_STATUS.error) {
    return (
      <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-danger">
        <TriangleAlert className="size-3" aria-hidden /> Save failed — retrying
      </span>
    );
  }
  return (
    <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-muted-foreground">
      <Check className="size-3" aria-hidden /> All changes saved
    </span>
  );
}

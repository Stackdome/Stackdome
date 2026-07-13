import { useState } from "react";
import { Loader2, AlertTriangle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, getErrorStatus } from "@/api/client";
import { deleteObjectStore } from "@/api/object-stores";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import type { ObjectStore } from "../types";

type Props = {
  store: ObjectStore | null;
  onOpenChange: (open: boolean) => void;
  onDeleted: () => void;
};

const IN_USE_PATTERN = /in use/i;

export function ObjectStoreDeleteDialog({ store, onOpenChange, onDeleted }: Props) {
  const { toast } = useToast();
  const { projectNameById } = useResourceProjects();
  const [submitting, setSubmitting] = useState(false);
  const [conflictMessage, setConflictMessage] = useState<string | null>(null);
  const open = !!store;

  async function handleConfirm() {
    if (!store?.id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      toast({
        title: "Could not delete Object Store",
        description: "No organization selected.",
        variant: "destructive",
      });
      return;
    }
    const projectName = projectNameById(store.project_id);
    if (!projectName) {
      toast({
        title: "Could not delete Object Store",
        description: "Could not resolve the project for this object store.",
        variant: "destructive",
      });
      return;
    }
    setSubmitting(true);
    setConflictMessage(null);
    try {
      await deleteObjectStore(orgId, projectName, store.id);
      toast({ title: "Object store deleted", variant: "success" });
      onDeleted();
      onOpenChange(false);
    } catch (e: unknown) {
      const status = getErrorStatus(e);
      const reason = getErrorMessage(e);
      // The backend currently returns HTTP 400 with a reason describing
      // that the store is in use; a future fix may switch to 409. Handle
      // both so the conflict is always surfaced in-dialog.
      const isReferencedConflict =
        status === 409 || (status === 400 && IN_USE_PATTERN.test(reason));
      if (isReferencedConflict) {
        setConflictMessage(
          reason || "This Object Store is in use by one or more Postgres add-ons.",
        );
      } else {
        toast({
          title: "Failed to delete",
          description: reason,
          variant: "destructive",
        });
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (submitting) return;
        if (!o) setConflictMessage(null);
        onOpenChange(o);
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Object Store</DialogTitle>
          <DialogDescription>
            Delete <span className="font-mono">{store?.name}</span>? Existing backup files in the
            destination are not removed.
          </DialogDescription>
        </DialogHeader>

        {conflictMessage && (
          <div className="flex items-start gap-2 rounded-md border border-warn-border bg-warn-bg px-3 py-2 text-sm text-warn">
            <AlertTriangle className="mt-0.5 h-4 w-4" />
            <span>{conflictMessage}</span>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={submitting || !!conflictMessage}
          >
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

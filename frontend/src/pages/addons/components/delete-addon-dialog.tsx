import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

interface DeleteAddonDialogProps {
  open: boolean;
  addonName?: string;
  loading: boolean;
  error: string | null;
  onConfirm: () => void;
  onCancel: () => void;
}

export function DeleteAddonDialog({
  open,
  addonName,
  loading,
  error,
  onConfirm,
  onCancel,
}: DeleteAddonDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && !loading && onCancel()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete addon</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 text-sm">
          <p>
            Are you sure you want to delete{" "}
            {addonName && <strong>"{addonName}"</strong>}? This action cannot be undone
            and the underlying database and storage will be removed.
          </p>
          {error && (
            <div className="text-destructive bg-destructive/10 p-3 rounded-md whitespace-pre-wrap">
              {error}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={loading}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={loading}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

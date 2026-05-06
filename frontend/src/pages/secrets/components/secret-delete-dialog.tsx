import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

interface SecretDeleteDialogProps {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  loading: boolean;
  secretName?: string;
}

export function SecretDeleteDialog({
  open,
  onConfirm,
  onCancel,
  loading,
  secretName,
}: SecretDeleteDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && !loading && onCancel()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Secret</DialogTitle>
        </DialogHeader>
        <div>
          Are you sure you want to delete the secret{" "}
          {secretName && <strong>"{secretName}"</strong>}? This action cannot be undone.
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={loading}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

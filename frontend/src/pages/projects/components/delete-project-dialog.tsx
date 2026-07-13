import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";

interface DeleteProjectDialogProps {
  open: boolean;
  projectName: string;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function DeleteProjectDialog({
  open,
  projectName,
  onOpenChange,
  onConfirm,
}: DeleteProjectDialogProps) {
  const [typed, setTyped] = useState("");

  // Reset typed value when dialog opens/closes
  useEffect(() => {
    if (!open) setTyped("");
  }, [open]);

  const canConfirm = typed === projectName;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-danger">Delete project</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            This action is <span className="font-semibold text-foreground">irreversible</span>.
            The project <span className="font-semibold text-foreground">&ldquo;{projectName}&rdquo;</span> will
            be permanently deleted.
          </p>

          <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            Deleting this project cannot be undone.
          </div>

          <FieldShell
            label="Type the project name to confirm"
            htmlFor="delete-project-confirm"
          >
            <Input
              id="delete-project-confirm"
              placeholder={projectName}
              value={typed}
              onChange={(e) => setTyped(e.target.value)}
            />
          </FieldShell>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={!canConfirm}
            onClick={() => {
              if (canConfirm) onConfirm();
            }}
            className="bg-danger text-white hover:bg-danger/90"
          >
            Delete project
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

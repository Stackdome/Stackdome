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
          <DialogTitle>Delete project?</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            The project &ldquo;{projectName}&rdquo; will be permanently deleted. This cannot be undone.
          </p>

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
            variant="destructive"
            disabled={!canConfirm}
            onClick={() => {
              if (canConfirm) onConfirm();
            }}
          >
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

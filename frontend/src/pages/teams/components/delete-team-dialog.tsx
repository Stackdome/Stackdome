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

interface DeleteTeamDialogProps {
  open: boolean;
  teamName: string;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function DeleteTeamDialog({
  open,
  teamName,
  onOpenChange,
  onConfirm,
}: DeleteTeamDialogProps) {
  const [typed, setTyped] = useState("");

  // Reset typed value when dialog opens/closes
  useEffect(() => {
    if (!open) setTyped("");
  }, [open]);

  const canConfirm = typed === teamName;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-danger">Delete team</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            This action is <span className="font-semibold text-foreground">irreversible</span>.
            The team <span className="font-semibold text-foreground">&ldquo;{teamName}&rdquo;</span> will
            be permanently deleted.
          </p>

          <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            Deleting this team cannot be undone.
          </div>

          <FieldShell
            label="Type the team name to confirm"
            htmlFor="delete-team-confirm"
          >
            <Input
              id="delete-team-confirm"
              placeholder={teamName}
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
            Delete team
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

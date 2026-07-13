import { useEffect, useState } from "react";
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

type ActionResult = { ok: true } | { ok: false; error: string };

interface RenameProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentName: string;
  onRename: (newName: string) => Promise<ActionResult>;
}

function toSlug(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

export function RenameProjectDialog({ open, onOpenChange, currentName, onRename }: RenameProjectDialogProps) {
  const [name, setName] = useState(currentName);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Re-seed the field whenever a different project is opened for rename.
  useEffect(() => {
    if (open) {
      setName(currentName);
      setError(null);
      setSubmitting(false);
    }
  }, [open, currentName]);

  const trimmed = name.trim();
  const slug = toSlug(name);
  const canSubmit = trimmed.length > 0 && trimmed !== currentName && !submitting;

  function handleOpenChange(val: boolean) {
    if (!val) {
      setError(null);
      setSubmitting(false);
    }
    onOpenChange(val);
  }

  async function handleSubmit() {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    const result = await onRename(trimmed);
    if (result.ok) {
      handleOpenChange(false);
    } else {
      setError(result.error);
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Rename project</DialogTitle>
        </DialogHeader>

        {error && (
          <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <FieldShell label="Name" htmlFor="rename-project-name" required>
            <Input
              id="rename-project-name"
              placeholder="e.g. Platform"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (error) setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") void handleSubmit();
              }}
              autoFocus
            />
          </FieldShell>

          {slug && (
            <p className="font-mono text-xs text-muted-foreground">
              Slug: <span className="text-foreground">{slug}</span>
            </p>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={submitting}>
            Cancel
          </Button>
          <Button onClick={() => void handleSubmit()} disabled={!canSubmit}>
            {submitting ? "Renaming…" : "Rename project"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

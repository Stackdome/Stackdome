import { useState } from "react";
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
import type { Project } from "@/api/projects";

type ActionResult = { ok: true } | { ok: false; error: string };

interface CreateProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreate: (name: string) => Promise<ActionResult>;
}

function toSlug(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

export function CreateProjectDialog({ open, onOpenChange, onCreate }: CreateProjectDialogProps) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const slug = toSlug(name);
  const canSubmit = name.trim().length > 0 && !submitting;

  function handleOpenChange(val: boolean) {
    if (!val) {
      setName("");
      setError(null);
      setSubmitting(false);
    }
    onOpenChange(val);
  }

  async function handleSubmit() {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    const result = await onCreate(name.trim());
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
          <DialogTitle>Create project</DialogTitle>
        </DialogHeader>

        {error && (
          <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <FieldShell label="Name" htmlFor="project-name" required>
            <Input
              id="project-name"
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
            {submitting ? "Creating…" : "Create project"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// Re-export Project type for convenience (avoids importing api/projects in the page just for type)
export type { Project };

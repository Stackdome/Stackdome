import { useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { syncPreviewEnv, type PreviewStack } from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";

interface SyncEnvDialogProps {
  env: PreviewStack | null;
  onOpenChange: (open: boolean) => void;
  onSynced: () => void;
}

export function SyncEnvDialog({ env, onOpenChange, onSynced }: SyncEnvDialogProps) {
  const { defaultTeamName } = useResourceTeams();
  const [commit, setCommit] = useState("");
  const [force, setForce] = useState(false);
  const [stackfileContent, setStackfileContent] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const submit = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !env?.id) return;
    setSaving(true);
    setError(null);
    try {
      await syncPreviewEnv(orgId, defaultTeamName, env.id, {
        ...(commit.trim() ? { commit: commit.trim() } : {}),
        ...(force ? { force_sync: true } : {}),
        ...(stackfileContent.trim() ? { stackfile_content: stackfileContent } : {}),
      });
      onSynced();
      onOpenChange(false);
      setCommit("");
      setForce(false);
      setStackfileContent("");
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={env != null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle>Sync preview environment</DialogTitle>
          <DialogDescription>
            Re-resolves {env?.branch} and redeploys PR #{env?.pr_number} at its
            latest commit.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="sync-commit">Pin to commit (optional)</Label>
            <Input
              id="sync-commit"
              placeholder="full or short SHA"
              value={commit}
              onChange={(e) => setCommit(e.target.value)}
              className="font-mono text-xs"
            />
          </div>
          {/*
            Brief calls for a "Force sync" Checkbox; @/components/ui/checkbox does not
            exist in this codebase and no @radix-ui/react-checkbox dependency is
            installed. Per task instructions ("report BLOCKED rather than hand-rolling
            one"), no new checkbox primitive was fabricated. Switch is the codebase's
            existing sanctioned primitive for a single boolean toggle (see
            add-cluster-dialog.tsx), so it is used here instead. Flagged as a concern
            in task-11-report.md — confirm before treating this as final.
          */}
          <div className="flex items-start gap-3">
            <Switch id="sync-force" checked={force} onCheckedChange={(v) => setForce(v === true)} className="mt-0.5" />
            <Label htmlFor="sync-force">Force sync even when nothing changed</Label>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="sync-stackfile">Stackfile content (optional)</Label>
            <Textarea
              id="sync-stackfile"
              rows={4}
              value={stackfileContent}
              onChange={(e) => setStackfileContent(e.target.value)}
              className="font-mono text-xs"
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Sync
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

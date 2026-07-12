import { useEffect, useState } from "react";
import { ChevronsUpDown } from "lucide-react";
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
import { parseImageOverrides } from "@/pages/previews/lib/parse-image-overrides";

interface SyncEnvDialogProps {
  env: PreviewStack | null;
  onOpenChange: (open: boolean) => void;
  onSynced: () => void;
}

export function SyncEnvDialog({ env, onOpenChange, onSynced }: SyncEnvDialogProps) {
  const { defaultTeamName } = useResourceTeams();
  const [commit, setCommit] = useState("");
  const [force, setForce] = useState(false);
  const [advanced, setAdvanced] = useState(false);
  const [stackfileContent, setStackfileContent] = useState("");
  const [overridesText, setOverridesText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const reset = () => {
    setCommit("");
    setForce(false);
    setAdvanced(false);
    setStackfileContent("");
    setOverridesText("");
    setError(null);
  };

  const envId = env?.id;
  useEffect(() => {
    if (envId == null) return;
    reset();
  }, [envId]);

  const submit = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !env?.id) return;
    setSaving(true);
    setError(null);
    try {
      const overrides = parseImageOverrides(overridesText);
      await syncPreviewEnv(orgId, defaultTeamName, env.id, {
        ...(commit.trim() ? { commit: commit.trim() } : {}),
        ...(force ? { force_sync: true } : {}),
        ...(stackfileContent.trim() ? { stackfile_content: stackfileContent } : {}),
        ...(overrides ? { image_overrides: overrides } : {}),
      });
      onSynced();
      onOpenChange(false);
      reset();
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
            <Label htmlFor="sync-commit">Pin to a specific commit (optional)</Label>
            <Input
              id="sync-commit"
              placeholder="full or short SHA"
              value={commit}
              onChange={(e) => setCommit(e.target.value)}
              className="font-mono text-xs"
            />
            <p className="text-xs text-muted-foreground">
              Leave empty to use the branch&apos;s latest commit.
            </p>
          </div>
          {/*
            "Force sync" uses a Switch rather than a Checkbox: @/components/ui/checkbox
            does not exist in this codebase and no @radix-ui/react-checkbox dependency is
            installed. Switch is the codebase's existing sanctioned primitive for a
            single boolean toggle (see add-cluster-dialog.tsx).
          */}
          <div className="flex items-start gap-3">
            <Switch id="sync-force" checked={force} onCheckedChange={setForce} className="mt-0.5" />
            <Label htmlFor="sync-force">Force sync even when nothing changed</Label>
          </div>

          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="px-0 text-muted-foreground"
            onClick={() => setAdvanced((v) => !v)}
          >
            <ChevronsUpDown className="h-4 w-4" />
            Advanced
          </Button>

          {advanced && (
            <div className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="sync-stackfile">Stackfile content (optional)</Label>
                <Textarea
                  id="sync-stackfile"
                  rows={6}
                  placeholder="Paste a stackfile to use instead of the one in the repository"
                  value={stackfileContent}
                  onChange={(e) => setStackfileContent(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="sync-overrides">Image overrides (optional)</Label>
                <Textarea
                  id="sync-overrides"
                  rows={3}
                  placeholder={"resource=registry/image:tag\none per line"}
                  value={overridesText}
                  onChange={(e) => setOverridesText(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
            </div>
          )}

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

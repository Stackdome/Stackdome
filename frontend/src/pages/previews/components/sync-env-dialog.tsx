import { useEffect, useState } from "react";
import { ChevronsUpDown } from "lucide-react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FieldShell, AlertBanner } from "@/components/branded";
import { Textarea } from "@/components/ui/textarea";
import { syncPreviewEnv, type PreviewStack } from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { parseImageOverrides } from "@/pages/previews/lib/parse-image-overrides";
import { syncEnvSchema, type SyncEnvValues } from "@/pages/previews/lib/form-schemas";

interface SyncEnvDialogProps {
  env: PreviewStack | null;
  onOpenChange: (open: boolean) => void;
  onSynced: () => void;
}

export function SyncEnvDialog({ env, onOpenChange, onSynced }: SyncEnvDialogProps) {
  const { defaultProjectName } = useResourceProjects();
  const [commit, setCommit] = useState("");
  const [force, setForce] = useState(false);
  const [advanced, setAdvanced] = useState(false);
  const [stackfileContent, setStackfileContent] = useState("");
  const [overridesText, setOverridesText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof SyncEnvValues, string>>>({});
  const [saving, setSaving] = useState(false);

  const reset = () => {
    setCommit("");
    setForce(false);
    setAdvanced(false);
    setStackfileContent("");
    setOverridesText("");
    setError(null);
    setFieldErrors({});
  };

  const envId = env?.id;
  useEffect(() => {
    if (envId == null) return;
    reset();
  }, [envId]);

  const submit = async () => {
    const parsed = syncEnvSchema.safeParse({ commit, overridesText });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({ commit: flat.commit?.[0], overridesText: flat.overridesText?.[0] });
      if (flat.overridesText) setAdvanced(true);
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultProjectName || !env?.id) return;
    setSaving(true);
    setError(null);
    try {
      const overrides = parseImageOverrides(parsed.data.overridesText);
      await syncPreviewEnv(orgId, defaultProjectName, env.id, {
        ...(parsed.data.commit ? { commit: parsed.data.commit } : {}),
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
          <FieldShell
            label="Pin to a specific commit (optional)"
            htmlFor="sync-commit"
            hint="Leave empty to use the branch's latest commit."
            error={fieldErrors.commit}
          >
            <Input
              id="sync-commit"
              placeholder="full or short SHA"
              value={commit}
              onChange={(e) => {
                setCommit(e.target.value);
                setFieldErrors((prev) => ({ ...prev, commit: undefined }));
              }}
              className="font-mono text-xs"
              aria-invalid={!!fieldErrors.commit}
            />
          </FieldShell>
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
              <FieldShell label="Stackfile content (optional)" htmlFor="sync-stackfile">
                <Textarea
                  id="sync-stackfile"
                  rows={6}
                  placeholder="Paste a stackfile to use instead of the one in the repository"
                  value={stackfileContent}
                  onChange={(e) => setStackfileContent(e.target.value)}
                  className="font-mono text-xs"
                />
              </FieldShell>
              <FieldShell
                label="Image overrides (optional)"
                htmlFor="sync-overrides"
                error={fieldErrors.overridesText}
              >
                <Textarea
                  id="sync-overrides"
                  rows={3}
                  placeholder={"resource=registry/image:tag\none per line"}
                  value={overridesText}
                  onChange={(e) => {
                    setOverridesText(e.target.value);
                    setFieldErrors((prev) => ({ ...prev, overridesText: undefined }));
                  }}
                  className="font-mono text-xs"
                />
              </FieldShell>
            </div>
          )}

          {error && <AlertBanner>{error}</AlertBanner>}
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

import { useEffect, useState } from "react";
import { ChevronsUpDown } from "lucide-react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { Textarea } from "@/components/ui/textarea";
import { createPreviewEnv } from "@/api/preview-envs";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import type { StackPreviewConfig } from "@/api/preview-configs";
import { parseImageOverrides } from "@/pages/previews/lib/parse-image-overrides";
import { newPreviewEnvSchema, type NewPreviewEnvValues } from "@/pages/previews/lib/form-schemas";

interface NewPreviewEnvModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: StackPreviewConfig;
  onCreated: () => void;
}

export function NewPreviewEnvModal({ open, onOpenChange, config, onCreated }: NewPreviewEnvModalProps) {
  const { defaultTeamName } = useResourceTeams();
  const [prNumber, setPrNumber] = useState("");
  const [branch, setBranch] = useState("");
  const [advanced, setAdvanced] = useState(false);
  const [stackfileContent, setStackfileContent] = useState("");
  const [overridesText, setOverridesText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof NewPreviewEnvValues, string>>>({});
  const [saving, setSaving] = useState(false);

  const reset = () => {
    setPrNumber("");
    setBranch("");
    setStackfileContent("");
    setOverridesText("");
    setError(null);
    setFieldErrors({});
    setAdvanced(false);
  };

  useEffect(() => {
    if (open) reset();
  }, [open]);

  const submit = async () => {
    const parsed = newPreviewEnvSchema.safeParse({ prNumber, branch, overridesText });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({
        prNumber: flat.prNumber?.[0],
        branch: flat.branch?.[0],
        overridesText: flat.overridesText?.[0],
      });
      if (flat.overridesText) setAdvanced(true);
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !config.id) return;
    setSaving(true);
    setError(null);
    try {
      const overrides = parseImageOverrides(parsed.data.overridesText);
      await createPreviewEnv(orgId, defaultTeamName, {
        config_id: config.id,
        pr_number: parsed.data.prNumber,
        branch: parsed.data.branch,
        ...(stackfileContent.trim() ? { stackfile_content: stackfileContent } : {}),
        ...(overrides ? { image_overrides: overrides } : {}),
      });
      onCreated();
      onOpenChange(false);
      reset();
    } catch (e) {
      setError(
        isErrorStatus(e, 409)
          ? `PR #${prNumber} already has an environment.`
          : getErrorMessage(e),
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>New preview environment</DialogTitle>
          <DialogDescription>
            Deploys the stackfile from a pull request branch of {config.name}.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <FieldShell label="PR number" htmlFor="env-pr" required error={fieldErrors.prNumber}>
            <Input
              id="env-pr"
              type="number"
              min={1}
              value={prNumber}
              onChange={(e) => {
                setPrNumber(e.target.value);
                setFieldErrors((prev) => ({ ...prev, prNumber: undefined }));
              }}
              aria-invalid={!!fieldErrors.prNumber}
            />
          </FieldShell>
          <FieldShell
            label="Branch"
            htmlFor="env-branch"
            required
            hint="The environment deploys this branch's latest commit; use Sync to pick up new commits later."
            error={fieldErrors.branch}
          >
            <Input
              id="env-branch"
              placeholder="feat/my-change"
              value={branch}
              onChange={(e) => {
                setBranch(e.target.value);
                setFieldErrors((prev) => ({ ...prev, branch: undefined }));
              }}
              aria-invalid={!!fieldErrors.branch}
            />
          </FieldShell>

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
              <FieldShell label="Stackfile content (optional)" htmlFor="env-stackfile">
                <Textarea
                  id="env-stackfile"
                  rows={6}
                  placeholder="Paste a stackfile to use instead of the one in the repository"
                  value={stackfileContent}
                  onChange={(e) => setStackfileContent(e.target.value)}
                  className="font-mono text-xs"
                />
              </FieldShell>
              <FieldShell
                label="Image overrides (optional)"
                htmlFor="env-overrides"
                error={fieldErrors.overridesText}
              >
                <Textarea
                  id="env-overrides"
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

          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Create environment
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

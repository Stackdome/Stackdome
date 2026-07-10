import { useEffect, useState } from "react";
import { ChevronsUpDown } from "lucide-react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { createPreviewEnv } from "@/api/preview-envs";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import type { StackPreviewConfig } from "@/api/preview-configs";

interface NewPreviewEnvModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: StackPreviewConfig;
  onCreated: () => void;
}

/** "K1=v1\nK2=v2" → { K1: "v1", K2: "v2" }; blank lines ignored. */
function parseOverrides(text: string): Record<string, string> | undefined {
  const entries = text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .map((l) => {
      const idx = l.indexOf("=");
      return idx > 0 ? ([l.slice(0, idx).trim(), l.slice(idx + 1).trim()] as const) : null;
    })
    .filter((e): e is readonly [string, string] => e != null);
  return entries.length > 0 ? Object.fromEntries(entries) : undefined;
}

export function NewPreviewEnvModal({ open, onOpenChange, config, onCreated }: NewPreviewEnvModalProps) {
  const { defaultTeamName } = useResourceTeams();
  const [prNumber, setPrNumber] = useState("");
  const [branch, setBranch] = useState("");
  const [advanced, setAdvanced] = useState(false);
  const [stackfileContent, setStackfileContent] = useState("");
  const [overridesText, setOverridesText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const reset = () => {
    setPrNumber("");
    setBranch("");
    setStackfileContent("");
    setOverridesText("");
    setError(null);
    setAdvanced(false);
  };

  useEffect(() => {
    if (open) reset();
  }, [open]);

  const submit = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !config.id) return;
    setSaving(true);
    setError(null);
    try {
      const overrides = parseOverrides(overridesText);
      await createPreviewEnv(orgId, defaultTeamName, {
        config_id: config.id,
        pr_number: prNumber.trim(),
        branch,
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
          <div className="space-y-1.5">
            <Label htmlFor="env-pr">PR number</Label>
            <Input
              id="env-pr"
              type="number"
              min={1}
              value={prNumber}
              onChange={(e) => setPrNumber(e.target.value)}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="env-branch">Branch</Label>
            <Input
              id="env-branch"
              placeholder="feat/my-change"
              value={branch}
              onChange={(e) => setBranch(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              The environment pins this branch&apos;s current commit; use Sync to
              pick up new commits later.
            </p>
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
                <Label htmlFor="env-stackfile">Stackfile content (optional)</Label>
                <Textarea
                  id="env-stackfile"
                  rows={6}
                  placeholder="Paste a stackfile to use instead of the one in the repository"
                  value={stackfileContent}
                  onChange={(e) => setStackfileContent(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="env-overrides">Image overrides (optional)</Label>
                <Textarea
                  id="env-overrides"
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
          <Button
            onClick={() => void submit()}
            disabled={saving || prNumber.trim() === "" || branch.trim() === ""}
          >
            Create environment
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

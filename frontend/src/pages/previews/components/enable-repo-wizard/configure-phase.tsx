import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { createPreviewConfig } from "@/api/preview-configs";
import { listRepositoryBranches } from "@/api/git-integrations";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import type { PickedRepo } from "./enable-repo-wizard";

interface ConfigurePhaseProps {
  repo: PickedRepo;
  onCreated: (configId: string) => void;
  onBack: () => void;
}

export function ConfigurePhase({ repo, onCreated, onBack }: ConfigurePhaseProps) {
  const { defaultTeamName } = useResourceTeams();
  const [name, setName] = useState(repo.fullName.split("/").pop() ?? "");
  const [baseBranch, setBaseBranch] = useState(repo.defaultBranch);
  const [branches, setBranches] = useState<string[]>([]);
  const [stackfilePath, setStackfilePath] = useState("stackfile.yaml");
  const [maxActive, setMaxActive] = useState(10);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !repo.integrationId) return;
    const [owner, repoName] = repo.fullName.split("/");
    listRepositoryBranches(orgId, repo.integrationId, owner, repoName)
      .then((b) => setBranches(b.items ?? []))
      .catch(() => {
        // fall back to free-text branch input
      });
  }, [repo]);

  const submit = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName) {
      setError("No organization or default team available.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const created = await createPreviewConfig(orgId, defaultTeamName, {
        name,
        git_repository: { repo_url: repo.cloneUrl, base_branch: baseBranch },
        stackfile_path: stackfilePath,
        max_active_previews: maxActive,
      });
      onCreated(created.id ?? "");
    } catch (e) {
      setError(
        isErrorStatus(e, 409)
          ? "A configuration with this name already exists."
          : getErrorMessage(e),
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 space-y-5 overflow-y-auto p-6">
        <div>
          <h3 className="font-mono text-xs">{repo.fullName}</h3>
          <p className="text-sm text-muted-foreground">
            Every pull request on this repository can get its own preview
            environment.
          </p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="cfg-name">Name</Label>
          <Input id="cfg-name" value={name} onChange={(e) => setName(e.target.value)} />
          <p className="text-xs text-muted-foreground">Cannot be changed later.</p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="cfg-branch">Base branch</Label>
          {branches.length > 0 ? (
            <Select value={baseBranch} onValueChange={setBaseBranch}>
              <SelectTrigger id="cfg-branch">
                <SelectValue placeholder="Select branch" />
              </SelectTrigger>
              <SelectContent>
                {branches.map((b) => (
                  <SelectItem key={b} value={b}>{b}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : (
            <Input
              id="cfg-branch"
              value={baseBranch}
              placeholder="main"
              onChange={(e) => setBaseBranch(e.target.value)}
            />
          )}
          <p className="text-xs text-muted-foreground">The branch pull requests target.</p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="cfg-stackfile">Stackfile path</Label>
          <Input
            id="cfg-stackfile"
            value={stackfilePath}
            onChange={(e) => setStackfilePath(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Checked when the first preview deploys — a wrong path shows up as a
            Failed environment.
          </p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="cfg-max">Max active previews</Label>
          <Input
            id="cfg-max"
            type="number"
            min={1}
            value={maxActive}
            onChange={(e) => setMaxActive(Number(e.target.value))}
            className="w-28"
          />
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>
      <WizardFooter
        onBack={onBack}
        onContinue={() => void submit()}
        continueLabel="Enable previews"
        continueDisabled={name.trim() === "" || baseBranch.trim() === ""}
        loading={saving}
      />
    </div>
  );
}

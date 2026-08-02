import { useState } from "react";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { BranchField } from "@/components/git-source-picker/branch-field";
import { WizardFooter } from "@/components/wizard-footer";
import { createPreviewConfig } from "@/api/preview-configs";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import {
  configurePhaseSchema, DEFAULT_STACKFILE_PATH, type ConfigurePhaseValues,
} from "@/pages/previews/lib/form-schemas";
import { EnvVarsEditor, type EnvVarFormRow } from "@/pages/previews/components/env-vars-editor";
import type { PickedRepo } from "./enable-repo-wizard";

interface ConfigurePhaseProps {
  repo: PickedRepo;
  onCreated: (configId: string) => void;
  onBack: () => void;
}

export function ConfigurePhase({ repo, onCreated, onBack }: ConfigurePhaseProps) {
  const { defaultProjectName } = useResourceProjects();
  const [name, setName] = useState(repo.fullName.split("/").pop() ?? "");
  const [baseBranch, setBaseBranch] = useState(repo.defaultBranch);
  const [stackfilePath, setStackfilePath] = useState(DEFAULT_STACKFILE_PATH);
  const [maxActive, setMaxActive] = useState(10);
  const [env, setEnv] = useState<EnvVarFormRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof ConfigurePhaseValues, string>>>({});
  const [saving, setSaving] = useState(false);

  const submit = async () => {
    const envRows = env.filter((row) => row.name.trim() !== "");
    const parsed = configurePhaseSchema.safeParse({ name, baseBranch, stackfilePath, maxActive, env: envRows });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({
        name: flat.name?.[0],
        baseBranch: flat.baseBranch?.[0],
        stackfilePath: flat.stackfilePath?.[0],
        maxActive: flat.maxActive?.[0],
        env: flat.env?.[0],
      });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultProjectName) {
      setError("No organization or default project available.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const created = await createPreviewConfig(orgId, defaultProjectName, {
        name: parsed.data.name,
        git_repository: { repo_url: repo.cloneUrl, base_branch: parsed.data.baseBranch },
        stackfile_path: parsed.data.stackfilePath || DEFAULT_STACKFILE_PATH,
        max_active_previews: parsed.data.maxActive,
        env: parsed.data.env,
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

        <FieldShell
          label="Name"
          htmlFor="cfg-name"
          required
          hint="Cannot be changed later."
          error={fieldErrors.name}
        >
          <Input
            id="cfg-name"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              setFieldErrors((prev) => ({ ...prev, name: undefined }));
            }}
            aria-invalid={!!fieldErrors.name}
          />
        </FieldShell>

        <FieldShell
          label="Base branch"
          htmlFor="cfg-branch"
          required
          hint="The branch pull requests target."
          error={fieldErrors.baseBranch}
        >
          <BranchField
            id="cfg-branch"
            value={baseBranch}
            onChange={(value) => {
              setBaseBranch(value);
              setFieldErrors((prev) => ({ ...prev, baseBranch: undefined }));
            }}
            integrationId={repo.integrationId}
            repoFullName={repo.fullName}
          />
        </FieldShell>

        <FieldShell
          label="Stackfile path"
          htmlFor="cfg-stackfile"
          hint="Defines the full stack (services, ports, env). Fetched from the repository on every deploy. A wrong path shows up as a Failed environment."
          error={fieldErrors.stackfilePath}
        >
          <Input
            id="cfg-stackfile"
            value={stackfilePath}
            onChange={(e) => {
              setStackfilePath(e.target.value);
              setFieldErrors((prev) => ({ ...prev, stackfilePath: undefined }));
            }}
            aria-invalid={!!fieldErrors.stackfilePath}
          />
        </FieldShell>

        <FieldShell label="Max active previews" htmlFor="cfg-max" error={fieldErrors.maxActive}>
          <Input
            id="cfg-max"
            type="number"
            min={1}
            value={maxActive}
            onChange={(e) => {
              const n = e.target.valueAsNumber;
              setMaxActive(Number.isNaN(n) ? 1 : Math.max(1, Math.floor(n)));
              setFieldErrors((prev) => ({ ...prev, maxActive: undefined }));
            }}
            className="w-28 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
          />
        </FieldShell>

        <FieldShell
          label="Environment variables (optional)"
          hint="Applied to every preview. Reference saved secrets with the Secret source instead of pasting raw values."
          error={fieldErrors.env}
        >
          <EnvVarsEditor
            value={env}
            onChange={(rows) => {
              setEnv(rows);
              setFieldErrors((prev) => ({ ...prev, env: undefined }));
            }}
          />
        </FieldShell>

        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>
      <WizardFooter
        onBack={onBack}
        onContinue={() => void submit()}
        continueLabel="Enable previews"
        loading={saving}
      />
    </div>
  );
}

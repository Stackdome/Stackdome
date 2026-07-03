import React from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { PlusCircle, GitBranch, Box, Trash2, Database, X } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { MultiSelect } from "@/components/multi-select";
import { DirtyField } from "@/pages/stacks/components/shared/dirty-field";
import { FieldShell } from "@/components/branded";

import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../hooks/use-secrets";

type Resource = Partial<FormStackResourceData>;
type VolumeMount = NonNullable<FormStackResourceData["volume_mounts"]>[number];
type Port = NonNullable<FormStackResourceData["ports"]>[number];

interface StackResourceConfigurationTabProps {
  index: number;
  /** A projected slice of `resource` containing only the fields this tab reads.
   *  Passed as `draft` to <DirtyField> calls; same shape as `baseline`. */
  draft: ConfigurationDraft;
  baseline: ConfigurationDraft | undefined;
  errors: { [field: string]: string | undefined };
  volumes: Partial<VolumeFormData>[];
  allResources?: { name: string; index: number }[];
  secrets: UseSecretsReturn;
  /** Per-field reset by dot-path. */
  onDiscardField?: (path: string) => void;
  /** Patch any subset of resource fields. Identity must be stable across renders. */
  onPatchResource: (patch: Partial<FormStackResourceData>) => void;
}

/** Subset of FormStackResourceData read by the Configuration tab. We pass this
 * projection (instead of the whole resource) so React.memo can skip the tab
 * when only Deployment- or Environment-tab fields change. */
export interface ConfigurationDraft {
  name?: Resource["name"];
  depends_on?: Resource["depends_on"];
  sourceType?: Resource["sourceType"];
  image_spec?: Resource["image_spec"];
  build_spec?: Resource["build_spec"];
  useImageSecret?: Resource["useImageSecret"];
  selectedImageSecretId?: Resource["selectedImageSecretId"];
  useGitSecret?: Resource["useGitSecret"];
  selectedGitSecretId?: Resource["selectedGitSecretId"];
  gitRevisionType?: Resource["gitRevisionType"];
  gitRevisionValue?: Resource["gitRevisionValue"];
  volume_mounts?: Resource["volume_mounts"];
  ports?: Resource["ports"];
}

/** Project a FormStackResourceData down to the keys the Configuration tab needs. */
export function pickConfigurationDraft(resource: Resource): ConfigurationDraft {
  return {
    name: resource.name,
    depends_on: resource.depends_on,
    sourceType: resource.sourceType,
    image_spec: resource.image_spec,
    build_spec: resource.build_spec,
    useImageSecret: resource.useImageSecret,
    selectedImageSecretId: resource.selectedImageSecretId,
    useGitSecret: resource.useGitSecret,
    selectedGitSecretId: resource.selectedGitSecretId,
    gitRevisionType: resource.gitRevisionType,
    gitRevisionValue: resource.gitRevisionValue,
    volume_mounts: resource.volume_mounts,
    ports: resource.ports,
  };
}

const getError = (errors: { [field: string]: string | undefined }, path: string) => {
  if (errors[path]) return errors[path];
  for (const key in errors) {
    if (key === path || key.startsWith(`${path}.`)) return errors[key];
    if (path.startsWith(`${key}.`)) return errors[key];
  }
  return undefined;
};

function StackResourceConfigurationTabImpl({
  index,
  draft,
  baseline,
  errors,
  volumes,
  allResources,
  secrets,
  onDiscardField,
  onPatchResource,
}: StackResourceConfigurationTabProps) {
  const update = onPatchResource;

  // Helper for updating nested build_spec
  const updateBuildSpec = (patch: Partial<NonNullable<FormStackResourceData["build_spec"]>>) => {
    const currentBuildSpec = draft.build_spec || {
      source_context: { git_repo: { repo_url: '' } },
      context_path_within_source: './',
      dockerfile_path: 'Dockerfile',
      image_repository: { external_image_ref: '' },
      insecure_registry: false,
      source_revision: { volume_source_revision: undefined, git_repo_revision: undefined },
    };
    const mergedSourceRevision = {
      volume_source_revision: patch.source_revision?.volume_source_revision ?? currentBuildSpec.source_revision?.volume_source_revision,
      git_repo_revision: patch.source_revision?.git_repo_revision ?? currentBuildSpec.source_revision?.git_repo_revision,
    };
    update({
      build_spec: {
        ...currentBuildSpec,
        ...patch,
        insecure_registry: patch.insecure_registry === undefined ? false : patch.insecure_registry,
        source_revision: mergedSourceRevision,
      },
      image_spec: undefined,
    });
  };

  const updateImageSpec = (patch: Partial<NonNullable<FormStackResourceData["image_spec"]>>) => {
    update({
      image_spec: { ...(draft.image_spec || { image: '' }), ...patch },
      build_spec: undefined,
    });
  };

  const updateDependsOn = (dependsOn: string[]) => {
    update({ depends_on: dependsOn });
  };

  const addVolumeMount = () => {
    update({
      volume_mounts: [
        ...(draft.volume_mounts || []),
        { source_volume_name: "", source_sub_path: "", target_path: "/mnt" },
      ],
    });
  };

  const updateVolumeMount = (
    vmIdx: number,
    patch: Partial<{ source_volume_name: string; source_sub_path: string; target_path: string }>,
  ) => {
    if (patch.target_path && draft.volume_mounts) {
      const isDuplicate = draft.volume_mounts.some(
        (vm: VolumeMount, i: number) => i !== vmIdx && vm.target_path === patch.target_path,
      );
      if (isDuplicate) {
        toast({
          title: "Duplicate Target Path",
          description: "Each volume mount must have a unique target path within a resource.",
          variant: "destructive",
        });
        return;
      }
    }
    update({
      volume_mounts: (draft.volume_mounts || []).map((vm: VolumeMount, i: number) =>
        i === vmIdx ? { ...vm, ...patch } : vm,
      ),
    });
  };

  const removeVolumeMount = (vmIdx: number) => {
    update({
      volume_mounts: (draft.volume_mounts || []).filter((_: VolumeMount, i: number) => i !== vmIdx),
    });
  };

  const addPort = () => {
    const existing = draft.ports || [];
    update({
      ports: [
        ...existing,
        { name: `port-${existing.length + 1}`, number: 80, protocol: "tcp", exposed_to_public: false },
      ],
    });
  };

  const updatePort = (
    pidx: number,
    patch: Partial<{ number: number; protocol: "http" | "tcp"; exposed_to_public: boolean; subdomain_prefix: string }>,
  ) => {
    update({
      ports: (draft.ports || []).map((port: Port, i: number) =>
        i === pidx ? { ...port, ...patch } : port,
      ),
    });
  };

  const removePort = (pidx: number) => {
    update({
      ports: (draft.ports || []).filter((_: Port, i: number) => i !== pidx),
    });
  };

  return (
    <TabsContent value="general" className="pt-4 space-y-6">
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">General</h3>
        <div className="grid gap-5 max-w-3xl">
          <FieldShell
            label="Resource Name"
            htmlFor={`resource-name-${index}`}
            required
            hint="A unique identifier within this stack — lowercase, no spaces."
            error={getError(errors, "name")}
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="name"
              onReset={onDiscardField ? () => onDiscardField("name") : undefined}
            >
              <Input
                id={`resource-name-${index}`}
                placeholder="e.g., api, database, frontend"
                value={draft.name || ""}
                onChange={(e) => update({ name: e.target.value })}
                className={`max-w-xl ${getError(errors, "name") ? "border-danger" : ""}`}
                required
                aria-invalid={!!getError(errors, "name")}
              />
            </DirtyField>
          </FieldShell>

          <FieldShell
            label="Depends On"
            hint="Other resources this service depends on. They are started first."
            error={errors["depends_on"]}
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="depends_on"
              onReset={onDiscardField ? () => onDiscardField("depends_on") : undefined}
            >
              {allResources ? (
                <MultiSelect
                  options={allResources
                    .filter((r) => r.index !== index && r.name && r.name.trim() !== "")
                    .map((r) => ({ label: r.name, value: r.name }))}
                  onValueChange={updateDependsOn}
                  defaultValue={draft.depends_on || []}
                  placeholder={allResources.length <= 1 ? "No other resources available" : "Select dependencies"}
                  disabled={allResources.length <= 1}
                  className="w-full"
                />
              ) : (
                <div className="text-sm text-muted-foreground">No dependency information available</div>
              )}
            </DirtyField>
          </FieldShell>

          <FieldShell
            label="Build From"
            htmlFor={`resource-source-${index}`}
            required
            hint="Build from a pre-built container image or from a Git repository."
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="sourceType"
              onReset={onDiscardField ? () => onDiscardField("sourceType") : undefined}
            >
              <Select
                value={draft.sourceType || "image"}
                onValueChange={(val) => update({ sourceType: val as "image" | "git" })}
              >
                <SelectTrigger className="w-[200px]">
                  <SelectValue placeholder="Select source type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="image">
                    <div className="flex items-center gap-2">
                      <Box size={16} />
                      <span>Container Image</span>
                    </div>
                  </SelectItem>
                  <SelectItem value="git">
                    <div className="flex items-center gap-2">
                      <GitBranch size={16} />
                      <span>Git Repository</span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </DirtyField>
          </FieldShell>
          {draft.sourceType === "image" ? (
            <div className="grid gap-5 max-w-3xl">
              <FieldShell
                label="Container Image"
                htmlFor={`container-image-${index}`}
                required
                hint="Docker image reference, e.g., nginx:latest or ghcr.io/org/app:1.2.3."
                error={getError(errors, "image_spec.image")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="image_spec.image"
                  onReset={onDiscardField ? () => onDiscardField("image_spec.image") : undefined}
                >
                  <Input
                    id={`container-image-${index}`}
                    placeholder="e.g., nginx:latest, redis:7"
                    value={draft.image_spec?.image || ""}
                    onChange={(e) => updateImageSpec({ image: e.target.value })}
                    className={`max-w-xl ${getError(errors, "image_spec.image") ? "border-danger" : ""}`}
                    required={draft.sourceType === "image"}
                    aria-invalid={!!getError(errors, "image_spec.image")}
                  />
                </DirtyField>
              </FieldShell>

              {/* Docker Registry Secret — sub-option of Container Image */}
              <div className="pl-3 border-l border-border space-y-3">
                <div className="flex items-center space-x-2">
                  <Switch
                    id={`use-image-secret-${index}`}
                    checked={draft.useImageSecret || false}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        update({ useImageSecret: checked });
                      } else {
                        update({
                          useImageSecret: false,
                          selectedImageSecretId: undefined,
                        });
                      }
                    }}
                    disabled={secrets.isLoading}
                  />
                  <Label htmlFor={`use-image-secret-${index}`} className="text-[12.5px] font-medium text-muted-foreground">
                    Use secret to pull this image
                  </Label>
                </div>

                {draft.useImageSecret && (
                  <FieldShell label="Select secret">
                    <Select
                      value={draft.selectedImageSecretId || ""}
                      onValueChange={(value) => update({ selectedImageSecretId: value })}
                      disabled={secrets.isLoading || secrets.dockerRegistrySecrets.length === 0}
                    >
                      <SelectTrigger className="w-full max-w-xl">
                        <SelectValue
                          placeholder={
                            secrets.dockerRegistrySecrets.length === 0
                              ? "No docker registry secrets available"
                              : "Select Docker registry secret"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {secrets.dockerRegistrySecrets.map((secret) => (
                          <SelectItem key={secret.id} value={secret.id!}>
                            {secret.name}
                            {secret.description && (
                              <span className="text-muted-foreground ml-2">
                                - {secret.description}
                              </span>
                            )}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FieldShell>
                )}
              </div>
            </div>
          ) : (
            <div className="grid gap-5 max-w-3xl">
              <FieldShell
                label="Git Repository URL"
                htmlFor={`git-repo-${index}`}
                required
                hint="HTTPS or SSH URL of the source repository."
                error={getError(errors, "build_spec.source_context.git_repo.repo_url")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="build_spec.source_context.git_repo.repo_url"
                  onReset={onDiscardField ? () => onDiscardField("build_spec.source_context.git_repo.repo_url") : undefined}
                >
                  <Input
                    id={`git-repo-${index}`}
                    value={draft.build_spec?.source_context?.git_repo?.repo_url || ""}
                    onChange={(e) => updateBuildSpec({ source_context: { git_repo: { repo_url: e.target.value } } })}
                    placeholder="https://github.com/username/repository.git"
                    className={`max-w-xl ${getError(errors, "build_spec.source_context.git_repo.repo_url") ? "border-danger" : ""}`}
                    required={draft.sourceType === "git"}
                    aria-invalid={!!getError(errors, "build_spec.source_context.git_repo.repo_url")}
                  />
                </DirtyField>
              </FieldShell>

              {/* Git Credentials — sub-option of Git Repository URL */}
              <div className="pl-3 border-l border-border space-y-3">
                <div className="flex items-center space-x-2">
                  <Switch
                    id={`use-git-secret-${index}`}
                    checked={draft.useGitSecret || false}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        update({ useGitSecret: checked });
                      } else {
                        update({
                          useGitSecret: false,
                          selectedGitSecretId: undefined,
                        });
                      }
                    }}
                    disabled={secrets.isLoading}
                  />
                  <Label htmlFor={`use-git-secret-${index}`} className="text-[12.5px] font-medium text-muted-foreground">
                    Use Git credentials secret
                  </Label>
                </div>

                {draft.useGitSecret && (
                  <FieldShell label="Select secret">
                    <Select
                      value={draft.selectedGitSecretId || ""}
                      onValueChange={(value) => update({ selectedGitSecretId: value })}
                      disabled={secrets.isLoading || secrets.gitSecrets.length === 0}
                    >
                      <SelectTrigger className="w-full max-w-xl">
                        <SelectValue
                          placeholder={
                            secrets.gitSecrets.length === 0
                              ? "No Git credentials secrets available"
                              : "Select Git credentials"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {secrets.gitSecrets.map((secret) => (
                          <SelectItem key={secret.id} value={secret.id!}>
                            {secret.name}
                            {secret.description && (
                              <span className="text-muted-foreground ml-2">
                                - {secret.description}
                              </span>
                            )}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FieldShell>
                )}
              </div>

              <FieldShell
                label="Image Repository URL"
                htmlFor={`external-image-repo-url-${index}`}
                required
                hint="External registry where built images are pushed, e.g., ghcr.io/your-org/your-image."
                error={getError(errors, "build_spec.image_repository.external_image_ref")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="build_spec.image_repository.external_image_ref"
                  onReset={onDiscardField ? () => onDiscardField("build_spec.image_repository.external_image_ref") : undefined}
                >
                  <Input
                    id={`external-image-repo-url-${index}`}
                    value={draft.build_spec?.image_repository?.external_image_ref || ""}
                    onChange={(e) => updateBuildSpec({ image_repository: { external_image_ref: e.target.value } })}
                    placeholder="e.g., ghcr.io/your-org/your-image"
                    className={`max-w-xl ${getError(errors, "build_spec.image_repository.external_image_ref") ? "border-danger" : ""}`}
                    required={draft.sourceType === "git"}
                    aria-invalid={!!getError(errors, "build_spec.image_repository.external_image_ref")}
                  />
                </DirtyField>
              </FieldShell>

              <FieldShell
                label="Git Revision Type"
                htmlFor={`git-revision-type-${index}`}
                required
                hint="Pin builds to a branch, a specific commit, or a tag."
                error={getError(errors, "gitRevisionType")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="gitRevisionType"
                  onReset={onDiscardField ? () => onDiscardField("gitRevisionType") : undefined}
                >
                  <Select
                    value={draft.gitRevisionType}
                    onValueChange={(val) => update({ gitRevisionType: val as "branch" | "commit" | "tag" })}
                  >
                    <SelectTrigger
                      id={`git-revision-type-${index}`}
                      className={`max-w-xl ${getError(errors, "gitRevisionType") ? "border-danger" : ""}`}
                    >
                      <SelectValue placeholder="Select revision type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="branch">Branch</SelectItem>
                      <SelectItem value="commit">Commit</SelectItem>
                      <SelectItem value="tag">Tag</SelectItem>
                    </SelectContent>
                  </Select>
                </DirtyField>
              </FieldShell>

              {draft.gitRevisionType && (
                <FieldShell
                  label={
                    draft.gitRevisionType === "branch"
                      ? "Branch Name"
                      : draft.gitRevisionType === "commit"
                        ? "Commit Hash"
                        : "Tag Name"
                  }
                  htmlFor={`git-revision-value-${index}`}
                  required
                  hint={
                    draft.gitRevisionType === "branch"
                      ? "The branch to check out, e.g., main."
                      : draft.gitRevisionType === "commit"
                        ? "The full commit SHA to check out."
                        : "The tag to check out, e.g., v1.0.0."
                  }
                  error={getError(errors, "gitRevisionValue")}
                >
                  <DirtyField
                    draft={draft}
                    baseline={baseline}
                    path="gitRevisionValue"
                    onReset={onDiscardField ? () => onDiscardField("gitRevisionValue") : undefined}
                  >
                    <Input
                      id={`git-revision-value-${index}`}
                      value={draft.gitRevisionValue || ""}
                      onChange={(e) => update({ gitRevisionValue: e.target.value })}
                      placeholder={
                        draft.gitRevisionType === "branch"
                          ? "e.g., main, develop"
                          : draft.gitRevisionType === "commit"
                            ? "e.g., a1b2c3d4e5..."
                            : "e.g., v1.0.0"
                      }
                      className={`max-w-xl ${getError(errors, "gitRevisionValue") ? "border-danger" : ""}`}
                      required={!!draft.gitRevisionType}
                      aria-invalid={!!getError(errors, "gitRevisionValue")}
                      onBlur={() => {
                        if (!draft.gitRevisionValue) {
                          update({ gitRevisionValue: "" });
                        }
                      }}
                    />
                  </DirtyField>
                </FieldShell>
              )}
            </div>
          )}
        </div>
      </div>
      <Separator className="my-6" />
      {/* Volume Mounts Section */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Volume Mounts</h3>
        <div className="grid gap-5 max-w-3xl">
          {(draft.volume_mounts || []).map((vm: VolumeMount, vmIdx: number) => (
            <DirtyField
              key={vmIdx}
              draft={draft}
              baseline={baseline}
              path={`volume_mounts.${vmIdx}`}
              onReset={onDiscardField ? () => onDiscardField(`volume_mounts.${vmIdx}`) : undefined}
              compact
            >
              <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_1fr_auto] gap-4 items-start border p-3 rounded-md bg-muted/10">
                <FieldShell
                  label="Volume"
                  htmlFor={`volume-name-${index}-${vmIdx}`}
                  required
                  error={getError(errors, `volume_mounts.${vmIdx}.source_volume_name`)}
                >
                  <Select
                    value={vm.source_volume_name || ""}
                    onValueChange={(value) => updateVolumeMount(vmIdx, { source_volume_name: value })}
                  >
                    <SelectTrigger
                      id={`volume-name-${index}-${vmIdx}`}
                      className={getError(errors, `volume_mounts.${vmIdx}.source_volume_name`) ? "border-danger" : ""}
                    >
                      <SelectValue placeholder="Select volume" />
                    </SelectTrigger>
                    <SelectContent>
                      {(volumes || []).filter((vol) => !!vol.name).length === 0 ? (
                        <div className="p-2 text-sm text-muted-foreground">No volumes available</div>
                      ) : (
                        (volumes || []).filter((vol) => !!vol.name).map((vol, vidx) => (
                          <SelectItem key={vidx} value={vol.name!}>
                            <div className="flex items-center gap-2">
                              <Database className="h-4 w-4" />
                              <span>{vol.name}</span>
                              {vol.spec?.size && <span className="ml-1 text-xs text-muted-foreground">({vol.spec.size})</span>}
                            </div>
                          </SelectItem>
                        ))
                      )}
                    </SelectContent>
                  </Select>
                </FieldShell>
                <FieldShell label="Sub Path" htmlFor={`volume-subpath-${index}-${vmIdx}`}>
                  <Input
                    id={`volume-subpath-${index}-${vmIdx}`}
                    value={vm.source_sub_path || ""}
                    onChange={(e) => updateVolumeMount(vmIdx, { source_sub_path: e.target.value })}
                    placeholder="e.g., data/config"
                  />
                </FieldShell>
                <FieldShell
                  label="Target Path"
                  htmlFor={`volume-target-${index}-${vmIdx}`}
                  required
                  error={getError(errors, `volume_mounts.${vmIdx}.target_path`)}
                >
                  <Input
                    id={`volume-target-${index}-${vmIdx}`}
                    value={vm.target_path || ""}
                    onChange={(e) => updateVolumeMount(vmIdx, { target_path: e.target.value })}
                    placeholder="e.g., /mnt/data"
                    className={getError(errors, `volume_mounts.${vmIdx}.target_path`) ? "border-danger" : ""}
                    required
                  />
                </FieldShell>
                <div className="pt-[26px]">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => removeVolumeMount(vmIdx)}
                    title="Remove volume mount"
                    className="text-danger hover:text-danger hover:bg-danger-bg"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </DirtyField>
          ))}
          <div>
            <Button
              variant="ghost"
              size="sm"
              onClick={addVolumeMount}
              disabled={(volumes || []).length === 0}
            >
              <PlusCircle className="h-4 w-4 mr-2" />Add mount
            </Button>
            {(volumes || []).length === 0 && (
              <p className="text-sm text-muted-foreground mt-2">No volumes available. Add volumes in the Volumes section below.</p>
            )}
          </div>
        </div>
      </div>
      <Separator className="my-6" />
      {/* Ports Section */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Ports</h3>
        <div className="grid gap-5 max-w-3xl">
          {(draft.ports || []).map((port: Port, pidx: number) => (
            <DirtyField
              key={pidx}
              draft={draft}
              baseline={baseline}
              path={`ports.${pidx}`}
              onReset={onDiscardField ? () => onDiscardField(`ports.${pidx}`) : undefined}
              compact
            >
              <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_1fr_auto] gap-4 items-start border p-3 rounded-md bg-muted/10">
                <FieldShell
                  label="Port Number"
                  htmlFor={`port-number-${index}-${pidx}`}
                  required
                  error={getError(errors, `ports.${pidx}.number`)}
                >
                  <Input
                    id={`port-number-${index}-${pidx}`}
                    type="number"
                    min="1"
                    max="65535"
                    value={port.number?.toString() || ""}
                    onChange={(e) => updatePort(pidx, { number: parseInt(e.target.value) || 0 })}
                    className={getError(errors, `ports.${pidx}.number`) ? "border-danger" : ""}
                    required
                  />
                </FieldShell>
                <FieldShell label="Protocol" htmlFor={`port-protocol-${index}-${pidx}`}>
                  <Select
                    value={port.protocol || "tcp"}
                    onValueChange={(value) => updatePort(pidx, { protocol: value as "tcp" | "http" })}
                  >
                    <SelectTrigger id={`port-protocol-${index}-${pidx}`}>
                      <SelectValue placeholder="Select protocol" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="tcp">TCP</SelectItem>
                      <SelectItem value="http">HTTP</SelectItem>
                    </SelectContent>
                  </Select>
                </FieldShell>
                <FieldShell label="Public Access" htmlFor={`port-expose-${index}-${pidx}`}>
                  <div className="flex items-center space-x-2 h-[40px]">
                    <Switch
                      id={`port-expose-${index}-${pidx}`}
                      checked={port.exposed_to_public || false}
                      onCheckedChange={(checked) => updatePort(pidx, { exposed_to_public: checked })}
                    />
                    <Label htmlFor={`port-expose-${index}-${pidx}`} className="text-[12.5px] text-muted-foreground cursor-pointer">
                      {port.exposed_to_public ? "Exposed" : "Internal Only"}
                    </Label>
                  </div>
                </FieldShell>
                <div className="self-center">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => removePort(pidx)}
                    title="Remove port"
                    aria-label="Remove port"
                    className="h-7 w-7 hover:bg-danger-bg hover:text-danger"
                  >
                    <X className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </DirtyField>
          ))}
          <div>
            <Button variant="ghost" size="sm" onClick={addPort}>
              <PlusCircle className="h-4 w-4 mr-2" />Add port
            </Button>
          </div>
        </div>
      </div>
    </TabsContent>
  );
}

/** Memoized so a keystroke in (say) the Environment tab does not re-render
 * Configuration. Parent must pass projected `draft` + stable `onPatchResource`. */
export const StackResourceConfigurationTab = React.memo(StackResourceConfigurationTabImpl);

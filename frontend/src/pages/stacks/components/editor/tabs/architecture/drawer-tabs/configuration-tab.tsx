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
import { PlusCircle, GitBranch, Box, Trash2, Database, X, ArrowUpRight, HardDrive } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { MultiSelect } from "@/components/multi-select";
import { DirtyField } from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/dirty-field";
import { FieldShell } from "@/components/branded";

import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";

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
  /** Per-field reset by dot-path. */
  onDiscardField?: (path: string) => void;
  /** Patch any subset of resource fields. Identity must be stable across renders. */
  onPatchResource: (patch: Partial<FormStackResourceData>) => void;
  /** When provided, the drawer offers inline volume creation (name+size+path)
   *  instead of only selecting a pre-existing volume. */
  onCreateVolume?: (input: { name: string; size: string; targetPath: string }) => void;
  /** When provided, mount rows show a navigate button that pushes the volume's drawer. */
  onOpenVolume?: (name: string) => void;
  /** When true, render mounts as read-only rows (canvas drives mounts via drag/connect). */
  mountsReadOnly?: boolean;
}

/** Subset of FormStackResourceData read by the Configuration tab. We pass this
 * projection (instead of the whole resource) so React.memo can skip the tab
 * when only Deployment- or Environment-tab fields change. */
export interface ConfigurationDraft {
  name?: Resource["name"];
  depends_on?: Resource["depends_on"];
  sourceType?: Resource["sourceType"];
  source?: Resource["source"];
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
    source: resource.source,
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
  onDiscardField,
  onPatchResource,
  onCreateVolume,
  onOpenVolume,
  mountsReadOnly = false,
}: StackResourceConfigurationTabProps) {
  const update = onPatchResource;

  type GitSource = NonNullable<NonNullable<FormStackResourceData["source"]>["git"]>;
  type ImageSource = NonNullable<NonNullable<FormStackResourceData["source"]>["image"]>;

  // Merge a patch into source.git. dockerfile_path/build_context carry the
  // API defaults (they are required on the resolved GitSource type).
  const updateGitSource = (patch: Partial<GitSource>) => {
    const current = draft.source?.git;
    update({
      source: {
        git: {
          repo_url: current?.repo_url ?? '',
          dockerfile_path: current?.dockerfile_path ?? 'Dockerfile',
          build_context: current?.build_context ?? '.',
          branch: current?.branch,
          tag: current?.tag,
          commit: current?.commit,
          push: current?.push,
          integration_id: current?.integration_id,
          ...patch,
        },
      },
    });
  };

  const updateImageSource = (patch: Partial<ImageSource>) => {
    const current = draft.source?.image;
    update({
      source: {
        image: {
          ref: current?.ref ?? '',
          registry_credentials_id: current?.registry_credentials_id,
          ...patch,
        },
      },
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
          title: "Duplicate target path",
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
            hint="A unique identifier within this stack. Use lowercase letters with no spaces."
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
                onValueChange={(val) => {
                  const sourceType = val as "image" | "git";
                  update({
                    sourceType,
                    source: sourceType === "git"
                      ? { git: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." } }
                      : { image: { ref: "" } },
                  });
                }}
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
                error={getError(errors, "source.image.ref")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.image.ref"
                  onReset={onDiscardField ? () => onDiscardField("source.image.ref") : undefined}
                >
                  <Input
                    id={`container-image-${index}`}
                    placeholder="e.g., nginx:latest, redis:7"
                    value={draft.source?.image?.ref || ""}
                    onChange={(e) => updateImageSource({ ref: e.target.value })}
                    className={`max-w-xl ${getError(errors, "source.image.ref") ? "border-danger" : ""}`}
                    required={draft.sourceType === "image"}
                    aria-invalid={!!getError(errors, "source.image.ref")}
                  />
                </DirtyField>
              </FieldShell>
            </div>
          ) : (
            <div className="grid gap-5 max-w-3xl">
              <FieldShell
                label="Git Repository URL"
                htmlFor={`git-repo-${index}`}
                required
                hint="HTTPS or SSH URL of the source repository."
                error={getError(errors, "source.git.repo_url")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.git.repo_url"
                  onReset={onDiscardField ? () => onDiscardField("source.git.repo_url") : undefined}
                >
                  <Input
                    id={`git-repo-${index}`}
                    value={draft.source?.git?.repo_url || ""}
                    onChange={(e) => updateGitSource({ repo_url: e.target.value })}
                    placeholder="https://github.com/username/repository.git"
                    className={`max-w-xl ${getError(errors, "source.git.repo_url") ? "border-danger" : ""}`}
                    required={draft.sourceType === "git"}
                    aria-invalid={!!getError(errors, "source.git.repo_url")}
                  />
                </DirtyField>
              </FieldShell>

              <FieldShell
                label="Push Repository"
                htmlFor={`push-repo-${index}`}
                hint="Optional. External registry to push built images to, e.g., ghcr.io/your-org/your-image. Leave blank to use the internal cluster registry."
                error={getError(errors, "source.git.push.repository")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.git.push.repository"
                  onReset={onDiscardField ? () => onDiscardField("source.git.push.repository") : undefined}
                >
                  <Input
                    id={`push-repo-${index}`}
                    value={draft.source?.git?.push?.repository || ""}
                    onChange={(e) =>
                      updateGitSource({ push: e.target.value ? { repository: e.target.value } : undefined })
                    }
                    placeholder="e.g., ghcr.io/your-org/your-image"
                    className="max-w-xl"
                  />
                </DirtyField>
              </FieldShell>

              <FieldShell
                label="Git Revision Type"
                htmlFor={`git-revision-type-${index}`}
                hint="Optional. Pin builds to a branch, commit, or tag. Defaults to the repository's default branch."
                error={getError(errors, "gitRevisionType")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="gitRevisionType"
                  onReset={onDiscardField ? () => onDiscardField("gitRevisionType") : undefined}
                >
                  <Select
                    value={draft.gitRevisionType ?? "default"}
                    onValueChange={(val) =>
                      val === "default"
                        ? update({ gitRevisionType: undefined, gitRevisionValue: undefined })
                        : update({ gitRevisionType: val as "branch" | "commit" | "tag" })
                    }
                  >
                    <SelectTrigger
                      id={`git-revision-type-${index}`}
                      className={`max-w-xl ${getError(errors, "gitRevisionType") ? "border-danger" : ""}`}
                    >
                      <SelectValue placeholder="Default branch" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="default">Default branch</SelectItem>
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
        {mountsReadOnly ? (
          <div className="grid gap-1.5 max-w-3xl">
            {(draft.volume_mounts || []).length === 0 && (
              <p className="text-sm text-muted-foreground">
                No volumes mounted. Add one from the canvas using “+ Add resource → Volume”.
              </p>
            )}
            {(draft.volume_mounts || []).map((vm: VolumeMount, vmIdx: number) => (
              <div
                key={vmIdx}
                className="flex items-center gap-2 rounded-md border border-border bg-muted/10 px-3 py-2"
              >
                <HardDrive className="h-3.5 w-3.5 shrink-0 text-fg-muted" aria-hidden />
                <span className="truncate font-mono text-[12.5px] text-foreground">{vm.source_volume_name}</span>
                <code className="ml-auto shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                  {vm.target_path}
                </code>
                {onOpenVolume && vm.source_volume_name && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-7 shrink-0 text-fg-muted hover:text-brand"
                    aria-label={`Open volume ${vm.source_volume_name}`}
                    title="Open volume settings"
                    onClick={() => onOpenVolume(vm.source_volume_name!)}
                  >
                    <ArrowUpRight className="size-3.5" />
                  </Button>
                )}
              </div>
            ))}
            <p className="mt-1 text-[12.5px] text-muted-foreground">
              Manage mounts on the canvas: right-click a volume to disconnect, drag a floating volume onto a service to attach.
            </p>
          </div>
        ) : (
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
                  <div className="flex items-end gap-1.5">
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
                          {/* A mount can reference a volume missing from the list
                            (dangling data). Render it as a disabled item so the
                            select still SHOWS the name instead of going blank. */}
                          {vm.source_volume_name &&
                          !(volumes || []).some((vol) => vol.name === vm.source_volume_name) && (
                            <SelectItem value={vm.source_volume_name} disabled>
                              <div className="flex items-center gap-2">
                                <Database className="h-4 w-4" />
                                <span>{vm.source_volume_name}</span>
                                <span className="ml-1 text-xs text-muted-foreground">(missing)</span>
                              </div>
                            </SelectItem>
                          )}
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
                    {onOpenVolume && vm.source_volume_name && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 shrink-0 self-end text-fg-muted hover:text-brand"
                        aria-label={`Open volume ${vm.source_volume_name}`}
                        title="Open volume settings"
                        onClick={() => onOpenVolume(vm.source_volume_name!)}
                      >
                        <ArrowUpRight className="size-3.5" />
                      </Button>
                    )}
                  </div>
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
            {onCreateVolume ? (
              <InlineVolumeAdder onCreate={onCreateVolume} />
            ) : (
              <div>
                <Button variant="ghost" size="sm" onClick={addVolumeMount} disabled={(volumes || []).length === 0}>
                  <PlusCircle className="h-4 w-4 mr-2" />Add mount
                </Button>
                {(volumes || []).length === 0 && (
                  <p className="text-sm text-muted-foreground mt-2">No volumes available. Add volumes in the Volumes section below.</p>
                )}
              </div>
            )}
          </div>
        )}
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
                <FieldShell
                  label="Protocol"
                  htmlFor={`port-protocol-${index}-${pidx}`}
                  error={getError(errors, `ports.${pidx}.protocol`)}
                >
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

function InlineVolumeAdder({ onCreate }: { onCreate: (i: { name: string; size: string; targetPath: string }) => void }) {
  const [name, setName] = React.useState("");
  const [size, setSize] = React.useState("1Gi");
  const [targetPath, setTargetPath] = React.useState("/mnt/data");
  const canAdd = name.trim() !== "" && size.trim() !== "" && targetPath.trim() !== "";
  return (
    <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_1fr_auto] gap-4 items-end border border-dashed p-3 rounded-md">
      <FieldShell label="Volume name" htmlFor="inline-vol-name">
        <Input id="inline-vol-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g., data" />
      </FieldShell>
      <FieldShell label="Size" htmlFor="inline-vol-size">
        <Input id="inline-vol-size" value={size} onChange={(e) => setSize(e.target.value)} placeholder="e.g., 1Gi" />
      </FieldShell>
      <FieldShell label="Mount path" htmlFor="inline-vol-path">
        <Input id="inline-vol-path" value={targetPath} onChange={(e) => setTargetPath(e.target.value)} placeholder="/mnt/data" />
      </FieldShell>
      <Button
        variant="ghost" size="sm" disabled={!canAdd}
        onClick={() => { onCreate({ name: name.trim(), size: size.trim(), targetPath: targetPath.trim() }); setName(""); setSize("1Gi"); setTargetPath("/mnt/data"); }}
      >
        <PlusCircle className="h-4 w-4 mr-2" />Add volume
      </Button>
    </div>
  );
}

/** Memoized so a keystroke in (say) the Environment tab does not re-render
 * Configuration. Parent must pass projected `draft` + stable `onPatchResource`. */
export const StackResourceConfigurationTab = React.memo(StackResourceConfigurationTabImpl);

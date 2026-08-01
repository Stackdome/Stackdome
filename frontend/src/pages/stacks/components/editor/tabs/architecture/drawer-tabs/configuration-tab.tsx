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
import { PlusCircle, GitBranch, Box, Trash2, Database, X, ArrowUpRight, HardDrive } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { MultiSelect } from "@/components/multi-select";
import { DirtyField } from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/dirty-field";
import {
  LedgerDisclosure,
  LedgerRow,
  LedgerSection,
  LedgerSegmented,
} from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/ledger";
import { FieldShell } from "@/components/branded";
import { RepoCombobox } from "@/components/git-source-picker/repo-combobox";
import { ImageRegistrySelect } from "./image-registry-select";
import { splitImageRef, joinImageRef } from "@/pages/stacks/lib/image-ref";

import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import { DEFAULT_BUILD_CONTEXT, DEFAULT_DOCKERFILE_PATH } from "@/pages/stacks/lib/git-build-defaults";

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
  gitCommitPin?: Resource["gitCommitPin"];
  // Abandoned-branch stash for the "Build from" toggle — see the handler below.
  stashedGitSource?: Resource["stashedGitSource"];
  stashedImageSource?: Resource["stashedImageSource"];
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
    gitCommitPin: resource.gitCommitPin,
    stashedGitSource: resource.stashedGitSource,
    stashedImageSource: resource.stashedImageSource,
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
          dockerfile_path: current?.dockerfile_path ?? DEFAULT_DOCKERFILE_PATH,
          build_context: current?.build_context ?? DEFAULT_BUILD_CONTEXT,
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
    const number = 80;
    update({
      ports: [
        ...existing,
        // Name derived from the number (port-<number>) so outputs read e.g.
        // url.port-8080 instead of a meaningless positional url.port-2. k8s port
        // names must contain a letter, so a bare number can't be the name.
        { name: `port-${number}`, number, protocol: "tcp", exposed_to_public: false },
      ],
    });
  };

  const updatePort = (
    pidx: number,
    patch: Partial<{ number: number; protocol: "http" | "tcp"; exposed_to_public: boolean; subdomain_prefix: string }>,
  ) => {
    update({
      ports: (draft.ports || []).map((port: Port, i: number) => {
        if (i !== pidx) return port;
        const next = { ...port, ...patch };
        // Re-derive the auto name (url.port-8080) only while it still matches the
        // auto PATTERN — not the current number, which a cleared field desyncs —
        // so a hand-set name is never overwritten.
        if (patch.number !== undefined && (!port.name || /^port-(\d+|undefined)$/.test(port.name))) {
          next.name = `port-${patch.number}`;
        }
        return next;
      }),
    });
  };

  const removePort = (pidx: number) => {
    update({
      ports: (draft.ports || []).filter((_: Port, i: number) => i !== pidx),
    });
  };

  const ports = draft.ports || [];
  const exposedCount = ports.filter((port: Port) => port.exposed_to_public).length;
  const portsMeta =
    ports.length === 0
      ? "none"
      : exposedCount > 0
        ? `${exposedCount} exposed`
        : `${ports.length} internal`;

  const mounts = draft.volume_mounts || [];
  const mountsMeta = mountsReadOnly
    ? mounts.length === 0
      ? "managed on canvas"
      : `${mounts.length} ${mounts.length === 1 ? "volume" : "volumes"} · managed on canvas`
    : `${mounts.length} ${mounts.length === 1 ? "volume" : "volumes"}`;

  return (
    <TabsContent value="general" className="pt-1">
      <LedgerSection label="General">
        <LedgerRow
          label="Name"
          htmlFor={`resource-name-${index}`}
          required
          meta="lowercase · unique in stack"
          error={getError(errors, "name")}
        >
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="name"
            compact
            onReset={onDiscardField ? () => onDiscardField("name") : undefined}
          >
            <Input
              id={`resource-name-${index}`}
              placeholder="e.g., api, database, frontend"
              value={draft.name || ""}
              onChange={(e) => update({ name: e.target.value })}
              className={`h-9 text-[13.5px] ${getError(errors, "name") ? "border-danger" : ""}`}
              required
              aria-invalid={!!getError(errors, "name")}
            />
          </DirtyField>
        </LedgerRow>

        <LedgerRow label="Depends on" meta="started first" error={errors["depends_on"]}>
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="depends_on"
            compact
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
        </LedgerRow>
      </LedgerSection>

      <LedgerSection
        label="Source"
        meta={draft.sourceType === "git" ? "git repository" : "container image"}
      >
        <LedgerRow label="Build from">
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="sourceType"
            compact
            onReset={onDiscardField ? () => onDiscardField("sourceType") : undefined}
          >
            <LedgerSegmented
              aria-label="Build from"
              value={draft.sourceType || "image"}
              onValueChange={(val) => {
                const sourceType = val as "image" | "git";
                // The API rejects a source with both `git` and `image` set
                // (source_conflict), so the abandoned branch can't stay live
                // in `source`. Stash it in a form-only field instead of
                // discarding it, and restore the other branch from its own
                // stash (falling back to fresh defaults the first time).
                if (sourceType === "git") {
                  update({
                    sourceType,
                    source: { git: draft.stashedGitSource ?? { repo_url: "", dockerfile_path: DEFAULT_DOCKERFILE_PATH, build_context: DEFAULT_BUILD_CONTEXT } },
                    stashedImageSource: draft.source?.image ?? draft.stashedImageSource,
                  });
                } else {
                  update({
                    sourceType,
                    source: { image: draft.stashedImageSource ?? { ref: "" } },
                    stashedGitSource: draft.source?.git ?? draft.stashedGitSource,
                  });
                }
              }}
              options={[
                { value: "image", label: "Container image", icon: <Box size={15} /> },
                { value: "git", label: "Git repository", icon: <GitBranch size={15} /> },
              ]}
            />
          </DirtyField>
        </LedgerRow>

        {draft.sourceType === "image" ? (
          <>
            <LedgerRow label="Registry" htmlFor={`image-registry-${index}`}>
              <DirtyField
                draft={draft}
                baseline={baseline}
                path="source.image"
                compact
                onReset={onDiscardField ? () => onDiscardField("source.image") : undefined}
              >
                <ImageRegistrySelect
                  id={`image-registry-${index}`}
                  imageRef={draft.source?.image?.ref || ""}
                  registryCredentialsId={draft.source?.image?.registry_credentials_id}
                  onChange={(patch) =>
                    updateImageSource({ ref: patch.ref, registry_credentials_id: patch.registry_credentials_id })
                  }
                />
              </DirtyField>
            </LedgerRow>

            <LedgerRow
              label="Image reference"
              htmlFor={`container-image-${index}`}
              required
              alignTop
              error={getError(errors, "source.image.ref")}
            >
              <DirtyField
                draft={draft}
                baseline={baseline}
                path="source.image"
                compact
                onReset={onDiscardField ? () => onDiscardField("source.image") : undefined}
              >
                {(() => {
                  const { host, remainder } = splitImageRef(draft.source?.image?.ref || "");
                  return (
                    <div className="flex items-center gap-1">
                      {host && (
                        <span className="rounded bg-muted px-1.5 py-1 font-mono text-[11px] text-muted-foreground">
                          {host}/
                        </span>
                      )}
                      <Input
                        id={`container-image-${index}`}
                        placeholder={host ? "e.g., acme/api:1.4.2" : "e.g., nginx:latest, redis:7"}
                        value={remainder}
                        onChange={(e) => {
                          const typed = e.target.value;
                          // A pasted full ref (with its own host) replaces the
                          // whole ref outright; otherwise compose against the
                          // active chip host as before.
                          const { host: typedHost } = splitImageRef(typed);
                          updateImageSource({ ref: typedHost ? typed : joinImageRef(host, typed) });
                        }}
                        className={`h-9 flex-1 font-mono text-[12.5px] ${getError(errors, "source.image.ref") ? "border-danger" : ""}`}
                        required={draft.sourceType === "image"}
                        aria-invalid={!!getError(errors, "source.image.ref")}
                      />
                    </div>
                  );
                })()}
              </DirtyField>
            </LedgerRow>
          </>
        ) : (
          <>
            <LedgerRow
              label="Repository"
              htmlFor={`git-repo-${index}`}
              required
              alignTop
              error={getError(errors, "source.git.repo_url")}
            >
              <DirtyField
                draft={draft}
                baseline={baseline}
                path="source.git.repo_url"
                compact
                onReset={onDiscardField ? () => {
                  onDiscardField("source.git.repo_url");
                  onDiscardField("source.git.integration_id");
                } : undefined}
              >
                <RepoCombobox
                  id={`git-repo-${index}`}
                  value={draft.source?.git?.repo_url || ""}
                  integrationId={draft.source?.git?.integration_id}
                  onChange={(pick) =>
                    updateGitSource({ repo_url: pick.repo_url, integration_id: pick.integration_id })
                  }
                  hasError={!!getError(errors, "source.git.repo_url")}
                />
              </DirtyField>
            </LedgerRow>

            <LedgerRow
              label="Revision"
              htmlFor={`git-revision-type-${index}`}
              error={getError(errors, "gitRevisionType")}
            >
              <DirtyField
                draft={draft}
                baseline={baseline}
                path="gitRevisionType"
                compact
                onReset={onDiscardField ? () => onDiscardField("gitRevisionType") : undefined}
              >
                <Select
                  value={draft.gitRevisionType ?? "default"}
                  onValueChange={(val) =>
                    val === "default"
                      ? update({ gitRevisionType: undefined, gitRevisionValue: undefined, gitCommitPin: undefined })
                      : update({ gitRevisionType: val as "branch" | "tag" })
                  }
                >
                  <SelectTrigger
                    id={`git-revision-type-${index}`}
                    className={`h-9 w-full text-[13px] ${getError(errors, "gitRevisionType") ? "border-danger" : ""}`}
                  >
                    <SelectValue placeholder="Default branch" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="default">Default branch</SelectItem>
                    <SelectItem value="branch">Branch</SelectItem>
                    <SelectItem value="tag">Tag</SelectItem>
                  </SelectContent>
                </Select>
              </DirtyField>
            </LedgerRow>

            {draft.gitRevisionType && (
              <LedgerRow
                label={draft.gitRevisionType === "branch" ? "Branch name" : "Tag name"}
                htmlFor={`git-revision-value-${index}`}
                required
                alignTop
                error={getError(errors, "gitRevisionValue")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="gitRevisionValue"
                  compact
                  onReset={onDiscardField ? () => onDiscardField("gitRevisionValue") : undefined}
                >
                  <Input
                    id={`git-revision-value-${index}`}
                    value={draft.gitRevisionValue || ""}
                    onChange={(e) => update({ gitRevisionValue: e.target.value })}
                    placeholder={draft.gitRevisionType === "branch" ? "e.g., main, develop" : "e.g., v1.0.0"}
                    className={`h-9 font-mono text-[12.5px] ${getError(errors, "gitRevisionValue") ? "border-danger" : ""}`}
                    required={!!draft.gitRevisionType}
                    aria-invalid={!!getError(errors, "gitRevisionValue")}
                    onBlur={() => {
                      if (!draft.gitRevisionValue) {
                        update({ gitRevisionValue: "" });
                      }
                    }}
                  />
                </DirtyField>
              </LedgerRow>
            )}

            <LedgerRow
              label="Pin to commit"
              htmlFor={`git-commit-pin-${index}`}
              hint="Optional commit SHA. Builds stay on this commit until unpinned."
              alignTop
              error={getError(errors, "gitCommitPin")}
            >
              <DirtyField
                draft={draft}
                baseline={baseline}
                path="gitCommitPin"
                compact
                onReset={onDiscardField ? () => onDiscardField("gitCommitPin") : undefined}
              >
                <Input
                  id={`git-commit-pin-${index}`}
                  value={draft.gitCommitPin || ""}
                  onChange={(e) => update({ gitCommitPin: e.target.value || undefined })}
                  placeholder="e.g., a1b2c3d4e5..."
                  disabled={!draft.gitRevisionType && !draft.gitCommitPin}
                  className={`h-9 font-mono text-[12.5px] ${getError(errors, "gitCommitPin") ? "border-danger" : ""}`}
                  aria-invalid={!!getError(errors, "gitCommitPin")}
                />
              </DirtyField>
            </LedgerRow>

            <LedgerDisclosure label="Advanced" meta="build & push">
              <LedgerRow
                label="Dockerfile path"
                htmlFor={`dockerfile-path-${index}`}
                hint="Relative to the build context."
                error={getError(errors, "source.git.dockerfile_path")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.git.dockerfile_path"
                  compact
                  onReset={onDiscardField ? () => onDiscardField("source.git.dockerfile_path") : undefined}
                >
                  <Input
                    id={`dockerfile-path-${index}`}
                    value={draft.source?.git?.dockerfile_path ?? ""}
                    onChange={(e) => updateGitSource({ dockerfile_path: e.target.value })}
                    onBlur={(e) => {
                      if (!e.target.value.trim()) updateGitSource({ dockerfile_path: DEFAULT_DOCKERFILE_PATH });
                    }}
                    placeholder="Dockerfile"
                    className="h-9 font-mono text-[12.5px]"
                  />
                </DirtyField>
              </LedgerRow>

              <LedgerRow
                label="Build context"
                htmlFor={`build-context-${index}`}
                hint="Directory passed to the image build."
                error={getError(errors, "source.git.build_context")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.git.build_context"
                  compact
                  onReset={onDiscardField ? () => onDiscardField("source.git.build_context") : undefined}
                >
                  <Input
                    id={`build-context-${index}`}
                    value={draft.source?.git?.build_context ?? ""}
                    onChange={(e) => updateGitSource({ build_context: e.target.value })}
                    onBlur={(e) => {
                      if (!e.target.value.trim()) updateGitSource({ build_context: DEFAULT_BUILD_CONTEXT });
                    }}
                    placeholder="."
                    className="h-9 font-mono text-[12.5px]"
                  />
                </DirtyField>
              </LedgerRow>

              <LedgerRow
                label="Push registry"
                htmlFor={`push-repo-${index}`}
                alignTop
                hint="Blank uses the internal cluster registry."
                error={getError(errors, "source.git.push.repository")}
              >
                <DirtyField
                  draft={draft}
                  baseline={baseline}
                  path="source.git.push.repository"
                  compact
                  onReset={onDiscardField ? () => onDiscardField("source.git.push.repository") : undefined}
                >
                  <Input
                    id={`push-repo-${index}`}
                    value={draft.source?.git?.push?.repository || ""}
                    onChange={(e) =>
                      updateGitSource({ push: e.target.value ? { repository: e.target.value } : undefined })
                    }
                    placeholder="e.g., ghcr.io/your-org/your-image"
                    className="h-9 font-mono text-[12.5px]"
                  />
                </DirtyField>
              </LedgerRow>
            </LedgerDisclosure>
          </>
        )}
      </LedgerSection>

      <LedgerSection label="Ports" meta={portsMeta}>
        {ports.map((port: Port, pidx: number) => (
          <LedgerRow
            key={pidx}
            label={`Port ${pidx + 1}`}
            htmlFor={`port-number-${index}-${pidx}`}
            error={
              getError(errors, `ports.${pidx}.number`) || getError(errors, `ports.${pidx}.protocol`)
            }
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path={`ports.${pidx}`}
              compact
              onReset={onDiscardField ? () => onDiscardField(`ports.${pidx}`) : undefined}
            >
              <div className="flex items-center gap-2.5">
                <Input
                  id={`port-number-${index}-${pidx}`}
                  inputMode="numeric"
                  value={port.number?.toString() ?? ""}
                  onChange={(e) => {
                    const digits = e.target.value.replace(/\D/g, "");
                    updatePort(pidx, { number: digits === "" ? undefined : parseInt(digits, 10) });
                  }}
                  className={`h-9 w-[84px] shrink-0 font-mono text-[13px] ${getError(errors, `ports.${pidx}.number`) ? "border-danger" : ""}`}
                  required
                />
                <Select
                  value={port.protocol || "tcp"}
                  onValueChange={(value) => updatePort(pidx, { protocol: value as "tcp" | "http" })}
                >
                  <SelectTrigger
                    aria-label="Protocol"
                    className="h-9 w-[92px] shrink-0 text-[13px]"
                  >
                    <SelectValue placeholder="Protocol" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="tcp">TCP</SelectItem>
                    <SelectItem value="http">HTTP</SelectItem>
                  </SelectContent>
                </Select>
                <div className="ml-auto flex items-center gap-2">
                  <Label
                    htmlFor={`port-expose-${index}-${pidx}`}
                    className="cursor-pointer font-mono text-[11px] text-muted-foreground"
                  >
                    {port.exposed_to_public ? "public" : "internal"}
                  </Label>
                  <Switch
                    id={`port-expose-${index}-${pidx}`}
                    checked={port.exposed_to_public || false}
                    onCheckedChange={(checked) => updatePort(pidx, { exposed_to_public: checked })}
                  />
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removePort(pidx)}
                  title="Remove port"
                  aria-label="Remove port"
                  className="h-7 w-7 shrink-0 text-fg-muted hover:bg-danger-bg hover:text-danger"
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </div>
            </DirtyField>
          </LedgerRow>
        ))}
        <Button
          variant="outline"
          size="sm"
          onClick={addPort}
          className="ml-1.5 mt-3 h-8 gap-1.5 rounded-md border-border font-mono text-[12px] font-normal text-muted-foreground hover:border-brand/50 hover:text-brand"
        >
          <PlusCircle className="h-3.5 w-3.5" />
          add port
        </Button>
      </LedgerSection>

      <LedgerSection label="Mounts" meta={mountsMeta}>
        {mountsReadOnly ? (
          <div>
            {mounts.length === 0 && (
              <p className="px-1.5 py-2.5 text-[12.5px] text-muted-foreground">
                No volumes mounted. Add one from the canvas using “+ Add resource → Volume”.
              </p>
            )}
            {mounts.map((vm: VolumeMount, vmIdx: number) => (
              <div key={vmIdx} className="border-b border-secondary/80 py-1">
                <div className="flex items-center gap-3 rounded-md px-1.5 py-1.5 transition-colors hover:bg-muted/20">
                  <div className="flex w-[150px] shrink-0 items-center gap-2 text-[13px] text-foreground/80 dark:text-fg-2">
                    <HardDrive className="h-3.5 w-3.5 shrink-0 text-fg-muted" aria-hidden />
                    <span className="truncate">{vm.source_volume_name}</span>
                  </div>
                  <code className="shrink-0 rounded-[3px] bg-secondary px-2 py-1 font-mono text-[11.5px] text-muted-foreground">
                    {vm.target_path}
                  </code>
                  <span className="ml-auto shrink-0 font-mono text-[10.5px] text-fg-muted/70">
                  drag a volume onto the node to attach
                  </span>
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
              </div>
            ))}
          </div>
        ) : (
          <div className="grid gap-5 pt-2">
            {mounts.map((vm: VolumeMount, vmIdx: number) => (
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
      </LedgerSection>
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

import React, { useMemo, useState } from "react";
import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PlusCircle, X, GitBranch, Box, Trash2, Database, Upload, FileText, Copy, Info } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { MultiSelect } from "@/components/multi-select";
import { ApiStackResourceStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { dirtyTabsForResource, envRowsDiff, isResourceDirty } from "@/pages/stacks/lib/stack-diff";
import { DirtyField } from "@/pages/stacks/components/shared/dirty-field";
import type { z } from "zod";
import { variantFromState } from "@/components/branded";

import type { FormStackResourceData, FormEnvVarData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../hooks/use-secrets";
import { EnvRow, type EnvFrom, type EnvRowErrors, type AddonBindingPatch } from "./env-row";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import type { PostgresAddon } from "@/api/addons";

export type AddonGroupStateMap = Map<string, "idle" | "editing-binding" | "detaching">;

interface StackResourceItemProps {
  resource: Partial<FormStackResourceData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  onChange: (index: number, updatedResource: Partial<FormStackResourceData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  volumes?: Partial<VolumeFormData>[];
  allResources?: { name: string; index: number }[];
  secrets: UseSecretsReturn;
  addons: PostgresAddon[];
  addonNameById: Map<string, string>;
  addonGroupState?: AddonGroupStateMap;
  onEditAddonBinding?: (addonId: string) => void;
  onDetachAddon?: (addonId: string) => void;
  onCancelDetachAddon?: (addonId: string) => void;
  /** Baseline snapshot of this resource. When provided, the component renders dirty visualization (modified row tints, tab dots, Modified pill) and exposes per-row reset. */
  baselineResource?: Partial<FormStackResourceData>;
  /** Reset a single env row to its baseline value. Required for the per-row reset arrow to render. */
  onDiscardEnvRow?: (envIdx: number) => void;
  /** Discard all changes for this resource. Required for the Modified pill ✕ affordance. */
  onDiscardResource?: () => void;
  /** Discard a single field on this resource by dot-path. Required for per-field reset arrows. */
  onDiscardField?: (path: string) => void;
}

const getError = (errors: { [field: string]: string | undefined }, path: string) => {
  // Check for direct match
  if (errors[path]) return errors[path];

  // Check if there's a nested error (handles nested objects like image_spec.image)
  // This makes the UI component work with both flattened and nested error structures
  for (const key in errors) {
    if (key === path || key.startsWith(`${path}.`)) {
      return errors[key];
    }

    // Handle the reverse case where path is more specific than the error key
    // For example, if error key is "image_spec" but we're checking for "image_spec.image"
    if (path.startsWith(`${key}.`)) {
      return errors[key];
    }
  }

  return undefined;
};

function StackResourceItemImpl({
  resource,
  index,
  itemRef,
  onChange,
  onRemove,
  errors,
  volumes = [],
  allResources: _allResources,
  secrets,
  addons,
  addonNameById,
  baselineResource,
  onDiscardEnvRow,
  onDiscardResource,
  onDiscardField,
}: StackResourceItemProps) {
  // Helper for updating resource fields
  const update = (patch: Partial<FormStackResourceData>) => {
    onChange(index, { ...resource, ...patch });
  };

  // Per-row diff status for env vars, used to tint modified rows + render reset arrow.
  const envRowStatuses = useMemo(() => {
    if (!baselineResource) return [];
    const draftRows = (resource.execution_config?.environment_variables || []) as Array<Record<string, unknown>>;
    const baselineRows = (baselineResource.execution_config?.environment_variables || []) as Array<Record<string, unknown>>;
    return envRowsDiff(draftRows, baselineRows);
  }, [resource.execution_config?.environment_variables, baselineResource]);

  // Per-tab dirty bucketing → renders the small brand dot next to each tab label.
  const dirtyTabs = useMemo(
    () => (baselineResource ? dirtyTabsForResource(resource, baselineResource) : { configuration: false, deployment: false, environment: false }),
    [resource, baselineResource],
  );

  const isDirty = baselineResource ? isResourceDirty(resource, baselineResource) : false;

  const [dirtyEnvRows, setDirtyEnvRows] = useState<Set<number>>(new Set());
  const markEnvRowDirty = (envIdx: number) => {
    setDirtyEnvRows((prev) => {
      if (prev.has(envIdx)) return prev;
      const next = new Set(prev);
      next.add(envIdx);
      return next;
    });
  };

  // Helper for updating nested build_spec
  const updateBuildSpec = (patch: Partial<NonNullable<FormStackResourceData["build_spec"]>>) => {
    const currentBuildSpec = resource.build_spec || {
      source_context: { git_repo: { repo_url: '' } },
      context_path_within_source: './',
      dockerfile_path: 'Dockerfile',
      image_repository: { external_image_repo_url: '' },
      insecure_registry: false,
      source_revision: { volume_source_revision: undefined, git_repo_revision: undefined },
    };
    // Always provide both keys for source_revision
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
  // Helper for updating nested image_spec
  const updateImageSpec = (patch: Partial<NonNullable<FormStackResourceData["image_spec"]>>) => {
    update({
      image_spec: { ...(resource.image_spec || { image: '' }), ...patch },
      build_spec: undefined,
    });
  };

  // Helper for adding an environment variable (defaults to a stack-literal row)
  const addEnvVar = (next: FormEnvVarData = { from: "stack", name: "", value: "" }) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [
          ...(resource.execution_config?.environment_variables || []),
          next,
        ],
      },
    });
  };

  // Helper for replacing an environment variable row entirely. Because rows
  // are a discriminated union, partial-merge is unsafe across `from` flips,
  // so callers pass in the full next row.
  const replaceEnvVar = (envIdx: number, next: FormEnvVarData) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: (resource.execution_config?.environment_variables || []).map((env, i) =>
          i === envIdx ? next : env
        ),
      },
    });
  };
  const removeEnvVar = (envIdx: number) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: (resource.execution_config?.environment_variables || []).filter((_, i) => i !== envIdx),
      },
    });
  };

  // Helper for switching a row's `from` discriminator. Each branch swaps in
  // an empty variant; the user fills in the bindings via the inline pickers.
  const switchRowFrom = (envIdx: number, from: EnvFrom) => {
    const current = resource.execution_config?.environment_variables?.[envIdx];
    if (!current) return;
    if (from === "stack") {
      replaceEnvVar(envIdx, { from: "stack", name: current.name, value: "" });
    } else if (from === "secret") {
      replaceEnvVar(envIdx, { from: "secret", name: current.name, secretId: "", secretKey: "" });
    } else if (from === "addon") {
      replaceEnvVar(envIdx, {
        from: "addon",
        name: current.name,
        addonType: "postgres",
        addonId: "",
        database: undefined,
        superuser: false,
        // credField is optional in the form schema; left undefined until the user picks one.
        credField: undefined,
      });
    }
  };

  // Apply a patch from the inline addon pickers. `database: null` means
  // "explicitly clear" (used by All databases), undefined means unchanged.
  const onChangeAddonForRow = (envIdx: number, patch: AddonBindingPatch) => {
    const current = resource.execution_config?.environment_variables?.[envIdx];
    if (!current || current.from !== "addon") return;
    const nextDatabase =
      patch.database === null
        ? undefined
        : patch.database === undefined
          ? current.database
          : patch.database;
    replaceEnvVar(envIdx, {
      ...current,
      addonId: patch.addonId ?? current.addonId,
      database: nextDatabase,
      superuser: patch.superuser ?? current.superuser,
      credField: patch.credField ?? current.credField,
    });
    markEnvRowDirty(envIdx);
  };

  const addVolumeMount = () => {
    update({
      volume_mounts: [
        ...(resource.volume_mounts || []),
        { source_volume_name: "", source_sub_path: "", target_path: "/mnt" },
      ],
    });
  };

  // Helper for adding multiple environment variables at once
  const addMultipleEnvVars = (envVars: Array<{name: string, value: string}>) => {
    // Filter out empty entries and duplicates
    const filteredVars = envVars.filter(env => env.name.trim() !== "");

    // Get current env vars
    const currentVars = resource.execution_config?.environment_variables || [];

    // Create a map of existing var names for quick lookup
    const existingVarNames = new Set(currentVars.map(env => env.name));

    // Filter out duplicates and add new vars as stack-literal rows
    const newVars: FormEnvVarData[] = filteredVars
      .filter(env => !existingVarNames.has(env.name))
      .map(env => ({
        from: "stack" as const,
        name: env.name,
        value: env.value,
      }));

    if (newVars.length === 0) {
      toast({
        title: "No new variables added",
        description: "All variables already exist or are invalid",
        variant: "destructive"
      });
      return;
    }

    // Update with combined variables
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [...currentVars, ...newVars]
      }
    });

    toast({
      title: "Environment variables added",
      description: `Added ${newVars.length} new environment variables`,
      variant: "default"
    });
  };

  // Parse environment variables in KEY=VALUE format with optional comments
  const parseEnvContent = (content: string): Array<{name: string, value: string}> => {
    return content.split('\n')
      .filter(line => line.trim() && !line.trim().startsWith('#'))
      .map(line => {
        const [name, ...valueParts] = line.split('=');
        return {
          name: name.trim(),
          value: valueParts.join('=').trim() // Rejoin in case value contains = characters
        };
      });
  };

  // Handler for uploading .env files
  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (!event.target.files || event.target.files.length === 0) return;

    const file = event.target.files[0];
    const reader = new FileReader();

    reader.onload = (e) => {
      if (!e.target?.result) return;

      const content = e.target.result.toString();
      const envVars = parseEnvContent(content);
      addMultipleEnvVars(envVars);
    };
    reader.readAsText(file);

    // Reset the input
    event.target.value = '';
  };



  const updateVolumeMount = (vmIdx: number, patch: Partial<{ source_volume_name: string, source_sub_path: string, target_path: string }>) => {
    // Check for duplicate target paths if updating target_path
    if (patch.target_path && resource.volume_mounts) {
      const isDuplicate = resource.volume_mounts.some(
        (vm, i) => i !== vmIdx && vm.target_path === patch.target_path
      );

      if (isDuplicate) {
        // Show an error using toast or modify the patch to trigger error display
        toast({
          title: "Duplicate Target Path",
          description: "Each volume mount must have a unique target path within a resource.",
          variant: "destructive",
        });
        return; // Don't update if duplicate
      }
    }

    update({
      volume_mounts: (resource.volume_mounts || []).map((vm, i) =>
        i === vmIdx ? { ...vm, ...patch } : vm
      ),
    });
  };

  const removeVolumeMount = (vmIdx: number) => {
    update({
      volume_mounts: (resource.volume_mounts || []).filter((_, i) => i !== vmIdx),
    });
  };

  // Helper for updating depends_on
  const updateDependsOn = (dependsOn: string[]) => {
    update({ depends_on: dependsOn });
  };

  // Helper for adding a port
  const addPort = () => {
    update({
      ports: [
        ...(resource.ports || []),
        { number: 80, protocol: "tcp", exposed_to_public: false },
      ],
    });
  };

  // Helper for updating a port
  const updatePort = (pidx: number, patch: Partial<{ number: number, protocol: "http" | "tcp", exposed_to_public: boolean, subdomain_prefix: string }>) => {
    update({
      ports: (resource.ports || []).map((port, i) =>
        i === pidx ? { ...port, ...patch } : port
      ),
    });
  };

  // Helper for removing a port
  const removePort = (pidx: number) => {
    update({
      ports: (resource.ports || []).filter((_, i) => i !== pidx),
    });
  };

  // Status semantics
  const statusObj = (resource.status ?? {}) as z.infer<typeof ApiStackResourceStatusSchema>;
  const statusVariant = variantFromState(statusObj.state);
  const statusDotColor =
    statusVariant === "ready" ? "bg-success"
    : statusVariant === "error" ? "bg-danger"
    : statusVariant === "pending" ? "bg-warn"
    : "bg-muted-foreground";

  return (
    <TooltipProvider>
      <AccordionItem value={String(index)} className="border-t border-border first:border-t-0">
        <AccordionTrigger
          ref={itemRef}
          className="px-4 py-3 hover:bg-muted/40 data-[state=open]:bg-muted/30 rounded-t-md [&[data-state=open]]:rounded-b-none"
        >
          <div className="flex items-center gap-3 text-left flex-grow">
            <Tooltip delayDuration={300}>
              <TooltipTrigger asChild>
                <span className={`h-2 w-2 rounded-full shrink-0 ${statusDotColor}`}></span>
              </TooltipTrigger>
              <TooltipContent side="top">
                <p className="capitalize">{statusObj.state || 'Pending'}</p>
              </TooltipContent>
            </Tooltip>
            <div className="flex flex-col flex-grow min-w-0">
              <span className="font-medium flex items-center gap-2">
                {resource.name || `Resource ${index + 1}`}
              </span>
              <span className="text-sm text-muted-foreground truncate">
                {resource.sourceType === "image" ? (
                  <span className="flex items-center gap-1.5">
                    <Box className="h-3.5 w-3.5" />
                    <span>{resource.image_spec?.image || "No image specified"}</span>
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5">
                    <GitBranch className="h-3.5 w-3.5" />
                    <span>
                      {resource.build_spec?.source_context?.git_repo?.repo_url || "No repository specified"}
                      {resource.gitRevisionType && resource.gitRevisionValue && (
                        <span className="ml-1 text-xs bg-muted/50 px-1.5 py-0.5 rounded-full">
                          {resource.gitRevisionType === "branch" && "Branch: "}
                          {resource.gitRevisionType === "tag" && "Tag: "}
                          {resource.gitRevisionType === "commit" && "SHA: "}
                          {resource.gitRevisionValue}
                        </span>
                      )}
                    </span>
                  </span>
                )}
              </span>
              {errors._form && (
                <span className="text-xs text-danger mt-0.5 pl-6">{errors._form}</span>
              )}
            </div>
            {isDirty && onDiscardResource && (
              <div className="ml-auto flex items-center shrink-0 mr-2" onClick={(e) => e.stopPropagation()}>
                <span className="inline-flex items-center gap-1.5 rounded-md border border-brand-border bg-brand-bg pl-2 pr-1 py-0.5 text-[11px] font-medium text-brand">
                  Modified
                  <button
                    type="button"
                    onClick={onDiscardResource}
                    aria-label="Discard changes to this resource"
                    title="Discard changes to this resource"
                    className="inline-flex h-4 w-4 items-center justify-center rounded hover:bg-brand/15"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              </div>
            )}
          </div>
        </AccordionTrigger>
        <AccordionContent className="bg-secondary border-t border-border pb-4 pt-4 px-1">
          <div className="px-4 space-y-4">
            <Tabs defaultValue="general" className="w-full">
              <div className="mt-1 mb-3">
                <TabsList className="w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-1 px-2">
                  <TabsTrigger value="general" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Configuration
                    {dirtyTabs.configuration && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="deployment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Deployment
                    {dirtyTabs.deployment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="environment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Environment
                    {dirtyTabs.environment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                </TabsList>
              </div>

              {/* General Section (always at top) */}
              <TabsContent value="general" className="pt-4 space-y-6">
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">General</h3>
                  <div className="grid gap-4 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`resource-name-${index}`} className="text-sm font-medium">
                          Resource Name <span className="text-danger">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">A unique name for this resource</TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField
                        draft={resource}
                        baseline={baselineResource}
                        path="name"
                        onReset={onDiscardField ? () => onDiscardField("name") : undefined}
                      >
                        <Input
                          id={`resource-name-${index}`}
                          placeholder="e.g., api, database, frontend"
                          value={resource.name || ""}
                          onChange={e => update({ name: e.target.value })}
                          className={`max-w-xl ${getError(errors, "name") ? "border-danger" : ""}`}
                          required
                          aria-invalid={!!getError(errors, "name")}
                        />
                      </DirtyField>
                      {getError(errors, "name") && (
                        <p className="text-sm text-danger mt-1">{getError(errors, "name")}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label>Depends On</Label>
                      <DirtyField
                        draft={resource}
                        baseline={baselineResource}
                        path="depends_on"
                        onReset={onDiscardField ? () => onDiscardField("depends_on") : undefined}
                      >
                        {_allResources ? (
                          <MultiSelect
                            options={_allResources
                              .filter((r) => r.index !== index && r.name && r.name.trim() !== "")
                              .map((r) => ({ label: r.name, value: r.name }))}
                            onValueChange={updateDependsOn}
                            defaultValue={resource.depends_on || []}
                            placeholder={_allResources.length <= 1 ? "No other resources available" : "Select dependencies"}
                            disabled={_allResources.length <= 1}
                            className="w-full"
                          />
                        ) : (
                          <div className="text-sm text-muted-foreground">No dependency information available</div>
                        )}
                      </DirtyField>
                      {errors["depends_on"] && (
                        <p className="text-sm text-danger">{errors["depends_on"]}</p>
                      )}
                      <p className="text-xs text-muted-foreground">Select resources this service depends on. They will be started first.</p>
                    </div>

                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`resource-source-${index}`} className="text-sm font-medium">
                      Build From <span className="text-danger">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top" className="max-w-xs">
                            <p>Select how this resource should be built: from a pre-built Box image or from a Git repository.</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField draft={resource} baseline={baselineResource} path="sourceType" onReset={onDiscardField ? () => onDiscardField("sourceType") : undefined}>
                        <Select
                          value={resource.sourceType || "image"}
                          onValueChange={val => update({ sourceType: val as "image" | "git" })}
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
                    </div>
                    {resource.sourceType === "image" ? (
                      <div className="grid gap-4 max-w-3xl">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`container-image-${index}`} className="text-sm font-medium">
                            Container Image <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">
                                <p>Docker image URL (e.g., nginx:latest)</p>
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <DirtyField
                            draft={resource}
                            baseline={baselineResource}
                            path="image_spec.image"
                            onReset={onDiscardField ? () => onDiscardField("image_spec.image") : undefined}
                          >
                            <Input
                              id={`container-image-${index}`}
                              placeholder="e.g., nginx:latest, redis:7"
                              value={resource.image_spec?.image || ""}
                              onChange={e => updateImageSpec({ image: e.target.value })}
                              className={`max-w-xl ${getError(errors, "image_spec.image") ? "border-danger" : ""}`}
                              required={resource.sourceType === "image"}
                              aria-invalid={!!getError(errors, "image_spec.image")}
                            />
                          </DirtyField>
                          {getError(errors, "image_spec.image") && (
                            <p className="text-sm text-danger mt-1">{getError(errors, "image_spec.image")}</p>
                          )}
                        </div>

                        {/* Docker Registry Secret Section */}
                        <div className="space-y-3">
                          <div className="flex items-center space-x-2">
                            <Switch
                              id={`use-image-secret-${index}`}
                              checked={resource.useImageSecret || false}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  update({ useImageSecret: checked });
                                } else {
                                  update({
                                    useImageSecret: false,
                                    selectedImageSecretId: undefined
                                  });
                                }
                              }}
                              disabled={secrets.isLoading}
                            />
                            <Label htmlFor={`use-image-secret-${index}`} className="text-sm font-medium">
                              Use secret
                            </Label>
                          </div>

                          {resource.useImageSecret && (
                            <div>
                              <Label className="text-sm font-medium mb-2 block">
                                Select secret
                              </Label>
                              <Select
                                value={resource.selectedImageSecretId || ""}
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
                            </div>
                          )}
                        </div>
                      </div>
                    ) : (
                      <div className="grid gap-4 max-w-3xl">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`git-repo-${index}`} className="text-sm font-medium">
                            Git Repository URL <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">URL to the Git repository for this resource</TooltipContent>
                            </Tooltip>
                          </div>
                          <DirtyField draft={resource} baseline={baselineResource} path="build_spec.source_context.git_repo.repo_url" onReset={onDiscardField ? () => onDiscardField("build_spec.source_context.git_repo.repo_url") : undefined}>
                            <Input
                              id={`git-repo-${index}`}
                              value={resource.build_spec?.source_context?.git_repo?.repo_url || ""}
                              onChange={e => updateBuildSpec({ source_context: { git_repo: { repo_url: e.target.value }}})}
                              placeholder="https://github.com/username/repository.git"
                              className={`max-w-xl ${getError(errors, "build_spec.source_context.git_repo.repo_url") ? "border-danger" : ""}`}
                              required={resource.sourceType === "git"}
                              aria-invalid={!!getError(errors, "build_spec.source_context.git_repo.repo_url")}
                            />
                          </DirtyField>
                          {getError(errors, "build_spec.source_context.git_repo.repo_url") && (
                            <p className="text-sm text-danger">{getError(errors, "build_spec.source_context.git_repo.repo_url")}</p>
                          )}
                        </div>

                        {/* Git Credentials Secret Section */}
                        <div className="space-y-3">
                          <div className="flex items-center space-x-2">
                            <Switch
                              id={`use-git-secret-${index}`}
                              checked={resource.useGitSecret || false}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  update({ useGitSecret: checked });
                                } else {
                                  update({
                                    useGitSecret: false,
                                    selectedGitSecretId: undefined
                                  });
                                }
                              }}
                              disabled={secrets.isLoading}
                            />
                            <Label htmlFor={`use-git-secret-${index}`} className="text-sm font-medium">
                              Use secret
                            </Label>
                          </div>

                          {resource.useGitSecret && (
                            <div>
                              <Label className="text-sm font-medium mb-2 block">
                                Select secret
                              </Label>
                              <Select
                                value={resource.selectedGitSecretId || ""}
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
                            </div>
                          )}
                        </div>
                        <div>
                          <div className="flex items-ce nter gap-1 mb-2">
                            <Label htmlFor={`external-image-repo-url-${index}`} className="text-sm font-medium">
                            Image Repository URL <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top" className="max-w-xs">
                                <p>The external container registry URL where images built from this Git repo will be pushed (e.g., ghcr.io/your-org/your-image).</p>
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <DirtyField draft={resource} baseline={baselineResource} path="build_spec.image_repository.external_image_repo_url" onReset={onDiscardField ? () => onDiscardField("build_spec.image_repository.external_image_repo_url") : undefined}>
                            <Input
                              id={`external-image-repo-url-${index}`}
                              value={resource.build_spec?.image_repository?.external_image_repo_url || ""}
                              onChange={e => updateBuildSpec({ image_repository: { external_image_repo_url: e.target.value } })}
                              placeholder="e.g., ghcr.io/your-org/your-image"
                              className={`max-w-xl ${getError(errors, "build_spec.image_repository.external_image_repo_url") ? "border-danger" : ""}`}
                              required={resource.sourceType === "git"}
                              aria-invalid={!!getError(errors, "build_spec.image_repository.external_image_repo_url")}
                            />
                          </DirtyField>
                          {getError(errors, "build_spec.image_repository.external_image_repo_url") && (
                            <p className="text-sm text-danger">{getError(errors, "build_spec.image_repository.external_image_repo_url")}</p>
                          )}
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`git-revision-type-${index}`} className="text-sm font-medium">
                            Git Revision Type <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Select the type of Git revision to use</TooltipContent>
                            </Tooltip>
                          </div>
                          <DirtyField draft={resource} baseline={baselineResource} path="gitRevisionType" onReset={onDiscardField ? () => onDiscardField("gitRevisionType") : undefined}>
                            <Select
                              value={resource.gitRevisionType}
                              onValueChange={val => update({ gitRevisionType: val as "branch" | "commit" | "tag" })}
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
                          {getError(errors, "gitRevisionType") && (
                            <p className="text-sm text-danger mt-1">{getError(errors, "gitRevisionType")}</p>
                          )}
                        </div>
                        {resource.gitRevisionType && (
                          <div>
                            <div className="flex items-center gap-1 mb-2">
                              <Label htmlFor={`git-revision-value-${index}`} className="text-sm font-medium">
                                {resource.gitRevisionType === "branch"
                                  ? "Branch Name"
                                  : resource.gitRevisionType === "commit"
                                    ? "Commit Hash"
                                    : "Tag Name"} <span className="text-danger">*</span>
                              </Label>
                              <Tooltip delayDuration={300}>
                                <TooltipTrigger asChild>
                                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                                </TooltipTrigger>
                                <TooltipContent side="top">
                                  {resource.gitRevisionType === "branch"
                                    ? "The branch to checkout (e.g., main)"
                                    : resource.gitRevisionType === "commit"
                                      ? "The full commit hash to checkout"
                                      : "The tag to checkout (e.g., v1.0.0)"}
                                </TooltipContent>
                              </Tooltip>
                            </div>
                            <DirtyField draft={resource} baseline={baselineResource} path="gitRevisionValue" onReset={onDiscardField ? () => onDiscardField("gitRevisionValue") : undefined}>
                              <Input
                                id={`git-revision-value-${index}`}
                                value={resource.gitRevisionValue || ""}
                                onChange={e => update({ gitRevisionValue: e.target.value })}
                                placeholder={
                                  resource.gitRevisionType === "branch"
                                    ? "e.g., main, develop"
                                    : resource.gitRevisionType === "commit"
                                      ? "e.g., a1b2c3d4e5..."
                                      : "e.g., v1.0.0"
                                }
                                className={`max-w-xl ${getError(errors, "gitRevisionValue") ? "border-danger" : ""}`}
                                required={!!resource.gitRevisionType}
                                aria-invalid={!!getError(errors, "gitRevisionValue")}
                                onBlur={() => {
                                // Mark as touched to trigger error display on submit
                                  if (!resource.gitRevisionValue) {
                                    update({ gitRevisionValue: "" });
                                  }
                                }}
                              />
                            </DirtyField>
                            {getError(errors, "gitRevisionValue") && (
                              <p className="text-sm text-danger mt-1">{getError(errors, "gitRevisionValue")}</p>
                            )}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
                <Separator className="my-4" />
                {/* Volume Mounts Section */}
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Volume Mounts</h3>
                  <div className="grid gap-6 max-w-3xl">
                    {(resource.volume_mounts || []).map((vm, vmIdx) => (
                      <DirtyField
                        key={vmIdx}
                        draft={resource}
                        baseline={baselineResource}
                        path={`volume_mounts.${vmIdx}`}
                        onReset={onDiscardField ? () => onDiscardField(`volume_mounts.${vmIdx}`) : undefined}
                        compact
                      >
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end border p-3 rounded-md bg-muted/10">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`volume-name-${index}-${vmIdx}`} className="text-sm font-medium">
                            Volume <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Select the volume to mount</TooltipContent>
                            </Tooltip>
                          </div>
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
                              {(volumes || []).filter(vol => !!vol.name).length === 0 ? (
                                <div className="p-2 text-sm text-muted-foreground">No volumes available</div>
                              ) : (
                                (volumes || []).filter(vol => !!vol.name).map((vol, vidx) => (
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
                          {getError(errors, `volume_mounts.${vmIdx}.source_volume_name`) && (
                            <p className="text-sm text-danger">{getError(errors, `volume_mounts.${vmIdx}.source_volume_name`)}</p>
                          )}
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`volume-subpath-${index}-${vmIdx}`} className="text-sm font-medium">
                            Sub Path
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Optional path within the volume</TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`volume-subpath-${index}-${vmIdx}`}
                            value={vm.source_sub_path || ""}
                            onChange={(e) => updateVolumeMount(vmIdx, { source_sub_path: e.target.value })}
                            placeholder="e.g., data/config"
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`volume-target-${index}-${vmIdx}`} className="text-sm font-medium">
                            Target Path <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Path in container to mount the volume</TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`volume-target-${index}-${vmIdx}`}
                            value={vm.target_path || ""}
                            onChange={(e) => updateVolumeMount(vmIdx, { target_path: e.target.value })}
                            placeholder="e.g., /mnt/data"
                            className={getError(errors, `volume_mounts.${vmIdx}.target_path`) ? "border-danger" : ""}
                            required
                          />
                          {getError(errors, `volume_mounts.${vmIdx}.target_path`) && (
                            <p className="text-sm text-danger">{getError(errors, `volume_mounts.${vmIdx}.target_path`)}</p>
                          )}
                        </div>
                        <div className="flex justify-end">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => removeVolumeMount(vmIdx)}
                            title="Remove volume mount"
                          >
                            <Trash2 className="h-5 w-5" />
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
                        <PlusCircle className="h-4 w-4 mr-2" />Add Mount
                      </Button>
                      {(volumes || []).length === 0 && (
                        <p className="text-sm text-muted-foreground mt-2">No volumes available. Add volumes in the Volumes section below.</p>
                      )}
                    </div>
                  </div>
                </div>
                <Separator className="my-4" />
                {/* Ports Section */}
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Ports</h3>
                  <div className="grid gap-6 max-w-3xl">
                    {(resource.ports || []).map((port, pidx) => (
                      <DirtyField
                        key={pidx}
                        draft={resource}
                        baseline={baselineResource}
                        path={`ports.${pidx}`}
                        onReset={onDiscardField ? () => onDiscardField(`ports.${pidx}`) : undefined}
                        compact
                      >
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end border p-3 rounded-md bg-muted/10">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`port-number-${index}-${pidx}`} className="text-sm font-medium">
                            Port Number <span className="text-danger">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Container port number</TooltipContent>
                            </Tooltip>
                          </div>
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
                          {getError(errors, `ports.${pidx}.number`) && (
                            <p className="text-sm text-danger">{getError(errors, `ports.${pidx}.number`)}</p>
                          )}
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`port-protocol-${index}-${pidx}`} className="text-sm font-medium">
                            Protocol
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Port communication protocol</TooltipContent>
                            </Tooltip>
                          </div>
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
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`port-expose-${index}-${pidx}`} className="text-sm font-medium">
                            Public Access
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Make this port accessible from outside the cluster</TooltipContent>
                            </Tooltip>
                          </div>
                          <div className="flex items-center space-x-2 h-[40px]"> {/* Match height with inputs */}
                            <Switch
                              id={`port-expose-${index}-${pidx}`}
                              checked={port.exposed_to_public || false}
                              onCheckedChange={(checked) => updatePort(pidx, { exposed_to_public: checked })}
                            />
                            <Label htmlFor={`port-expose-${index}-${pidx}`} className="text-sm font-medium cursor-pointer">
                              {port.exposed_to_public ? "Exposed" : "Internal Only"}
                            </Label>
                          </div>
                        </div>
                        <div className="flex justify-end">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => removePort(pidx)}
                            title="Remove port"
                          >
                            <Trash2 className="h-5 w-5" />
                          </Button>
                        </div>
                      </div>
                      </DirtyField>
                    ))}
                    <div>
                      <Button variant="ghost" size="sm" onClick={addPort}>
                        <PlusCircle className="h-4 w-4 mr-2" />Add Port
                      </Button>
                    </div>
                  </div>
                </div>
              </TabsContent>

              {/* Deployment Tab */}
              <TabsContent value="deployment" className="pt-4 space-y-6">
                {/* Pre-Deploy Section (Init) */}
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Pre-Deployment step</h3>
                  <div className="grid gap-4 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`init-command-${index}`} className="text-sm font-medium">
                      Init Command
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">Pre-deployment init command (comma separated)</TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField draft={resource} baseline={baselineResource} path="init_spec.command" onReset={onDiscardField ? () => onDiscardField("init_spec.command") : undefined}>
                        <Input
                          id={`init-command-${index}`}
                          value={resource.init_spec?.command?.join(",") || ""}
                          onChange={e => update({ init_spec: { ...resource.init_spec, command: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                          placeholder="e.g., sh,/scripts/init.sh"
                        />
                      </DirtyField>
                    </div>
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`init-args-${index}`} className="text-sm font-medium">
                      Init Arguments
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">Pre-deployment arguments (comma separated)</TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField draft={resource} baseline={baselineResource} path="init_spec.args" onReset={onDiscardField ? () => onDiscardField("init_spec.args") : undefined}>
                        <Input
                          id={`init-args-${index}`}
                          value={resource.init_spec?.args?.join(",") || ""}
                          onChange={e => update({ init_spec: { ...resource.init_spec, args: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                          placeholder="e.g., arg1,arg2,arg3"
                        />
                      </DirtyField>
                    </div>
                  </div>
                </div>
                <Separator className="my-4" />
                {/* Post-Deploy Section (Execution) */}
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Main container step</h3>
                  <div className="grid gap-4 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`exec-command-${index}`} className="text-sm font-medium">
                      Command
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">Container runtime command (comma separated)</TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField draft={resource} baseline={baselineResource} path="execution_config.command" onReset={onDiscardField ? () => onDiscardField("execution_config.command") : undefined}>
                        <Input
                          id={`exec-command-${index}`}
                          value={resource.execution_config?.command?.join(",") || ""}
                          onChange={e => update({ execution_config: { ...resource.execution_config, command: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                          placeholder="e.g., node,server.js"
                        />
                      </DirtyField>
                    </div>
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`exec-args-${index}`} className="text-sm font-medium">
                      Arguments
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">Container runtime arguments (comma separated)</TooltipContent>
                        </Tooltip>
                      </div>
                      <DirtyField draft={resource} baseline={baselineResource} path="execution_config.args" onReset={onDiscardField ? () => onDiscardField("execution_config.args") : undefined}>
                        <Input
                          id={`exec-args-${index}`}
                          value={resource.execution_config?.args?.join(",") || ""}
                          onChange={e => update({ execution_config: { ...resource.execution_config, args: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                          placeholder="e.g., --port=3000,--verbose"
                        />
                      </DirtyField>
                    </div>
                  </div>
                </div>
              </TabsContent>

              {/* Environment Variables Tab */}
              <TabsContent value="environment" className="pt-4">
                <div className="flex items-center mb-3">
                  <h3 className="text-lg font-medium">Environment Variables</h3>
                  <div className="ml-auto flex gap-2">
                    <Button variant="ghost" className="text-danger hover:text-danger hover:bg-danger-bg" size="sm" onClick={() => {
                      if (resource.execution_config?.environment_variables?.length) {
                        update({
                          execution_config: {
                            ...resource.execution_config,
                            environment_variables: []
                          }
                        });
                        toast({
                          title: "Environment variables cleared",
                          description: "All environment variables have been removed",
                        });
                      }
                    }} disabled={!(resource.execution_config?.environment_variables?.length)}>
                      <X className="h-4 w-4 mr-1" />
                      <span>Clear All</span>
                    </Button>
                    {/* Paste Variables button */}
                    <Dialog>
                      <DialogTrigger asChild>
                        <Button variant="ghost" size="sm" className="gap-2">
                          <Copy className="h-4 w-4" />
                          <span>Paste Variables</span>
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="w-[95vw] max-w-4xl p-0 overflow-auto">
                        <div className="p-6">
                          <DialogHeader>
                            <DialogTitle className="text-lg font-medium">
                            Paste Environment Variables
                            </DialogTitle>
                          </DialogHeader>
                          <div className="space-y-4">
                            <div className="space-y-2">
                              <Label htmlFor={`env-paste-${index}`} className="text-sm font-medium">
                              Paste in KEY=VALUE format (one per line)
                              </Label>
                              <div className="relative">
                                <Textarea
                                  id={`env-paste-${index}`}
                                  placeholder={
                                    'DATABASE_URL=postgres://user:pass@localhost:5432/db\n' +
                                  'API_KEY=your_api_key\n' +
                                  '# NODE_ENV=development'
                                  }
                                  className="font-mono text-sm min-h-[180px] w-full"
                                />
                              </div>
                              <p className="text-xs text-muted-foreground">
                              Lines starting with # will be ignored as comments
                              </p>
                            </div>
                            <Button
                              onClick={() => {
                                const textarea = document.getElementById(
                                  `env-paste-${index}`
                                ) as HTMLTextAreaElement | null;
                                if (textarea) {
                                  const content = textarea.value.trim();
                                  if (content) {
                                    const envVars = parseEnvContent(content);
                                    addMultipleEnvVars(envVars);
                                  }
                                }
                              }}
                            >
                            Add Variables
                            </Button>
                          </div>
                        </div>
                      </DialogContent>
                    </Dialog>
                    {/* Import from file button */}
                    <Dialog>
                      <DialogTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <Upload className="h-4 w-4 mr-2" /> Import File
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="sm:max-w-md">
                        <DialogHeader>
                          <DialogTitle>Import Environment Variables</DialogTitle>
                        </DialogHeader>
                        <div className="space-y-4 py-4">
                          <div className="flex flex-col gap-2">
                            <Label htmlFor={`env-file-upload-${index}`} className="text-sm font-medium">
                            Upload .env File
                            </Label>
                            <div className="flex items-center justify-center w-full">
                              <label htmlFor={`env-file-upload-${index}`} className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer bg-muted/20 hover:bg-muted/30">
                                <div className="flex flex-col items-center justify-center pt-5 pb-6">
                                  <FileText className="w-8 h-8 mb-2 text-muted-foreground" />
                                  <p className="mb-2 text-sm text-muted-foreground">Click to upload or drag and drop</p>
                                  <p className="text-xs text-muted-foreground">Supports .env files</p>
                                </div>
                                <input
                                  id={`env-file-upload-${index}`}
                                  type="file"
                                  accept=".env,text/plain"
                                  className="hidden"
                                  onChange={handleFileUpload}
                                />
                              </label>
                            </div>
                          </div>
                        </div>
                      </DialogContent>
                    </Dialog>
                  </div>
                </div>
                <div className="border border-muted rounded-md">
                  {/* Header Row */}
                  <div className="grid grid-cols-12 gap-2 p-3 border-b bg-muted/30 text-sm font-medium">
                    <div className="col-span-3">Key</div>
                    <div className="col-span-6">Value</div>
                    <div className="col-span-2 text-center">From</div>
                    <div className="col-span-1"></div>
                  </div>

                  {/* Environment Variables Rows */}
                  {(() => {
                    const envVars = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];
                    // Live duplicate detection (always on, regardless of dirty state)
                    const nameCounts = new Map<string, number>();
                    envVars.forEach((r) => {
                      const k = r.name?.trim();
                      if (!k) return;
                      nameCounts.set(k, (nameCounts.get(k) ?? 0) + 1);
                    });
                    const rowErrorsForIndex = (envIdx: number): EnvRowErrors | undefined => {
                      const r = envVars[envIdx];
                      if (!r) return undefined;
                      const out: EnvRowErrors = {};
                      if (r.name && (nameCounts.get(r.name.trim()) ?? 0) > 1) {
                        out.duplicate = `Duplicate name "${r.name}"`;
                      }
                      const dirty = dirtyEnvRows.has(envIdx);
                      const errPath = (field: string) =>
                        errors[`execution_config.environment_variables.${envIdx}.${field}`];
                      if (r.from === "addon") {
                        if ((dirty || errPath("addonId")) && !r.addonId) out.addonId = "Pick an addon";
                        if ((dirty || errPath("database")) && !r.superuser && !r.database) out.database = "Pick a database";
                        if ((dirty || errPath("credField")) && !r.credField) out.credField = "Pick a field";
                      }
                      if ((dirty || errPath("name")) && !r.name) out.name = "Required";
                      return Object.keys(out).length === 0 ? undefined : out;
                    };

                    if (envVars.length === 0) {
                      return (
                        <div className="p-8 text-center text-muted-foreground">
                          No environment variables defined
                        </div>
                      );
                    }

                    type PlainGroup = { kind: "plain"; items: { env: FormEnvVarData; envIdx: number }[] };
                    type AddonGroup = { kind: "addon"; addonId: string; database: string; items: { env: FormEnvVarData; envIdx: number }[] };
                    type Group = PlainGroup | AddonGroup;
                    const groups: Group[] = [];
                    const addonGroupByKey = new Map<string, AddonGroup>();
                    envVars.forEach((env, envIdx) => {
                      if (env.from === "addon") {
                        const aid = env.addonId || "";
                        const db = env.database || "";
                        const key = `${aid}|${db}`;
                        let g = addonGroupByKey.get(key);
                        if (!g) {
                          g = { kind: "addon", addonId: aid, database: db, items: [] };
                          addonGroupByKey.set(key, g);
                          groups.push(g);
                        }
                        g.items.push({ env, envIdx });
                      } else {
                        const last = groups[groups.length - 1];
                        if (last && last.kind === "plain") {
                          last.items.push({ env, envIdx });
                        } else {
                          groups.push({ kind: "plain", items: [{ env, envIdx }] });
                        }
                      }
                    });

                    const renderRow = ({ env, envIdx }: { env: FormEnvVarData; envIdx: number }) => (
                      <EnvRow
                        key={envIdx}
                        row={env}
                        index={envIdx}
                        resourceIndex={index}
                        secrets={secrets.secrets}
                        secretsLoading={secrets.isLoading}
                        addonNameById={addonNameById}
                        rowErrors={rowErrorsForIndex(envIdx)}
                        status={envRowStatuses[envIdx] ?? "unchanged"}
                        onReset={onDiscardEnvRow ? () => onDiscardEnvRow(envIdx) : undefined}
                        onChangeAddon={(patch) => onChangeAddonForRow(envIdx, patch)}
                        onChangeName={(name) => {
                          replaceEnvVar(envIdx, { ...env, name });
                        }}
                        onChangeValue={(value) => {
                          if (env.from === "stack") {
                            replaceEnvVar(envIdx, { ...env, value });
                          }
                        }}
                        onChangeFrom={(from) => {
                          switchRowFrom(envIdx, from);
                          markEnvRowDirty(envIdx);
                        }}
                        onChangeSecret={(secretId, secretKey) =>
                          replaceEnvVar(envIdx, {
                            from: "secret",
                            name: env.name,
                            secretId,
                            secretKey,
                          })
                        }
                        onBlur={() => markEnvRowDirty(envIdx)}
                        onRemove={() => removeEnvVar(envIdx)}
                      />
                    );

                    return groups.map((g, gIdx) => {
                      if (g.kind === "plain") {
                        return <div key={`p-${gIdx}`}>{g.items.map(renderRow)}</div>;
                      }
                      const aid = g.addonId;
                      const db = g.database;
                      const selectedAddon = addons.find((a) => a.id === aid);
                      const databases = ((selectedAddon?.spec as unknown as { databases?: { name?: string }[] })
                        ?.databases ?? []) as { name?: string }[];
                      const name = aid ? (addonNameById?.get(aid) ?? aid) : null;

                      const updateAllInGroup = (patch: { addonId?: string; database?: string | undefined }) => {
                        const current = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];
                        const next = current.map((e, i) => {
                          if (g.items.some((it) => it.envIdx === i) && e.from === "addon") {
                            return { ...e, ...patch };
                          }
                          return e;
                        });
                        update({
                          execution_config: {
                            ...resource.execution_config,
                            environment_variables: next,
                          },
                        });
                      };

                      // Disallow picking an (addon, db) combo that already has its own group.
                      const usedDbsByAddon = new Map<string, Set<string>>();
                      const usedAddonsByDb = new Map<string, Set<string>>();
                      for (const og of groups) {
                        if (og.kind !== "addon" || og === g) continue;
                        if (og.addonId) {
                          if (!usedDbsByAddon.has(og.addonId)) usedDbsByAddon.set(og.addonId, new Set());
                          if (og.database) usedDbsByAddon.get(og.addonId)!.add(og.database);
                        }
                        if (og.database) {
                          if (!usedAddonsByDb.has(og.database)) usedAddonsByDb.set(og.database, new Set());
                          if (og.addonId) usedAddonsByDb.get(og.database)!.add(og.addonId);
                        }
                      }
                      const dbBlocked = (dbName: string) =>
                        aid !== "" && (usedDbsByAddon.get(aid)?.has(dbName) ?? false);
                      const addonBlocked = (addonId: string) =>
                        db !== "" && (usedAddonsByDb.get(db)?.has(addonId) ?? false);

                      const handleAddonChange = (newAid: string) => {
                        const a = addons.find((x) => x.id === newAid);
                        const dbs = ((a?.spec as unknown as { databases?: { name?: string }[] })?.databases ?? []) as { name?: string }[];
                        const usedDbs = usedDbsByAddon.get(newAid) ?? new Set();
                        const firstFreeDb = dbs.find((d) => d.name && !usedDbs.has(d.name))?.name;
                        const newDb = dbs.length === 1 ? dbs[0].name : firstFreeDb;
                        updateAllInGroup({ addonId: newAid, database: newDb });
                      };
                      const handleDbChange = (newDb: string) => {
                        updateAllInGroup({ database: newDb });
                      };
                      const handleAddBinding = () => {
                        addEnvVar({
                          from: "addon",
                          name: "",
                          addonType: "postgres",
                          addonId: aid,
                          database: db || undefined,
                          superuser: false,
                          credField: undefined,
                        });
                      };

                      return (
                        <div
                          key={`a-${gIdx}-${aid}-${db}`}
                          className="rounded-md border border-dashed border-foreground/25 my-2"
                        >
                          <div className="flex items-center gap-2 px-3 pt-2 pb-1">
                            <Select value={aid || undefined} onValueChange={handleAddonChange}>
                              <SelectTrigger className="h-7 w-[200px] text-[12.5px] font-semibold gap-2">
                                <span className="flex items-center gap-2 min-w-0">
                                  {aid && <AddonTypeIcon type="postgres" size={14} />}
                                  <SelectValue placeholder="Pick addon">
                                    {aid ? name : undefined}
                                  </SelectValue>
                                </span>
                              </SelectTrigger>
                              <SelectContent>
                                {addons.length === 0 ? (
                                  <div className="px-3 py-2 text-xs text-muted-foreground">
                                    No addons linked. Add one from the bottom panel.
                                  </div>
                                ) : (
                                  addons.map((a) => (
                                    <SelectItem
                                      key={a.id}
                                      value={a.id!}
                                      disabled={a.id !== aid && addonBlocked(a.id!)}
                                    >
                                      <span className="flex items-center gap-2">
                                        <AddonTypeIcon type="postgres" size={14} />
                                        <span>{a.name}</span>
                                        {a.id !== aid && addonBlocked(a.id!) && (
                                          <span className="ml-1 text-[10px] text-muted-foreground">
                                            in use
                                          </span>
                                        )}
                                      </span>
                                    </SelectItem>
                                  ))
                                )}
                              </SelectContent>
                            </Select>
                            {aid && (
                              <>
                                <span className="text-muted-foreground/60">·</span>
                                {databases.length > 1 ? (
                                  <Select value={db || undefined} onValueChange={handleDbChange}>
                                    <SelectTrigger className="h-7 w-[160px] text-[12.5px]">
                                      <SelectValue placeholder="Pick database" />
                                    </SelectTrigger>
                                    <SelectContent>
                                      {databases.map((d) =>
                                        d.name ? (
                                          <SelectItem
                                            key={d.name}
                                            value={d.name}
                                            disabled={d.name !== db && dbBlocked(d.name)}
                                          >
                                            {d.name}
                                            {d.name !== db && dbBlocked(d.name) && (
                                              <span className="ml-2 text-[10px] text-muted-foreground">
                                                in use
                                              </span>
                                            )}
                                          </SelectItem>
                                        ) : null,
                                      )}
                                    </SelectContent>
                                  </Select>
                                ) : (
                                  <span className="text-[12px] text-muted-foreground">
                                    db: {db || databases[0]?.name || "—"}
                                  </span>
                                )}
                              </>
                            )}
                          </div>
                          {g.items.map(renderRow)}
                          <div className="px-3 py-1.5 flex justify-end">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={handleAddBinding}
                              className="h-7 text-[12.5px]"
                            >
                              <PlusCircle className="h-3 w-3 mr-1" /> Add binding
                            </Button>
                          </div>
                        </div>
                      );
                    });
                  })()}
                </div>
                {/* Add Variable button */}
                <div className="flex justify-end mt-2">
                  <Button variant="ghost" size="sm" onClick={() => addEnvVar()}>
                    <PlusCircle className="h-4 w-4 mr-2" /> Add Variable
                  </Button>
                </div>
              </TabsContent>
            </Tabs>

            <div className="flex justify-center items-center mt-8">
              <span className="flex items-center justify-center w-full py-3 rounded-md bg-muted/70">
                <Button
                  type="button"
                  variant="ghost"
                  className="text-danger hover:text-danger hover:bg-danger-bg focus-visible:bg-danger-bg"
                  onClick={() => onRemove(index)}
                >
                  <Trash2 className="h-4 w-4 mr-1" />
                Remove Resource
                </Button>
              </span>
            </div>
          </div>
        </AccordionContent>
      </AccordionItem>
    </TooltipProvider>
  );
}

/**
 * The form is heavy (~1500 lines of JSX) and Radix Accordion keeps closed
 * items mounted, so a keystroke in one resource would otherwise re-render
 * every other resource. React.memo with the default shallow compare keeps
 * unchanged items idle as long as the parent passes stable refs (handled by
 * useCallback in detail/index.tsx and useRef'd EMPTY_ERRORS in
 * resource-form-list.tsx).
 */
const StackResourceItem = React.memo(StackResourceItemImpl);
export default StackResourceItem;

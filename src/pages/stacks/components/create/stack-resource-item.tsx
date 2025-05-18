import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
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
import { Plus, X, GitBranch, ImageIcon, Trash2, Database, Upload, FileText, Copy } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { MultiSelect } from "@/components/multi-select";

import type { StackResourceData, VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";

interface StackResourceItemProps {
  resource: Partial<StackResourceData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  isOnlyResource: boolean;
  onChange: (index: number, updatedResource: Partial<StackResourceData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  volumes?: Partial<VolumeFormData>[];
  allResources: { name: string; index: number }[];
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

export default function StackResourceItem({
  resource,
  index,
  itemRef,
  isOnlyResource,
  onChange,
  onRemove,
  errors,
  volumes = [],
  allResources
}: StackResourceItemProps) {
  // Helper for updating resource fields
  const update = (patch: Partial<StackResourceData>) => {
    onChange(index, { ...resource, ...patch });
  };

  // Helper for updating nested build_spec
  const updateBuildSpec = (patch: Partial<NonNullable<StackResourceData["build_spec"]>>) => {
    const currentBuildSpec = resource.build_spec || {
      source_context: { git_repo: { repo_url: '' } },
      context_path_within_source: './',
      dockerfile_path: 'Dockerfile',
      image_repository_url: { url: '', cluster_registry_id: '' },
      insecure_registry: false
    };

    update({
      build_spec: { ...currentBuildSpec, ...patch },
      image_spec: undefined,
    });
  };
  // Helper for updating nested image_spec
  const updateImageSpec = (patch: Partial<NonNullable<StackResourceData["image_spec"]>>) => {
    const currentImageSpec = resource.image_spec || { image: '' };

    update({
      image_spec: { ...currentImageSpec, ...patch },
      build_spec: undefined,
    });
  };

  // Helper for updating ports
  const updatePort = (pidx: number, patch: Partial<NonNullable<StackResourceData["ports"]>[number]>) => {
    update({
      ports: (resource.ports || []).map((pt, i) => (i === pidx ? { ...pt, ...patch } : pt)),
    });
  };

  // Helper for adding/removing ports
  const addPort = () => {
    update({
      ports: [
        ...(resource.ports || []),
        { number: 80, protocol: "http", exposed_to_public: false, subdomain_prefix: "" },
      ],
    });
  };
  const removePort = (pidx: number) => {
    update({
      ports: (resource.ports || []).filter((_, i) => i !== pidx),
    });
  };

  // Helper for updating environment variables
  const addEnvVar = () => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [
          ...(resource.execution_config?.environment_variables || []),
          { name: "", value: "" },
        ],
      },
    });
  };
  const updateEnvVar = (envIdx: number, name: string, value: string) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: (resource.execution_config?.environment_variables || []).map((env, i) =>
          i === envIdx ? { name, value } : env
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

    // Filter out duplicates and add new vars
    const newVars = filteredVars.filter(env => !existingVarNames.has(env.name));

    if (newVars.length === 0) {
      toast({
        title: "No new variables added",
        description: "All variables already exist or are invalid",
        variant: "destructive"
      });
      return;
    }

    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [
          ...currentVars,
          ...newVars
        ],
      },
    });

    toast({
      title: "Environment variables added",
      description: `${newVars.length} new variable${newVars.length === 1 ? '' : 's'} added successfully`,
    });
  };

  // Parse .env file content or pasted text
  const parseEnvContent = (content: string): Array<{name: string, value: string}> => {
    if (!content.trim()) return [];

    const lines = content.split('\n');
    const envVars: Array<{name: string, value: string}> = [];

    for (const line of lines) {
      // Skip comments and empty lines
      const trimmedLine = line.trim();
      if (!trimmedLine || trimmedLine.startsWith('#')) continue;

      // Parse KEY=VALUE format
      const match = trimmedLine.match(/^([^=]+)=(.*)$/);
      if (match) {
        const [, name, value] = match;
        envVars.push({
          name: name.trim(),
          value: value.trim().replace(/^['"](.*)['"]/g, '$1') // Remove surrounding quotes if present
        });
      }
    }

    return envVars;
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
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

  // Prepare options for depends_on (exclude self, only named resources)
  const dependsOnOptions = allResources
    .filter((r) => r.index !== index && r.name && r.name.trim() !== "")
    .map((r) => ({ label: r.name, value: r.name }));

  return (
    <AccordionItem value={String(index)} className="border-0">
      <AccordionTrigger
        ref={itemRef}
        className="px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-2 text-left">
          {resource.sourceType === "git" ? (
            <GitBranch className="h-5 w-5 text-muted-foreground shrink-0" />
          ) : (
            <ImageIcon className="h-5 w-5 text-muted-foreground shrink-0" />
          )}
          <div>
            <span className="font-medium">{resource.name || `Resource ${index + 1}`}</span>
            {resource.name === "" && (
              <span className="ml-2 text-sm text-muted-foreground">(unnamed)</span>
            )}
            {errors._form && (
              <div className="text-sm text-destructive mt-1">{errors._form}</div>
            )}
          </div>
        </div>
      </AccordionTrigger>

      <AccordionContent className="pb-4 pt-2">
        <div className="px-4 space-y-4">
          <Tabs defaultValue="configuration" className="w-full">
            <div className="mt-1 mb-3">
              <TabsList className="grid grid-cols-3 w-full">
                <TabsTrigger value="configuration">Configuration</TabsTrigger>
                <TabsTrigger value="deployment">Deployment</TabsTrigger>
                <TabsTrigger value="environment">Environment Variables</TabsTrigger>
              </TabsList>
            </div>

            {/* General Section (always at top) */}
            <TabsContent value="configuration" className="pt-4 space-y-6">
              <div>
                <h3 className="text-lg font-medium mb-3">General</h3>
                <div className="grid gap-4 max-w-3xl">
                  <div>
                    <div className="flex items-center gap-1 mb-2">
                      <Label htmlFor={`name-${index}`} className="text-sm font-medium">
                      Resource Name <span className="text-red-500">*</span>
                      </Label>
                      <Tooltip delayDuration={300}>
                        <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                        <TooltipContent side="top" className="max-w-xs">
                          <p>A unique name for this resource in your stack</p>
                        </TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      id={`name-${index}`}
                      value={resource.name || ""}
                      onChange={e => update({ name: e.target.value })}
                      required
                      className={`max-w-md ${errors.name ? "border-destructive" : ""}`}
                      placeholder="e.g., web-server, database, cache"
                      aria-invalid={!!errors.name}
                      onBlur={() => {
                        if (!resource.name) {
                          update({ name: "" });
                        }
                      }}
                    />
                    {errors.name && <p className="text-sm text-destructive mt-1">{errors.name}</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>Depends On</Label>
                    <MultiSelect
                      options={dependsOnOptions}
                      onValueChange={updateDependsOn}
                      defaultValue={resource.depends_on || []}
                      placeholder={dependsOnOptions.length === 0 ? "No other resources available" : "Select dependencies"}
                      disabled={dependsOnOptions.length === 0}
                      className="w-full"
                    />
                    {errors["depends_on"] && (
                      <p className="text-sm text-destructive">{errors["depends_on"]}</p>
                    )}
                    <p className="text-xs text-muted-foreground">Select resources this service depends on. They will be started first.</p>
                  </div>
                </div>
              </div>
              <Separator className="my-4" />
              {/* Source Configuration Section */}
              <div>
                <h3 className="text-lg font-medium mb-3">Source Configuration</h3>
                <div className="mb-4">
                  <div className="flex items-center gap-1 mb-2">
                    <Label htmlFor={`source-type-${index}`} className="text-sm font-medium">
                    Build From <span className="text-red-500">*</span>
                    </Label>
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                      <TooltipContent side="top" className="max-w-xs">
                        <p>Select how this resource should be built: from a pre-built container image or from a Git repository.</p>
                      </TooltipContent>
                    </Tooltip>
                  </div>
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
                          <ImageIcon size={16} />
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
                </div>
                {resource.sourceType === "image" ? (
                  <div className="grid gap-4 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`container-image-${index}`} className="text-sm font-medium">
                        Container Image URL <span className="text-red-500">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent side="top" className="max-w-xs">
                            <p>The full URL of the container image (e.g., nginx:latest, gcr.io/project/image:tag).</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                      <Input
                        id={`container-image-${index}`}
                        value={resource.image_spec?.image || ""}
                        onChange={e => updateImageSpec({ image: e.target.value })}
                        placeholder="e.g., nginx:latest, your-registry/your-image:v1.0"
                        className={`max-w-xl ${getError(errors, "image_spec.image") ? "border-destructive" : ""}`}
                        required={resource.sourceType === "image"}
                        aria-invalid={!!getError(errors, "image_spec.image")}
                        onBlur={() => {
                          // Mark as touched to trigger error display on submit
                          if (resource.sourceType === "image") {
                            updateImageSpec({ image: resource.image_spec?.image || "" });
                          }
                        }}
                      />
                      {getError(errors, "image_spec.image") && (
                        <p className="text-sm text-destructive">{getError(errors, "image_spec.image")}</p>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="grid gap-6 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`git-repo-${index}`} className="text-sm font-medium">
                          Git Repository URL <span className="text-red-500">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent side="top" className="max-w-xs">
                            <p>The HTTPS or SSH URL of the Git repository.</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                      <Input
                        id={`git-repo-${index}`}
                        value={resource.build_spec?.source_context?.git_repo?.repo_url || ""}
                        onChange={e => updateBuildSpec({ source_context: { git_repo: { repo_url: e.target.value }}})}
                        placeholder="https://github.com/username/repository.git"
                        className={`max-w-xl ${getError(errors, "build_spec.source_context.git_repo.repo_url") ? "border-destructive" : ""}`}
                        required={resource.sourceType === "git"}
                        aria-invalid={!!getError(errors, "build_spec.source_context.git_repo.repo_url")}
                      />
                      {getError(errors, "build_spec.source_context.git_repo.repo_url") && (
                        <p className="text-sm text-destructive">{getError(errors, "build_spec.source_context.git_repo.repo_url")}</p>
                      )}
                    </div>
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`git-revision-type-${index}`} className="text-sm font-medium">
                          Git Revision Type <span className="text-red-500">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent side="top" className="max-w-xs">
                            <p>Choose whether to use a commit SHA, branch name, or tag for this build.</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                      <Select
                        value={resource.gitRevisionType || ""}
                        onValueChange={val => update({ gitRevisionType: val as typeof resource.gitRevisionType })}
                      >
                        <SelectTrigger
                          id={`git-revision-type-${index}`}
                          className={`w-[200px] ${getError(errors, "gitRevisionType") ? "border-destructive" : ""}`}
                        >
                          <SelectValue placeholder="Select revision type" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="commit">Commit</SelectItem>
                          <SelectItem value="branch">Branch</SelectItem>
                          <SelectItem value="tag">Tag</SelectItem>
                        </SelectContent>
                      </Select>
                      {getError(errors, "gitRevisionType") && (
                        <p className="text-sm text-destructive mt-1">{getError(errors, "gitRevisionType")}</p>
                      )}
                    </div>
                    {resource.gitRevisionType && (
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`git-revision-value-${index}`} className="text-sm font-medium">
                            {resource.gitRevisionType === "commit" ? "Commit SHA" : resource.gitRevisionType === "branch" ? "Branch Name" : "Tag Name"} <span className="text-red-500">*</span>
                          </Label>
                        </div>
                        <Input
                          id={`git-revision-value-${index}`}
                          value={resource.gitRevisionValue || ""}
                          onChange={e => update({ gitRevisionValue: e.target.value })}
                          placeholder={resource.gitRevisionType === "commit" ? "e.g., 1a2b3c4d" : resource.gitRevisionType === "branch" ? "e.g., main" : "e.g., v1.0.0"}
                          className={`max-w-xl ${getError(errors, "gitRevisionValue") ? "border-destructive" : ""}`}
                          required={!!resource.gitRevisionType}
                          aria-invalid={!!getError(errors, "gitRevisionValue")}
                          onBlur={() => {
                            // Mark as touched to trigger error display on submit
                            if (!resource.gitRevisionValue) {
                              update({ gitRevisionValue: "" });
                            }
                          }}
                        />
                        {getError(errors, "gitRevisionValue") && (
                          <p className="text-sm text-destructive mt-1">{getError(errors, "gitRevisionValue")}</p>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
              <Separator className="my-4" />
              {/* Volume Mounts Section */}
              <div>
                <h3 className="text-lg font-medium mb-3">Volume Mounts</h3>
                <div className="grid gap-6 max-w-3xl">
                  {(resource.volume_mounts || []).map((vm, vmIdx) => (
                    <div key={vmIdx} className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end border p-3 rounded-md bg-muted/10">
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`volume-name-${index}-${vmIdx}`} className="text-sm font-medium">
                            Volume <span className="text-red-500">*</span>
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">Select the volume to mount</TooltipContent>
                          </Tooltip>
                        </div>
                        <Select
                          value={vm.source_volume_name || ""}
                          onValueChange={(value) => updateVolumeMount(vmIdx, { source_volume_name: value })}
                        >
                          <SelectTrigger
                            id={`volume-name-${index}-${vmIdx}`}
                            className={getError(errors, `volume_mounts.${vmIdx}.source_volume_name`) ? "border-destructive" : ""}
                          >
                            <SelectValue placeholder="Select volume" />
                          </SelectTrigger>
                          <SelectContent>
                            {(volumes || []).length === 0 ? (
                              <div className="p-2 text-sm text-muted-foreground">No volumes available</div>
                            ) : (
                              (volumes || []).map((vol, vidx) => (
                                <SelectItem key={vidx} value={vol.name || ""}>
                                  <div className="flex items-center gap-2">
                                    <Database className="h-4 w-4" />
                                    <span>{vol.name || `Volume ${vidx + 1}`}</span>
                                    {vol.spec?.size && <span className="ml-1 text-xs text-muted-foreground">({vol.spec.size})</span>}
                                  </div>
                                </SelectItem>
                              ))
                            )}
                          </SelectContent>
                        </Select>
                        {getError(errors, `volume_mounts.${vmIdx}.source_volume_name`) && (
                          <p className="text-sm text-destructive">{getError(errors, `volume_mounts.${vmIdx}.source_volume_name`)}</p>
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`volume-subpath-${index}-${vmIdx}`} className="text-sm font-medium">
                            Sub Path
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
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
                            Target Path <span className="text-red-500">*</span>
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">Path in container to mount the volume</TooltipContent>
                          </Tooltip>
                        </div>
                        <Input
                          id={`volume-target-${index}-${vmIdx}`}
                          value={vm.target_path || ""}
                          onChange={(e) => updateVolumeMount(vmIdx, { target_path: e.target.value })}
                          placeholder="e.g., /mnt/data"
                          className={getError(errors, `volume_mounts.${vmIdx}.target_path`) ? "border-destructive" : ""}
                          required
                        />
                        {getError(errors, `volume_mounts.${vmIdx}.target_path`) && (
                          <p className="text-sm text-destructive">{getError(errors, `volume_mounts.${vmIdx}.target_path`)}</p>
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
                  ))}
                  <div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={addVolumeMount}
                      disabled={(volumes || []).length === 0}
                    >
                      <Plus className="h-4 w-4 mr-2" />Add Volume Mount
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
                <h3 className="text-lg font-medium mb-3">Ports Configuration</h3>
                <div className="grid gap-6 max-w-3xl">
                  {(resource.ports || []).map((port, pidx) => (
                    <div key={pidx} className="grid grid-cols-2 md:grid-cols-5 gap-4 items-end border p-3 rounded-md bg-muted/10">
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`port-number-${index}-${pidx}`} className="text-sm font-medium">
                          Port Number <span className="text-red-500">*</span>
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">Port number</TooltipContent>
                          </Tooltip>
                        </div>
                        <Input
                          id={`port-number-${index}-${pidx}`}
                          type="number"
                          value={port.number}
                          onChange={e => updatePort(pidx, { number: Number(e.target.value) })}
                          required
                        />
                      </div>
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`port-protocol-${index}-${pidx}`} className="text-sm font-medium">
                          Protocol
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">tcp or http</TooltipContent>
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
                          <Label htmlFor={`port-public-${index}-${pidx}`} className="text-sm font-medium">
                          Expose Publicly
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">Expose to public internet</TooltipContent>
                          </Tooltip>
                        </div>
                        <div className="pt-2 pl-2">
                          <Switch
                            id={`port-public-${index}-${pidx}`}
                            checked={!!port.exposed_to_public}
                            onCheckedChange={v => updatePort(pidx, { exposed_to_public: v })}
                          />
                        </div>
                      </div>
                      <div>
                        <div className="flex items-center gap-1 mb-2">
                          <Label htmlFor={`port-subdomain-${index}-${pidx}`} className="text-sm font-medium">
                          Subdomain Prefix
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top">Prefix for public URL</TooltipContent>
                          </Tooltip>
                        </div>
                        <Input
                          id={`port-subdomain-${index}-${pidx}`}
                          value={port.subdomain_prefix || ""}
                          onChange={e => updatePort(pidx, { subdomain_prefix: e.target.value })}
                        />
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
                  ))}
                  <div>
                    <Button variant="outline" size="sm" onClick={addPort}>
                      <Plus className="h-4 w-4 mr-2" />Add Port
                    </Button>
                  </div>
                </div>
              </div>
            </TabsContent>

            {/* Deployment Tab */}
            <TabsContent value="deployment" className="pt-4 space-y-6">
              {/* Pre-Deploy Section (Init) */}
              <div>
                <h3 className="text-lg font-medium mb-3">Pre-Deployment Configuration</h3>
                <p className="text-sm text-muted-foreground mb-3">
                Commands to run before the main container starts
                </p>
                <div className="grid gap-4 max-w-3xl">
                  <div>
                    <div className="flex items-center gap-1 mb-2">
                      <Label htmlFor={`init-command-${index}`} className="text-sm font-medium">
                      Init Command
                      </Label>
                      <Tooltip delayDuration={300}>
                        <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                        <TooltipContent side="top">Pre-deployment command (comma separated)</TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      id={`init-command-${index}`}
                      value={resource.init_spec?.command?.join(",") || ""}
                      onChange={e => update({ init_spec: { ...resource.init_spec, command: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                      placeholder="e.g., sh,/scripts/init.sh"
                    />
                  </div>
                  <div>
                    <div className="flex items-center gap-1 mb-2">
                      <Label htmlFor={`init-args-${index}`} className="text-sm font-medium">
                      Init Arguments
                      </Label>
                      <Tooltip delayDuration={300}>
                        <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                        <TooltipContent side="top">Pre-deployment arguments (comma separated)</TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      id={`init-args-${index}`}
                      value={resource.init_spec?.args?.join(",") || ""}
                      onChange={e => update({ init_spec: { ...resource.init_spec, args: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                      placeholder="e.g., arg1,arg2,arg3"
                    />
                  </div>
                </div>
              </div>
              <Separator className="my-4" />
              {/* Post-Deploy Section (Execution) */}
              <div>
                <h3 className="text-lg font-medium mb-3">Main Container Configuration</h3>
                <p className="text-sm text-muted-foreground mb-3">
                Main container runtime settings
                </p>
                <div className="grid gap-4 max-w-3xl">
                  <div>
                    <div className="flex items-center gap-1 mb-2">
                      <Label htmlFor={`exec-command-${index}`} className="text-sm font-medium">
                      Command
                      </Label>
                      <Tooltip delayDuration={300}>
                        <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                        <TooltipContent side="top">Container runtime command (comma separated)</TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      id={`exec-command-${index}`}
                      value={resource.execution_config?.command?.join(",") || ""}
                      onChange={e => update({ execution_config: { ...resource.execution_config, command: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                      placeholder="e.g., npm,start"
                    />
                  </div>
                  <div>
                    <div className="flex items-center gap-1 mb-2">
                      <Label htmlFor={`exec-args-${index}`} className="text-sm font-medium">
                      Arguments
                      </Label>
                      <Tooltip delayDuration={300}>
                        <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                        <TooltipContent side="top">Container runtime arguments (comma separated)</TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      id={`exec-args-${index}`}
                      value={resource.execution_config?.args?.join(",") || ""}
                      onChange={e => update({ execution_config: { ...resource.execution_config, args: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                      placeholder="e.g., --port=3000,--verbose"
                    />
                  </div>
                </div>
              </div>
            </TabsContent>

            {/* Environment Variables Tab */}
            <TabsContent value="environment" className="pt-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-lg font-medium">Environment Variables</h3>
                <div className="flex gap-2">
                  {/* Clear all variables button */}
                  {/* Clear all variables button */}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      if (resource.execution_config?.environment_variables?.length > 0) {
                        update({
                          execution_config: {
                            ...resource.execution_config,
                            environment_variables: []
                          }
                        });
                        toast({
                          title: "Cleared all environment variables",
                          description: "All environment variables have been removed",
                        });
                      }
                    }}
                    disabled={!resource.execution_config?.environment_variables?.length}
                    className="text-destructive hover:text-destructive/90 hover:bg-destructive/10"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Clear All
                  </Button>

                  {/* Add single variable button */}
                  <Button variant="ghost" size="sm" onClick={addEnvVar}>
                    <Plus className="h-4 w-4 mr-2" /> Add Variable
                  </Button>

                  {/* Paste variables dialog */}
                  <Dialog>
                    <DialogTrigger asChild>
                      <Button variant="ghost" size="sm" className="gap-2">
                        <Copy className="h-4 w-4" />
                        <span>Paste Variables</span>
                      </Button>
                    </DialogTrigger>
                    <DialogContent className="w-[95vw] max-w-4xl p-0 overflow-auto">
                      <div className="p-6">
                        {/* Dialog Header */}
                        <DialogHeader>
                          <DialogTitle className="text-lg font-medium">
                            Paste Environment Variables
                          </DialogTitle>
                        </DialogHeader>

                        {/* Dialog Content */}
                        <div className="space-y-4">
                          <div className="space-y-2">
                            <Label
                              htmlFor={`env-paste-${index}`}
                              className="text-sm font-medium"
                            >
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
                              ) as HTMLTextAreaElement;
                              const content = textarea.value;
                              const envVars = parseEnvContent(content);

                              if (envVars.length > 0) {
                                addMultipleEnvVars(envVars);
                                textarea.value = '';
                                // Close the dialog after adding
                                const closeButton = document.querySelector(
                                  'button[data-state="open"] + div [data-radix-collection-item]'
                                ) as HTMLButtonElement;
                                closeButton?.click();
                              } else {
                                toast({
                                  title: 'No valid variables found',
                                  description: 'Please enter variables in KEY=VALUE format',
                                  variant: 'destructive',
                                });
                              }
                            }}
                            className="w-full bg-black hover:bg-gray-800"
                          >
                            Add Variables
                          </Button>
                        </div>
                      </div>
                    </DialogContent>
                  </Dialog>

                  {/* Import from file dialog */}
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

              <div className="overflow-x-auto">
                <table className="min-w-full border border-muted rounded-md">
                  <thead className="bg-muted/30">
                    <tr>
                      <th className="text-left px-6 py-3 text-sm">Key</th>
                      <th className="text-left px-6 py-3 text-sm">Value</th>
                      <th className="w-12"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {(resource.execution_config?.environment_variables || []).length ? (
                      (resource.execution_config?.environment_variables || []).map((env, envIdx) => (
                        <tr key={envIdx} className="border-t border-muted">
                          <td className="px-6 py-2 align-middle">
                            <Input
                              id={`env-name-${index}-${envIdx}`}
                              value={env.name}
                              onChange={e => updateEnvVar(envIdx, e.target.value, env.value)}
                              placeholder="VARIABLE_NAME"
                              className="bg-muted/30 font-mono text-xs"
                            />
                          </td>
                          <td className="px-6 py-2 align-middle">
                            <Input
                              id={`env-value-${index}-${envIdx}`}
                              value={env.value}
                              onChange={e => updateEnvVar(envIdx, env.name, e.target.value)}
                              placeholder="value"
                              className="bg-muted/30 font-mono text-xs"
                            />
                          </td>
                          <td className="px-2 py-2 align-middle text-center">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => removeEnvVar(envIdx)}
                              title="Remove variable"
                            >
                              <X className="h-5 w-5" />
                            </Button>
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan={3} className="text-center text-muted-foreground py-8">
                        No environment variables defined
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </TabsContent>
          </Tabs>

          {/* Delete button at bottom */}
          {!isOnlyResource && (
            <div className="pt-4 border-t">
              <Button
                type="button"
                variant="ghost"
                className="text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={() => onRemove(index)}
              >
                <Trash2 className="h-4 w-4 mr-1" />
                Remove Resource
              </Button>
            </div>
          )}
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

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
import type { z } from "zod";

import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../hooks/use-secrets";

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
  onChange,
  onRemove,
  errors,
  volumes = [],
  allResources: _allResources,
  secrets,
}: StackResourceItemProps) {
  // Helper for updating resource fields
  const update = (patch: Partial<FormStackResourceData>) => {
    onChange(index, { ...resource, ...patch });
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

  // Helper for adding an environment variable
  const addEnvVar = () => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: [
          ...(resource.execution_config?.environment_variables || []),
          { name: "", value: "", useSecret: false, selectedSecretId: undefined, selectedSecretKey: undefined },
        ],
      },
    });
  };

  // Helper for updating an environment variable
  const updateEnvVar = (envIdx: number, updates: Partial<{ name: string; value: string; useSecret: boolean; selectedSecretId: string; selectedSecretKey: string }>) => {
    update({
      execution_config: {
        ...resource.execution_config,
        environment_variables: (resource.execution_config?.environment_variables || []).map((env, i) =>
          i === envIdx ? { ...env, ...updates } : env
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

    // Filter out duplicates and add new vars with default secret fields
    const newVars = filteredVars
      .filter(env => !existingVarNames.has(env.name))
      .map(env => ({
        ...env,
        useSecret: false,
        selectedSecretId: undefined,
        selectedSecretKey: undefined,
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

  // Status color logic
  const statusObj = (resource.status ?? {}) as z.infer<typeof ApiStackResourceStatusSchema>;
  const status = statusObj.state?.toLowerCase() || 'pending';
  let statusColor = 'bg-yellow-500';
  if (status === 'ready' || status === 'running') {
    statusColor = 'bg-green-500';
  } else if (status === 'failed') {
    statusColor = 'bg-red-500';
  }

  return (
    <TooltipProvider>
      <AccordionItem value={String(index)} className="border-0">
        <AccordionTrigger
          ref={itemRef}
          className="px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
        >
          <div className="flex items-center gap-2 text-left flex-grow">
            <div className="flex flex-col flex-grow min-w-0">
              <span className="font-medium flex items-center gap-2">
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <span className={`h-2 w-2 rounded-full ${statusColor}`}></span>
                  </TooltipTrigger>
                  <TooltipContent side="top">
                    <p className="capitalize">{statusObj.state || 'Pending'}</p>
                  </TooltipContent>
                </Tooltip>
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
                <span className="text-xs text-destructive mt-0.5 pl-6">{errors._form}</span>
              )}
            </div>
          </div>
        </AccordionTrigger>
        <AccordionContent className="pb-4 pt-2">
          <div className="px-4 space-y-4">
            <Tabs defaultValue="general" className="w-full">
              <div className="mt-1 mb-3">
                <TabsList className="grid grid-cols-3 w-full">
                  <TabsTrigger value="general">General</TabsTrigger>
                  <TabsTrigger value="deployment">Deployment</TabsTrigger>
                  <TabsTrigger value="environment">Environment Variables</TabsTrigger>
                </TabsList>
              </div>

              {/* General Section (always at top) */}
              <TabsContent value="general" className="pt-4 space-y-6">
                <div>
                  <h3 className="text-lg font-medium mb-3">General</h3>
                  <div className="grid gap-4 max-w-3xl">
                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`resource-name-${index}`} className="text-sm font-medium">
                          Resource Name <span className="text-red-500">*</span>
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
                          <TooltipContent side="top">A unique name for this resource</TooltipContent>
                        </Tooltip>
                      </div>
                      <Input
                        id={`resource-name-${index}`}
                        placeholder="e.g., api, database, frontend"
                        value={resource.name || ""}
                        onChange={e => update({ name: e.target.value })}
                        className={`max-w-xl ${getError(errors, "name") ? "border-destructive" : ""}`}
                        required
                        aria-invalid={!!getError(errors, "name")}
                      />
                      {getError(errors, "name") && (
                        <p className="text-sm text-destructive mt-1">{getError(errors, "name")}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label>Depends On</Label>
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
                      {errors["depends_on"] && (
                        <p className="text-sm text-destructive">{errors["depends_on"]}</p>
                      )}
                      <p className="text-xs text-muted-foreground">Select resources this service depends on. They will be started first.</p>
                    </div>

                    <div>
                      <div className="flex items-center gap-1 mb-2">
                        <Label htmlFor={`resource-source-${index}`} className="text-sm font-medium">
                      Build From <span className="text-red-500">*</span>
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
                    </div>
                    {resource.sourceType === "image" ? (
                      <div className="grid gap-4 max-w-3xl">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`container-image-${index}`} className="text-sm font-medium">
                            Container Image <span className="text-red-500">*</span>
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
                          <Input
                            id={`container-image-${index}`}
                            placeholder="e.g., nginx:latest, redis:7"
                            value={resource.image_spec?.image || ""}
                            onChange={e => updateImageSpec({ image: e.target.value })}
                            className={`max-w-xl ${getError(errors, "image_spec.image") ? "border-destructive" : ""}`}
                            required={resource.sourceType === "image"}
                            aria-invalid={!!getError(errors, "image_spec.image")}
                          />
                          {getError(errors, "image_spec.image") && (
                            <p className="text-sm text-destructive mt-1">{getError(errors, "image_spec.image")}</p>
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
                            Git Repository URL <span className="text-red-500">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">URL to the Git repository for this resource</TooltipContent>
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
                            Image Repository URL <span className="text-red-500">*</span>
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
                          <Input
                            id={`external-image-repo-url-${index}`}
                            value={resource.build_spec?.image_repository?.external_image_repo_url || ""}
                            onChange={e => updateBuildSpec({ image_repository: { external_image_repo_url: e.target.value } })}
                            placeholder="e.g., ghcr.io/your-org/your-image"
                            className={`max-w-xl ${getError(errors, "build_spec.image_repository.external_image_repo_url") ? "border-destructive" : ""}`}
                            required={resource.sourceType === "git"}
                            aria-invalid={!!getError(errors, "build_spec.image_repository.external_image_repo_url")}
                          />
                          {getError(errors, "build_spec.image_repository.external_image_repo_url") && (
                            <p className="text-sm text-destructive">{getError(errors, "build_spec.image_repository.external_image_repo_url")}</p>
                          )}
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`git-revision-type-${index}`} className="text-sm font-medium">
                            Git Revision Type <span className="text-red-500">*</span>
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                              </TooltipTrigger>
                              <TooltipContent side="top">Select the type of Git revision to use</TooltipContent>
                            </Tooltip>
                          </div>
                          <Select
                            value={resource.gitRevisionType}
                            onValueChange={val => update({ gitRevisionType: val as "branch" | "commit" | "tag" })}
                          >
                            <SelectTrigger
                              id={`git-revision-type-${index}`}
                              className={`max-w-xl ${getError(errors, "gitRevisionType") ? "border-destructive" : ""}`}
                            >
                              <SelectValue placeholder="Select revision type" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="branch">Branch</SelectItem>
                              <SelectItem value="commit">Commit</SelectItem>
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
                                {resource.gitRevisionType === "branch"
                                  ? "Branch Name"
                                  : resource.gitRevisionType === "commit"
                                    ? "Commit Hash"
                                    : "Tag Name"} <span className="text-red-500">*</span>
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
                              className={getError(errors, `volume_mounts.${vmIdx}.source_volume_name`) ? "border-destructive" : ""}
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
                            <p className="text-sm text-destructive">{getError(errors, `volume_mounts.${vmIdx}.source_volume_name`)}</p>
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
                            Target Path <span className="text-red-500">*</span>
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
                  <h3 className="text-lg font-medium mb-3">Ports</h3>
                  <div className="grid gap-6 max-w-3xl">
                    {(resource.ports || []).map((port, pidx) => (
                      <div key={pidx} className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end border p-3 rounded-md bg-muted/10">
                        <div>
                          <div className="flex items-center gap-1 mb-2">
                            <Label htmlFor={`port-number-${index}-${pidx}`} className="text-sm font-medium">
                            Port Number <span className="text-red-500">*</span>
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
                            className={getError(errors, `ports.${pidx}.number`) ? "border-destructive" : ""}
                            required
                          />
                          {getError(errors, `ports.${pidx}.number`) && (
                            <p className="text-sm text-destructive">{getError(errors, `ports.${pidx}.number`)}</p>
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
                  <h3 className="text-lg font-medium mb-3">Pre-Deployment step</h3>
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
                          <TooltipTrigger asChild>
                            <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                          </TooltipTrigger>
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
                  <h3 className="text-lg font-medium mb-3">Main container step</h3>
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
                      <Input
                        id={`exec-command-${index}`}
                        value={resource.execution_config?.command?.join(",") || ""}
                        onChange={e => update({ execution_config: { ...resource.execution_config, command: e.target.value.split(",").map(s => s.trim()).filter(Boolean) } })}
                        placeholder="e.g., node,server.js"
                      />
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
                <div className="flex items-center mb-3">
                  <h3 className="text-lg font-medium">Environment Variables</h3>
                  <div className="ml-auto flex gap-2">
                    <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => {
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
                    <div className="col-span-2 text-center">Use Secret</div>
                    <div className="col-span-1"></div>
                  </div>

                  {/* Environment Variables Rows */}
                  {(resource.execution_config?.environment_variables || []).length ? (
                    (resource.execution_config?.environment_variables || []).map((env, envIdx) => (
                      <div key={envIdx} className="grid grid-cols-12 gap-2 p-3 border-b last:border-b-0 items-start">
                        {/* Key Input - Fixed width */}
                        <div className="col-span-3">
                          <Input
                            value={env.name || ""}
                            onChange={(e) => updateEnvVar(envIdx, { name: e.target.value })}
                            className="w-full text-sm font-mono"
                            placeholder="KEY"
                          />
                        </div>

                        {/* Value Input/Secret Selection - Fixed width */}
                        <div className="col-span-6">
                          {env.useSecret ? (
                            <div className="space-y-2">
                              <Select
                                value={env.selectedSecretId || ""}
                                onValueChange={(value) => updateEnvVar(envIdx, { selectedSecretId: value, selectedSecretKey: undefined })}
                                disabled={secrets.isLoading || secrets.secrets.filter(s => s.type === 'Generic').length === 0}
                              >
                                <SelectTrigger className="w-full">
                                  <SelectValue placeholder={
                                    secrets.secrets.filter(s => s.type === 'Generic').length === 0
                                      ? "No generic secrets available"
                                      : "select secret..."
                                  } />
                                </SelectTrigger>
                                <SelectContent>
                                  {secrets.secrets.filter(s => s.type === 'Generic').map((secret) => (
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
                              {env.selectedSecretId && (() => {
                                const selectedSecret = secrets.secrets.find(s => s.id === env.selectedSecretId);
                                const availableKeys = selectedSecret?.data?.map(d => d.key) || [];

                                return (
                                  <Select
                                    value={env.selectedSecretKey || ""}
                                    onValueChange={(value) => updateEnvVar(envIdx, { selectedSecretKey: value })}
                                    disabled={availableKeys.length === 0}
                                  >
                                    <SelectTrigger className="w-full">
                                      <SelectValue placeholder={
                                        availableKeys.length === 0
                                          ? "No keys available in secret"
                                          : "select key..."
                                      } />
                                    </SelectTrigger>
                                    <SelectContent>
                                      {availableKeys.map((key) => (
                                        <SelectItem key={key} value={key}>
                                          {key}
                                        </SelectItem>
                                      ))}
                                    </SelectContent>
                                  </Select>
                                );
                              })()}
                            </div>
                          ) : (
                            <Input
                              value={env.value || ""}
                              onChange={(e) => updateEnvVar(envIdx, { value: e.target.value })}
                              className="w-full text-sm font-mono"
                              placeholder="VALUE"
                            />
                          )}
                        </div>

                        {/* Use Secret Toggle - Fixed width */}
                        <div className="col-span-2 flex justify-center items-start pt-2">
                          <Switch
                            checked={env.useSecret || false}
                            onCheckedChange={(checked) => {
                              if (checked) {
                                updateEnvVar(envIdx, { useSecret: true, value: '' });
                              } else {
                                updateEnvVar(envIdx, {
                                  useSecret: false,
                                  selectedSecretId: undefined,
                                  selectedSecretKey: undefined
                                });
                              }
                            }}
                            disabled={secrets.isLoading}
                          />
                        </div>

                        {/* Remove Button - Fixed width */}
                        <div className="col-span-1 flex justify-center items-start pt-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 hover:bg-destructive/10 hover:text-destructive"
                            onClick={() => removeEnvVar(envIdx)}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="p-8 text-center text-muted-foreground">
                      No environment variables defined
                    </div>
                  )}
                </div>
                {/* Add Variable button */}
                <div className="flex justify-end mt-2">
                  <Button variant="ghost" size="sm" onClick={addEnvVar}>
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
                  className="text-destructive hover:text-destructive hover:bg-destructive/10 focus-visible:bg-destructive/10"
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

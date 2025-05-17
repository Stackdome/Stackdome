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
import { Plus, X, GitBranch, ImageIcon, Trash2 } from "lucide-react";
import type { StackResourceData } from "@/pages/stacks/schemas/stack-create-schema";

interface StackResourceItemProps {
  resource: Partial<StackResourceData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  isOnlyResource: boolean;
  onChange: (index: number, updatedResource: Partial<StackResourceData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
}

export default function StackResourceItem({
  resource,
  index,
  itemRef,
  isOnlyResource,
  onChange,
  onRemove,
  errors,
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
                  />
                  {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
                </div>
                <div>
                  <div className="flex items-center gap-1 mb-2">
                    <Label htmlFor={`depends-on-${index}`} className="text-sm font-medium">
                      Depends On
                    </Label>
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                      <TooltipContent side="top" className="max-w-xs">
                        <p>Comma separated list of resource names this depends on. Ensure these resources are defined in your stack.</p>
                      </TooltipContent>
                    </Tooltip>
                  </div>
                  <Input
                    id={`depends-on-${index}`}
                    value={(resource.depends_on || []).join(", ")}
                    onChange={e => update({ depends_on: e.target.value.split(",").map(s => s.trim()).filter(Boolean) })}
                    placeholder="e.g., database, redis"
                    className="max-w-md"
                  />
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
                      className={`max-w-xl ${errors["image_spec.image"] ? "border-destructive" : ""}`}
                      required={resource.sourceType === "image"}
                      aria-invalid={!!errors["image_spec.image"]}
                    />
                    {errors["image_spec.image"] && (
                      <p className="text-sm text-destructive">{errors["image_spec.image"]}</p>
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
                      className={`max-w-xl ${errors["build_spec.source_context.git_repo.repo_url"] ? "border-destructive" : ""}`}
                      required={resource.sourceType === "git"}
                      aria-invalid={!!errors["build_spec.source_context.git_repo.repo_url"]}
                    />
                    {errors["build_spec.source_context.git_repo.repo_url"] && (
                      <p className="text-sm text-destructive">{errors["build_spec.source_context.git_repo.repo_url"]}</p>
                    )}
                  </div>
                  {/* Add more build_spec fields as needed */}
                </div>
              )}
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
                          <SelectItem value="udp">UDP</SelectItem>
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
                    <Label htmlFor={`init-image-${index}`} className="text-sm font-medium">
                      Init Image
                    </Label>
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                      <TooltipContent side="top">Optional image for initialization step</TooltipContent>
                    </Tooltip>
                  </div>
                  <Input
                    id={`init-image-${index}`}
                    value={resource.init_spec?.image_spec?.image || ""}
                    onChange={e => update({ init_spec: { ...resource.init_spec, image_spec: { image: e.target.value } } })}
                    placeholder="e.g., busybox:latest"
                    className="max-w-xl"
                  />
                </div>
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
              <Button variant="outline" size="sm" onClick={addEnvVar}>
                <Plus className="h-4 w-4 mr-2" /> Add Variable
              </Button>
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
              variant="destructive"
              onClick={() => onRemove(index)}
            >
              <Trash2 className="h-5 w-5 mr-1" />
              Remove Resource
            </Button>
          </div>
        )}
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

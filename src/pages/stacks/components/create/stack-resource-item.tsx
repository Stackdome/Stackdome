import React from "react";
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
  const updateBuildSpec = (patch: any) => {
    update({
      build_spec: { ...resource.build_spec, ...patch },
      image_spec: undefined,
    });
  };
  // Helper for updating nested image_spec
  const updateImageSpec = (patch: any) => {
    update({
      image_spec: { ...resource.image_spec, ...patch },
      build_spec: undefined,
    });
  };

  // Helper for updating ports
  const updatePort = (pidx: number, patch: any) => {
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
    <AccordionItem value={String(index)} className="border rounded-md last:border-b">
      <AccordionTrigger className="px-4 py-4 hover:bg-muted/10 data-[state=open]:bg-muted/40 rounded-t-lg transition-colors duration-200" ref={itemRef}>
        <div className="flex items-center gap-3 w-full justify-between">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary/15 text-primary font-medium shadow-sm">
              {index + 1}
            </div>
            <span className="font-medium text-lg">
              {resource.name || `Resource ${index + 1}`}
            </span>
            {resource.name === "" && (
              <span className="ml-2 text-xs text-muted-foreground font-normal inline-block align-middle">
                (unnamed)
              </span>
            )}
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="p-0 rounded-b-lg overflow-hidden">
        <Tabs defaultValue="configuration" className="w-full">
          <div className="border-b px-4 mt-2">
            <TabsList className="mb-0 h-12">
              <TabsTrigger value="configuration" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Configuration</TabsTrigger>
              <TabsTrigger value="deployment" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Deployment</TabsTrigger>
              <TabsTrigger value="environment" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Environment Variables</TabsTrigger>
            </TabsList>
          </div>

          {/* General Section (always at top) */}
          <TabsContent value="configuration" className="p-6 space-y-8">
            <div>
              <h3 className="text-lg font-semibold mb-4">General</h3>
              <div className="grid gap-6 max-w-3xl">
                <div>
                  <div className="flex items-center gap-1 mb-1">
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
                    className="max-w-md"
                    placeholder="e.g., web-server, database, cache"
                  />
                  {errors.name && <p className="text-sm text-red-500 mt-1">{errors.name}</p>}
                </div>
                <div>
                  <div className="flex items-center gap-1 mb-1">
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
            <Separator />
            {/* Source Configuration Section */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Source Configuration</h3>
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
                    <div className="flex items-center gap-1 mb-1">
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
                      className="max-w-xl"
                      required={resource.sourceType === "image"}
                    />
                  </div>
                </div>
              ) : (
                <div className="grid gap-6 max-w-3xl">
                  <div>
                    <div className="flex items-center gap-1 mb-1">
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
                      className="max-w-xl"
                      required={resource.sourceType === "git"}
                    />
                  </div>
                  {/* Add more build_spec fields as needed */}
                </div>
              )}
            </div>
            <Separator />
            {/* Ports Section */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Ports Configuration</h3>
              <div className="grid gap-6 max-w-3xl">
                {(resource.ports || []).map((port, pidx) => (
                  <div key={pidx} className="grid grid-cols-2 md:grid-cols-5 gap-4 items-end border p-3 rounded-md bg-muted/10">
                    <div>
                      <div className="flex items-center gap-1 mb-1">
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
                      <div className="flex items-center gap-1 mb-1">
                        <Label htmlFor={`port-protocol-${index}-${pidx}`} className="text-sm font-medium">
                          Protocol
                        </Label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent side="top">tcp or http</TooltipContent>
                        </Tooltip>
                      </div>
                      <Input
                        id={`port-protocol-${index}-${pidx}`}
                        value={port.protocol || ""}
                        onChange={e => updatePort(pidx, { protocol: e.target.value })}
                        placeholder="http"
                      />
                    </div>
                    <div>
                      <div className="flex items-center gap-1 mb-1">
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
                      <div className="flex items-center gap-1 mb-1">
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
                  <Button variant="outline" onClick={addPort}>
                    <Plus className="h-4 w-4 mr-2" />Add Port
                  </Button>
                </div>
              </div>
            </div>
          </TabsContent>

          {/* Deployment Tab */}
          <TabsContent value="deployment" className="p-6 space-y-8">
            {/* Pre-Deploy Section (Init) */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Pre-Deployment Configuration</h3>
              <p className="text-sm text-muted-foreground mb-4">
                Commands to run before the main container starts
              </p>
              <div className="grid gap-6 max-w-3xl">
                <div>
                  <div className="flex items-center gap-1 mb-1">
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
                  <div className="flex items-center gap-1 mb-1">
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
                  <div className="flex items-center gap-1 mb-1">
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
            <Separator />
            {/* Post-Deploy Section (Execution) */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Main Container Configuration</h3>
              <p className="text-sm text-muted-foreground mb-4">
                Main container runtime settings
              </p>
              <div className="grid gap-6 max-w-3xl">
                <div>
                  <div className="flex items-center gap-1 mb-1">
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
                  <div className="flex items-center gap-1 mb-1">
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
          <TabsContent value="environment" className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Environment Variables</h3>
              <Button variant="outline" size="sm" onClick={addEnvVar}>
                <Plus className="h-4 w-4 mr-2" /> Add Variable
              </Button>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full border border-muted rounded-md">
                <thead className="bg-muted/30">
                  <tr>
                    <th className="text-left px-6 py-3 font-semibold text-sm">Key</th>
                    <th className="text-left px-6 py-3 font-semibold text-sm">Value</th>
                    <th className="w-12"></th>
                  </tr>
                </thead>
                <tbody>
                  {(resource.execution_config?.environment_variables || []).length ? (
                    resource.execution_config?.environment_variables.map((env, envIdx) => (
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

          <div className="flex justify-end mt-4 px-6 py-4 border-t bg-muted/30">
            <Button
              variant="destructive"
              onClick={() => onRemove(index)}
              disabled={isOnlyResource}
              size="sm"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Remove Resource
            </Button>
          </div>
        </Tabs>
      </AccordionContent>
    </AccordionItem>
  );
}

import { useState, useEffect, useRef } from "react";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Plus, X, GitBranch, Image as ImageIcon, Trash2 } from "lucide-react";

// Types from OpenAPI
import type { components } from "@/api/types/openapi";

type StackResource = components["schemas"]["StackResource"];

type StackResourceForm = Omit<
  StackResource,
  | "id"
  | "stack_id"
  | "version"
  | "status"
  | "labels"
  | "annotations"
  | "volume_mounts"
  | "lifecycle_config"
  | "stateful"
>;

// Source type constants
const SourceType = {
  Image: "image" as const,
  Git: "git" as const
};
type SourceTypeValue = typeof SourceType.Image | typeof SourceType.Git;

const LOCAL_STORAGE_KEY = "stackdome-stack-resources-form";

function getDefaultResource(): StackResourceForm {
  return {
    name: "",
    build_spec: undefined,
    image_spec: undefined,
    init_spec: undefined,
    execution_config: undefined,
    depends_on: [],
    ports: [],
  };
}

// Helper to ensure build_spec is always fully valid
function getValidBuildSpec(
  current: Partial<components["schemas"]["StackResourceBuildSpec"]>,
  patch: Partial<components["schemas"]["StackResourceBuildSpec"]>
): components["schemas"]["StackResourceBuildSpec"] {
  // Always provide all required fields, never undefined
  return {
    source_context:
      patch.source_context ||
      current?.source_context ||
      { git_repo: { repo_url: "" } },
    context_path_within_source:
      patch.context_path_within_source ??
      current?.context_path_within_source ??
      "./",
    dockerfile_path:
      patch.dockerfile_path ??
      current?.dockerfile_path ??
      "Dockerfile",
    source_revision:
      patch.source_revision ||
      current?.source_revision ||
      {},
    image_repository_url: {
      url:
        patch.image_repository_url?.url ??
        current?.image_repository_url?.url ??
        "",
      cluster_registry_id:
        patch.image_repository_url?.cluster_registry_id ??
        current?.image_repository_url?.cluster_registry_id ??
        "",
    },
    insecure_registry:
      patch.insecure_registry ??
      current?.insecure_registry ??
      false,
  };
}

export default function StackResourcesForm() {
  const [resources, setResources] = useState<StackResourceForm[]>(() => {
    const saved = localStorage.getItem(LOCAL_STORAGE_KEY);
    return saved ? JSON.parse(saved) : [getDefaultResource()];
  });
  
  // Track the source type for each resource (image or git)
  const [sourceTypes, setSourceTypes] = useState<Record<number, SourceTypeValue>>({});
  
  const [openAccordions, setOpenAccordions] = useState<string[]>(["0"]);
  const itemRefs = useRef<Array<HTMLDivElement | null>>([]);
  
  // Initialize source types when resources change
  useEffect(() => {
    setSourceTypes(prev => {
      const newSourceTypes = { ...prev };
      resources.forEach((resource, idx) => {
        if (!newSourceTypes[idx]) {
          // Default to Git if there's a build_spec, otherwise Image
          newSourceTypes[idx] = resource.build_spec ? SourceType.Git : SourceType.Image;
        }
      });
      return newSourceTypes;
    });
  }, [resources]);

  // Persist to localStorage
  useEffect(() => {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(resources));
  }, [resources]);

  // Remove resource
  const removeResource = (idx: number) => {
    setResources((prev) => prev.filter((_, i) => i !== idx));
    setOpenAccordions((prev) => prev.filter((id) => id !== String(idx)));
    
    // Also remove from sourceTypes
    setSourceTypes((prev) => {
      const newSourceTypes = { ...prev };
      delete newSourceTypes[idx];
      return newSourceTypes;
    });
  };

  // Update resource
  const updateResource = (idx: number, patch: Partial<StackResourceForm>) => {
    setResources((prev) =>
      prev.map((r, i) => (i === idx ? { ...r, ...patch } : r))
    );
  };

  // Update source type
  const updateSourceType = (idx: number, sourceType: SourceTypeValue) => {
    setSourceTypes((prev) => ({
      ...prev,
      [idx]: sourceType
    }));
  };

  // Add environment variable
  const addEnvironmentVariable = (idx: number) => {
    updateResource(idx, {
      execution_config: {
        ...resources[idx].execution_config,
        environment_variables: [
          ...(resources[idx].execution_config?.environment_variables || []),
          { name: '', value: '' }
        ]
      }
    });
  };

  // Update environment variable
  const updateEnvironmentVariable = (idx: number, envIdx: number, name: string, value: string) => {
    updateResource(idx, {
      execution_config: {
        ...resources[idx].execution_config,
        environment_variables: resources[idx].execution_config?.environment_variables?.map(
          (env, i) => i === envIdx ? { name, value } : env
        ) || []
      }
    });
  };

  // Remove environment variable
  const removeEnvironmentVariable = (idx: number, envIdx: number) => {
    updateResource(idx, {
      execution_config: {
        ...resources[idx].execution_config,
        environment_variables: resources[idx].execution_config?.environment_variables?.filter(
          (_, i) => i !== envIdx
        ) || []
      }
    });
  };

  // The scrollable container for the form
  return (
    <div className="flex flex-col h-full">
      <div className="overflow-y-auto flex-1 scrollbar-hide px-6 pt-6 pb-6">
        <Accordion
          type="multiple"
          value={openAccordions}
          onValueChange={setOpenAccordions}
          className="w-full space-y-4"
        >
          {resources.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 border-2 border-dashed rounded-lg border-muted">
              <div className="mb-3 p-3 rounded-full bg-muted">
                <Plus className="h-6 w-6 text-muted-foreground" />
              </div>
              <p className="text-lg text-muted-foreground">No resources defined yet</p>
              <p className="text-sm text-muted-foreground mb-4">Add your first stack resource to get started</p>
              <Button 
                onClick={() => {
                  const newResource = getDefaultResource();
                  const newIndex = 0;
                  setResources([newResource]);
                  setOpenAccordions([String(newIndex)]);
                }}
              >
                <Plus className="w-4 h-4 mr-2" />
                Add First Resource
              </Button>
            </div>
          )}
          {resources.map((resource, idx) => (
            <AccordionItem
              key={idx}
              value={String(idx)}
              className="border rounded-md last:border-b"
              ref={el => { itemRefs.current[idx] = el; }}
            >
              <AccordionTrigger className="px-4 py-4 hover:bg-muted/10 data-[state=open]:bg-muted/40 rounded-t-lg transition-colors duration-200">
                <div className="flex items-center gap-3 w-full justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary/15 text-primary font-medium shadow-sm">
                      {idx + 1}
                    </div>
                    <span className="font-medium text-lg">
                      {resource.name || `Resource ${idx + 1}`}
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
                  
                  {/* Configuration Tab (combines General + Build/Image + Ports) */}
                  <TabsContent value="configuration" className="p-6 space-y-8">
                    {/* Source Information Section */}
                    <div>
                      <h3 className="text-lg font-semibold mb-4">Source Information</h3>
                      <div className="grid gap-6 max-w-3xl">
                        <div>
                          <div className="flex items-center gap-1 mb-1">
                            <Label htmlFor={`name-${idx}`} className="text-sm font-medium">
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
                            id={`name-${idx}`}
                            value={resource.name}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, { name: e.target.value })
                            }
                            required
                            className="max-w-md"
                            placeholder="e.g., web-server, database, cache"
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-1">
                            <Label htmlFor={`depends-on-${idx}`} className="text-sm font-medium">
                              Depends On
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top" className="max-w-xs">
                                <p>Comma separated list of resource names this depends on</p>
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`depends-on-${idx}`}
                            value={(resource.depends_on || []).join(",")}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                depends_on: e.target.value
                                  .split(",")
                                  .map((s) => s.trim())
                                  .filter(Boolean),
                              })
                            }
                            placeholder="e.g., database, redis, cache"
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
                          <Label htmlFor={`source-type-${idx}`} className="text-sm font-medium">
                            Build From <span className="text-red-500">*</span>
                          </Label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent side="top" className="max-w-xs">
                              <p>Select how this resource should be built</p>
                            </TooltipContent>
                          </Tooltip>
                        </div>
                        <Select 
                          value={sourceTypes[idx] || SourceType.Image} 
                          onValueChange={(value) => updateSourceType(idx, value as SourceTypeValue)}
                        >
                          <SelectTrigger className="w-[180px]">
                            <SelectValue placeholder="Select source" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value={SourceType.Image}>
                              <div className="flex items-center gap-2">
                                <ImageIcon size={16} />
                                <span>Container Image</span>
                              </div>
                            </SelectItem>
                            <SelectItem value={SourceType.Git}>
                              <div className="flex items-center gap-2">
                                <GitBranch size={16} />
                                <span>Git Repository</span>
                              </div>
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      {/* Conditional content based on source type */}
                      {sourceTypes[idx] === SourceType.Image ? (
                        <div className="grid gap-4 max-w-3xl">
                          <div>
                            <div className="flex items-center gap-1 mb-1">
                              <Label htmlFor={`container-image-${idx}`} className="text-sm font-medium">
                                Container Image <span className="text-red-500">*</span>
                              </Label>
                              <Tooltip delayDuration={300}>
                                <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                <TooltipContent side="top" className="max-w-xs">
                                  <p>Specify the container image to use (e.g., nginx:latest, redis:alpine)</p>
                                </TooltipContent>
                              </Tooltip>
                            </div>
                            <Input
                              id={`container-image-${idx}`}
                              value={resource.image_spec?.image || ""}
                              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                updateResource(idx, {
                                  image_spec: { image: e.target.value },
                                  // Clear build_spec when switching to image mode
                                  build_spec: undefined
                                })
                              }
                              placeholder="e.g., nginx:latest, postgres:13"
                              className="max-w-xl"
                            />
                          </div>
                        </div>
                      ) : (
                        <div className="grid gap-6 max-w-3xl">
                          <div>
                            <div className="flex items-center gap-1 mb-1">
                              <Label htmlFor={`git-repo-${idx}`} className="text-sm font-medium">
                                Git Repository URL <span className="text-red-500">*</span>
                              </Label>
                              <Tooltip delayDuration={300}>
                                <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                <TooltipContent side="top" className="max-w-xs">
                                  <p>The Git repository URL containing the source code to build</p>
                                </TooltipContent>
                              </Tooltip>
                            </div>
                            <Input
                              id={`git-repo-${idx}`}
                              value={
                                resource.build_spec?.source_context?.git_repo
                                  ?.repo_url || ""
                              }
                              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                updateResource(idx, {
                                  build_spec: getValidBuildSpec(resource.build_spec || {}, {
                                    source_context: {
                                      git_repo: { repo_url: e.target.value },
                                    },
                                  }),
                                  // Clear image_spec when switching to git mode
                                  image_spec: undefined
                                })
                              }
                              placeholder="https://github.com/username/repository.git"
                              className="max-w-xl"
                            />
                          </div>

                          <div className="flex gap-4">
                            <div className="flex-1">
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`context-path-${idx}`} className="text-sm font-medium">
                                  Context Path
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Directory to build from (default: ./)
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`context-path-${idx}`}
                                value={
                                  resource.build_spec?.context_path_within_source ||
                                  ""
                                }
                                onChange={
                                  (e: React.ChangeEvent<HTMLInputElement>) =>
                                    updateResource(idx, {
                                      build_spec: getValidBuildSpec(
                                        resource.build_spec || {},
                                        { context_path_within_source: e.target.value }
                                      ),
                                    })
                                }
                                placeholder="./"
                              />
                            </div>
                            <div className="flex-1">
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`dockerfile-path-${idx}`} className="text-sm font-medium">
                                  Dockerfile Path
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Path to Dockerfile (default: Dockerfile)
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`dockerfile-path-${idx}`}
                                value={
                                  resource.build_spec?.dockerfile_path || ""
                                }
                                onChange={
                                  (e: React.ChangeEvent<HTMLInputElement>) =>
                                    updateResource(idx, {
                                      build_spec: getValidBuildSpec(
                                        resource.build_spec || {},
                                        { dockerfile_path: e.target.value }
                                      ),
                                    })
                                }
                                placeholder="Dockerfile"
                              />
                            </div>
                          </div>

                          <div className="flex gap-4">
                            <div className="flex-1">
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`image-repo-url-${idx}`} className="text-sm font-medium">
                                  Image Repository URL <span className="text-red-500">*</span>
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Where to push built images
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`image-repo-url-${idx}`}
                                value={
                                  resource.build_spec?.image_repository_url?.url ||
                                  ""
                                }
                                onChange={
                                  (e: React.ChangeEvent<HTMLInputElement>) =>
                                    updateResource(idx, {
                                      build_spec: getValidBuildSpec(
                                        resource.build_spec || {},
                                        {
                                          image_repository_url: {
                                            url: e.target.value,
                                            cluster_registry_id:
                                              resource.build_spec?.image_repository_url
                                                ?.cluster_registry_id || "",
                                          },
                                        }
                                      ),
                                    })
                                }
                              />
                            </div>
                            <div className="flex-1">
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`cluster-registry-${idx}`} className="text-sm font-medium">
                                  Cluster Registry ID
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Cluster registry identifier
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`cluster-registry-${idx}`}
                                value={
                                  resource.build_spec?.image_repository_url
                                    ?.cluster_registry_id || ""
                                }
                                onChange={
                                  (e: React.ChangeEvent<HTMLInputElement>) =>
                                    updateResource(idx, {
                                      build_spec: getValidBuildSpec(
                                        resource.build_spec || {},
                                        {
                                          image_repository_url: {
                                            url:
                                              resource.build_spec?.image_repository_url
                                                ?.url || "",
                                            cluster_registry_id: e.target.value,
                                          },
                                        }
                                      ),
                                    })
                                }
                              />
                            </div>
                          </div>

                          <div>
                            <div className="flex items-center gap-2">
                              <Label htmlFor={`insecure-registry-${idx}`} className="text-sm font-medium">
                                Insecure Registry
                              </Label>
                              <Tooltip delayDuration={300}>
                                <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                <TooltipContent side="top">
                                  Enable if registry is not HTTPS
                                </TooltipContent>
                              </Tooltip>
                              <Switch
                                id={`insecure-registry-${idx}`}
                                checked={!!resource.build_spec?.insecure_registry}
                                onCheckedChange={
                                  (v: boolean) =>
                                    updateResource(idx, {
                                      build_spec: getValidBuildSpec(
                                        resource.build_spec || {},
                                        { insecure_registry: v }
                                      ),
                                    })
                                }
                              />
                            </div>
                          </div>
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
                                <Label htmlFor={`port-number-${idx}-${pidx}`} className="text-sm font-medium">
                                  Port Number <span className="text-red-500">*</span>
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">Port number</TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`port-number-${idx}-${pidx}`}
                                type="number"
                                value={port.number}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                                  const n = Number(e.target.value);
                                  updateResource(idx, {
                                    ports: (resource.ports || []).map(
                                      (pt, i) =>
                                        i === pidx
                                          ? { ...pt, number: n }
                                          : pt
                                    ),
                                  });
                                }}
                                required
                              />
                            </div>
                            <div>
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`port-protocol-${idx}-${pidx}`} className="text-sm font-medium">
                                  Protocol
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">tcp or http</TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`port-protocol-${idx}-${pidx}`}
                                value={port.protocol || ""}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                  updateResource(idx, {
                                    ports: (resource.ports || []).map(
                                      (pt, i) =>
                                        i === pidx
                                          ? { ...pt, protocol: e.target.value }
                                          : pt
                                    ),
                                  })
                                }
                                placeholder="http"
                              />
                            </div>
                            <div>
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`port-public-${idx}-${pidx}`} className="text-sm font-medium">
                                  Expose Publicly
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Expose to public internet
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <div className="pt-2 pl-2">
                                <Switch
                                  id={`port-public-${idx}-${pidx}`}
                                  checked={!!port.exposed_to_public}
                                  onCheckedChange={(v) =>
                                    updateResource(idx, {
                                      ports: (resource.ports || []).map(
                                        (pt, i) =>
                                          i === pidx
                                            ? { ...pt, exposed_to_public: v }
                                            : pt
                                      ),
                                    })
                                  }
                                />
                              </div>
                            </div>
                            <div>
                              <div className="flex items-center gap-1 mb-1">
                                <Label htmlFor={`port-subdomain-${idx}-${pidx}`} className="text-sm font-medium">
                                  Subdomain Prefix
                                </Label>
                                <Tooltip delayDuration={300}>
                                  <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                                  <TooltipContent side="top">
                                    Prefix for public URL
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                              <Input
                                id={`port-subdomain-${idx}-${pidx}`}
                                value={port.subdomain_prefix || ""}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                  updateResource(idx, {
                                    ports: (resource.ports || []).map(
                                      (pt, i) =>
                                        i === pidx
                                          ? {
                                              ...pt,
                                              subdomain_prefix: e.target.value,
                                            }
                                          : pt
                                    ),
                                  })
                                }
                              />
                            </div>
                            <div className="flex justify-end">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() =>
                                  updateResource(idx, {
                                    ports: (resource.ports || []).filter(
                                      (_, i) => i !== pidx
                                    ),
                                  })
                                }
                                title="Remove port"
                              >
                                <Trash2 className="h-5 w-5" />
                              </Button>
                            </div>
                          </div>
                        ))}
                        <div>
                          <Button
                            variant="outline"
                            onClick={() =>
                              updateResource(idx, {
                                ports: [
                                  ...(resource.ports || []),
                                  {
                                    number: 80,
                                    protocol: "http",
                                    exposed_to_public: false,
                                    subdomain_prefix: ""
                                  },
                                ],
                              })
                            }
                          >
                            <Plus className="h-4 w-4 mr-2" /> Add Port
                          </Button>
                        </div>
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Deployment Tab (combines Init and Execution) */}
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
                            <Label htmlFor={`init-image-${idx}`} className="text-sm font-medium">
                              Init Image
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top">
                                Optional image for initialization step
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`init-image-${idx}`}
                            value={resource.init_spec?.image_spec?.image || ""}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                init_spec: {
                                  ...resource.init_spec,
                                  image_spec: { image: e.target.value },
                                },
                              })
                            }
                            placeholder="e.g., busybox:latest"
                            className="max-w-xl"
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-1">
                            <Label htmlFor={`init-command-${idx}`} className="text-sm font-medium">
                              Init Command
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top">
                                Pre-deployment command (comma separated)
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`init-command-${idx}`}
                            value={resource.init_spec?.command?.join(",") || ""}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                init_spec: {
                                  ...resource.init_spec,
                                  command: e.target.value
                                    .split(",")
                                    .map((s) => s.trim())
                                    .filter(Boolean),
                                },
                              })
                            }
                            placeholder="e.g., sh,/scripts/init.sh"
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-1">
                            <Label htmlFor={`init-args-${idx}`} className="text-sm font-medium">
                              Init Arguments
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top">
                                Pre-deployment arguments (comma separated)
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`init-args-${idx}`}
                            value={resource.init_spec?.args?.join(",") || ""}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                init_spec: {
                                  ...resource.init_spec,
                                  args: e.target.value
                                    .split(",")
                                    .map((s) => s.trim())
                                    .filter(Boolean),
                                },
                              })
                            }
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
                            <Label htmlFor={`exec-command-${idx}`} className="text-sm font-medium">
                              Command
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top">
                                Container runtime command (comma separated)
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`exec-command-${idx}`}
                            value={
                              resource.execution_config?.command?.join(",") ||
                              ""
                            }
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                execution_config: {
                                  ...resource.execution_config,
                                  command: e.target.value
                                    .split(",")
                                    .map((s) => s.trim())
                                    .filter(Boolean),
                                },
                              })
                            }
                            placeholder="e.g., npm,start"
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-1 mb-1">
                            <Label htmlFor={`exec-args-${idx}`} className="text-sm font-medium">
                              Arguments
                            </Label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent side="top">
                                Container runtime arguments (comma separated)
                              </TooltipContent>
                            </Tooltip>
                          </div>
                          <Input
                            id={`exec-args-${idx}`}
                            value={
                              resource.execution_config?.args?.join(",") || ""
                            }
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                              updateResource(idx, {
                                execution_config: {
                                  ...resource.execution_config,
                                  args: e.target.value
                                    .split(",")
                                    .map((s) => s.trim())
                                    .filter(Boolean),
                                },
                              })
                            }
                            placeholder="e.g., --port=3000,--verbose"
                          />
                        </div>
                      </div>
                    </div>
                  </TabsContent>

                  {/* Environment Variables Tab */}
                  <TabsContent value="environment" className="p-6">
                    <div className="flex items-center justify-end mb-4">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => addEnvironmentVariable(idx)}
                      >
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
                          {resource.execution_config?.environment_variables?.length ? (
                            resource.execution_config.environment_variables.map((env, envIdx) => (
                              <tr key={envIdx} className="border-t border-muted">
                                <td className="px-6 py-2 align-middle">
                                  <Input
                                    id={`env-name-${idx}-${envIdx}`}
                                    value={env.name}
                                    onChange={(e) => updateEnvironmentVariable(idx, envIdx, e.target.value, env.value)}
                                    placeholder="VARIABLE_NAME"
                                    className="bg-muted/30 font-mono text-xs"
                                  />
                                </td>
                                <td className="px-6 py-2 align-middle">
                                  <EnvValueCell
                                    value={env.value}
                                    onChange={(v) => updateEnvironmentVariable(idx, envIdx, env.name, v)}
                                  />
                                </td>
                                <td className="px-2 py-2 align-middle text-center">
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => removeEnvironmentVariable(idx, envIdx)}
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
                  
                  <div className="flex justify-between mt-4 px-6 py-4 border-t bg-muted/30">
                    <Button
                      variant="destructive"
                      onClick={() => removeResource(idx)}
                      disabled={resources.length === 1}
                    >
                      Remove Resource
                    </Button>
                  </div>
                </Tabs>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      </div>

      {resources.length > 0 && (
        <div className="sticky bottom-0 left-0 w-full bg-background/95 border-t py-6 px-6 backdrop-blur-sm shadow-md">
          <Button
            variant="outline"
            size="lg"
            className="border-dashed border-2 bg-background hover:bg-muted/20 text-foreground w-full max-w-md mx-auto flex items-center justify-center"
            onClick={() => {
              const newResource = getDefaultResource();
              const newIndex = resources.length;
              setResources([...resources, newResource]);
              setOpenAccordions((prev) => [...prev, String(newIndex)]);
              setTimeout(() => {
                itemRefs.current[newIndex]?.scrollIntoView({ behavior: "smooth", block: "center" });
              }, 50);
            }}
          >
            <Plus className="w-4 h-4 mr-2" />
            Add Another Resource
          </Button>
        </div>
      )}
    </div>
  );
}

// Helper component for value cell (no show/hide, always masked)
function EnvValueCell({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <Input
      type="password"
      value={value}
      onChange={e => onChange(e.target.value)}
      className="bg-muted/30 font-mono text-xs"
    />
  );
}

import { useState, useEffect, useRef } from "react";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Plus } from "lucide-react";

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
  const [openAccordions, setOpenAccordions] = useState<string[]>(["0"]);
  const itemRefs = useRef<Array<HTMLDivElement | null>>([]);

  // Persist to localStorage
  useEffect(() => {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(resources));
  }, [resources]);

  // Remove resource
  const removeResource = (idx: number) => {
    setResources((prev) => prev.filter((_, i) => i !== idx));
    setOpenAccordions((prev) => prev.filter((id) => id !== String(idx)));
  };

  // Update resource
  const updateResource = (idx: number, patch: Partial<StackResourceForm>) => {
    setResources((prev) =>
      prev.map((r, i) => (i === idx ? { ...r, ...patch } : r))
    );
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
              className="border rounded-md overflow-hidden"
              ref={el => { itemRefs.current[idx] = el; }}
            >
              <AccordionTrigger className="px-4 py-4 hover:bg-muted/40 data-[state=open]:bg-muted/40 rounded-t-lg transition-colors duration-200">
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
                <Tabs defaultValue="general" className="w-full">
                  <div className="border-b bg-muted/30 px-4">
                    <TabsList className="mb-0 h-12">
                      <TabsTrigger value="general" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">General</TabsTrigger>
                      <TabsTrigger value="build" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Build</TabsTrigger>
                      <TabsTrigger value="image" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Image</TabsTrigger>
                      <TabsTrigger value="init" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Init</TabsTrigger>
                      <TabsTrigger value="execution" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Execution</TabsTrigger>
                      <TabsTrigger value="ports" className="data-[state=active]:bg-background rounded-t-md rounded-b-none">Ports</TabsTrigger>
                    </TabsList>
                  </div>
                  {/* General Tab */}
                  <TabsContent value="general" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Name <span className="text-red-500">*</span>
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent className="max-w-xs" side="right">
                            <p>A unique name for this resource in your stack</p>
                          </TooltipContent>
                        </Tooltip>
                        <Input
                          value={resource.name}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                            updateResource(idx, { name: e.target.value })
                          }
                          required
                          disabled={false}
                          className="max-w-md"
                          placeholder="e.g., web-server, database, cache"
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Depends On (comma separated)
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent className="max-w-xs" side="right">
                            <p>Comma separated list of resource names this depends on</p>
                          </TooltipContent>
                        </Tooltip>
                        <Input
                          value={(resource.depends_on || []).join(",")}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                            updateResource(idx, {
                              depends_on: e.target.value
                                .split(",")
                                .map((s) => s.trim())
                                .filter(Boolean),
                            })
                          }
                          disabled={false}
                          placeholder="e.g., database, redis, cache"
                          className="max-w-md"
                        />
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Build Tab */}
                  <TabsContent value="build" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Git Repository URL
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent className="max-w-xs" side="right">
                            <p>The Git repository URL containing the source code to build</p>
                          </TooltipContent>
                        </Tooltip>
                        <Input
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
                            })
                          }
                          disabled={false}
                          placeholder="https://github.com/username/repository.git"
                          className="max-w-xl"
                        />
                      </div>
                      <div className="flex gap-2">
                        <div className="flex-1">
                          <label className="text-sm font-medium flex items-center gap-1 mb-2">
                            Context Path
                          </label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent>
                              Directory to build from (default: ./)
                            </TooltipContent>
                          </Tooltip>
                          <Input
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
                            disabled={false}
                          />
                        </div>
                        <div className="flex-1">
                          <label className="text-sm font-medium flex items-center gap-1 mb-2">
                            Dockerfile Path
                          </label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent>
                              Path to Dockerfile (default: Dockerfile)
                            </TooltipContent>
                          </Tooltip>
                          <Input
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
                            disabled={false}
                          />
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <div className="flex-1">
                          <label className="text-sm font-medium flex items-center gap-1 mb-2">
                            Image Repository URL
                          </label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent>
                              Where to push built images
                            </TooltipContent>
                          </Tooltip>
                          <Input
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
                            disabled={false}
                          />
                        </div>
                        <div className="flex-1">
                          <label className="text-sm font-medium flex items-center gap-1 mb-2">
                            Cluster Registry ID
                          </label>
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                            <TooltipContent>
                              Cluster registry identifier
                            </TooltipContent>
                          </Tooltip>
                          <Input
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
                            disabled={false}
                          />
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <label className="text-sm font-medium flex items-center gap-1">
                          Insecure Registry
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Enable if registry is not HTTPS
                          </TooltipContent>
                        </Tooltip>
                        <Switch
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
                          disabled={false}
                        />
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Image Tab */}
                  <TabsContent value="image" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Container Image
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent className="max-w-xs" side="right">
                            <p>Specify the container image to use (e.g., nginx:latest, redis:alpine)</p>
                          </TooltipContent>
                        </Tooltip>
                        <Input
                          value={resource.image_spec?.image || ""}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                            updateResource(idx, {
                              image_spec: { image: e.target.value },
                            })
                          }
                          disabled={false}
                          placeholder="e.g., nginx:latest, postgres:13"
                          className="max-w-xl"
                        />
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Init Tab */}
                  <TabsContent value="init" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Init Image
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Optional image for init step
                          </TooltipContent>
                        </Tooltip>
                        <Input
                          value={resource.init_spec?.image_spec?.image || ""}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                            updateResource(idx, {
                              init_spec: {
                                ...resource.init_spec,
                                image_spec: { image: e.target.value },
                              },
                            })
                          }
                          disabled={false}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Command
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Init command (comma separated)
                          </TooltipContent>
                        </Tooltip>
                        <Input
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
                          disabled={false}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Args
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Init args (comma separated)
                          </TooltipContent>
                        </Tooltip>
                        <Input
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
                          disabled={false}
                        />
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Execution Tab */}
                  <TabsContent value="execution" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Command
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Container command (comma separated)
                          </TooltipContent>
                        </Tooltip>
                        <Input
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
                          disabled={false}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Args
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            Container args (comma separated)
                          </TooltipContent>
                        </Tooltip>
                        <Input
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
                          disabled={false}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium flex items-center gap-1 mb-2">
                          Environment Variables
                        </label>
                        <Tooltip delayDuration={300}>
                          <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                          <TooltipContent>
                            KEY=VALUE, comma separated
                          </TooltipContent>
                        </Tooltip>
                        <Input
                          value={
                            resource.execution_config?.environment_variables
                              ?.map((ev) => `${ev.name}=${ev.value}`)
                              .join(",") || ""
                          }
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                            updateResource(idx, {
                              execution_config: {
                                ...resource.execution_config,
                                environment_variables: e.target.value
                                  .split(",")
                                  .map((pair) => {
                                    const [name, ...rest] = pair.split("=");
                                    return name && rest.length
                                      ? {
                                          name: name.trim(),
                                          value: rest.join("=").trim(),
                                        }
                                      : null;
                                  })
                                  .filter(Boolean) as {
                                  name: string;
                                  value: string;
                                }[],
                              },
                            })
                          }
                          disabled={false}
                        />
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Ports Tab */}
                  <TabsContent value="ports" className="p-6">
                    <div className="grid gap-6 max-w-3xl">
                      {(resource.ports || []).map((port, pidx) => (
                        <div key={pidx} className="flex gap-2 items-end">
                          <div className="flex-1">
                            <label className="text-sm font-medium flex items-center gap-1 mb-2">
                              Port Number <span className="text-red-500">*</span>
                            </label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent>Port number</TooltipContent>
                            </Tooltip>
                            <Input
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
                              disabled={false}
                            />
                          </div>
                          <div className="flex-1">
                            <label className="text-sm font-medium flex items-center gap-1 mb-2">
                              Protocol
                            </label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent>tcp or http</TooltipContent>
                            </Tooltip>
                            <Input
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
                              disabled={false}
                            />
                          </div>
                          <div className="flex-1">
                            <label className="text-sm font-medium flex items-center gap-1 mb-2">
                              Expose Publicly
                            </label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent>
                                Expose to public (for http)
                              </TooltipContent>
                            </Tooltip>
                            <div className="pt-2">
                              <Switch
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
                                disabled={false}
                              />
                            </div>
                          </div>
                          <div className="flex-1">
                            <label className="text-sm font-medium flex items-center gap-1 mb-2">
                              Subdomain Prefix
                            </label>
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                              <TooltipContent>
                                Prefix for public URL
                              </TooltipContent>
                            </Tooltip>
                            <Input
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
                              disabled={false}
                            />
                          </div>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() =>
                              updateResource(idx, {
                                ports: (resource.ports || []).filter(
                                  (_, i) => i !== pidx
                                ),
                              })
                            }
                            disabled={false}
                            className="mb-0.5"
                          >
                            Remove
                          </Button>
                        </div>
                      ))}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          updateResource(idx, {
                            ports: [
                              ...(resource.ports || []),
                              {
                                number: 80,
                                protocol: "http",
                                exposed_to_public: false,
                              },
                            ],
                          })
                        }
                        disabled={false}
                      >
                        Add Port
                      </Button>
                    </div>
                  </TabsContent>
                  
                  <div className="flex justify-between mt-4 px-6 py-4 border-t bg-muted/30">
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => removeResource(idx)}
                      disabled={resources.length === 1 || false}
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
        <div className="sticky bottom-6 left-0 w-full bg-background/95 border-t py-4 px-6 backdrop-blur-sm rounded-b-lg shadow-md">
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

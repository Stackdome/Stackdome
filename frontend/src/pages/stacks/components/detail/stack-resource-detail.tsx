import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Box, GitBranch } from "lucide-react";
import type { FormStackResourceData as StackResourceData } from "@/pages/stacks/schemas/form-schema";

interface StackResourceDetailProps {
  resource: Partial<StackResourceData>;
  index: number;
}

export default function StackResourceDetail({
  resource,
  index,
}: StackResourceDetailProps) {
  return (
    <AccordionItem value={String(index)} className="border-0">
      <AccordionTrigger
        className="px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-2 text-left flex-grow">
          <div className="flex flex-col flex-grow min-w-0">
            <span className="font-medium flex items-center gap-2">
              {resource.name || `Resource ${index + 1}`}
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <span className="h-2 w-2 rounded-full bg-blue-500 cursor-help"></span>
                </TooltipTrigger>
                <TooltipContent side="top">Resource is active</TooltipContent>
              </Tooltip>
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
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="pb-4 pt-2">
        <div className="px-4 space-y-4">
          <Tabs defaultValue="configuration" className="w-full">
            <TabsList className="w-full justify-start">
              <TabsTrigger value="configuration" className="flex-1">Configuration</TabsTrigger>
              <TabsTrigger value="deployment" className="flex-1">Deployment</TabsTrigger>
              <TabsTrigger value="environment" className="flex-1">Environment</TabsTrigger>
            </TabsList>
            <TabsContent value="configuration" className="pt-4 space-y-6">
              <div>
                <h3 className="text-lg font-medium mb-3">General</h3>
                <div className="grid gap-4 max-w-3xl">
                  <div>
                    <div className="mb-1 text-sm font-medium">Resource Name</div>
                    <div className="p-2 bg-muted/30 rounded-md">{resource.name}</div>
                  </div>

                  {resource.depends_on && resource.depends_on.length > 0 && (
                    <div>
                      <div className="mb-1 text-sm font-medium">Depends On</div>
                      <div className="p-2 bg-muted/30 rounded-md">
                        {resource.depends_on.join(", ")}
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">These resources will start first.</p>
                    </div>
                  )}

                  <div>
                    <div className="mb-1 text-sm font-medium">Build From</div>
                    <div className="p-2 bg-muted/30 rounded-md flex items-center gap-2">
                      {resource.sourceType === "image" ? (
                        <>
                          <Box className="h-4 w-4" />
                          <span>Container Image</span>
                        </>
                      ) : (
                        <>
                          <GitBranch className="h-4 w-4" />
                          <span>Git Repository</span>
                        </>
                      )}
                    </div>
                  </div>
                  {resource.sourceType === "image" ? (
                    <div>
                      <div className="mb-1 text-sm font-medium">Container Image</div>
                      <div className="p-2 bg-muted/30 rounded-md">
                        {resource.image_spec?.image || "Not specified"}
                      </div>
                    </div>
                  ) : (
                    <>
                      <div>
                        <div className="mb-1 text-sm font-medium">Git Repository URL</div>
                        <div className="p-2 bg-muted/30 rounded-md overflow-x-auto">
                          {resource.build_spec?.source_context?.git_repo?.repo_url || "Not specified"}
                        </div>
                      </div>
                      <div>
                        <div className="mb-1 text-sm font-medium">Image Repository URL</div>
                        <div className="p-2 bg-muted/30 rounded-md overflow-x-auto">
                          {resource.build_spec?.image_repository?.external_image_repo_url || "Not specified"}
                        </div>
                      </div>
                      {resource.gitRevisionType && (
                        <>
                          <div>
                            <div className="mb-1 text-sm font-medium">Git Revision Type</div>
                            <div className="p-2 bg-muted/30 rounded-md">
                              {resource.gitRevisionType.charAt(0).toUpperCase() + resource.gitRevisionType.slice(1)}
                            </div>
                          </div>
                          <div>
                            <div className="mb-1 text-sm font-medium">
                              {resource.gitRevisionType === "branch"
                                ? "Branch Name"
                                : resource.gitRevisionType === "commit"
                                  ? "Commit Hash"
                                  : "Tag Name"}
                            </div>
                            <div className="p-2 bg-muted/30 rounded-md">
                              {resource.gitRevisionValue || "Not specified"}
                            </div>
                          </div>
                        </>
                      )}
                    </>
                  )}
                </div>
              </div>
              <Separator className="my-4" />

              {/* Volume Mounts Section */}
              <div>
                <h3 className="text-lg font-medium mb-3">Volume Mounts</h3>
                {resource.volume_mounts && resource.volume_mounts.length > 0 ? (
                  <div className="grid gap-3 max-w-3xl">
                    {resource.volume_mounts.map((vm, vmIdx) => (
                      <div key={vmIdx} className="grid grid-cols-1 md:grid-cols-3 gap-4 p-3 rounded-md bg-muted/10 border">
                        <div>
                          <div className="mb-1 text-sm font-medium">Volume</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {vm.source_volume_name}
                          </div>
                        </div>
                        <div>
                          <div className="mb-1 text-sm font-medium">Sub Path</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {vm.source_sub_path || "-"}
                          </div>
                        </div>
                        <div>
                          <div className="mb-1 text-sm font-medium">Target Path</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {vm.target_path}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-sm text-muted-foreground">No volume mounts configured</div>
                )}
              </div>
              <Separator className="my-4" />

              {/* Ports Section */}
              <div>
                <h3 className="text-lg font-medium mb-3">Ports</h3>
                {resource.ports && resource.ports.length > 0 ? (
                  <div className="grid gap-3 max-w-3xl">
                    {resource.ports.map((port, pidx) => (
                      <div key={pidx} className="grid grid-cols-1 md:grid-cols-3 gap-4 p-3 rounded-md bg-muted/10 border">
                        <div>
                          <div className="mb-1 text-sm font-medium">Port</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {port.number}
                          </div>
                        </div>
                        <div>
                          <div className="mb-1 text-sm font-medium">Protocol</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {port.protocol || "TCP"}
                          </div>
                        </div>
                        <div>
                          <div className="mb-1 text-sm font-medium">Access</div>
                          <div className="p-2 bg-muted/30 rounded-md">
                            {port.exposed_to_public ? "Exposed" : "Internal Only"}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-sm text-muted-foreground">No ports configured</div>
                )}
              </div>
            </TabsContent>

            <TabsContent value="deployment" className="pt-4 space-y-6">
              {/* Pre-deploy Section (Init) */}
              <div>
                <h3 className="text-lg font-medium mb-3">Pre-deployment step</h3>
                <div className="grid gap-4 max-w-3xl">
                  {resource.init_spec ? (
                    <>
                      <div>
                        <div className="mb-1 text-sm font-medium">Command</div>
                        <div className="p-2 bg-muted/30 rounded-md">
                          {resource.init_spec.command ? resource.init_spec.command.join(' ') : "-"}
                        </div>
                      </div>
                      <div>
                        <div className="mb-1 text-sm font-medium">Arguments</div>
                        <div className="p-2 bg-muted/30 rounded-md">
                          {resource.init_spec.args ? resource.init_spec.args.join(', ') : "-"}
                        </div>
                      </div>
                    </>
                  ) : (
                    <div className="text-sm text-muted-foreground">No pre-deployment step configured</div>
                  )}
                </div>
              </div>
              <Separator className="my-4" />

              {/* Main Container (Execution) */}
              <div>
                <h3 className="text-lg font-medium mb-3">Main container step</h3>
                <div className="grid gap-4 max-w-3xl">
                  {resource.execution_config ? (
                    <>
                      <div>
                        <div className="mb-1 text-sm font-medium">Command</div>
                        <div className="p-2 bg-muted/30 rounded-md">
                          {resource.execution_config.command ? resource.execution_config.command.join(' ') : "-"}
                        </div>
                      </div>
                      <div>
                        <div className="mb-1 text-sm font-medium">Arguments</div>
                        <div className="p-2 bg-muted/30 rounded-md">
                          {resource.execution_config.args ? resource.execution_config.args.join(', ') : "-"}
                        </div>
                      </div>
                    </>
                  ) : (
                    <div className="text-sm text-muted-foreground">No main container configuration</div>
                  )}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="environment" className="pt-4 space-y-6">
              <div>
                <h3 className="text-lg font-medium mb-3">Environment Variables</h3>
                {resource.execution_config?.environment_variables &&
                resource.execution_config.environment_variables.length > 0 ? (
                    <div className="overflow-x-auto">
                      <table className="min-w-full border border-muted rounded-md">
                        <thead className="bg-muted/30">
                          <tr>
                            <th className="text-left px-6 py-3 text-sm">Name</th>
                            <th className="text-left px-6 py-3 text-sm">Value</th>
                          </tr>
                        </thead>
                        <tbody>
                          {resource.execution_config.environment_variables.map((env, idx) => (
                            <tr key={idx} className="border-t border-muted">
                              <td className="px-6 py-2 text-sm">{env.name}</td>
                              <td className="px-6 py-2 text-sm font-mono">{env.value}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="text-sm text-muted-foreground">No environment variables configured</div>
                  )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

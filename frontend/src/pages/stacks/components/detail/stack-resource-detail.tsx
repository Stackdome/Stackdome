import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { useToast } from "@/components/ui/use-toast";
import { Box, GitBranch, ExternalLink, Copy, Pencil } from "lucide-react";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { z } from "zod";
import type { ApiStackResourceStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { variantFromState } from "@/components/branded";

interface StackResourceDetailProps {
  resource: Partial<FormStackResourceData>;
  index: number;
  isSessionActive?: boolean;
  isDirty?: boolean;
  dirtyCount?: number;
  onEdit?: () => void;
  onDiscard?: () => void;
  /** Map of `${resourceIdx}::${envName}` → provenance for converted env rows. */
  detachedProvenance?: Map<string, { addonName: string; credField?: string }>;
  /** addonId → display name, for read-mode addon group headers. */
  addonNameById?: Map<string, string>;
}

export default function StackResourceDetail({
  resource,
  index,
  isSessionActive = false,
  isDirty = false,
  dirtyCount = 0,
  onEdit,
  onDiscard,
  detachedProvenance,
  addonNameById,
}: StackResourceDetailProps) {
  const { toast } = useToast();
  const statusObj = (resource.status ?? {}) as z.infer<typeof ApiStackResourceStatusSchema>;
  const status = statusObj.state?.toLowerCase() || 'pending';
  const statusVariant = variantFromState(statusObj.state);
  const statusDotColor =
    statusVariant === "ready" ? "bg-success"
      : statusVariant === "error" ? "bg-danger"
        : statusVariant === "pending" ? "bg-warn"
          : "bg-muted-foreground";

  const ensureAbsoluteUrl = (url: string): string => {
    if (!url) return url;
    if (url.startsWith('http://') || url.startsWith('https://')) {
      return url;
    }
    // If URL starts with '//', assume it's protocol-relative
    if (url.startsWith('//')) {
      return `https:${url}`;
    }
    // If URL looks like a domain, add https://
    if (url.includes('.') && !url.startsWith('/')) {
      return `https://${url}`;
    }
    return url;
  };

  const publicUrls = (statusObj.public_ingress || [])
    .map((ingress: { url?: string }) => ingress.url)
    .filter((url: string | undefined): url is string => url !== undefined && url.trim() !== '')
    .map(ensureAbsoluteUrl);

  const firstPublicUrl = publicUrls[0];

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({
        description: "Copied to clipboard",
      });
    } catch (err) {
      console.error('Failed to copy text: ', err);
      toast({
        description: "Failed to copy to clipboard",
        variant: "destructive",
      });
    }
  };

  const getUrlForPort = (portNumber: number): string | undefined => {
    const url = (statusObj.public_ingress || [])
      .find((ingress: { target_port?: number; url?: string }) =>
        ingress.target_port === portNumber
      )?.url;
    return url ? ensureAbsoluteUrl(url) : undefined;
  };

  return (
    <AccordionItem value={String(index)} className="border-t border-border first:border-t-0">
      <AccordionTrigger
        className="group/row px-4 py-3 hover:bg-muted/40 data-[state=open]:bg-muted/30 rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-3 text-left flex-grow">
          <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
              <span className={`h-2 w-2 rounded-full shrink-0 ${statusDotColor}`}></span>
            </TooltipTrigger>
            <TooltipContent side="top">
              <p className="capitalize">{status}</p>
            </TooltipContent>
          </Tooltip>
          <div className="flex flex-col flex-grow min-w-0">
            <span className="font-medium flex items-center gap-2">
              {resource.name || `Resource ${index + 1}`}
              {firstPublicUrl && (
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <a
                      href={firstPublicUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-muted-foreground hover:text-foreground transition-colors"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  </TooltipTrigger>
                  <TooltipContent side="top">
                    <p>Open URL: {firstPublicUrl}</p>
                  </TooltipContent>
                </Tooltip>
              )}
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
          <div className="flex items-center gap-3 shrink-0 mr-2">
            {!isSessionActive && onEdit && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-xs text-muted-foreground border border-border opacity-0 group-hover/row:opacity-100 focus:opacity-100 transition-opacity"
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit();
                }}
              >
                <Pencil className="h-3 w-3" />
                Edit
              </Button>
            )}
            {isSessionActive && isDirty && (
              <>
                <span className="text-[11.5px] font-medium text-foreground bg-brand-bg px-2 py-0.5 rounded-sm">
                  {dirtyCount} changed
                </span>
                {onDiscard && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-xs hover:bg-danger-bg hover:text-danger hover:border-danger border border-transparent"
                    onClick={(e) => {
                      e.stopPropagation();
                      onDiscard();
                    }}
                  >
                    Discard
                  </Button>
                )}
              </>
            )}
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="bg-background dark:bg-secondary border-t border-border pb-4 pt-4 px-1">
        <div className="px-4 space-y-4">
          <Tabs defaultValue="configuration" className="w-full">
            <TabsList className="w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-1 px-2">
              <TabsTrigger value="configuration" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">Configuration</TabsTrigger>
              <TabsTrigger value="deployment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">Deployment</TabsTrigger>
              <TabsTrigger value="environment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">Environment</TabsTrigger>
            </TabsList>
            <TabsContent value="configuration" className="pt-4 space-y-6">
              <div>
                <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">General</h3>
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

              {/* Ingress Section */}
              <div>
                <h3 className="text-sm font-semibold text-foreground mb-3">Ingress</h3>
                {resource.ports && resource.ports.length > 0 ? (
                  <div className="space-y-2">
                    {resource.ports.map((port, pidx) => {
                      const portUrl = getUrlForPort(port.number);
                      return (
                        <div key={pidx} className="rounded-md border border-border bg-muted/10 p-3">
                          <div className="flex items-center gap-2">
                            <ExternalLink className="h-3.5 w-3.5 text-brand shrink-0" />
                            {portUrl ? (
                              <>
                                <a
                                  href={portUrl}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="font-mono text-[13px] text-brand hover:text-brand-press hover:underline truncate min-w-0 flex-1"
                                  title={portUrl}
                                >
                                  {portUrl}
                                </a>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="shrink-0 h-6 w-6 text-muted-foreground hover:text-foreground"
                                  onClick={() => copyToClipboard(portUrl)}
                                  title="Copy URL"
                                >
                                  <Copy className="h-3 w-3" />
                                </Button>
                              </>
                            ) : (
                              <span className="font-mono text-[13px] text-muted-foreground italic">
                                {port.exposed_to_public ? "Pending…" : "Not exposed"}
                              </span>
                            )}
                          </div>
                          <div className="font-mono text-[12px] text-muted-foreground pl-5 mt-1">
                            port {port.number}
                            <span className="mx-1.5 text-muted-foreground/60">·</span>
                            {port.protocol || "tcp"}
                            <span className="mx-1.5 text-muted-foreground/60">·</span>
                            {port.exposed_to_public ? "exposed" : "internal only"}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">No ports configured</p>
                )}
              </div>
              <Separator className="my-4" />

              {/* Volume Mounts Section */}
              <div>
                <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Volume Mounts</h3>
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
            </TabsContent>

            <TabsContent value="deployment" className="pt-4 space-y-6">
              {/* Pre-deploy Section (Init) */}
              <div>
                <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Pre-deployment step</h3>
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
                <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Main container step</h3>
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
                <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Environment variables</h3>
                {(() => {
                  const envs = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];
                  if (envs.length === 0) {
                    return (
                      <div className="text-sm text-muted-foreground">No environment variables configured</div>
                    );
                  }

                  // Render the read-only (masked) value for a single env row.
                  // Reference values are masked; literal stack values are shown.
                  const renderValue = (env: FormEnvVarData) => {
                    switch (env.from) {
                      case "stack":
                        return <span className="text-foreground">{env.value}</span>;
                      case "secret":
                        return (
                          <span className="text-muted-foreground">
                            ●●●●●●●● · {env.secretKey}
                          </span>
                        );
                      case "addon":
                        return (
                          <span className="text-muted-foreground">
                            ●●● · {env.credField}
                          </span>
                        );
                      case "resource":
                        return (
                          <span className="text-foreground">
                            {env.resourceName} · {env.output}
                          </span>
                        );
                      case "self":
                        return <span className="text-foreground">{env.selfOutput}</span>;
                      default:
                        return null;
                    }
                  };

                  // Compact "From" source pill label per arm.
                  const fromPillLabel = (env: FormEnvVarData): string => {
                    switch (env.from) {
                      case "stack":
                        return "Plain text";
                      case "secret":
                        return "Secret";
                      case "resource":
                        return "Resource";
                      case "self":
                        return "Self";
                      case "addon":
                        return `Addon · ${env.credField}`;
                      default:
                        return "";
                    }
                  };

                  const renderRow = ({ env, idx }: { env: FormEnvVarData; idx: number }) => {
                    const prov = env.name
                      ? detachedProvenance?.get(`${index}::${env.name}`)
                      : undefined;
                    return (
                      <div
                        key={idx}
                        className="grid grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_auto] gap-4 items-center px-3 py-2 border-t border-border first:border-t-0"
                      >
                        <div className="text-sm font-mono truncate text-foreground">{env.name}</div>
                        <div className="text-sm font-mono truncate">
                          {renderValue(env)}
                          {prov && (
                            <div className="mt-1 text-[11px] font-sans text-muted-foreground not-italic">
                              ↶ was bound to {prov.addonName}
                              {prov.credField ? `.${prov.credField}` : ""}
                            </div>
                          )}
                        </div>
                        <span className="inline-flex items-center gap-1 w-fit text-[12px] text-muted-foreground bg-secondary border border-border rounded px-2 py-[3px] whitespace-nowrap">
                          {fromPillLabel(env)}
                        </span>
                      </div>
                    );
                  };

                  // Partition rows so all ungrouped (non-addon) rows render
                  // FIRST (original order, in a single bordered container), then
                  // addon groups keyed by (addonId, database) render LAST (in
                  // first-appearance order), mirroring edit mode.
                  type AddonGroup = {
                    addonId: string;
                    database: string;
                    items: { env: FormEnvVarData; idx: number }[];
                  };
                  const plainItems: { env: FormEnvVarData; idx: number }[] = [];
                  const addonGroups: AddonGroup[] = [];
                  const addonGroupByKey = new Map<string, AddonGroup>();
                  envs.forEach((env, idx) => {
                    if (env.from === "addon") {
                      const aid = env.addonId || "";
                      const db = env.database || "";
                      const key = `${aid}|${db}`;
                      let g = addonGroupByKey.get(key);
                      if (!g) {
                        g = { addonId: aid, database: db, items: [] };
                        addonGroupByKey.set(key, g);
                        addonGroups.push(g);
                      }
                      g.items.push({ env, idx });
                    } else {
                      plainItems.push({ env, idx });
                    }
                  });

                  return (
                    <div className="rounded-md border border-border bg-background">
                      {plainItems.map(renderRow)}
                      {addonGroups.map((g, gIdx) => {
                        const name = g.addonId
                          ? (addonNameById?.get(g.addonId) ?? g.addonId)
                          : g.addonId;
                        return (
                          <div
                            key={`a-${gIdx}-${g.addonId}-${g.database}`}
                            className="relative border border-dashed border-border-strong rounded mx-3 mt-3.5 mb-1.5 px-2.5 pt-2 pb-1"
                            data-testid="env-addon-group"
                          >
                            <div className="absolute -top-2.5 left-3 inline-flex items-center gap-1.5 bg-background px-2 text-[11.5px]">
                              <AddonTypeIcon type="postgres" size={14} />
                              <span className="font-medium text-foreground truncate">{name}</span>
                              <span className="text-muted-foreground">· Postgres</span>
                              {g.database && (
                                <span className="text-muted-foreground">· db: {g.database}</span>
                              )}
                            </div>
                            {g.items.map(renderRow)}
                          </div>
                        );
                      })}
                    </div>
                  );
                })()}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

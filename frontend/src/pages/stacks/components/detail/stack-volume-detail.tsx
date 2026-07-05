import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { HardDrive, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData   } from "@/pages/stacks/schemas/form-schema";
import type { z } from "zod";
import { ApiVolumeStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { volumeDotClass } from "@/pages/stacks/lib/volume-status";

interface StackVolumeDetailProps {
  volume: Partial<VolumeFormData>;
  index: number;
  allStackResources?: Partial<FormStackResourceData>[];
  isSessionActive?: boolean;
  isDirty?: boolean;
  dirtyCount?: number;
  onEdit?: () => void;
  onDiscard?: () => void;
}

export default function StackVolumeDetail({
  volume,
  index,
  allStackResources = [],
  isSessionActive = false,
  isDirty = false,
  dirtyCount = 0,
  onEdit,
  onDiscard,
}: StackVolumeDetailProps) {
  // Determine status color based on volume.status.phase
  const statusObj = (volume.status ?? {}) as z.infer<typeof ApiVolumeStatusSchema>;
  const status = statusObj.phase?.toLowerCase() || 'pending';
  const statusColor = volumeDotClass(statusObj.phase);

  // Helper to find resources that mount this volume
  const mountingInfo = volume.name
    ? allStackResources
      .map(resource => {
        if (!resource.name || !resource.volume_mounts) return null;
        const mountDetail = resource.volume_mounts.find(
          vm => vm.source_volume_name === volume.name
        );
        return mountDetail ? { resourceName: resource.name, targetPath: mountDetail.target_path } : null;
      })
      .filter(Boolean) as { resourceName: string; targetPath: string }[]
    : [];

  return (
    <AccordionItem value={String(index)} className="border-t border-border first:border-t-0">
      <AccordionTrigger
        className="group/row px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-2 text-left flex-grow">
          <div className="flex flex-col flex-grow min-w-0">
            <span className="font-medium flex items-center gap-2">
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <span className={`h-2 w-2 rounded-full ${statusColor}`}></span>
                </TooltipTrigger>
                <TooltipContent side="top">
                  <p className="capitalize">{status}</p>
                </TooltipContent>
              </Tooltip>
              {volume.name || `Volume ${index + 1}`}
            </span>
            <span className="text-sm text-muted-foreground truncate">
              <span className="flex items-center gap-1.5">
                <HardDrive className="h-3.5 w-3.5" />
                <span>{volume.spec?.size || "Not specified"}</span>
              </span>
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
          {/* Basic info section */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="mb-1 text-sm font-medium">Name</div>
              <div className="p-2 bg-muted/30 rounded-md">
                {volume.name || "Not specified"}
              </div>
            </div>

            <div className="space-y-2">
              <div className="mb-1 text-sm font-medium">Size</div>
              <div className="p-2 bg-muted/30 rounded-md">
                {volume.spec?.size || "Not specified"}
              </div>
              <p className="text-xs text-muted-foreground">Volume size (e.g., 1Gi, 500Mi)</p>
            </div>
          </div>

          {/* Access Mode */}
          <div className="space-y-2">
            <div className="mb-1 text-sm font-medium">Access Mode</div>
            <div className="p-2 bg-muted/30 rounded-md">
              ReadWriteOnce (RWO)
            </div>
            <p className="text-xs text-muted-foreground">ReadWriteOnce: Can be mounted by a single resource for read/write.</p>
          </div>

          {/* Mount Details Section */}
          {mountingInfo.length > 0 && (
            <div className="pt-4 border-t">
              <h3 className="text-base font-semibold mb-2 text-foreground">Mount Details</h3>
              <div className="space-y-1">
                {mountingInfo.map((mount, mountIdx) => (
                  <div key={mountIdx} className="text-sm text-muted-foreground">
                    Mounted by <span className="font-medium text-foreground">{mount.resourceName}</span> at path: <code className="text-xs bg-muted text-muted-foreground p-1 rounded">{mount.targetPath}</code>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Source configuration, if any */}
          {volume.sourceType && volume.sourceType !== "None" && (
            <div className="pt-4 border-t">
              <h3 className="text-base font-semibold mb-2 text-foreground">Volume Source</h3>
              <div className="mb-1 text-sm font-medium">Source Type</div>
              <div className="p-2 bg-muted/30 rounded-md">
                {volume.sourceType}
              </div>

              {/* Show relevant source details based on type */}
              {volume.sourceType === "GitRepo" && volume.spec?.source?.git_repo_source && (
                <>
                  <div className="mt-2 mb-1 text-sm font-medium">Repository URL</div>
                  <div className="p-2 bg-muted/30 rounded-md">
                    {volume.spec.source.git_repo_source.repo_url}
                  </div>
                  <div className="mt-2 mb-1 text-sm font-medium">Repository Revision</div>
                  <div className="p-2 bg-muted/30 rounded-md">
                    {volume.spec.source.git_repo_source.revision?.branch &&
                      `Branch: ${volume.spec.source.git_repo_source.revision.branch}`}
                    {volume.spec.source.git_repo_source.revision?.commit &&
                      `Commit: ${volume.spec.source.git_repo_source.revision.commit}`}
                    {volume.spec.source.git_repo_source.revision?.tag &&
                      `Tag: ${volume.spec.source.git_repo_source.revision.tag}`}
                    {!volume.spec.source.git_repo_source.revision?.branch &&
                     !volume.spec.source.git_repo_source.revision?.commit &&
                     !volume.spec.source.git_repo_source.revision?.tag && "main"}
                  </div>
                </>
              )}

              {volume.sourceType === "RemoteDir" && volume.spec?.source?.remote_source && (
                <>
                  <div className="mt-2 mb-1 text-sm font-medium">Remote Path</div>
                  <div className="p-2 bg-muted/30 rounded-md">
                    {volume.spec.source.remote_source.path || "Not specified"}
                  </div>
                </>
              )}

              {volume.sourceType === "BuildArtifact" && volume.spec?.source?.build_source && (
                <>
                  <div className="mt-2 mb-1 text-sm font-medium">Build Artifacts</div>
                  <div className="p-2 bg-muted/30 rounded-md">
                    {Array.isArray(volume.spec.source.build_source)
                      ? volume.spec.source.build_source.map(item =>
                        item.resource_ref).join(", ")
                      : "No build artifacts specified"}
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

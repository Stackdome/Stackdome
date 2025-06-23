import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { HardDrive, Trash2 } from "lucide-react";
import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData   } from "@/pages/stacks/schemas/form-schema";
import type { z } from "zod";
import { ApiVolumeStatusSchema } from "@/pages/stacks/schemas/api-schema";

interface StackVolumeItemProps {
  volume: Partial<VolumeFormData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  onChange: (index: number, updatedVolume: Partial<VolumeFormData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  allVolumes: Partial<VolumeFormData>[];
  allStackResources?: Partial<FormStackResourceData>[];
}

export default function StackVolumeItem({
  volume,
  index,
  itemRef,
  onChange,
  onRemove,
  errors,
  allVolumes,
  allStackResources = [],
}: StackVolumeItemProps) {
  // Helper for updating volume fields
  const update = (patch: Partial<VolumeFormData>) => {
    onChange(index, { ...volume, ...patch });
  };

  // Check for duplicate name
  const isDuplicate = allVolumes.filter((v) => v.name?.length && v.name === volume.name).length > 1;

  // Determine status color based on volume.status.phase
  const statusObj = (volume.status ?? {}) as z.infer<typeof ApiVolumeStatusSchema>;
  const status = statusObj.phase?.toLowerCase() || 'pending';
  let statusColor = 'bg-yellow-500';

  if (status === 'ready') {
    statusColor = 'bg-green-500';
  } else if (status === 'failed') {
    statusColor = 'bg-red-500';
  }

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
            {errors._form && (
              <span className="text-xs text-destructive mt-0.5 pl-6">{errors._form}</span>
            )}
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="pb-4 pt-2">
        <div className="px-4 space-y-4">
          {/* Basic info section */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor={`volume-name-${index}`}>
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id={`volume-name-${index}`}
                placeholder="Volume name"
                value={volume.name || ""}
                onChange={(e) => update({ name: e.target.value })}
                className={errors.name || isDuplicate ? "border-destructive" : ""}
                aria-invalid={!!errors.name || isDuplicate}
              />
              {(errors.name || isDuplicate) && (
                <p className="text-sm text-destructive">
                  {errors.name || (isDuplicate ? "Volume name must be unique" : "")}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor={`volume-size-${index}`}>
                Size <span className="text-destructive">*</span>
              </Label>
              <Input
                id={`volume-size-${index}`}
                placeholder="e.g., 1Gi, 500Mi"
                value={volume.spec?.size || ""}
                onChange={(e) => update({
                  spec: {
                    ...volume.spec,
                    size: e.target.value,
                    needs_sync_before_use: volume.spec?.needs_sync_before_use ?? false,
                    access_mode: volume.spec?.access_mode ?? "ReadWriteOnce",
                  }
                })}
                className={errors["spec.size"] ? "border-destructive" : ""}
                aria-invalid={!!errors["spec.size"]}
              />
              {errors["spec.size"] && <p className="text-sm text-destructive">{errors["spec.size"]}</p>}
              <p className="text-xs text-muted-foreground">Volume size (e.g., 1Gi, 500Mi)</p>
            </div>
          </div>

          {/* Access Mode (RWO only, disabled) */}
          <div className="space-y-2">
            <Label>Access Mode</Label>
            <Input value="ReadWriteOnce (RWO)" disabled className="bg-muted" />
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

          {/* Remove button at the end, styled like stack resource item */}
          <div className="flex justify-center items-center mt-8">
            <span className="flex items-center justify-center w-full py-3 rounded-md bg-muted/70">
              <Button
                type="button"
                variant="ghost"
                className="text-destructive hover:text-destructive hover:bg-destructive/10 focus-visible:bg-destructive/10"
                onClick={() => onRemove(index)}
                title="Remove volume"
              >
                <Trash2 className="h-4 w-4 mr-1" />
                Remove Volume
              </Button>
            </span>
          </div>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

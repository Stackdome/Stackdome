import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { HardDrive } from "lucide-react";
import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData } from "@/pages/stacks/schemas/form-schema";
import type { z } from "zod";
import { ApiVolumeStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { VolumeFields, volumeMountingInfo } from "./volume-fields";

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
  const statusObj = (volume.status ?? {}) as z.infer<typeof ApiVolumeStatusSchema>;
  const status = statusObj.phase?.toLowerCase() || 'pending';
  let statusColor = 'bg-warn';
  if (status === 'ready') {
    statusColor = 'bg-success';
  } else if (status === 'failed') {
    statusColor = 'bg-danger';
  }

  const mountingInfo = volumeMountingInfo(volume, allStackResources);
  const accessMode = volume.spec?.access_mode || "ReadWriteOnce";
  const mountedBy = mountingInfo.map(m => m.resourceName).join(", ");

  return (
    <AccordionItem value={String(index)} className="border-t border-border first:border-t-0">
      <AccordionTrigger
        ref={itemRef}
        className="px-4 py-3 hover:bg-muted/40 data-[state=open]:bg-muted/30 rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-3 text-left flex-grow">
          <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
              <span className={`h-2 w-2 rounded-full shrink-0 ${statusColor}`}></span>
            </TooltipTrigger>
            <TooltipContent side="top">
              <p className="capitalize">{status}</p>
            </TooltipContent>
          </Tooltip>
          <div className="flex flex-col flex-grow min-w-0">
            <span className="font-medium flex items-center gap-2">
              {volume.name || `Volume ${index + 1}`}
            </span>
            <span className="text-sm text-muted-foreground truncate">
              <span className="flex items-center gap-1.5">
                <HardDrive className="h-3.5 w-3.5" />
                <span>
                  {volume.spec?.size || "Not specified"}
                  <span className="mx-1.5 text-muted-foreground/60">·</span>
                  {accessMode}
                  {mountedBy && (
                    <>
                      <span className="mx-1.5 text-muted-foreground/60">·</span>
                      mounted by {mountedBy}
                    </>
                  )}
                </span>
              </span>
            </span>
            {errors._form && (
              <span className="text-xs text-danger mt-0.5">{errors._form}</span>
            )}
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="bg-background dark:bg-secondary border-t border-border pb-4 pt-4 px-1">
        <div className="px-4 space-y-6">
          <VolumeFields
            volume={volume}
            index={index}
            onChange={onChange}
            onRemove={onRemove}
            errors={errors}
            allVolumes={allVolumes}
            allStackResources={allStackResources}
          />
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

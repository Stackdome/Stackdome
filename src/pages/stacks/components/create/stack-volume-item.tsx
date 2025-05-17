import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Trash2 } from "lucide-react";
import type { VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";

interface StackVolumeItemProps {
  volume: Partial<VolumeFormData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  isOnlyVolume: boolean;
  onChange: (index: number, updatedVolume: Partial<VolumeFormData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
}

export default function StackVolumeItem({
  volume,
  index,
  itemRef,
  onChange,
  onRemove,
  errors,
}: StackVolumeItemProps) {
  // Helper for updating volume fields
  const update = (patch: Partial<VolumeFormData>) => {
    onChange(index, { ...volume, ...patch });
  };


  return (
    <AccordionItem value={String(index)} className="border-0">
      <AccordionTrigger
        ref={itemRef}
        className="px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-2 text-left">
          <div>
            <span className="font-medium">{volume.name || `Volume ${index + 1}`}</span>
            {volume.spec?.size && (
              <span className="ml-2 text-sm text-muted-foreground">({volume.spec.size})</span>
            )}
            {errors._form && (
              <div className="text-sm text-destructive mt-1">{errors._form}</div>
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
                className={errors.name ? "border-destructive" : ""}
                aria-invalid={!!errors.name}
              />
              {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
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
            <p className="text-xs text-muted-foreground">ReadWriteOnce: Can be mounted by a single node for read/write.</p>
          </div>

          {/* Remove button always visible at the bottom */}
          <div className="pt-4 border-t">
            <Button
              type="button"
              variant="ghost"
              className="text-destructive hover:text-destructive hover:bg-destructive/10"
              onClick={() => onRemove(index)}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              Remove Volume
            </Button>
          </div>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData as StackResourceData } from "@/pages/stacks/schemas/form-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackVolumeItem from "@/pages/stacks/components/shared/stack-volume-item";
import { Database } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";

interface StackVolumesFormProps {
  volumes: Partial<VolumeFormData>[];
  onVolumesChange: (updatedVolumes: Partial<VolumeFormData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  stackResources?: Partial<StackResourceData>[];
  accordionDefaultOpen?: boolean; // If false, all collapsed by default
}

function getDefaultVolume(): Partial<VolumeFormData> {
  return {
    name: "",
    sourceType: "None",
    labels: [],
    spec: {
      size: "1Gi",
      access_mode: "ReadWriteOnce",
      needs_sync_before_use: false
    }
  };
}

export default function StackVolumesForm({
  volumes,
  onVolumesChange,
  errors,
  stackResources = [],
  accordionDefaultOpen = true,
}: StackVolumesFormProps) {
  const [pendingRemoveIdx, setPendingRemoveIdx] = useState<number | null>(null);

  // Helper function to check if a volume is being used by resources
  const isVolumeUsed = (volumeName: string) => {
    return stackResources.some(res =>
      res.volume_mounts?.some(vm => vm.source_volume_name === volumeName)
    );
  };

  // Helper function to check if a volume has data
  const isVolumeFilled = (vol: Partial<VolumeFormData>) => {
    return !!(vol.name || vol.spec?.size || vol.labels?.length || vol.spec?.needs_sync_before_use || vol.spec?.access_mode !== undefined);
  };

  return (
    <div className="space-y-3">
      <ResourceFormList<Partial<VolumeFormData>>
        renderItem={({ item, index, itemRef, onChange, onRemove, errors }) => (
          <StackVolumeItem
            volume={item}
            index={index}
            itemRef={itemRef}
            onChange={onChange}
            onRemove={(idx) => {
              if (item.name && isVolumeUsed(item.name)) {
                // Show error dialog
                console.error(`Volume ${item.name} is in use and cannot be removed.`);
                return;
              }

              if (isVolumeFilled(item)) {
                setPendingRemoveIdx(idx);
              } else {
                onRemove(idx);
              }
            }}
            errors={errors}
            allVolumes={volumes}
            allStackResources={stackResources}
          />
        )}
        items={volumes}
        onItemsChange={onVolumesChange}
        createDefaultItem={getDefaultVolume}
        errors={errors}
        emptyText="No volumes added."
        emptyIcon={<Database className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
      <div className="flex justify-center mt-4">
        <Button
          type="button"
          variant="ghost"
          onClick={() => onVolumesChange([...volumes, getDefaultVolume()])}
        >
          + Add Volume
        </Button>
      </div>
      <Dialog open={pendingRemoveIdx !== null} onOpenChange={open => !open && setPendingRemoveIdx(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Volume?</DialogTitle>
          </DialogHeader>
          <div>Are you sure you want to remove this volume? This action cannot be undone.</div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setPendingRemoveIdx(null)}>Cancel</Button>
            <Button variant="destructive" onClick={() => {
              if (pendingRemoveIdx !== null) {
                onVolumesChange(volumes.filter((_, i) => i !== pendingRemoveIdx));
                setPendingRemoveIdx(null);
              }
            }}>Remove Volume</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

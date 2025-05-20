import type { VolumeFormData, StackResourceData } from "@/pages/stacks/schemas/stack-create-schema";
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
  readOnly?: boolean;
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
  readOnly = false,
  accordionDefaultOpen = true,
}: StackVolumesFormProps) {
  const [pendingRemoveIdx, setPendingRemoveIdx] = useState<number | null>(null);

  const isVolumeFilled = (vol: Partial<VolumeFormData>) => {
    return !!(vol.name || vol.spec?.size || vol.labels?.length || vol.spec?.needs_sync_before_use || vol.spec?.access_mode !== undefined);
  };

  const handleRemove = (idx: number) => {
    if (isVolumeFilled(volumes[idx])) {
      setPendingRemoveIdx(idx);
    } else {
      onVolumesChange(volumes.filter((_, i) => i !== idx));
    }
  };

  const createDefaultVolumeWithWorkspace = () => {
    return getDefaultVolume();
  };

  return (
    <div>
      <ResourceFormList<VolumeFormData>
        items={volumes}
        onItemsChange={onVolumesChange}
        errors={errors}
        createDefaultItem={createDefaultVolumeWithWorkspace}
        renderItem={({ item, index, itemRef, onChange, errors }) => (
          <StackVolumeItem
            key={index}
            volume={item}
            index={index}
            itemRef={itemRef}
            onChange={onChange}
            onRemove={() => handleRemove(index)}
            errors={errors}
            allVolumes={volumes}
            allStackResources={stackResources}
            readOnly={readOnly}
          />
        )}
        emptyText="No volumes added."
        emptyIcon={<Database className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        readOnly={readOnly}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
      {/* Only show add button if not readOnly */}
      {!readOnly && (
        <div className="flex justify-center mt-4">
          <Button
            type="button"
            variant="ghost"
            onClick={() => onVolumesChange([...volumes, getDefaultVolume()])}
            disabled={readOnly}
          >
            + Add Volume
          </Button>
        </div>
      )}
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

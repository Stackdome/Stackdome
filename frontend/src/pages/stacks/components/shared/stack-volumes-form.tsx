import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData   } from "@/pages/stacks/schemas/form-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackVolumeItem from "@/pages/stacks/components/shared/stack-volume-item";
import { Database, PlusCircle } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";

interface StackVolumesFormProps {
  volumes: Partial<VolumeFormData>[];
  onVolumesChange: (updatedVolumes: Partial<VolumeFormData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  stackResources?: Partial<FormStackResourceData>[];
  /** Baseline snapshot — when provided, removing a volume that exists in
   *  baseline shows a confirm dialog. Newly-added drafts are removed silently. */
  baselineVolumes?: Partial<VolumeFormData>[];
  accordionDefaultOpen?: boolean; // If false, all collapsed by default
  defaultOpenVolumeIdx?: number | null;
}

export function getDefaultVolume(): Partial<VolumeFormData> {
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
  baselineVolumes,
  accordionDefaultOpen = true,
  defaultOpenVolumeIdx = null,
}: StackVolumesFormProps) {
  const [pendingRemoveIdx, setPendingRemoveIdx] = useState<number | null>(null);

  // Removing a volume is destructive only when it exists in the baseline —
  // i.e., it was already deployed. Newly-added drafts are removed silently.
  const wasInBaseline = (vol: Partial<VolumeFormData>) =>
    !!(vol.name && baselineVolumes?.some(b => b.name === vol.name));

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
              if (wasInBaseline(item)) {
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
        emptyTitle="No volumes added"
        emptyOptional
        emptyDescription="Add a persistent volume if your services need to keep data across restarts. Otherwise, you can skip this."
        emptyCtaLabel="Add volume"
        emptyOnAdd={() => onVolumesChange([...volumes, getDefaultVolume()])}
        emptyIcon={<Database className="h-6 w-6" />}
        defaultAllCollapsed={!accordionDefaultOpen}
        defaultOpenIndex={defaultOpenVolumeIdx}
      />
      {volumes.length > 0 && (
        <div className="flex justify-center mt-4">
          <Button
            type="button"
            variant="ghost"
            onClick={() => onVolumesChange([...volumes, getDefaultVolume()])}
          >
            <PlusCircle className="h-4 w-4" />
            Add Volume
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

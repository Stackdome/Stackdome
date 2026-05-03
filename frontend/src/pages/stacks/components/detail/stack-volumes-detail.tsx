import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import ResourceDetailList from "@/pages/stacks/components/detail/resource-detail-list";
import StackVolumeDetail from "@/pages/stacks/components/detail/stack-volume-detail";
import { Database } from "lucide-react";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";

interface StackVolumesDetailProps {
  volumes: Partial<VolumeFormData>[];
  stackResources?: Partial<FormStackResourceData>[];
  accordionDefaultOpen?: boolean;
  isSessionActive?: boolean;
  dirty?: StackDiff;
  onEditVolume?: (idx: number) => void;
  onDiscardVolume?: (idx: number) => void;
}

export default function StackVolumesDetail({
  volumes,
  stackResources = [],
  accordionDefaultOpen = true,
  isSessionActive = false,
  dirty,
  onEditVolume,
  onDiscardVolume,
}: StackVolumesDetailProps) {
  return (
    <div>
      <ResourceDetailList<Partial<VolumeFormData>>
        items={volumes}
        renderItem={({ item, index }) => {
          const isDirty = dirty?.dirtyVolumeIdx.has(index) ?? false;
          const perV = dirty?.perVolumeDirty.get(index);
          const dirtyCount = perV?.fieldsChanged || 0;
          return (
            <StackVolumeDetail
              key={index}
              volume={item}
              index={index}
              allStackResources={stackResources}
              isSessionActive={isSessionActive}
              isDirty={isDirty}
              dirtyCount={dirtyCount}
              onEdit={onEditVolume ? () => onEditVolume(index) : undefined}
              onDiscard={onDiscardVolume ? () => onDiscardVolume(index) : undefined}
            />
          );
        }}
        emptyText="No volumes added."
        emptyIcon={<Database className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
    </div>
  );
}

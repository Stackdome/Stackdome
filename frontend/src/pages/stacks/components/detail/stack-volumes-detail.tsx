import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import ResourceDetailList from "@/pages/stacks/components/detail/resource-detail-list";
import StackVolumeDetail from "@/pages/stacks/components/detail/stack-volume-detail";
import { Database } from "lucide-react";

interface StackVolumesDetailProps {
  volumes: Partial<VolumeFormData>[];
  stackResources?: Partial<FormStackResourceData>[];
  accordionDefaultOpen?: boolean;
}

export default function StackVolumesDetail({
  volumes,
  stackResources = [],
  accordionDefaultOpen = true,
}: StackVolumesDetailProps) {
  return (
    <div>
      <ResourceDetailList<Partial<VolumeFormData>>
        items={volumes}
        renderItem={({ item, index }) => (
          <StackVolumeDetail
            key={index}
            volume={item}
            index={index}
            allStackResources={stackResources}
          />
        )}
        emptyText="No volumes added."
        emptyIcon={<Database className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
    </div>
  );
}

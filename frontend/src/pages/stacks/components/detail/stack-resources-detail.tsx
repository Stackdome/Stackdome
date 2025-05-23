import type { FormStackResourceData   } from "@/pages/stacks/schemas/form-schema";
import ResourceDetailList from "@/pages/stacks/components/detail/resource-detail-list";
import StackResourceDetail from "@/pages/stacks/components/detail/stack-resource-detail";
import { Container } from "lucide-react";

interface StackResourcesDetailProps {
  resources: Partial<FormStackResourceData>[];
  accordionDefaultOpen?: boolean;
}

export default function StackResourcesDetail({
  resources,
  accordionDefaultOpen = true,
}: StackResourcesDetailProps) {
  return (
    <div>
      <ResourceDetailList<Partial<FormStackResourceData>>
        items={resources}
        renderItem={({ item, index }) => (
          <StackResourceDetail
            key={index}
            resource={item}
            index={index}
          />
        )}
        emptyText="No Resources added."
        emptyIcon={<Container className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
    </div>
  );
}

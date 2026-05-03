import type { FormStackResourceData   } from "@/pages/stacks/schemas/form-schema";
import ResourceDetailList from "@/pages/stacks/components/detail/resource-detail-list";
import StackResourceDetail from "@/pages/stacks/components/detail/stack-resource-detail";
import { Container } from "lucide-react";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";

interface StackResourcesDetailProps {
  resources: Partial<FormStackResourceData>[];
  accordionDefaultOpen?: boolean;
  isSessionActive?: boolean;
  dirty?: StackDiff;
  onEditResource?: (idx: number) => void;
  onDiscardResource?: (idx: number) => void;
  /** Map of `${resourceIdx}::${envName}` → provenance for converted env rows. */
  detachedProvenance?: Map<string, { addonName: string; credField?: string }>;
}

export default function StackResourcesDetail({
  resources,
  accordionDefaultOpen = true,
  isSessionActive = false,
  dirty,
  onEditResource,
  onDiscardResource,
  detachedProvenance,
}: StackResourcesDetailProps) {
  return (
    <div>
      <ResourceDetailList<Partial<FormStackResourceData>>
        items={resources}
        renderItem={({ item, index }) => {
          const isDirty = dirty?.dirtyResourceIdx.has(index) ?? false;
          const perR = dirty?.perResourceDirty.get(index);
          const dirtyCount = perR?.rowsChanged || perR?.fieldsChanged || 0;
          return (
            <StackResourceDetail
              key={index}
              resource={item}
              index={index}
              isSessionActive={isSessionActive}
              isDirty={isDirty}
              dirtyCount={dirtyCount}
              onEdit={onEditResource ? () => onEditResource(index) : undefined}
              onDiscard={onDiscardResource ? () => onDiscardResource(index) : undefined}
              detachedProvenance={detachedProvenance}
            />
          );
        }}
        emptyText="No Resources added."
        emptyIcon={<Container className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />}
        defaultAllCollapsed={!accordionDefaultOpen}
      />
    </div>
  );
}

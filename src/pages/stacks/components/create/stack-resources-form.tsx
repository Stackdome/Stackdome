import type { StackResourceData, VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackResourceItem from "./stack-resource-item";

// Props interface for StackResourcesForm
interface StackResourcesFormProps {
  resources: Partial<StackResourceData>[];
  onResourcesChange: (updatedResources: Partial<StackResourceData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  volumes?: Partial<VolumeFormData>[];
}  // Helper to create a default new resource, aligning with StackResourceData
function getDefaultResource(): Partial<StackResourceData> {
  return {
    name: "",
    sourceType: "image",
    labels: [],
    depends_on: [],
    ports: [],
    volume_mounts: [],
    execution_config: { environment_variables: [] },
    build_spec: undefined,
    image_spec: { image: "" },
  };
}

export default function StackResourcesForm({
  resources,
  onResourcesChange,
  errors,
  volumes = []
}: StackResourcesFormProps) {
  const allResources = resources.map((r, idx) => ({ name: r.name || `Resource ${idx + 1}`, index: idx }));

  return (
    <ResourceFormList<StackResourceData>
      items={resources}
      onItemsChange={onResourcesChange}
      errors={errors}
      createDefaultItem={getDefaultResource}
      renderItem={({ item, index, itemRef, isOnlyItem, onChange, onRemove, errors }) => (
        <StackResourceItem
          key={index}
          resource={item}
          index={index}
          itemRef={itemRef}
          isOnlyResource={isOnlyItem}
          onChange={onChange}
          onRemove={onRemove}
          errors={errors}
          volumes={volumes}
          allResources={allResources}
        />
      )}
      addButtonText="Add Another Resource"
    />
  );
}

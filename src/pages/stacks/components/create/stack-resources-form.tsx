import type { StackResourceData } from "@/pages/stacks/schemas/stack-create-schema";
import ResourceFormList from "@/components/resource-form-list";
import StackResourceItem from "./stack-resource-item";

// Props interface for StackResourcesForm
interface StackResourcesFormProps {
  resources: Partial<StackResourceData>[];
  onResourcesChange: (updatedResources: Partial<StackResourceData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
}

// Helper to create a default new resource, aligning with StackResourceData
function getDefaultResource(): Partial<StackResourceData> {
  return {
    name: "",
    sourceType: "image",
    labels: [],
    depends_on: [],
    ports: [],
    execution_config: { environment_variables: [] },
    build_spec: undefined,
    image_spec: { image: "" },
  };
}

export default function StackResourcesForm({
  resources,
  onResourcesChange,
  errors,
}: StackResourcesFormProps) {
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
        />
      )}
      addButtonText="Add Another Resource"
    />
  );
}

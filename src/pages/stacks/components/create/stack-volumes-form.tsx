import type { VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackVolumeItem from "./stack-volume-item";

// Props interface for StackVolumesForm
interface StackVolumesFormProps {
  volumes: Partial<VolumeFormData>[];
  onVolumesChange: (updatedVolumes: Partial<VolumeFormData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  workspace?: string;  // For pre-filling workspace_name
}

// Helper to create a default new volume, aligning with VolumeFormData
function getDefaultVolume(workspace?: string): Partial<VolumeFormData> {
  return {
    name: "",
    workspace_name: workspace || "",
    sourceType: "None",
    labels: [],
    spec: {
      size: "1Gi", // Default size
      access_mode: "ReadWriteOnce", // Default access mode
      needs_sync_before_use: false
    }
  };
}

export default function StackVolumesForm({
  volumes,
  onVolumesChange,
  errors,
  workspace
}: StackVolumesFormProps) {
  // Create a default volume with the workspace name
  const createDefaultVolumeWithWorkspace = () => {
    return getDefaultVolume(workspace);
  };

  return (
    <ResourceFormList<VolumeFormData>
      items={volumes}
      onItemsChange={onVolumesChange}
      errors={errors}
      createDefaultItem={createDefaultVolumeWithWorkspace}
      renderItem={({ item, index, itemRef, isOnlyItem, onChange, onRemove, errors }) => (
        <StackVolumeItem
          key={index}
          volume={item}
          index={index}
          itemRef={itemRef}
          isOnlyVolume={isOnlyItem}
          onChange={onChange}
          onRemove={onRemove}
          errors={errors}
        />
      )}
      addButtonText="Add Volume"
    />
  );
}

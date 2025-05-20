import type { StackResourceData, VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackResourceItem from "@/pages/stacks/components/shared/stack-resource-item";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useState } from "react";

interface StackResourcesFormProps {
  resources: Partial<StackResourceData>[];
  onResourcesChange: (updatedResources: Partial<StackResourceData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  volumes?: Partial<VolumeFormData>[];
  readOnly?: boolean;
  addButtonText?: string;
  autoAddFirstItem?: boolean;
}

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
  volumes = [],
  readOnly = false,
  addButtonText = "Add Resource",
  autoAddFirstItem = false
}: StackResourcesFormProps) {
  const [pendingRemoveIdx, setPendingRemoveIdx] = useState<number | null>(null);

  const isResourceFilled = (res: Partial<StackResourceData>) => {
    return !!(res.name || res.ports?.length || res.volume_mounts?.length || res.labels?.length || res.depends_on?.length || res.execution_config?.environment_variables?.length || res.build_spec || (res.image_spec && res.image_spec.image));
  };

  const handleRemove = (idx: number) => {
    if (isResourceFilled(resources[idx])) {
      setPendingRemoveIdx(idx);
    } else {
      onResourcesChange(resources.filter((_, i) => i !== idx));
    }
  };

  const confirmRemove = () => {
    if (pendingRemoveIdx !== null) {
      onResourcesChange(resources.filter((_, i) => i !== pendingRemoveIdx));
      setPendingRemoveIdx(null);
    }
  };

  const allResources = resources.map((r, idx) => ({ name: r.name || `Resource ${idx + 1}`, index: idx }));

  return (
    <>
      <ResourceFormList<StackResourceData>
        items={resources}
        onItemsChange={onResourcesChange}
        errors={errors}
        createDefaultItem={getDefaultResource}
        renderItem={({ item, index, itemRef, onChange, errors }) => (
          <StackResourceItem
            key={index}
            resource={item}
            index={index}
            itemRef={itemRef}
            onChange={onChange}
            errors={errors}
            volumes={volumes}
            allResources={allResources}
            readOnly={readOnly}
            onRemove={handleRemove}
          />
        )}
        addButtonText={addButtonText}
        autoAddFirstItem={autoAddFirstItem}
      />
      <Dialog open={pendingRemoveIdx !== null} onOpenChange={open => !open && setPendingRemoveIdx(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Resource?</DialogTitle>
          </DialogHeader>
          <div>Are you sure you want to remove this resource? This action cannot be undone.</div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setPendingRemoveIdx(null)}>Cancel</Button>
            <Button variant="destructive" onClick={confirmRemove}>Remove Resource</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

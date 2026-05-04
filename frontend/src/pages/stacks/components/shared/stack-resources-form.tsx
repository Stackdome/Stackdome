import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackResourceItem from "@/pages/stacks/components/shared/stack-resource-item";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useMemo, useState } from "react";
import { PlusCircle } from "lucide-react";
import { useSecrets } from "../../hooks/use-secrets";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import type { AddonGroupStateMap } from "./stack-resource-item";

interface StackResourcesFormProps {
  resources: Partial<FormStackResourceData>[];
  onResourcesChange: (updatedResources: Partial<FormStackResourceData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  volumes?: Partial<VolumeFormData>[];
  accordionDefaultOpen?: boolean; // If false, all collapsed by default
  defaultOpenResourceIdx?: number | null;
  addonGroupState?: AddonGroupStateMap;
  onEditAddonBinding?: (addonId: string) => void;
  onDetachAddon?: (addonId: string) => void;
  onCancelDetachAddon?: (addonId: string) => void;
  availableAddonIds?: Set<string>;
}

function getDefaultResource(): Partial<FormStackResourceData> {
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
  accordionDefaultOpen = true,
  defaultOpenResourceIdx = null,
  addonGroupState,
  onEditAddonBinding,
  onDetachAddon,
  onCancelDetachAddon,
  availableAddonIds,
}: StackResourcesFormProps) {
  const [pendingRemoveIdx, setPendingRemoveIdx] = useState<number | null>(null);
  const secrets = useSecrets();
  const { addons: allAddons } = usePostgresAddons();
  const addons = useMemo(
    () =>
      availableAddonIds
        ? allAddons.filter((a: PostgresAddon) => a.id && availableAddonIds.has(a.id))
        : allAddons,
    [allAddons, availableAddonIds],
  );
  const addonNameById = useMemo(
    () => new Map(allAddons.filter((a: PostgresAddon) => a.id).map((a: PostgresAddon) => [a.id!, a.name])),
    [allAddons],
  );

  const isResourceFilled = (res: Partial<FormStackResourceData>) => {
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

  return (
    <div>
      <ResourceFormList<Partial<FormStackResourceData>>
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
            allResources={resources.map((r, i) => ({ name: r.name || `Resource ${i + 1}`, index: i }))}
            onRemove={handleRemove}
            secrets={secrets}
            addons={addons}
            addonNameById={addonNameById}
            addonGroupState={addonGroupState}
            onEditAddonBinding={onEditAddonBinding}
            onDetachAddon={onDetachAddon}
            onCancelDetachAddon={onCancelDetachAddon}
          />
        )}
        defaultAllCollapsed={!accordionDefaultOpen}
        defaultOpenIndex={defaultOpenResourceIdx}
        emptyTitle="No resources added yet"
        emptyDescription="Add a service to start running it. Each resource is a container or build that becomes part of this stack."
        emptyCtaLabel="Add resource"
        emptyOnAdd={() => onResourcesChange([...resources, getDefaultResource()])}
      />
      {resources.length > 0 && (
        <div className="flex justify-center mt-4">
          <Button
            type="button"
            variant="addAction"
            onClick={() => onResourcesChange([...resources, getDefaultResource()])}
          >
            <PlusCircle className="h-4 w-4" />
            Add Resource
          </Button>
        </div>
      )}
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
    </div>
  );
}

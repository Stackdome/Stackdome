import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import ResourceFormList from "@/pages/stacks/components/shared/resource-form-list";
import StackResourceItem from "@/pages/stacks/components/shared/stack-resource-item";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useCallback, useMemo, useRef, useState } from "react";
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
  /** Baseline resources (immutable session snapshot) — when provided, dirty visualization renders on each row. */
  baselineResources?: Partial<FormStackResourceData>[];
  /** Discard a single env row in a resource. Required for the per-row reset arrow. */
  onDiscardEnvRow?: (resourceIdx: number, envIdx: number) => void;
  /** Discard all changes for a resource. Required for the Modified pill ✕. */
  onDiscardResource?: (resourceIdx: number) => void;
  /** Discard a single field on a resource by dot-path. Required for per-field reset arrows. */
  onDiscardResourceField?: (resourceIdx: number, path: string) => void;
  /** When set, replaces the empty-state title with this message in danger color. */
  emptyError?: string;
}

export function getDefaultResource(): Partial<FormStackResourceData> {
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
  baselineResources,
  onDiscardEnvRow,
  onDiscardResource,
  onDiscardResourceField,
  emptyError,
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

  // Use refs so handleRemove identity stays stable across keystrokes —
  // otherwise it'd be a fresh closure every render and break React.memo
  // on every StackResourceItem child.
  const resourcesRef = useRef(resources);
  resourcesRef.current = resources;
  const onResourcesChangeRef = useRef(onResourcesChange);
  onResourcesChangeRef.current = onResourcesChange;
  const handleRemove = useCallback((idx: number) => {
    const cur = resourcesRef.current;
    if (isResourceFilled(cur[idx])) {
      setPendingRemoveIdx(idx);
    } else {
      onResourcesChangeRef.current(cur.filter((_, i) => i !== idx));
    }
  }, []);

  const confirmRemove = () => {
    if (pendingRemoveIdx !== null) {
      onResourcesChange(resources.filter((_, i) => i !== pendingRemoveIdx));
      setPendingRemoveIdx(null);
    }
  };

  // Stable reference for the depends-on options. Recomputed only when the
  // names or count of resources change, NOT on every keystroke into a body
  // field. Without this, every keystroke would hand a new array to each
  // StackResourceItem and break React.memo.
  const allResourcesRef = useMemo(
    () => resources.map((r, i) => ({ name: r.name || `Resource ${i + 1}`, index: i })),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- name list is the only structural input
    [resources.length, ...resources.map((r) => r.name)],
  );

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
            allResources={allResourcesRef}
            onRemove={handleRemove}
            secrets={secrets}
            addons={addons}
            addonNameById={addonNameById}
            addonGroupState={addonGroupState}
            onEditAddonBinding={onEditAddonBinding}
            onDetachAddon={onDetachAddon}
            onCancelDetachAddon={onCancelDetachAddon}
            baselineResource={baselineResources?.[index]}
            onDiscardEnvRow={onDiscardEnvRow ? (envIdx) => onDiscardEnvRow(index, envIdx) : undefined}
            onDiscardResource={onDiscardResource ? () => onDiscardResource(index) : undefined}
            onDiscardField={onDiscardResourceField ? (path) => onDiscardResourceField(index, path) : undefined}
          />
        )}
        defaultAllCollapsed={!accordionDefaultOpen}
        defaultOpenIndex={defaultOpenResourceIdx}
        emptyTitle="No resources added yet"
        emptyDescription="Add a service to start running it. Each resource is a container or build that becomes part of this stack."
        emptyCtaLabel="Add resource"
        emptyError={emptyError}
        emptyOnAdd={() => onResourcesChange([...resources, getDefaultResource()])}
      />
      {resources.length > 0 && (
        <div className="flex justify-center mt-4">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => onResourcesChange([...resources, getDefaultResource()])}
          >
            <PlusCircle className="h-4 w-4 mr-2" />
            Add resource
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

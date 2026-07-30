import type { ReactNode } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { CornerDownRight, Trash2 } from "lucide-react";
import {
  LedgerRow,
  LedgerSection,
} from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/ledger";
import type { FormVolumeExtendedData as VolumeFormData, FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

/** Derive the list of resources that mount this volume. */
export function volumeMountingInfo(
  volume: Partial<VolumeFormData>,
  allStackResources: Partial<FormStackResourceData>[],
): { resourceName: string; targetPath: string }[] {
  if (!volume.name) return [];
  return allStackResources
    .map((resource) => {
      if (!resource.name || !resource.volume_mounts) return null;
      const mountDetail = resource.volume_mounts.find(
        (vm) => vm.source_volume_name === volume.name,
      );
      return mountDetail ? { resourceName: resource.name, targetPath: mountDetail.target_path } : null;
    })
    .filter(Boolean) as { resourceName: string; targetPath: string }[];
}

interface VolumeFieldsProps {
  volume: Partial<VolumeFormData>;
  index: number;
  onChange: (index: number, updated: Partial<VolumeFormData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  allVolumes: Partial<VolumeFormData>[];
  allStackResources?: Partial<FormStackResourceData>[];
  /**
   * Render the Name input read-only. Used by VolumeDrawer, where drawer
   * entries are keyed by volume name — a live rename would orphan the open
   * entry and close the drawer mid-edit.
   */
  nameReadOnly?: boolean;
  /**
   * Render the spec inputs (size) read-only. Used for server-persisted
   * volumes: PVC size is immutable once provisioned, and the autosave engine
   * has no volume-update op — an edit here would silently go nowhere.
   */
  specReadOnly?: boolean;
}

/** Shared form body for a volume: spec ledger, mount details, and remove footer. */
export function VolumeFields({
  volume,
  index,
  onChange,
  onRemove,
  errors,
  allVolumes,
  allStackResources = [],
  nameReadOnly = false,
  specReadOnly = false,
}: VolumeFieldsProps): ReactNode {
  const update = (patch: Partial<VolumeFormData>) => {
    onChange(index, { ...volume, ...patch });
  };

  const isDuplicate = allVolumes.filter((v) => v.name?.length && v.name === volume.name).length > 1;

  const mountingInfo = volumeMountingInfo(volume, allStackResources);

  return (
    <>
      <LedgerSection label="Specification">
        <LedgerRow
          label="Name"
          htmlFor={`volume-name-${index}`}
          required
          meta={nameReadOnly ? "rename from volumes list" : "unique in stack"}
          error={errors.name || (isDuplicate ? "Volume name must be unique" : undefined)}
        >
          <Input
            id={`volume-name-${index}`}
            placeholder="Volume name"
            value={volume.name || ""}
            onChange={(e) => update({ name: e.target.value })}
            disabled={nameReadOnly}
            className={`h-9 font-mono text-[12.5px] ${errors.name || isDuplicate ? "border-danger" : ""}`}
            aria-invalid={!!errors.name || isDuplicate}
          />
        </LedgerRow>
        <LedgerRow
          label="Size"
          htmlFor={`volume-size-${index}`}
          required
          meta={specReadOnly ? "fixed once provisioned" : "e.g. 1Gi, 500Mi"}
          error={errors["spec.size"]}
        >
          <Input
            id={`volume-size-${index}`}
            placeholder="e.g., 1Gi, 500Mi"
            value={volume.spec?.size || ""}
            onChange={(e) =>
              update({
                spec: {
                  ...volume.spec,
                  size: e.target.value,
                  needs_sync_before_use: volume.spec?.needs_sync_before_use ?? false,
                  access_mode: volume.spec?.access_mode ?? "ReadWriteOnce",
                },
              })
            }
            disabled={specReadOnly}
            className={`h-9 font-mono text-[12.5px] ${errors["spec.size"] ? "border-danger" : ""}`}
            aria-invalid={!!errors["spec.size"]}
          />
        </LedgerRow>
        <LedgerRow label="Access mode" meta="single resource · read/write">
          <code className="inline-block rounded-[3px] bg-secondary px-2 py-1 font-mono text-[11.5px] text-muted-foreground">
            ReadWriteOnce (RWO)
          </code>
        </LedgerRow>
      </LedgerSection>

      {mountingInfo.length > 0 && (
        <LedgerSection
          label="Mounts"
          meta={`${mountingInfo.length} ${mountingInfo.length === 1 ? "resource" : "resources"}`}
        >
          {mountingInfo.map((mount, mountIdx) => (
            <div key={mountIdx} className="border-b border-secondary/80 py-1">
              <div className="flex items-center gap-3 rounded-md px-1.5 py-1.5 transition-colors hover:bg-muted/20">
                <div className="flex w-[150px] shrink-0 items-center gap-2 text-[13px] text-foreground/80 dark:text-fg-2">
                  <CornerDownRight className="h-3.5 w-3.5 shrink-0 text-fg-muted" aria-hidden />
                  <span className="truncate">{mount.resourceName}</span>
                </div>
                <code className="shrink-0 rounded-[3px] bg-secondary px-2 py-1 font-mono text-[11.5px] text-muted-foreground">
                  {mount.targetPath}
                </code>
              </div>
            </div>
          ))}
        </LedgerSection>
      )}

      <div className="mt-5 flex justify-end border-t border-border pt-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 px-2 text-[12.5px] text-danger hover:bg-danger-bg hover:text-danger"
          onClick={() => onRemove(index)}
          title="Remove volume"
        >
          <Trash2 className="h-3.5 w-3.5" />
          Remove volume
        </Button>
      </div>
    </>
  );
}

import type { ReactNode } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { CornerDownRight, Trash2 } from "lucide-react";
import { FieldShell } from "@/components/branded";
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
}

/** Shared form body for a volume: spec grid, mount details, and remove footer. */
export function VolumeFields({
  volume,
  index,
  onChange,
  onRemove,
  errors,
  allVolumes,
  allStackResources = [],
}: VolumeFieldsProps): ReactNode {
  const update = (patch: Partial<VolumeFormData>) => {
    onChange(index, { ...volume, ...patch });
  };

  const isDuplicate = allVolumes.filter((v) => v.name?.length && v.name === volume.name).length > 1;

  const mountingInfo = volumeMountingInfo(volume, allStackResources);

  return (
    <>
      {/* Volume specification */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Volume specification</h3>
        <div className="grid gap-5">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FieldShell
              label="Name"
              htmlFor={`volume-name-${index}`}
              required
              error={errors.name || (isDuplicate ? "Volume name must be unique" : undefined)}
            >
              <Input
                id={`volume-name-${index}`}
                placeholder="Volume name"
                value={volume.name || ""}
                onChange={(e) => update({ name: e.target.value })}
                className={`font-mono ${errors.name || isDuplicate ? "border-danger" : ""}`}
                aria-invalid={!!errors.name || isDuplicate}
              />
            </FieldShell>
            <FieldShell
              label="Size"
              htmlFor={`volume-size-${index}`}
              required
              hint="e.g., 1Gi, 500Mi."
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
                className={`font-mono ${errors["spec.size"] ? "border-danger" : ""}`}
                aria-invalid={!!errors["spec.size"]}
              />
            </FieldShell>
          </div>
          <FieldShell
            label="Access mode"
            hint="Mounted by a single resource for read/write."
          >
            <Input value="ReadWriteOnce (RWO)" disabled className="bg-muted font-mono" />
          </FieldShell>
        </div>
      </div>

      {/* Mount details */}
      {mountingInfo.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-foreground mb-3">Mount details</h3>
          <div className="space-y-1.5">
            {mountingInfo.map((mount, mountIdx) => (
              <div key={mountIdx} className="flex items-center gap-1.5 text-[13px] text-muted-foreground">
                <CornerDownRight className="h-3.5 w-3.5 shrink-0" />
                <span>
                  Mounted by <span className="font-medium text-foreground">{mount.resourceName}</span>
                  <span className="mx-1.5">at</span>
                  <code className="font-mono text-[12px] bg-muted text-muted-foreground px-1.5 py-0.5 rounded">
                    {mount.targetPath}
                  </code>
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Remove */}
      <div className="pt-3 border-t border-border flex justify-end">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-[12.5px] text-muted-foreground/70 hover:text-danger hover:bg-danger-bg"
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

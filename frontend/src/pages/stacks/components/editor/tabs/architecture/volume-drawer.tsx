import { useCallback } from "react";
import { X, HardDrive } from "lucide-react";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import type { FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import { VolumeFields } from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/volume-fields";
import { removeMountsOf } from "@/pages/stacks/lib/canvas/volume-ops";

interface VolumeDrawerProps {
  /** Draft volume name — the stack entry's identity. */
  volumeName: string;
  session: UseStackEditSession;
  onClose: () => void;
  /** When provided, "Remove volume" defers to a caller-owned confirm dialog
   *  instead of removing immediately (e.g. StackCanvasTab's shared confirm). */
  onRequestRemove?: (name: string) => void;
  /** True when the volume already exists server-side — its spec (size) is
   *  immutable once the PVC is provisioned. */
  persisted?: boolean;
}

/** Drawer body for a volume pushed from a service's mount row. */
export function VolumeDrawer({ volumeName, session, onClose, onRequestRemove, persisted = false }: VolumeDrawerProps) {
  const volumes = session.draft.volumes;
  const index = volumes.findIndex((v) => v.name === volumeName);
  const volume = (volumes[index] ?? {}) as Partial<VolumeFormData>;

  const onChange = useCallback(
    (idx: number, updated: Partial<VolumeFormData>) => {
      session.updateVolumes((prev) => prev.map((v, i) => (i === idx ? updated : v)));
    },
    [session],
  );

  const onRemove = useCallback(
    (idx: number) => {
      const name = volumes[idx]?.name;
      if (onRequestRemove) {
        if (name) onRequestRemove(name);
        return;
      }
      session.updateVolumes((prev) => prev.filter((_, i) => i !== idx));
      if (name) session.updateResources((prev) => removeMountsOf(prev, name));
      onClose();
    },
    [session, volumes, onClose, onRequestRemove],
  );

  if (index < 0) return null;

  return (
    <div className="flex h-full w-full flex-col bg-background" data-testid="volume-drawer">
      <div className="flex flex-none items-center gap-3 border-b border-border px-4 py-[15px]">
        <HardDrive className="size-[19px] shrink-0 text-brand" />
        <div className="min-w-0 flex-1 leading-tight">
          <div className="truncate text-base font-medium text-foreground">{volume.name || volumeName}</div>
          <div className="truncate font-mono text-[11px] text-fg-muted">
            {volume.spec?.size || "size unset"} · {volume.spec?.access_mode || "ReadWriteOnce"}
          </div>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="shrink-0 rounded p-1 text-fg-muted hover:bg-muted hover:text-foreground"
        >
          <X className="size-[18px]" />
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
        <VolumeFields
          volume={volume}
          index={index}
          onChange={onChange}
          onRemove={onRemove}
          errors={{}}
          allVolumes={volumes}
          allStackResources={session.draft.resources}
          nameReadOnly
          specReadOnly={persisted}
        />
      </div>
    </div>
  );
}

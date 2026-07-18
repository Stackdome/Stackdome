import { HardDrive, Link2, Plug, Settings2, Trash2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export type CanvasMenuTarget =
  | { kind: "resource"; resourceIdx: number; resourceName: string; x: number; y: number }
  | { kind: "volume-chip"; volumeName: string; x: number; y: number }
  | { kind: "volume-node"; volumeName: string; x: number; y: number };

interface CanvasContextMenuProps {
  target: CanvasMenuTarget | null;
  onClose: () => void;
  onOpenResource: (resourceIdx: number) => void;
  onAddVolumeToResource: (resourceIdx: number) => void;
  onDeleteResource: (resourceName: string) => void;
  onDisconnectVolume: (volumeName: string) => void;
  onOpenVolume: (volumeName: string) => void;
  onRequestDeleteVolume: (volumeName: string) => void;
  onRequestAttach: (volumeName: string) => void;
}

const DANGER_ITEM = "text-danger focus:text-danger focus:bg-danger-bg";

/** Right-click menu for canvas nodes: one component, item-set keyed by target kind. */
export function CanvasContextMenu({
  target,
  onClose,
  onOpenResource,
  onAddVolumeToResource,
  onDeleteResource,
  onDisconnectVolume,
  onOpenVolume,
  onRequestDeleteVolume,
  onRequestAttach,
}: CanvasContextMenuProps) {
  if (!target) return null;

  const volumeItems = (volumeName: string, mounted: boolean) => (
    <>
      {mounted ? (
        <DropdownMenuItem onSelect={() => onDisconnectVolume(volumeName)}>
          <Plug className="size-4" /> Disconnect volume
        </DropdownMenuItem>
      ) : (
        <DropdownMenuItem onSelect={() => onRequestAttach(volumeName)}>
          <Link2 className="size-4" /> Attach to service…
        </DropdownMenuItem>
      )}
      <DropdownMenuItem onSelect={() => onOpenVolume(volumeName)}>
        <Settings2 className="size-4" /> Volume settings
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem className={DANGER_ITEM} onSelect={() => onRequestDeleteVolume(volumeName)}>
        <Trash2 className="size-4" /> Delete volume
      </DropdownMenuItem>
    </>
  );

  return (
    <DropdownMenu open onOpenChange={(open) => !open && onClose()}>
      <DropdownMenuTrigger asChild>
        <span style={{ position: "fixed", left: target.x, top: target.y, width: 0, height: 0 }} aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={2} className="w-52">
        {target.kind === "resource" ? (
          <>
            <DropdownMenuItem onSelect={() => onOpenResource(target.resourceIdx)}>
              <Settings2 className="size-4" /> Open settings
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => onAddVolumeToResource(target.resourceIdx)}>
              <HardDrive className="size-4" /> Add volume…
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className={DANGER_ITEM} onSelect={() => onDeleteResource(target.resourceName)}>
              <Trash2 className="size-4" /> Delete service
            </DropdownMenuItem>
          </>
        ) : (
          volumeItems(target.volumeName, target.kind === "volume-chip")
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

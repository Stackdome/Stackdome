import { useEffect, useState } from "react";
import { HardDrive } from "lucide-react";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import type { ResourceArr, VolumeArr } from "@/pages/stacks/lib/stack-diff";
import {
  suggestVolumeName,
  validateTargetPath,
  validateVolumeName,
} from "@/pages/stacks/lib/canvas/volume-ops";

const DEFAULT_SIZE = "1Gi";

interface AddVolumeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  resources: ResourceArr;
  volumes: VolumeArr;
  /** Pre-select a service (resource-card "Add volume…" menu entry). */
  initialResourceIdx?: number | null;
  onCreate: (input: { name: string; size: string; resourceIdx: number; targetPath: string }) => void;
}

/** Create-a-volume dialog: the volume is born attached to the chosen service. */
export function AddVolumeDialog({ open, onOpenChange, resources, volumes, initialResourceIdx, onCreate }: AddVolumeDialogProps) {
  const [name, setName] = useState("");
  const [size, setSize] = useState(DEFAULT_SIZE);
  const [resourceIdx, setResourceIdx] = useState<number | null>(null);
  const [targetPath, setTargetPath] = useState("");
  const [errors, setErrors] = useState<{ name?: string; resource?: string; targetPath?: string }>({});

  // Re-seed the form each time the dialog opens.
  useEffect(() => {
    if (!open) return;
    setName(suggestVolumeName(volumes));
    setSize(DEFAULT_SIZE);
    setResourceIdx(initialResourceIdx ?? (resources.length === 1 ? 0 : null));
    setTargetPath("");
    setErrors({});
    // eslint-disable-next-line react-hooks/exhaustive-deps -- seed on open only
  }, [open]);

  const submit = () => {
    const next = {
      name: validateVolumeName(name, volumes),
      resource: resourceIdx == null ? "Required" : undefined,
      targetPath: validateTargetPath(targetPath, resourceIdx == null ? undefined : resources[resourceIdx]),
    };
    setErrors(next);
    if (next.name || next.resource || next.targetPath) return;
    onCreate({ name, size: size.trim() || DEFAULT_SIZE, resourceIdx: resourceIdx!, targetPath });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="size-[18px] text-brand" /> Add volume
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <div className="grid grid-cols-2 gap-4">
            <FieldShell label="Name" htmlFor="add-volume-name" required error={errors.name}>
              <Input
                id="add-volume-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={`font-mono ${errors.name ? "border-danger" : ""}`}
                aria-invalid={!!errors.name}
              />
            </FieldShell>
            <FieldShell label="Size" htmlFor="add-volume-size" hint="e.g., 1Gi, 500Mi.">
              <Input id="add-volume-size" value={size} onChange={(e) => setSize(e.target.value)} className="font-mono" />
            </FieldShell>
          </div>
          <FieldShell label="Attach to service" htmlFor="add-volume-service" required error={errors.resource}>
            <Select
              value={resourceIdx == null ? "" : String(resourceIdx)}
              onValueChange={(v) => setResourceIdx(Number(v))}
            >
              <SelectTrigger id="add-volume-service" className={errors.resource ? "border-danger" : ""}>
                <SelectValue placeholder="Select service" />
              </SelectTrigger>
              <SelectContent>
                {resources.map((r, i) =>
                  r.name ? (
                    <SelectItem key={r.name} value={String(i)}>
                      {r.name}
                    </SelectItem>
                  ) : null,
                )}
              </SelectContent>
            </Select>
          </FieldShell>
          <FieldShell
            label="Mount path"
            htmlFor="add-volume-path"
            required
            hint="Path inside the service, starting with /, e.g., /var/lib/data."
            error={errors.targetPath}
          >
            <Input
              id="add-volume-path"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="/var/lib/data"
              className={`font-mono ${errors.targetPath ? "border-danger" : ""}`}
              aria-invalid={!!errors.targetPath}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={submit}>Add volume</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

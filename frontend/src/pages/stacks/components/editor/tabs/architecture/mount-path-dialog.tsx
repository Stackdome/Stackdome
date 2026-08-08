import { useEffect, useState } from "react";
import { HardDrive } from "lucide-react";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import type { ResourceArr } from "@/pages/stacks/lib/stack-diff";
import { validateTargetPath } from "@/pages/stacks/lib/canvas/volume-ops";

interface MountPathDialogProps {
  volumeName: string | null;
  resources: ResourceArr;
  resourceIdx: number | null;
  onCancel: () => void;
  onAttach: (input: { volumeName: string; resourceIdx: number; targetPath: string }) => void;
}

/** Attach an existing volume: fixed target after a drag-drop, picker from the menu. */
export function MountPathDialog({ volumeName, resources, resourceIdx, onCancel, onAttach }: MountPathDialogProps) {
  const open = volumeName != null;
  const [pickedIdx, setPickedIdx] = useState<number | null>(null);
  const [targetPath, setTargetPath] = useState("");
  const [errors, setErrors] = useState<{ resource?: string; targetPath?: string }>({});

  useEffect(() => {
    if (!open) return;
    setPickedIdx(resourceIdx ?? (resources.length === 1 ? 0 : null));
    setTargetPath("");
    setErrors({});
    // eslint-disable-next-line react-hooks/exhaustive-deps -- seed on open only
  }, [open, resourceIdx]);

  const submit = () => {
    const idx = resourceIdx ?? pickedIdx;
    const next = {
      resource: idx == null ? "Required" : undefined,
      targetPath: validateTargetPath(targetPath, idx == null ? undefined : resources[idx]),
    };
    setErrors(next);
    if (next.resource || next.targetPath || idx == null || !volumeName) return;
    onAttach({ volumeName, resourceIdx: idx, targetPath });
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onCancel()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="size-[18px] text-brand" /> Attach “{volumeName}”
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          {resourceIdx == null && (
            <FieldShell label="Service" htmlFor="attach-service" required error={errors.resource}>
              <Select value={pickedIdx == null ? "" : String(pickedIdx)} onValueChange={(v) => setPickedIdx(Number(v))}>
                <SelectTrigger id="attach-service" className={errors.resource ? "border-danger" : ""}>
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
          )}
          <FieldShell
            label="Mount path"
            htmlFor="attach-path"
            required
            hint="Absolute path inside the service."
            error={errors.targetPath}
          >
            <Input
              id="attach-path"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="/var/lib/data"
              className={`font-mono ${errors.targetPath ? "border-danger" : ""}`}
              aria-invalid={!!errors.targetPath}
              autoFocus
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button shape="flat" variant="ghost" onClick={onCancel}>
            Cancel
          </Button>
          <Button onClick={submit}>Attach</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

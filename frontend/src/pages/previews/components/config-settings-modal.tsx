import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/components/ui/use-toast";
import {
  updatePreviewConfig, deletePreviewConfig,
  type StackPreviewConfig,
} from "@/api/preview-configs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";

interface ConfigSettingsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: StackPreviewConfig;
  onSaved: (updated: StackPreviewConfig) => void;
  onDeleted: () => void;
}

export function ConfigSettingsModal({ open, onOpenChange, config, onSaved, onDeleted }: ConfigSettingsModalProps) {
  const { toast } = useToast();
  const { defaultTeamName } = useResourceTeams();

  const [baseBranch, setBaseBranch] = useState("");
  const [stackfilePath, setStackfilePath] = useState("");
  const [maxActive, setMaxActive] = useState(10);
  const [saving, setSaving] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  // Re-seed local state from the latest config every time the modal opens —
  // discards unsaved edits on cancel and picks up changes saved elsewhere.
  useEffect(() => {
    if (!open) return;
    setBaseBranch(config.git_repository?.base_branch ?? "");
    setStackfilePath(config.stackfile_path ?? "stackfile.yaml");
    setMaxActive(config.max_active_previews ?? 10);
  }, [open, config]);

  const save = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !config.id) return;
    setSaving(true);
    try {
      // PUT is full-replace server-side; echo unchanged fields so they survive the save.
      const updated = await updatePreviewConfig(orgId, defaultTeamName, config.id, {
        git_repository: {
          repo_url: config.git_repository?.repo_url ?? "",
          base_branch: baseBranch,
        },
        stackfile_path: stackfilePath,
        max_active_previews: maxActive,
        ...(config.description != null ? { description: config.description } : {}),
        ...(config.labels != null ? { labels: config.labels } : {}),
        ...(config.annotations != null ? { annotations: config.annotations } : {}),
      });
      toast({ title: "Configuration saved" });
      onSaved(updated);
      onOpenChange(false);
    } catch (e) {
      toast({ title: "Save failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  const confirmDelete = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !config.id) return;
    setDeleting(true);
    try {
      await deletePreviewConfig(orgId, defaultTeamName, config.id);
      toast({ title: "Configuration deleted" });
      setDeleteConfirmOpen(false);
      onOpenChange(false);
      onDeleted();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setDeleting(false);
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Repository settings</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="cs-branch">Base branch</Label>
              <Input id="cs-branch" value={baseBranch} onChange={(e) => setBaseBranch(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="cs-stackfile">Stackfile path</Label>
              <Input id="cs-stackfile" value={stackfilePath} onChange={(e) => setStackfilePath(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="cs-max">Max active previews</Label>
              <Input
                id="cs-max"
                type="number"
                min={1}
                value={maxActive}
                onChange={(e) => {
                  const n = e.target.valueAsNumber;
                  setMaxActive(Number.isNaN(n) ? 1 : Math.max(1, Math.floor(n)));
                }}
                className="w-28"
              />
            </div>
            <Button onClick={() => void save()} disabled={saving}>Save</Button>
          </div>

          <Separator className="my-2" />

          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-danger">Danger zone</h3>
            <p className="text-sm text-muted-foreground">
              Deleting the configuration stops new previews for this repository.
            </p>
            <Button variant="destructive" size="sm" onClick={() => setDeleteConfirmOpen(true)}>
              Delete configuration
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Rendered as a sibling of the settings Dialog rather than nested
          inside DialogContent (with an AlertDialogTrigger) — that composition
          hits a Radix focus-restore bug where closing the AlertDialog leaves
          focus lost instead of returning it to the settings Dialog. */}
      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {config.name}?</AlertDialogTitle>
            <AlertDialogDescription>
              Existing preview environments must be deleted first; the backend
              rejects the delete otherwise.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => void confirmDelete()} disabled={deleting}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

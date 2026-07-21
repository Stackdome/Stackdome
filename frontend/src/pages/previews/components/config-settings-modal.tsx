import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/components/ui/use-toast";
import {
  updatePreviewConfig, deletePreviewConfig,
  type StackPreviewConfig,
} from "@/api/preview-configs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import {
  configSettingsSchema, DEFAULT_STACKFILE_PATH, type ConfigSettingsValues,
} from "@/pages/previews/lib/form-schemas";

interface ConfigSettingsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: StackPreviewConfig;
  onSaved: (updated: StackPreviewConfig) => void;
  onDeleted: () => void;
}

export function ConfigSettingsModal({ open, onOpenChange, config, onSaved, onDeleted }: ConfigSettingsModalProps) {
  const { toast } = useToast();
  const { defaultProjectName } = useResourceProjects();

  const [baseBranch, setBaseBranch] = useState("");
  const [stackfilePath, setStackfilePath] = useState("");
  const [maxActive, setMaxActive] = useState(10);
  const [saving, setSaving] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof ConfigSettingsValues, string>>>({});
  const [deleting, setDeleting] = useState(false);
  // True from clicking "Delete configuration" until the flow settles — keeps
  // the settings Dialog from self-dismissing while the confirm is up.
  const [deleteFlowActive, setDeleteFlowActive] = useState(false);
  const confirm = useConfirm();

  // Re-seed local state from the latest config every time the modal opens —
  // discards unsaved edits on cancel and picks up changes saved elsewhere.
  useEffect(() => {
    if (!open) return;
    setBaseBranch(config.git_repository?.base_branch ?? "");
    setStackfilePath(config.stackfile_path ?? DEFAULT_STACKFILE_PATH);
    setMaxActive(config.max_active_previews ?? 10);
    setFieldErrors({});
  }, [open, config]);

  const save = async () => {
    const parsed = configSettingsSchema.safeParse({ baseBranch, stackfilePath, maxActive });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({
        baseBranch: flat.baseBranch?.[0],
        stackfilePath: flat.stackfilePath?.[0],
        maxActive: flat.maxActive?.[0],
      });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultProjectName || !config.id) return;
    setSaving(true);
    try {
      // PUT is full-replace server-side; echo unchanged fields so they survive the save.
      const updated = await updatePreviewConfig(orgId, defaultProjectName, config.id, {
        git_repository: {
          repo_url: config.git_repository?.repo_url ?? "",
          base_branch: parsed.data.baseBranch,
        },
        stackfile_path: parsed.data.stackfilePath,
        max_active_previews: parsed.data.maxActive,
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

  const requestDelete = async () => {
    setDeleteFlowActive(true);
    try {
      const ok = await confirm({
        title: `Delete ${config.name}?`,
        description: "Delete this configuration's preview environments first.",
        confirmLabel: "Delete",
        variant: "destructive",
      });
      if (!ok) return;
      const orgId = getCurrentOrganizationId();
      if (!orgId || !defaultProjectName || !config.id) return;
      setDeleting(true);
      try {
        await deletePreviewConfig(orgId, defaultProjectName, config.id);
        toast({ title: "Configuration deleted" });
        onOpenChange(false);
        onDeleted();
      } catch (e) {
        toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
      } finally {
        setDeleting(false);
      }
    } finally {
      setDeleteFlowActive(false);
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        {/* The app-root confirm dialog portals as a sibling, so its clicks land
            "outside" this Dialog's content — without these guards Radix
            dismisses the settings modal underneath the confirm. */}
        <DialogContent
          className="sm:max-w-md"
          onInteractOutside={(e) => {
            if (deleteFlowActive) e.preventDefault();
          }}
          onEscapeKeyDown={(e) => {
            if (deleteFlowActive) e.preventDefault();
          }}
        >
          <DialogHeader>
            <DialogTitle>Repository settings</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <FieldShell label="Base branch" htmlFor="cs-branch" required error={fieldErrors.baseBranch}>
              <Input
                id="cs-branch"
                value={baseBranch}
                onChange={(e) => {
                  setBaseBranch(e.target.value);
                  setFieldErrors((prev) => ({ ...prev, baseBranch: undefined }));
                }}
                aria-invalid={!!fieldErrors.baseBranch}
              />
            </FieldShell>
            <FieldShell
              label="Stackfile path"
              htmlFor="cs-stackfile"
              required
              hint="Defines the full stack (services, ports, env). Fetched from the repository on every deploy."
              error={fieldErrors.stackfilePath}
            >
              <Input
                id="cs-stackfile"
                value={stackfilePath}
                onChange={(e) => {
                  setStackfilePath(e.target.value);
                  setFieldErrors((prev) => ({ ...prev, stackfilePath: undefined }));
                }}
                aria-invalid={!!fieldErrors.stackfilePath}
              />
            </FieldShell>
            <FieldShell label="Max active previews" htmlFor="cs-max" error={fieldErrors.maxActive}>
              <Input
                id="cs-max"
                type="number"
                min={1}
                value={maxActive}
                onChange={(e) => {
                  const n = e.target.valueAsNumber;
                  setMaxActive(Number.isNaN(n) ? 1 : Math.max(1, Math.floor(n)));
                  setFieldErrors((prev) => ({ ...prev, maxActive: undefined }));
                }}
                className="w-28"
              />
            </FieldShell>
            <Button onClick={() => void save()} disabled={saving}>Save</Button>
          </div>

          <Separator className="my-2" />

          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-danger">Danger zone</h3>
            <p className="text-sm text-muted-foreground">
              Deleting the configuration stops new previews for this repository.
            </p>
            <Button
              variant="destructive"
              size="sm"
              disabled={deleting}
              onClick={() => void requestDelete()}
            >
              Delete configuration
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

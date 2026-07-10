import { useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/components/ui/use-toast";
import { deletePreviewEnv, type PreviewStack } from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import type { StackPreviewConfig } from "@/api/preview-configs";
import { usePreviewEnvs } from "../hooks/use-preview-envs";
import { PreviewEnvRow } from "./preview-env-row";
import { NewPreviewEnvModal } from "./new-preview-env-modal";
import { SyncEnvDialog } from "./sync-env-dialog";

interface ConfigEnvsSectionProps {
  config: StackPreviewConfig;
}

export function ConfigEnvsSection({ config }: ConfigEnvsSectionProps) {
  const { envs, loading, error, refresh } = usePreviewEnvs(config.id);
  const { defaultTeamName } = useResourceTeams();
  const { toast } = useToast();
  const [createOpen, setCreateOpen] = useState(false);
  const [syncing, setSyncing] = useState<PreviewStack | null>(null);
  const [deleting, setDeleting] = useState<PreviewStack | null>(null);

  const confirmDelete = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !deleting?.id) return;
    try {
      await deletePreviewEnv(orgId, defaultTeamName, deleting.id);
      toast({ title: `Deleting PR #${deleting.pr_number} environment` });
      await refresh();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div className="space-y-2">
      {loading && <p className="text-sm text-muted-foreground">Loading environments…</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}

      {!loading && envs.length === 0 && (
        <p className="rounded-md border border-dashed px-3 py-4 text-sm text-muted-foreground">
          No preview environments yet.
        </p>
      )}

      {envs.map((env) => (
        <PreviewEnvRow key={env.id} env={env} onSync={setSyncing} onDelete={setDeleting} />
      ))}

      <Button variant="outline" size="sm" onClick={() => setCreateOpen(true)}>
        <Plus className="h-4 w-4" />
        New preview environment
      </Button>

      <NewPreviewEnvModal
        open={createOpen}
        onOpenChange={setCreateOpen}
        config={config}
        onCreated={() => void refresh()}
      />
      <SyncEnvDialog
        env={syncing}
        onOpenChange={(o) => !o && setSyncing(null)}
        onSynced={() => void refresh()}
      />
      <AlertDialog open={deleting != null} onOpenChange={(o) => !o && setDeleting(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete PR #{deleting?.pr_number} environment?</AlertDialogTitle>
            <AlertDialogDescription>
              The environment&apos;s stack and resources are torn down. This cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => void confirmDelete()}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

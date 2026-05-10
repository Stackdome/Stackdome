import { useState } from "react";
import { Loader2, PlayCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import { Panel, EmptyState } from "@/components/branded";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { triggerPostgresBackup } from "@/api/postgres-backups";
import type { PostgresAddon } from "@/api/addons";
import { usePostgresBackups } from "../hooks/use-postgres-backups";
import { BackupConfigEditor } from "./backup-config-editor";
import { BackupsList } from "./backups-list";

type Props = { addon: PostgresAddon; onAddonChanged: () => void };

export function BackupsTab({ addon, onAddonChanged }: Props) {
  const { toast } = useToast();
  const { backups, loading, error, refetch } = usePostgresBackups(addon.id);
  const [triggering, setTriggering] = useState(false);
  const hasDestination = !!addon.spec?.backup?.object_store_id;

  async function handleTrigger() {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !addon.id) return;
    setTriggering(true);
    try {
      await triggerPostgresBackup(orgId, addon.id);
      toast({ title: "Backup triggered" });
      await refetch();
    } catch (e) {
      toast({
        title: "Failed to trigger backup",
        description: getErrorMessage(e),
        variant: "destructive",
      });
    } finally {
      setTriggering(false);
    }
  }

  const triggerButton = (
    <Button
      onClick={handleTrigger}
      disabled={triggering || !hasDestination}
      title={!hasDestination ? "Configure an Object Store first" : undefined}
    >
      {triggering ? (
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
      ) : (
        <PlayCircle className="mr-2 h-4 w-4" />
      )}
      Run backup now
    </Button>
  );

  const isEmpty = !error && !loading && backups.length === 0;

  return (
    <div className="flex flex-col gap-6">
      <BackupConfigEditor addon={addon} onSaved={onAddonChanged} />

      <div>
        <div className="mb-3 flex items-center justify-end">
          {triggerButton}
        </div>
        <Panel
          title="Backups history"
          count={backups.length}
          bodyClassName={isEmpty || error || (loading && backups.length === 0) ? "p-5" : "p-0"}
        >
          {error ? (
            <div className="text-sm text-danger">{error}</div>
          ) : loading && backups.length === 0 ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : backups.length === 0 ? (
            <EmptyState
              title="No backups yet"
              description={
                hasDestination
                  ? "Click Run backup now to create your first backup."
                  : "Configure an Object Store above, then click Run backup now."
              }
            />
          ) : (
            <BackupsList backups={backups} />
          )}
        </Panel>
      </div>
    </div>
  );
}

import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2, PlayCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Panel, EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import {
  deletePostgresAddon,
  getPostgresAddon,
  type PostgresAddon,
} from "@/api/addons";
import { triggerPostgresBackup } from "@/api/postgres-backups";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useObjectStores } from "@/pages/object-stores/hooks/use-object-stores";
import { PostgresDetailHeader } from "./components/postgres-detail-header";
import { BackupsList } from "./components/backups-list";
import { DeleteAddonDialog } from "./components/delete-addon-dialog";
import { usePostgresBackups } from "./hooks/use-postgres-backups";

export default function PostgresDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [addon, setAddon] = useState<PostgresAddon | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { objectStores } = useObjectStores();
  const {
    backups,
    loading: backupsLoading,
    error: backupsError,
    refetch: refetchBackups,
  } = usePostgresBackups(id);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [triggering, setTriggering] = useState(false);

  const refetch = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !id) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getPostgresAddon(orgId, id);
      setAddon(data);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void refetch();
  }, [refetch]);

  useEffect(() => {
    if (!id) return;
    const path = `/addons/postgres/${id}`;
    setCustomLabel(path, addon?.name ?? "Postgres add-on");
    setPathLoading(path, loading);
  }, [id, addon?.name, loading, setCustomLabel, setPathLoading]);

  if (loading && !addon) {
    return (
      <div className="flex items-center justify-center p-12 text-sm text-muted-foreground">
        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Loading…
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <p className="text-sm text-danger">{error}</p>
      </div>
    );
  }

  if (!addon) return null;

  const b = addon.spec?.backup;
  const hasDestination = !!b?.object_store_id;
  const storeName = b?.object_store_id
    ? objectStores.find((s) => s.id === b.object_store_id)?.name ?? b.object_store_id
    : "—";

  async function handleTrigger() {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !addon?.id) return;
    setTriggering(true);
    try {
      await triggerPostgresBackup(orgId, addon.id);
      toast({ title: "Backup triggered" });
      await refetchBackups();
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

  async function handleDelete() {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !addon?.id) return;
    setDeleting(true);
    setDeleteError(null);
    try {
      await deletePostgresAddon(orgId, addon.id);
      toast({
        title: "Addon deleted",
        description: `"${addon.name}" is being torn down.`,
        variant: "destructive",
      });
      navigate("/addons", { replace: true });
    } catch (e) {
      setDeleteError(
        isErrorStatus(e, 409)
          ? `${getErrorMessage(e)}\n\nRemove the stack references first, then try again.`
          : getErrorMessage(e),
      );
    } finally {
      setDeleting(false);
    }
  }

  return (
    <TooltipProvider>
      <div className="flex flex-col gap-6 p-6">
        <PostgresDetailHeader addon={addon} onDelete={() => setDeleteOpen(true)} />

        <Panel title="Addon Information">
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm max-w-3xl">
            <div>
              <dt className="text-muted-foreground">Name</dt>
              <dd className="font-mono">{addon.name}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Type</dt>
              <dd>Postgres</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Created</dt>
              <dd>
                {addon.created_at
                  ? new Date(addon.created_at).toLocaleString()
                  : "—"}
              </dd>
            </div>
          </dl>
        </Panel>

        <Panel title="Configuration">
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm max-w-3xl">
            <div>
              <dt className="text-muted-foreground">Version</dt>
              <dd>PG {addon.spec.version.major}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Storage</dt>
              <dd className="font-mono">{addon.spec.storage.size ?? "—"}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Instances</dt>
              <dd>{addon.spec.instances.count}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Superuser access</dt>
              <dd>
                {addon.spec.configuration?.enable_superuser_access
                  ? "On"
                  : "Off"}
              </dd>
            </div>
          </dl>
        </Panel>

        <Panel title="Backups">
          <div className="flex flex-col gap-5">
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm max-w-3xl">
              <div>
                <dt className="text-muted-foreground">Scheduled backups</dt>
                <dd>{b?.enabled ? "On" : "Off"}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Object Store</dt>
                <dd>{storeName}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Schedule</dt>
                <dd className="font-mono">{b?.schedule || "—"}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">WAL archiving</dt>
                <dd>{b?.wal_archiving ? "On" : "Off"}</dd>
              </div>
            </dl>
            {b?.enabled && (
              <>
                <div className="flex items-center justify-between border-t border-border pt-4">
                  <span className="text-sm text-muted-foreground">
                    {backups.length} recent run{backups.length === 1 ? "" : "s"}
                  </span>
                  <Button
                    onClick={handleTrigger}
                    disabled={triggering || !hasDestination}
                    title={
                      !hasDestination
                        ? "Configure an Object Store via Edit first"
                        : undefined
                    }
                  >
                    {triggering ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <PlayCircle className="mr-2 h-4 w-4" />
                    )}
                    Run backup now
                  </Button>
                </div>
                {backupsError ? (
                  <div className="text-sm text-danger">{backupsError}</div>
                ) : backupsLoading && backups.length === 0 ? (
                  <div className="text-sm text-muted-foreground">Loading…</div>
                ) : backups.length === 0 ? (
                  <EmptyState
                    title="No backups yet"
                    description="Click Run backup now to create your first backup."
                  />
                ) : (
                  <BackupsList backups={backups} />
                )}
              </>
            )}
          </div>
        </Panel>

        <DeleteAddonDialog
          open={deleteOpen}
          addonName={addon.name}
          loading={deleting}
          error={deleteError}
          onConfirm={handleDelete}
          onCancel={() => {
            if (!deleting) {
              setDeleteOpen(false);
              setDeleteError(null);
            }
          }}
        />
      </div>
    </TooltipProvider>
  );
}

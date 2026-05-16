import { useCallback, useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  Loader2,
  PlayCircle,
  ChevronLeft,
  ChevronRight,
  Info,
} from "lucide-react";
import cronstrue from "cronstrue";
import { Button } from "@/components/ui/button";
import {
  TooltipProvider,
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from "@/components/ui/tooltip";
import { Panel, EmptyState, FieldShell } from "@/components/branded";
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
import { detectPlan } from "./lib/payload";
import { PLAN_PRESETS } from "./lib/plan-presets";
import { PostgresDetailHeader } from "./components/postgres-detail-header";
import { BackupsList } from "./components/backups-list";
import { DeleteAddonDialog } from "./components/delete-addon-dialog";
import { usePostgresBackups } from "./hooks/use-postgres-backups";

// Read-only mirror of the edit form's FieldShell row, so the view page shares
// the create/edit label/value rhythm (visual parity, no inputs).
function ReadField({
  label,
  children,
}: {
  label: ReactNode;
  children: ReactNode;
}) {
  return (
    <FieldShell label={label}>
      <div className="text-sm text-foreground">{children}</div>
    </FieldShell>
  );
}

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
    total: backupsTotal,
    page: backupsPage,
    pageCount: backupsPageCount,
    pageSize: backupsPageSize,
    setPage: setBackupsPage,
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

  function describeSchedule(expr?: string): string | null {
    if (!expr) return null;
    try {
      return cronstrue.toString(expr, { use24HourTimeFormat: true });
    } catch {
      return null;
    }
  }

  const planId = detectPlan(addon.spec.resources);
  const planPreset = PLAN_PRESETS.find((p) => p.id === planId);
  const planLabel = planPreset?.label ?? planId;
  const res = addon.spec.resources;
  const resourcesLine =
    planId === "custom"
      ? res &&
        [res.cpu?.request, res.cpu?.limit, res.memory?.request, res.memory?.limit]
          .some(Boolean)
        ? `CPU ${res.cpu?.request ?? "—"}/${res.cpu?.limit ?? "—"} · Mem ${res.memory?.request ?? "—"}/${res.memory?.limit ?? "—"}`
        : null
      : planPreset
        ? `${planPreset.cpu} · ${planPreset.memory}`
        : null;

  const b = addon.spec?.backup;
  // Backend stores spec.backup only when enabled and omits the `enabled` flag
  // from responses, so presence of the object means backups are enabled.
  const backupEnabled = !!b && (b.enabled ?? true);
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

        <Panel title="Configuration">
          <div className="flex flex-col gap-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl">
              <ReadField label="Plan">
                {planLabel}
                {resourcesLine && (
                  <span className="block font-mono text-xs text-muted-foreground">
                    {resourcesLine}
                  </span>
                )}
              </ReadField>
              <ReadField label="Version">
                PG {addon.spec.version.major}
              </ReadField>
              <ReadField label="Storage">
                <span className="font-mono">
                  {addon.spec.storage.size ?? "—"}
                </span>
              </ReadField>
              <ReadField label="Instances">
                {addon.spec.instances.count}
              </ReadField>
              <ReadField label="Superuser access">
                {addon.spec.configuration?.enable_superuser_access
                  ? "On"
                  : "Off"}
              </ReadField>
            </div>

            <div className="border-t border-border pt-5">
              <h3 className="text-sm font-semibold text-foreground mb-3">
                Backups
              </h3>
              <div className="flex flex-col gap-5">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl">
                  <ReadField label="Scheduled backups">
                    {backupEnabled ? "On" : "Off"}
                  </ReadField>
                  <ReadField label="Object Store">{storeName}</ReadField>
                  <ReadField label="Schedule">
                    {b?.schedule ? (
                      <span className="inline-flex items-center gap-1.5">
                        {describeSchedule(b.schedule) ?? b.schedule}
                        <span className="text-muted-foreground/70">· UTC</span>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button
                              type="button"
                              aria-label="Show cron expression"
                              className="text-muted-foreground hover:text-foreground"
                            >
                              <Info className="h-3.5 w-3.5" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent className="font-mono text-xs">
                            {b.schedule}
                          </TooltipContent>
                        </Tooltip>
                      </span>
                    ) : (
                      "—"
                    )}
                  </ReadField>
                  <ReadField label="WAL archiving">
                    {b?.wal_archiving ? "On" : "Off"}
                  </ReadField>
                </div>

                {hasDestination ? (
                  <div className="flex items-center justify-end border-t border-border pt-4">
                    <Button onClick={handleTrigger} disabled={triggering}>
                      {triggering ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <PlayCircle className="h-4 w-4" />
                      )}
                      Run backup now
                    </Button>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground max-w-3xl border-t border-border pt-4">
                    No backup destination configured. Use{" "}
                    <Link
                      to={`/addons/postgres/${addon.id}/edit`}
                      className="text-brand hover:underline"
                    >
                      Edit configuration
                    </Link>{" "}
                    to set an object store — required for scheduled and manual
                    backups. Previous backup runs remain listed below.
                  </p>
                )}

                {backupsError ? (
                  <div className="text-sm text-danger">{backupsError}</div>
                ) : backupsLoading && backups.length === 0 ? (
                  <div className="text-sm text-muted-foreground">Loading…</div>
                ) : backups.length === 0 ? (
                  hasDestination ? (
                    <EmptyState
                      title="No backups yet"
                      description="Click Run backup now to create your first backup."
                    />
                  ) : null
                ) : (
                  <>
                    <BackupsList backups={backups} />
                    {backupsPageCount > 1 && (
                      <div className="flex items-center justify-between border-t border-border pt-3 text-sm text-muted-foreground">
                        <span className="tabular-nums">
                          {backupsPage * backupsPageSize + 1}–
                          {Math.min(
                            (backupsPage + 1) * backupsPageSize,
                            backupsTotal,
                          )}{" "}
                          of {backupsTotal}
                        </span>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setBackupsPage(backupsPage - 1)}
                            disabled={backupsPage === 0 || backupsLoading}
                          >
                            <ChevronLeft className="h-4 w-4" />
                            Prev
                          </Button>
                          <span className="tabular-nums">
                            Page {backupsPage + 1} of {backupsPageCount}
                          </span>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setBackupsPage(backupsPage + 1)}
                            disabled={
                              backupsPage + 1 >= backupsPageCount ||
                              backupsLoading
                            }
                          >
                            Next
                            <ChevronRight className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
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

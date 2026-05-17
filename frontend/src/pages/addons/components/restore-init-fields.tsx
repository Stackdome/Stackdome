import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import type { PostgresAddonFormValues } from "../schemas/form-schema";
import type { PostgresAddon } from "@/api/addons";
import type { ObjectStore } from "@/api/object-stores";
import { useSourceBackups } from "../hooks/use-source-backups";

type Initialization = PostgresAddonFormValues["initialization"];
type RestoreStore = Extract<
  Initialization,
  { type: "restore_from_object_store" }
>;
type RestoreBackup = Extract<
  Initialization,
  { type: "restore_from_backup" }
>;

type Props = {
  init: Initialization;
  restoreSources: PostgresAddon[];
  objectStores: ObjectStore[];
  errors: Partial<Record<string, string>>;
  onChange: (next: Initialization) => void;
};

const isRestore = (i: Initialization) =>
  i.type === "restore_from_object_store" || i.type === "restore_from_backup";

export function RestoreInitFields({
  init,
  restoreSources,
  objectStores,
  errors,
  onChange,
}: Props) {
  const sel = isRestore(init)
    ? (init as RestoreStore | RestoreBackup)
    : null;
  const sourceAddonId = sel?.sourceAddonId ?? "";
  const objectStoreId = sel?.objectStoreId ?? "";
  const resolvedStoreName =
    sel &&
    (objectStores.find((s) => s.id === objectStoreId)?.name ??
      objectStoreId);

  const selectedSource = restoreSources.find((a) => a.id === sourceAddonId);
  const sourceWalOn =
    selectedSource?.spec?.backup?.wal_archiving === true;

  const {
    backups: sourceBackups,
    loading: backupsLoading,
    error: backupsError,
  } = useSourceBackups(
    sel && !sourceWalOn && sourceAddonId ? sourceAddonId : undefined,
  );

  const storeArm = init.type === "restore_from_object_store" ? init : null;
  const backupArm = init.type === "restore_from_backup" ? init : null;

  const updateStore = (patch: Partial<Omit<RestoreStore, "type">>) =>
    onChange({
      type: "restore_from_object_store",
      sourceAddonId,
      objectStoreId,
      recoveryTargetTime: storeArm?.recoveryTargetTime,
      ...patch,
    });

  const updateBackup = (patch: Partial<Omit<RestoreBackup, "type">>) =>
    onChange({
      type: "restore_from_backup",
      sourceAddonId,
      objectStoreId,
      backupId: backupArm?.backupId ?? "",
      ...patch,
    });

  const onPickSource = (addonId: string) => {
    const src = restoreSources.find((a) => a.id === addonId);
    const osId = src?.spec?.backup?.object_store_id ?? "";
    if (src?.spec?.backup?.wal_archiving === true) {
      onChange({
        type: "restore_from_object_store",
        sourceAddonId: addonId,
        objectStoreId: osId,
      });
    } else {
      onChange({
        type: "restore_from_backup",
        sourceAddonId: addonId,
        objectStoreId: osId,
        backupId: "",
      });
    }
  };

  return (
    <>
      <FieldShell
        label="Create"
        htmlFor="init-mode"
        hint="Start empty, or restore from an existing addon's backups."
      >
        <Select
          value={isRestore(init) ? "restore" : "new"}
          onValueChange={(v) =>
            onChange(
              v === "restore"
                ? {
                  type: "restore_from_object_store",
                  sourceAddonId: "",
                  objectStoreId: "",
                }
                : { type: "new" },
            )
          }
        >
          <SelectTrigger id="init-mode" className="max-w-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="new">New empty database</SelectItem>
            <SelectItem value="restore">Restore from backup</SelectItem>
          </SelectContent>
        </Select>
      </FieldShell>

      {sel && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl">
          <FieldShell
            label="Source addon"
            htmlFor="restore-source"
            hint="Addons that have an object store for backups."
            error={errors["initialization.sourceAddonId"]}
          >
            {restoreSources.length === 0 ? (
              <p className="text-[12px] text-muted-foreground">
                No addon has backups in an object store to restore from.
              </p>
            ) : (
              <Select
                value={sourceAddonId || undefined}
                onValueChange={onPickSource}
              >
                <SelectTrigger id="restore-source">
                  <SelectValue placeholder="Select an addon" />
                </SelectTrigger>
                <SelectContent>
                  {restoreSources.map((a) => (
                    <SelectItem key={a.id ?? ""} value={a.id ?? ""}>
                      {a.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FieldShell>

          <FieldShell label="Object store">
            <div className="text-sm text-muted-foreground h-9 flex items-center">
              {resolvedStoreName || "—"}
            </div>
          </FieldShell>

          {sourceAddonId && sourceWalOn && storeArm && (
            <>
              <FieldShell
                label="Recover to"
                hint="Off restores to the latest archived WAL."
              >
                <div className="flex items-center gap-3 h-9">
                  <Switch
                    id="restore-pit"
                    checked={storeArm.recoveryTargetTime != null}
                    onCheckedChange={(on) =>
                      updateStore({
                        recoveryTargetTime: on
                          ? new Date().toISOString().slice(0, 19) + "Z"
                          : undefined,
                      })
                    }
                  />
                  <span className="text-sm text-foreground">
                    {storeArm.recoveryTargetTime != null
                      ? "Specific time (UTC)"
                      : "Latest"}
                  </span>
                </div>
              </FieldShell>

              {storeArm.recoveryTargetTime != null && (
                <FieldShell
                  label="Target time (UTC)"
                  htmlFor="restore-time"
                  error={errors["initialization.recoveryTargetTime"]}
                >
                  <Input
                    id="restore-time"
                    type="datetime-local"
                    className="font-mono"
                    value={storeArm.recoveryTargetTime?.replace(
                      /:\d{2}Z$/,
                      "",
                    )}
                    onChange={(e) =>
                      updateStore({
                        recoveryTargetTime: e.target.value
                          ? `${e.target.value}:00Z`
                          : undefined,
                      })
                    }
                  />
                </FieldShell>
              )}
            </>
          )}

          {sourceAddonId && !sourceWalOn && backupArm && (
            <FieldShell
              label="Backup"
              htmlFor="restore-backup"
              hint="WAL archiving was off, so point-in-time isn't available. Pick a completed backup."
              error={errors["initialization.backupId"]}
            >
              {backupsLoading ? (
                <p className="text-[12px] text-muted-foreground">Loading…</p>
              ) : backupsError ? (
                <p className="text-[12px] text-danger">{backupsError}</p>
              ) : sourceBackups.length === 0 ? (
                <p className="text-[12px] text-muted-foreground">
                  This addon has no completed backups to restore from.
                </p>
              ) : (
                <Select
                  value={backupArm?.backupId || undefined}
                  onValueChange={(backupId) => updateBackup({ backupId })}
                >
                  <SelectTrigger id="restore-backup">
                    <SelectValue placeholder="Select a backup" />
                  </SelectTrigger>
                  <SelectContent>
                    {sourceBackups.map((bk) => (
                      <SelectItem key={bk.id ?? ""} value={bk.id ?? ""}>
                        {bk.name ?? bk.id}
                        {bk.completed_at
                          ? ` · ${new Date(bk.completed_at).toLocaleString()}`
                          : ""}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FieldShell>
          )}
        </div>
      )}
    </>
  );
}

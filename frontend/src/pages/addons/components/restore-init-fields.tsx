import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
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

type Initialization = PostgresAddonFormValues["initialization"];
type RestoreInit = Extract<
  Initialization,
  { type: "restore_from_object_store" }
>;

type Props = {
  init: Initialization;
  restoreSources: PostgresAddon[];
  objectStores: ObjectStore[];
  errors: Partial<Record<string, string>>;
  onChange: (next: Initialization) => void;
};

export function RestoreInitFields({
  init,
  restoreSources,
  objectStores,
  errors,
  onChange,
}: Props) {
  const restoreSel =
    init.type === "restore_from_object_store" ? init : null;
  const resolvedStoreName =
    restoreSel &&
    (objectStores.find((s) => s.id === restoreSel.objectStoreId)?.name ??
      restoreSel.objectStoreId);

  const updateRestore = (
    patch: Partial<Omit<RestoreInit, "type">>,
  ) =>
    onChange({
      type: "restore_from_object_store",
      sourceAddonId: restoreSel?.sourceAddonId ?? "",
      objectStoreId: restoreSel?.objectStoreId ?? "",
      recoveryTargetTime: restoreSel?.recoveryTargetTime,
      ...patch,
    });

  return (
    <div className="mb-6 border-b border-border pb-6">
      <h3 className="text-sm font-semibold text-foreground mb-3">
        Initialization
      </h3>
      <RadioGroup
        value={
          init.type === "restore_from_object_store" ? "restore" : "new"
        }
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
        className="flex flex-col gap-2 max-w-3xl"
      >
        <label
          htmlFor="init-new"
          className="flex items-center gap-2 cursor-pointer text-sm text-foreground"
        >
          <RadioGroupItem id="init-new" value="new" />
          New empty database
        </label>
        <label
          htmlFor="init-restore"
          className="flex items-center gap-2 cursor-pointer text-sm text-foreground"
        >
          <RadioGroupItem id="init-restore" value="restore" />
          Restore from point in time
        </label>
      </RadioGroup>

      {restoreSel && (
        <div className="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl">
          <FieldShell
            label="Source addon"
            htmlFor="restore-source"
            hint="Only addons with an object store and WAL archiving."
            error={errors["initialization.sourceAddonId"]}
          >
            {restoreSources.length === 0 ? (
              <p className="text-[12px] text-muted-foreground">
                No addon has WAL-archived backups to restore from.
              </p>
            ) : (
              <Select
                value={restoreSel.sourceAddonId || undefined}
                onValueChange={(addonId) => {
                  const src = restoreSources.find((a) => a.id === addonId);
                  updateRestore({
                    sourceAddonId: addonId,
                    objectStoreId:
                      src?.spec?.backup?.object_store_id ?? "",
                  });
                }}
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

          <FieldShell
            label="Recover to"
            hint="Off restores to the latest archived WAL."
          >
            <div className="flex items-center gap-3 h-9">
              <Switch
                id="restore-pit"
                checked={restoreSel.recoveryTargetTime != null}
                onCheckedChange={(on) =>
                  updateRestore({
                    recoveryTargetTime: on
                      ? new Date().toISOString().slice(0, 19) + "Z"
                      : undefined,
                  })
                }
              />
              <span className="text-sm text-foreground">
                {restoreSel.recoveryTargetTime != null
                  ? "Specific time (UTC)"
                  : "Latest"}
              </span>
            </div>
          </FieldShell>

          {restoreSel.recoveryTargetTime != null && (
            <FieldShell
              label="Target time (UTC)"
              htmlFor="restore-time"
              error={errors["initialization.recoveryTargetTime"]}
            >
              <Input
                id="restore-time"
                type="datetime-local"
                className="font-mono"
                value={restoreSel.recoveryTargetTime?.replace(
                  /:\d{2}Z$/,
                  "",
                )}
                onChange={(e) =>
                  updateRestore({
                    recoveryTargetTime: e.target.value
                      ? `${e.target.value}:00Z`
                      : undefined,
                  })
                }
              />
            </FieldShell>
          )}
        </div>
      )}
    </div>
  );
}

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
import { Link } from "react-router-dom";
import { CronPresets } from "./cron-presets";
import type { BackupConfigFormValues } from "../schemas/backup-config-schema";
import type { ObjectStore } from "@/api/object-stores";

type Props = {
  values: BackupConfigFormValues;
  errors: Partial<Record<"objectStoreId" | "schedule", string>>;
  objectStores: ObjectStore[];
  storesLoading: boolean;
  onChange: (next: BackupConfigFormValues) => void;
};

export function BackupConfigFields({
  values,
  errors,
  objectStores,
  storesLoading,
  onChange,
}: Props) {
  const set = <K extends keyof BackupConfigFormValues>(
    k: K,
    val: BackupConfigFormValues[K],
  ) => onChange({ ...values, [k]: val });

  const noStores = !storesLoading && objectStores.length === 0;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl">
      <FieldShell
        label="Enable scheduled backups"
        htmlFor="bk-enabled"
        hint="When off, only manual backups can run."
      >
        <div className="flex items-center h-10">
          <Switch
            id="bk-enabled"
            checked={values.enabled}
            onCheckedChange={(c) => set("enabled", c)}
          />
        </div>
      </FieldShell>

      <FieldShell
        label="WAL archiving"
        htmlFor="bk-wal"
        hint="Continuously ships WAL segments. Required for point-in-time recovery."
      >
        <div className="flex items-center h-10">
          <Switch
            id="bk-wal"
            checked={values.walArchiving}
            onCheckedChange={(c) => set("walArchiving", c)}
          />
        </div>
      </FieldShell>

      <FieldShell
        label="Object Store"
        htmlFor="bk-objstore"
        error={errors.objectStoreId}
        hint={
          noStores ? (
            <>
              No Object Stores yet.{" "}
              <Link
                to="/object-stores"
                target="_blank"
                rel="noreferrer noopener"
                className="text-brand hover:underline"
              >
                Create one
              </Link>{" "}
              to enable backups.
            </>
          ) : undefined
        }
      >
        <Select
          value={values.objectStoreId}
          onValueChange={(v) => set("objectStoreId", v)}
          disabled={storesLoading || noStores}
        >
          <SelectTrigger id="bk-objstore">
            <SelectValue
              placeholder={storesLoading ? "Loading…" : "Select a destination"}
            />
          </SelectTrigger>
          <SelectContent>
            {objectStores
              .filter((s): s is typeof s & { id: string } => !!s.id)
              .map((s) => (
                <SelectItem key={s.id} value={s.id}>
                  {s.name}
                </SelectItem>
              ))}
          </SelectContent>
        </Select>
      </FieldShell>

      <FieldShell
        label="Schedule (6-field Quartz cron)"
        htmlFor="bk-schedule"
        error={errors.schedule}
      >
        <Input
          id="bk-schedule"
          className="font-mono"
          value={values.schedule}
          disabled={!values.enabled}
          onChange={(e) => set("schedule", e.target.value)}
          placeholder="0 0 3 * * *"
        />
        <div className="mt-2">
          <CronPresets
            value={values.schedule}
            disabled={!values.enabled}
            onChange={(expr) => set("schedule", expr)}
          />
        </div>
      </FieldShell>
    </div>
  );
}

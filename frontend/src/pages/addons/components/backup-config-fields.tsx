import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
    <div className="flex flex-col gap-5 max-w-3xl">
      <div className="flex items-center justify-between">
        <div>
          <Label htmlFor="bk-enabled">Enable scheduled backups</Label>
          <p className="text-xs text-muted-foreground">
            When off, only manual backups can run.
          </p>
        </div>
        <Switch
          id="bk-enabled"
          checked={values.enabled}
          onCheckedChange={(c) => set("enabled", c)}
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="bk-objstore">Object Store</Label>
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
        {noStores && (
          <p className="text-xs text-muted-foreground">
            No Object Stores yet.{" "}
            <Link to="/object-stores" className="text-brand hover:underline">
              Create one
            </Link>{" "}
            to enable backups.
          </p>
        )}
        {errors.objectStoreId && (
          <p className="text-xs text-danger">{errors.objectStoreId}</p>
        )}
      </div>

      <div className="grid gap-2">
        <Label htmlFor="bk-schedule">Schedule (6-field Quartz cron)</Label>
        <Input
          id="bk-schedule"
          className="font-mono"
          value={values.schedule}
          disabled={!values.enabled}
          onChange={(e) => set("schedule", e.target.value)}
          placeholder="0 0 3 * * *"
        />
        <CronPresets
          value={values.schedule}
          disabled={!values.enabled}
          onChange={(expr) => set("schedule", expr)}
        />
        {errors.schedule && (
          <p className="text-xs text-danger">{errors.schedule}</p>
        )}
      </div>

      <div className="flex items-center justify-between">
        <div>
          <Label htmlFor="bk-wal">WAL archiving</Label>
          <p className="text-xs text-muted-foreground">
            Continuously ships WAL segments. Required for point-in-time recovery.
          </p>
        </div>
        <Switch
          id="bk-wal"
          checked={values.walArchiving}
          onCheckedChange={(c) => set("walArchiving", c)}
        />
      </div>
    </div>
  );
}

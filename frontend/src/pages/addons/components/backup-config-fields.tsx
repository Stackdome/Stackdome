import { useEffect, useRef, useState } from "react";
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
import cronstrue from "cronstrue";
import {
  buildCron,
  parseCron,
  normalizeCron,
  isValidCronArity,
  type Frequency,
  type ScheduleParts,
} from "../lib/cron-builder";
import type { BackupConfigFormValues } from "../schemas/backup-config-schema";
import type { ObjectStore } from "@/api/object-stores";

const FREQUENCIES: { value: Frequency; label: string }[] = [
  { value: "hourly", label: "Hourly" },
  { value: "daily", label: "Daily" },
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" },
  { value: "custom", label: "Custom (cron)" },
];

const WEEKDAYS = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

const pad = (n: number) => String(n).padStart(2, "0");

function describeCron(expr: string): string | null {
  try {
    return cronstrue.toString(expr, { use24HourTimeFormat: true });
  } catch {
    return null;
  }
}

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
  const disabled = !values.enabled;
  const parsed = parseCron(values.schedule);
  const scheduleArityOk = isValidCronArity(values.schedule);
  const description = scheduleArityOk ? describeCron(values.schedule) : null;
  const scheduleError =
    errors.schedule ||
    (values.schedule.trim() && !scheduleArityOk
      ? "Use 6 space-separated fields: sec min hour day-of-month month day-of-week (or @daily)."
      : undefined);

  // `frequency` is a UI mode that must survive even when the resulting cron
  // happens to match a builder shape (otherwise "Custom" snaps back). It only
  // re-syncs from the schedule on *external* changes (e.g. edit hydration),
  // tracked via the last value this component emitted.
  const lastEmitted = useRef<string | null>(null);
  const [mode, setMode] = useState<Frequency>(() => parsed.frequency);

  useEffect(() => {
    if (values.schedule !== lastEmitted.current) {
      setMode(parseCron(values.schedule).frequency);
      lastEmitted.current = values.schedule;
    }
  }, [values.schedule]);

  const parts: ScheduleParts = { ...parsed, frequency: mode };

  const clamp = (raw: string, min: number, max: number) => {
    const n = Number(raw);
    if (!Number.isFinite(n)) return min;
    return Math.min(max, Math.max(min, Math.trunc(n)));
  };

  const emit = (schedule: string) => {
    lastEmitted.current = schedule;
    onChange({ ...values, schedule });
  };

  const applyParts = (next: ScheduleParts) => {
    setMode(next.frequency);
    if (next.frequency === "custom") {
      // Switching into Custom keeps the current expression for editing.
      emit(parsed.custom);
    } else {
      emit(buildCron(next));
    }
  };

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
                className="font-medium text-foreground underline-offset-2 hover:underline"
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
        label="Schedule"
        htmlFor="bk-frequency"
        error={scheduleError}
        hint={
          mode === "custom" && description ? (
            <span>
              {description} <span className="text-muted-foreground/70">(UTC)</span>
            </span>
          ) : undefined
        }
      >
        <div className="flex flex-col gap-2">
          <Select
            value={parts.frequency}
            onValueChange={(v) => applyParts({ ...parts, frequency: v as Frequency })}
            disabled={disabled}
          >
            <SelectTrigger id="bk-frequency">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {FREQUENCIES.map((f) => (
                <SelectItem key={f.value} value={f.value}>
                  {f.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {parts.frequency === "hourly" && (
            <label className="flex items-center gap-2 text-sm text-muted-foreground">
              at minute
              <Input
                type="number"
                min={0}
                max={59}
                className="w-20 font-mono"
                value={parts.minute}
                disabled={disabled}
                onChange={(e) =>
                  applyParts({ ...parts, minute: clamp(e.target.value, 0, 59) })
                }
              />
            </label>
          )}

          {(parts.frequency === "daily" ||
            parts.frequency === "weekly" ||
            parts.frequency === "monthly") && (
            <div className="flex flex-wrap items-center gap-2">
              {parts.frequency === "weekly" && (
                <Select
                  value={String(parts.dayOfWeek)}
                  onValueChange={(v) =>
                    applyParts({ ...parts, dayOfWeek: Number(v) })
                  }
                  disabled={disabled}
                >
                  <SelectTrigger className="w-36">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {WEEKDAYS.map((d, i) => (
                      <SelectItem key={d} value={String(i)}>
                        {d}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              {parts.frequency === "monthly" && (
                <label className="flex items-center gap-2 text-sm text-muted-foreground">
                  day
                  <Input
                    type="number"
                    min={1}
                    max={31}
                    className="w-20 font-mono"
                    value={parts.dayOfMonth}
                    disabled={disabled}
                    onChange={(e) =>
                      applyParts({
                        ...parts,
                        dayOfMonth: clamp(e.target.value, 1, 31),
                      })
                    }
                  />
                </label>
              )}
              <label className="flex items-center gap-2 text-sm text-muted-foreground">
                at
                <Input
                  type="time"
                  className="w-32 font-mono"
                  value={`${pad(parts.hour)}:${pad(parts.minute)}`}
                  disabled={disabled}
                  onChange={(e) => {
                    const [h, m] = e.target.value.split(":");
                    applyParts({
                      ...parts,
                      hour: clamp(h, 0, 23),
                      minute: clamp(m, 0, 59),
                    });
                  }}
                />
              </label>
            </div>
          )}

          {parts.frequency === "custom" && (
            <Input
              id="bk-schedule"
              className="font-mono"
              value={values.schedule}
              disabled={disabled}
              onChange={(e) => emit(e.target.value)}
              onBlur={(e) => emit(normalizeCron(e.target.value))}
              placeholder="0 0 3 * * *  (sec min hour dom mon dow)"
            />
          )}
        </div>
      </FieldShell>
    </div>
  );
}

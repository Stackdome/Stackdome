import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
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
import { useToast } from "@/components/ui/use-toast";
import { Panel } from "@/components/branded";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useObjectStores } from "@/pages/object-stores/hooks/use-object-stores";
import {
  backupConfigSchema,
  toApiBackupConfig,
  type BackupConfigFormValues,
} from "../schemas/backup-config-schema";
import type { PostgresAddon } from "@/api/addons";
import { updatePostgresAddon } from "@/api/addons";
import { CronPresets } from "./cron-presets";

type Props = {
  addon: PostgresAddon;
  onSaved: () => void;
};

const DEFAULT_SCHEDULE = "0 0 3 * * *";

function fromAddon(addon: PostgresAddon): BackupConfigFormValues {
  const b = addon.spec?.backup;
  return {
    enabled: b?.enabled ?? false,
    objectStoreId: b?.object_store_id ?? "",
    schedule: b?.schedule || DEFAULT_SCHEDULE,
    walArchiving: b?.wal_archiving ?? false,
  };
}

export function BackupConfigEditor({ addon, onSaved }: Props) {
  const { toast } = useToast();
  const { objectStores, loading: storesLoading } = useObjectStores();
  const [values, setValues] = useState<BackupConfigFormValues>(fromAddon(addon));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setValues(fromAddon(addon));
    setErrors({});
  }, [addon]);

  function clearError(key: string) {
    setErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  async function handleSave() {
    const parsed = backupConfigSchema.safeParse(values);
    if (!parsed.success) {
      const errs: Record<string, string> = {};
      for (const issue of parsed.error.issues) errs[issue.path.join(".")] = issue.message;
      setErrors(errs);
      return;
    }
    const orgId = getCurrentOrganizationId();
    if (!orgId || !addon.id) {
      toast({
        title: "Cannot save",
        description: "Missing org or addon id",
        variant: "destructive",
      });
      return;
    }
    setSubmitting(true);
    try {
      const apiBackup = toApiBackupConfig(parsed.data);
      const payload = {
        name: addon.name,
        spec: { ...addon.spec, backup: apiBackup },
      };
      await updatePostgresAddon(orgId, addon.id, payload);
      toast({ title: "Backup configuration saved" });
      onSaved();
    } catch (e) {
      toast({
        title: "Failed to save backup configuration",
        description: getErrorMessage(e),
        variant: "destructive",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Panel title="Backup configuration">
      <div className="flex flex-col gap-5">
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
            onCheckedChange={(checked) => {
              clearError("enabled");
              setValues((v) => ({ ...v, enabled: checked }));
            }}
          />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="bk-objstore">Object Store</Label>
          <Select
            value={values.objectStoreId}
            onValueChange={(v) => {
              clearError("objectStoreId");
              setValues((vs) => ({ ...vs, objectStoreId: v }));
            }}
            disabled={storesLoading}
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
              {!storesLoading && objectStores.length === 0 && (
                <div className="px-3 py-2 text-xs text-muted-foreground">
                  No Object Stores. Create one first.
                </div>
              )}
            </SelectContent>
          </Select>
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
            onChange={(e) => {
              clearError("schedule");
              setValues((v) => ({ ...v, schedule: e.target.value }));
            }}
            placeholder="0 0 3 * * *"
          />
          <CronPresets
            value={values.schedule}
            disabled={!values.enabled}
            onChange={(expr) => {
              clearError("schedule");
              setValues((v) => ({ ...v, schedule: expr }));
            }}
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
            onCheckedChange={(checked) =>
              setValues((v) => ({ ...v, walArchiving: checked }))
            }
          />
        </div>

        <div className="flex justify-end">
          <Button onClick={handleSave} disabled={submitting}>
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save backup configuration
          </Button>
        </div>
      </div>
    </Panel>
  );
}

import { z } from "zod";
import { schemas } from "@/api/zod-schemas";

type ApiPostgresBackupConfig = z.infer<typeof schemas.PostgresBackupConfig>;

// 6-field Quartz cron (seconds minutes hours day-of-month month day-of-week).
// Each field is non-whitespace; we leave deep semantic validation to the backend.
const quartzCronRegex = /^\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+$/;

export const backupConfigSchema = z
  .object({
    enabled: z.boolean(),
    objectStoreId: z.string(),
    schedule: z
      .string()
      .regex(quartzCronRegex, { message: "Use a 6-field Quartz cron (sec min hour dom mon dow)" }),
    walArchiving: z.boolean(),
  })
  .superRefine((val, ctx) => {
    if (val.enabled && !val.objectStoreId) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["objectStoreId"],
        message: "Pick an Object Store to enable scheduled backups",
      });
    }
  });

export type BackupConfigFormValues = z.infer<typeof backupConfigSchema>;

export function toApiBackupConfig(values: BackupConfigFormValues): ApiPostgresBackupConfig {
  return {
    enabled: values.enabled,
    schedule: values.schedule,
    wal_archiving: values.walArchiving,
    ...(values.objectStoreId ? { object_store_id: values.objectStoreId } : {}),
  } satisfies ApiPostgresBackupConfig;
}

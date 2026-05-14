import { z } from "zod";
import type { components } from "@/api/types/openapi";

// Anchor the converter return type against the openapi-typescript-generated type
// instead of the zod-schemas one. `openapi-zod-client` makes everything
// `.partial().passthrough()`, which over-loosens required fields; the strict
// shape from openapi.d.ts matches what `PostgresAddonCreateInput["spec"]["backup"]`
// expects at the API boundary. The runtime form validation still goes through
// `backupConfigSchema` below — only the converter's output type is anchored here.
type ApiPostgresBackupConfig = NonNullable<components["schemas"]["PostgresBackupConfig"]>;

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

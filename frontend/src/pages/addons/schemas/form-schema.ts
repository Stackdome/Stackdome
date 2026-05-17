import { z } from "zod";
import { schemas } from "@/api/zod-schemas";
import type { PlanId } from "../lib/plan-presets";

const NAME_PATTERN = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;

// 6-field Quartz cron (sec min hour dom mon dow). Deep semantics left to backend.
const QUARTZ_CRON = /^\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+$/;

const BackupFormSchema = z.object({
  enabled: z.boolean(),
  objectStoreId: z.string(),
  schedule: z
    .string()
    .regex(QUARTZ_CRON, "Use a 6-field Quartz cron (sec min hour dom mon dow)"),
  walArchiving: z.boolean(),
});

// Variant B: extend the generated PostgresDatabase shape with form-level
// validation (required name, required extensions array). If the backend
// adds/removes an extension literal, TS will flag this site.
const DatabaseFormItemSchema = schemas.PostgresDatabase.extend({
  name: z.string().min(1, "Required"),
  extensions: z.array(z.literal("vector")),
});

const PostgresAddonFormBase = z.object({
  name: z
    .string()
    .min(1, "Required")
    .max(63, "Name must be 63 characters or fewer")
    .regex(
      NAME_PATTERN,
      "Use lowercase letters, numbers, and hyphens; must start and end with a letter or number",
    ),
  clusterId: z.string(),
  plan: z.enum(["basic", "starter", "launch", "scale", "performance", "custom"]),
  customCpuRequest: z.string().optional(),
  customCpuLimit: z.string().optional(),
  customMemoryRequest: z.string().optional(),
  customMemoryLimit: z.string().optional(),
  storageGB: z.number().int().min(1, "Storage must be at least 1 GB"),
  versionMajor: z.number().int(),
  highAvailability: z.boolean(),
  replicasOverride: z.number().int().min(1).optional(),
  autoMinorUpgrade: z.boolean(),
  autoMajorUpgrade: z.boolean(),
  superuserAccess: z.boolean(),
  databases: z.array(DatabaseFormItemSchema),
  initialization: z.discriminatedUnion("type", [
    z.object({ type: z.literal("new") }),
    z.object({
      type: z.literal("restore_from_backup"),
      backupId: z.string().min(1, "Pick a backup"),
      // UI context for the source picker; backend only consumes backupId.
      sourceAddonId: z.string().min(1, "Pick a source addon"),
      objectStoreId: z.string().optional(),
    }),
    z.object({
      type: z.literal("restore_from_object_store"),
      sourceAddonId: z.string().min(1, "Pick a source addon"),
      objectStoreId: z.string().min(1, "Object store could not be resolved"),
      // ISO-8601; absent = restore to latest. datetime-local input is
      // serialised to `${value}:00Z` before it lands here.
      recoveryTargetTime: z
        .string()
        .datetime({ offset: true })
        .optional(),
    }),
  ]),
  advancedJson: z.string(),
  backup: BackupFormSchema,
});

export const PostgresAddonFormSchema = PostgresAddonFormBase.superRefine((val, ctx) => {
  if (val.backup.enabled && !val.backup.objectStoreId) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["backup", "objectStoreId"],
      message: "Pick an Object Store to enable scheduled backups",
    });
  }
});

export type PostgresAddonFormValues = z.infer<typeof PostgresAddonFormSchema>;

export const DEFAULT_ADVANCED_JSON = "";

export function defaultFormValues(clusterId: string): PostgresAddonFormValues {
  return {
    name: "",
    clusterId,
    plan: "basic" as PlanId,
    storageGB: 10,
    versionMajor: 17,
    highAvailability: false,
    autoMinorUpgrade: true,
    autoMajorUpgrade: false,
    superuserAccess: false,
    databases: [],
    initialization: { type: "new" },
    advancedJson: DEFAULT_ADVANCED_JSON,
    backup: {
      enabled: false,
      objectStoreId: "",
      schedule: "0 0 3 * * *",
      walArchiving: false,
    },
  };
}

import { z } from "zod";
import { schemas } from "@/api/zod-schemas";
import type { PlanId } from "../lib/plan-presets";

const NAME_PATTERN = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;

// Variant B: extend the generated PostgresDatabase shape with form-level
// validation (required name, required extensions array). If the backend
// adds/removes an extension literal, TS will flag this site.
const DatabaseFormItemSchema = schemas.PostgresDatabase.extend({
  name: z.string().min(1, "Required"),
  extensions: z.array(z.literal("vector")),
});

export const PostgresAddonFormSchema = z.object({
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
    }),
  ]),
  advancedJson: z.string(),
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
  };
}

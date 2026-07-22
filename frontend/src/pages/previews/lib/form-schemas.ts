import { z } from "zod";
import { isOverrideLine } from "@/pages/previews/lib/parse-image-overrides";

/** Full 40-char or short (min 7) hex commit SHA, case-insensitive. */
const COMMIT_SHA_REGEX = /^[0-9a-f]{7,40}$/i;

/** Applied when the user leaves "Stackfile path" blank. */
export const DEFAULT_STACKFILE_PATH = "stackfile.yaml";

/** True when every non-blank line matches the `resource=image` shape —
 *  checked per line (via `isOverrideLine`, the same rule `parseImageOverrides`
 *  uses) rather than by comparing line count to parsed-key count, since
 *  duplicate resource keys collapse to one key and would otherwise be
 *  wrongly rejected. */
function isValidOverridesText(text: string): boolean {
  return text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .every(isOverrideLine);
}

/** Shared by every "Image overrides" textarea: optional, but any non-blank
 *  line must be a well-formed `resource=image` pair. */
export const overridesTextSchema = z.string().refine(isValidOverridesText, {
  message: "Each line must be resource=image (e.g. web=registry/web:tag).",
});

/** One row of the env-vars editor. Callers filter out blank-named rows
 *  before validating, so by the time this schema sees a row its name is
 *  expected to be non-empty; the check here just guards that invariant. */
const envVarRowSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  value: z.string().refine((v) => !/^\{\{\s*secret\.\s*\}\}$/.test(v), {
    message: "Pick a secret for the Secret-sourced variable.",
  }),
});
export type EnvVarFormRow = z.infer<typeof envVarRowSchema>;

function duplicateEnvName(rows: EnvVarFormRow[]): string | undefined {
  const seen = new Set<string>();
  for (const row of rows) {
    if (seen.has(row.name)) return row.name;
    seen.add(row.name);
  }
  return undefined;
}

/** Env-var overrides shared by the configure phase and the settings modal:
 *  every row needs a name, and names must be unique. */
export const envVarsSchema = z.array(envVarRowSchema).refine(
  (rows) => duplicateEnvName(rows) === undefined,
  (rows) => ({ message: `Duplicate variable name: ${duplicateEnvName(rows)}.` }),
);

/** Configure phase of the enable-previews wizard. */
export const configurePhaseSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  baseBranch: z.string().trim().min(1, "Base branch is required."),
  stackfilePath: z.string().trim().optional(),
  maxActive: z.coerce.number().int("Must be a whole number.").min(1, "Must be at least 1."),
  env: envVarsSchema.optional(),
});
export type ConfigurePhaseValues = z.infer<typeof configurePhaseSchema>;

/** Repository settings modal for an existing preview config. Stackfile path
 *  is required here (unlike the configure phase) — an empty path would
 *  break every preview built from this config. */
export const configSettingsSchema = z.object({
  baseBranch: z.string().trim().min(1, "Base branch is required."),
  stackfilePath: z.string().trim().min(1, "Stackfile path is required."),
  maxActive: z.coerce.number().int("Must be a whole number.").min(1, "Must be at least 1."),
  env: envVarsSchema.optional(),
});
export type ConfigSettingsValues = z.infer<typeof configSettingsSchema>;

/** New preview environment modal. */
export const newPreviewEnvSchema = z.object({
  prNumber: z
    .string()
    .trim()
    .min(1, "PR number is required.")
    .regex(/^\d+$/, "PR number must be digits only."),
  branch: z.string().trim().min(1, "Branch is required."),
  overridesText: overridesTextSchema,
});
export type NewPreviewEnvValues = z.infer<typeof newPreviewEnvSchema>;

/** Sync preview environment dialog — everything is optional; the commit,
 *  when present, must look like a real SHA. */
export const syncEnvSchema = z.object({
  commit: z.string().trim().refine((v) => v === "" || COMMIT_SHA_REGEX.test(v), {
    message: "Enter a valid commit SHA (7-40 hex characters).",
  }),
  overridesText: overridesTextSchema,
});
export type SyncEnvValues = z.infer<typeof syncEnvSchema>;

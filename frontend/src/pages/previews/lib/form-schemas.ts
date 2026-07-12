import { z } from "zod";
import { parseImageOverrides } from "@/pages/previews/lib/parse-image-overrides";

/** Full 40-char or short (min 7) hex commit SHA, case-insensitive. */
const COMMIT_SHA_REGEX = /^[0-9a-f]{7,40}$/i;

/** True when every non-blank line parses as `resource=image` —
 *  i.e. `parseImageOverrides` didn't have to silently drop a line. */
function isValidOverridesText(text: string): boolean {
  const nonBlankLines = text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean);
  if (nonBlankLines.length === 0) return true;
  const parsed = parseImageOverrides(text);
  return Object.keys(parsed ?? {}).length === nonBlankLines.length;
}

/** Shared by every "Image overrides" textarea: optional, but any non-blank
 *  line must be a well-formed `resource=image` pair. */
export const overridesTextSchema = z.string().refine(isValidOverridesText, {
  message: "Each line must be resource=image (e.g. web=registry/web:tag).",
});

/** Configure phase of the enable-previews wizard. */
export const configurePhaseSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  baseBranch: z.string().trim().min(1, "Base branch is required."),
  stackfilePath: z.string().trim().optional(),
  maxActive: z.coerce.number().int("Must be a whole number.").min(1, "Must be at least 1."),
});
export type ConfigurePhaseValues = z.infer<typeof configurePhaseSchema>;

/** Repository settings modal for an existing preview config. Stackfile path
 *  is required here (unlike the configure phase) — an empty path would
 *  break every preview built from this config. */
export const configSettingsSchema = z.object({
  baseBranch: z.string().trim().min(1, "Base branch is required."),
  stackfilePath: z.string().trim().min(1, "Stackfile path is required."),
  maxActive: z.coerce.number().int("Must be a whole number.").min(1, "Must be at least 1."),
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

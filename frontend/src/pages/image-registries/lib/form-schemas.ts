import { z } from "zod";

/** Password is write-only in the API — required on create, never echoed back. */
export const createRegistrySchema = z.object({
  host: z.string().trim().min(1, "Host is required."),
  username: z.string().trim().min(1, "Username is required."),
  password: z.string().trim().min(1, "Password is required."),
});
export type CreateRegistryValues = z.infer<typeof createRegistrySchema>;

/** Rotation form: host/purpose are immutable, only the login pair changes. */
export const rotateRegistrySchema = createRegistrySchema.omit({ host: true });
export type RotateRegistryValues = z.infer<typeof rotateRegistrySchema>;

/** Verify form: repository to check access against. Path-only values
 *  ("acme/app") are allowed — the dialog prefixes the credential's host
 *  before calling the API, which requires a fully-qualified reference. */
export const verifyRegistrySchema = z.object({
  repository: z.string().trim().min(1, "Repository is required."),
});
export type VerifyRegistryValues = z.infer<typeof verifyRegistrySchema>;

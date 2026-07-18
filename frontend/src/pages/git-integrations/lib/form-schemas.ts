import { z } from "zod";

/** Credentials phase of the add-integration wizard: host + token are
 *  required for every provider; username is only needed for basic-auth
 *  hosts (e.g. Bitbucket app passwords) so it stays optional. */
export const credentialsFormSchema = z.object({
  host: z.string().trim().min(1, "Host is required."),
  username: z.string().trim().optional(),
  token: z.string().trim().min(1, "Access token is required."),
});
export type CredentialsFormValues = z.infer<typeof credentialsFormSchema>;

/** Credential-rotation dialog: same auth model as the wizard's credentials
 *  phase but without host — host is immutable on update. */
export const updateCredentialsFormSchema = credentialsFormSchema.omit({ host: true });
export type UpdateCredentialsFormValues = z.infer<typeof updateCredentialsFormSchema>;

function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

/** Verify-repository-access dialog: the URL must be non-empty and parse as
 *  an http(s) URL — the backend clones it, so anything else can't work. */
export const verifyIntegrationFormSchema = z.object({
  repoUrl: z
    .string()
    .trim()
    .min(1, "Repository URL is required.")
    .refine(isHttpUrl, "Enter a valid http(s) URL (e.g. https://github.com/acme/webapp)."),
});
export type VerifyIntegrationFormValues = z.infer<typeof verifyIntegrationFormSchema>;

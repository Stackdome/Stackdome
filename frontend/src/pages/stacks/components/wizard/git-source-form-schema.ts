import { z } from "zod";

/** Service form of the git-source wizard panel. `dockerfilePath` and
 *  `buildContext` are optional here — `buildGitSeed` falls back to
 *  "Dockerfile" / "." when they're blank, matching the API defaults. */
export const gitSourceFormSchema = z.object({
  serviceName: z.string().trim().min(1, "Service name is required."),
  branch: z.string().trim().min(1, "Branch is required."),
  port: z
    .string()
    .trim()
    .min(1, "Port is required.")
    .regex(/^\d+$/, "Port must be a whole number.")
    .refine((v) => Number(v) > 0 && Number(v) < 65536, "Port must be between 1 and 65535."),
  dockerfilePath: z.string().trim().optional(),
  buildContext: z.string().trim().optional(),
});
export type GitSourceFormValues = z.infer<typeof gitSourceFormSchema>;

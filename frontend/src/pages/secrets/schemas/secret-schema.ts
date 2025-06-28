import { z } from "zod";

export const SecretTypeSchema = z.enum([
  "Generic",
  "DockerRegistry",
  "GitCredentials",
  "UsernamePassword",
  "Token",
  "SSHKey"
]);

export const SecretDataSchema = z.object({
  key: z.string().min(1, "Key is required"),
  value: z.string().min(1, "Value is required"),
});

const BaseSecretSchema = z.object({
  name: z.string().min(1, "Secret name is required").max(255, "Name too long"),
  description: z.string().optional(),
});

export const GenericSecretSchema = BaseSecretSchema.extend({
  type: z.literal("Generic"),
  data: z.array(SecretDataSchema).min(1, "At least one key-value pair is required"),
});

export const DockerRegistrySecretSchema = BaseSecretSchema.extend({
  type: z.literal("DockerRegistry"),
  registry: z.string().min(1, "Registry URL is required"),
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

export const GitCredentialsSecretSchema = BaseSecretSchema.extend({
  type: z.literal("GitCredentials"),
  username: z.string().optional(),
  password: z.string().optional(),
  token: z.string().optional(),
});

export const UsernamePasswordSecretSchema = BaseSecretSchema.extend({
  type: z.literal("UsernamePassword"),
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

export const TokenSecretSchema = BaseSecretSchema.extend({
  type: z.literal("Token"),
  token: z.string().min(8, "Token must be at least 8 characters long"),
});

export const SSHKeySecretSchema = BaseSecretSchema.extend({
  type: z.literal("SSHKey"),
  sshPrivateKey: z.string()
    .min(1, "SSH private key is required")
    .refine(
      (key) => key.includes("-----BEGIN") && key.includes("-----END"),
      "SSH private key must be in PEM format"
    ),
});

// Union schema for form validation
export const SecretFormSchema = z.discriminatedUnion("type", [
  GenericSecretSchema,
  DockerRegistrySecretSchema,
  GitCredentialsSecretSchema,
  UsernamePasswordSecretSchema,
  TokenSecretSchema,
  SSHKeySecretSchema,
]).refine(
  (data) => {
    // Additional validation for GitCredentials
    if (data.type === "GitCredentials") {
      const hasUsernamePassword = data.username?.trim() && data.password?.trim();
      const hasToken = data.token?.trim();
      return hasUsernamePassword || hasToken;
    }
    return true;
  },
  {
    message: "Either username/password or token is required",
    path: ["credentials"], // This will show the error under a "credentials" field
  }
);

export const SecretApiSchema = z.object({
  id: z.string().optional(),
  name: z.string(),
  description: z.string().optional(),
  organisation_id: z.string().optional(),
  type: SecretTypeSchema,
  data: z.array(SecretDataSchema),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export type SecretFormData = z.infer<typeof SecretFormSchema>;
export type SecretApiData = z.infer<typeof SecretApiSchema>;
export type SecretType = z.infer<typeof SecretTypeSchema>;

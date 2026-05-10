import { z } from "zod";
import type { ObjectStoreCreatePayload } from "../types";

const retentionPolicyRegex = /^[1-9]\d*[dhw]$/;

const secretReferenceSchema = z.object({
  secret_id: z.string().uuid({ message: "Pick a secret" }),
  key: z.string().min(1, { message: "Pick a key from the secret" }),
});

const s3Schema = z.object({
  region: z.string().min(1, { message: "Region is required" }),
  endpointUrl: z.string().optional().default(""),
  accessKeyId: secretReferenceSchema,
  secretAccessKey: secretReferenceSchema,
});

const azureSchema = z.object({
  storageAccountName: z.string().optional().default(""),
  connectionString: secretReferenceSchema,
});

const gcsSchema = z.object({
  serviceAccountCredentials: secretReferenceSchema,
});

export const objectStoreFormSchema = z
  .object({
    name: z.string().min(1, { message: "Name is required" }),
    destinationPath: z.string().min(1, { message: "Destination path is required" }),
    retentionPolicy: z
      .string()
      .regex(retentionPolicyRegex, { message: "Use a value like 7d, 24h, 4w (no zeros)" }),
    provider: z.enum(["s3", "azure", "gcs"]),
    s3: s3Schema.optional(),
    azure: azureSchema.optional(),
    gcs: gcsSchema.optional(),
  })
  .superRefine((val, ctx) => {
    if (val.provider === "s3" && !val.s3) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["s3"], message: "Required" });
    }
    if (val.provider === "azure" && !val.azure) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["azure"], message: "Required" });
    }
    if (val.provider === "gcs" && !val.gcs) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["gcs"], message: "Required" });
    }
  });

export type ObjectStoreFormValues = z.infer<typeof objectStoreFormSchema>;

export function toApiPayload(values: ObjectStoreFormValues): ObjectStoreCreatePayload {
  const configuration: ObjectStoreCreatePayload["spec"]["configuration"] = {};
  if (values.provider === "s3" && values.s3) {
    configuration.s3_credentials = {
      region: values.s3.region,
      access_key_id: values.s3.accessKeyId,
      secret_access_key: values.s3.secretAccessKey,
      ...(values.s3.endpointUrl.trim() ? { endpoint_url: values.s3.endpointUrl.trim() } : {}),
    };
  }
  if (values.provider === "azure" && values.azure) {
    configuration.azure_credentials = {
      connection_string: values.azure.connectionString,
      ...(values.azure.storageAccountName.trim()
        ? { storage_account_name: values.azure.storageAccountName.trim() }
        : {}),
    };
  }
  if (values.provider === "gcs" && values.gcs) {
    configuration.gcs_credentials = {
      service_account_credentials: values.gcs.serviceAccountCredentials,
    };
  }

  return {
    name: values.name,
    spec: {
      configuration,
      destination_path: values.destinationPath,
      retention_policy: values.retentionPolicy,
    },
  };
}

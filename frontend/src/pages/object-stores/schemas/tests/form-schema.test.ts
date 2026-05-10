import { describe, it, expect } from "vitest";
import {
  objectStoreFormSchema,
  toApiPayload,
  type ObjectStoreFormValues,
} from "../form-schema";

const validS3Form: ObjectStoreFormValues = {
  name: "minio-local",
  destinationPath: "s3://stackdome-backups/db1",
  retentionPolicy: "7d",
  provider: "s3",
  s3: {
    region: "us-east-1",
    endpointUrl: "http://localhost:9000",
    accessKeyId: { secret_id: "11111111-1111-1111-1111-111111111111", key: "accessKeyId" },
    secretAccessKey: { secret_id: "11111111-1111-1111-1111-111111111111", key: "secretAccessKey" },
  },
};

describe("objectStoreFormSchema", () => {
  it("accepts a valid S3 form", () => {
    const result = objectStoreFormSchema.safeParse(validS3Form);
    expect(result.success).toBe(true);
  });

  it("rejects empty name", () => {
    const result = objectStoreFormSchema.safeParse({ ...validS3Form, name: "" });
    expect(result.success).toBe(false);
  });

  it("rejects S3 form without an access key reference", () => {
    const result = objectStoreFormSchema.safeParse({
      ...validS3Form,
      s3: { ...validS3Form.s3!, accessKeyId: { secret_id: "", key: "" } },
    });
    expect(result.success).toBe(false);
  });

  it("rejects unknown retention format", () => {
    const result = objectStoreFormSchema.safeParse({ ...validS3Form, retentionPolicy: "forever" });
    expect(result.success).toBe(false);
  });

  it("rejects s3 provider with no s3 block", () => {
    const { s3, ...rest } = validS3Form;
    const result = objectStoreFormSchema.safeParse(rest);
    expect(result.success).toBe(false);
  });

  it("rejects retention of 0d", () => {
    const result = objectStoreFormSchema.safeParse({ ...validS3Form, retentionPolicy: "0d" });
    expect(result.success).toBe(false);
  });

  it("rejects retention with a 'm' suffix (ambiguous)", () => {
    const result = objectStoreFormSchema.safeParse({ ...validS3Form, retentionPolicy: "5m" });
    expect(result.success).toBe(false);
  });
});

describe("toApiPayload", () => {
  it("converts S3 form values into the API payload shape", () => {
    const parsed = objectStoreFormSchema.parse(validS3Form);
    const payload = toApiPayload(parsed);
    expect(payload.name).toBe("minio-local");
    expect(payload.spec.destination_path).toBe("s3://stackdome-backups/db1");
    expect(payload.spec.retention_policy).toBe("7d");
    expect(payload.spec.configuration.s3_credentials).toMatchObject({
      region: "us-east-1",
      endpoint_url: "http://localhost:9000",
      access_key_id: { secret_id: validS3Form.s3!.accessKeyId.secret_id, key: "accessKeyId" },
      secret_access_key: { secret_id: validS3Form.s3!.secretAccessKey.secret_id, key: "secretAccessKey" },
    });
    expect(payload.spec.configuration.azure_credentials).toBeUndefined();
    expect(payload.spec.configuration.gcs_credentials).toBeUndefined();
  });

  it("omits endpoint_url when empty", () => {
    const payload = toApiPayload({
      ...validS3Form,
      s3: { ...validS3Form.s3!, endpointUrl: "" },
    });
    expect(payload.spec.configuration.s3_credentials?.endpoint_url).toBeUndefined();
  });

  it("converts azure provider values into azure_credentials only", () => {
    const azureValues: ObjectStoreFormValues = {
      name: "blob",
      destinationPath: "https://acct.blob.core.windows.net/c/p",
      retentionPolicy: "7d",
      provider: "azure",
      azure: {
        storageAccountName: "acct",
        connectionString: {
          secret_id: "22222222-2222-2222-2222-222222222222",
          key: "connectionString",
        },
      },
    };
    const payload = toApiPayload(azureValues);
    expect(payload.spec.configuration.azure_credentials).toMatchObject({
      connection_string: { secret_id: azureValues.azure!.connectionString.secret_id, key: "connectionString" },
      storage_account_name: "acct",
    });
    expect(payload.spec.configuration.s3_credentials).toBeUndefined();
    expect(payload.spec.configuration.gcs_credentials).toBeUndefined();
  });

  it("converts gcs provider values into gcs_credentials only", () => {
    const gcsValues: ObjectStoreFormValues = {
      name: "g",
      destinationPath: "gs://bucket/path",
      retentionPolicy: "7d",
      provider: "gcs",
      gcs: {
        serviceAccountCredentials: {
          secret_id: "33333333-3333-3333-3333-333333333333",
          key: "serviceAccount",
        },
      },
    };
    const payload = toApiPayload(gcsValues);
    expect(payload.spec.configuration.gcs_credentials).toMatchObject({
      service_account_credentials: { secret_id: gcsValues.gcs!.serviceAccountCredentials.secret_id, key: "serviceAccount" },
    });
    expect(payload.spec.configuration.s3_credentials).toBeUndefined();
    expect(payload.spec.configuration.azure_credentials).toBeUndefined();
  });

  it("omits whitespace-only endpoint_url", () => {
    const payload = toApiPayload({
      ...validS3Form,
      s3: { ...validS3Form.s3!, endpointUrl: "   " },
    });
    expect(payload.spec.configuration.s3_credentials?.endpoint_url).toBeUndefined();
  });
});

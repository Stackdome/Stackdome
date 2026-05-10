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
});

describe("toApiPayload", () => {
  it("converts S3 form values into the API payload shape", () => {
    const payload = toApiPayload(validS3Form);
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
});

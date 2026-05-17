import { describe, it, expect } from "vitest";
import {
  backupConfigSchema,
  toApiBackupConfig,
  type BackupConfigFormValues,
} from "../backup-config-schema";

const validForm: BackupConfigFormValues = {
  enabled: true,
  objectStoreId: "11111111-1111-1111-1111-111111111111",
  schedule: "0 0 0 * * 0",
  walArchiving: false,
};

describe("backupConfigSchema", () => {
  it("accepts a valid form", () => {
    expect(backupConfigSchema.safeParse(validForm).success).toBe(true);
  });

  it("requires an object store when enabled", () => {
    const r = backupConfigSchema.safeParse({ ...validForm, objectStoreId: "" });
    expect(r.success).toBe(false);
  });

  it("accepts disabled form without an object store", () => {
    const r = backupConfigSchema.safeParse({
      ...validForm,
      enabled: false,
      objectStoreId: "",
    });
    expect(r.success).toBe(true);
  });

  it("rejects a 5-field cron (Unix-style)", () => {
    const r = backupConfigSchema.safeParse({ ...validForm, schedule: "0 0 * * 0" });
    expect(r.success).toBe(false);
  });

  it("accepts a 6-field Quartz cron", () => {
    const r = backupConfigSchema.safeParse({ ...validForm, schedule: "0 30 5 * * *" });
    expect(r.success).toBe(true);
  });
});

describe("toApiBackupConfig", () => {
  it("converts form values to the API shape", () => {
    const out = toApiBackupConfig(validForm);
    expect(out).toEqual({
      enabled: true,
      object_store_id: "11111111-1111-1111-1111-111111111111",
      schedule: "0 0 0 * * 0",
      wal_archiving: false,
    });
  });

  it("omits object_store_id when empty", () => {
    const out = toApiBackupConfig({ ...validForm, enabled: false, objectStoreId: "" });
    expect(out.object_store_id).toBeUndefined();
  });
});

// @vitest-environment node
import { describe, it, expect } from "vitest";
import { PostgresAddonFormSchema, defaultFormValues } from "../form-schema";

const base = { ...defaultFormValues("cluster-1"), name: "db1" };

describe("initialization: restore_from_object_store", () => {
  it("accepts restore with source + object store, no target time (latest)", () => {
    const v = { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1" } };
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(true);
  });

  it("rejects restore missing sourceAddonId", () => {
    const v = { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "", objectStoreId: "os1" } };
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(false);
  });

  it("rejects restore missing objectStoreId", () => {
    const v = { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "" } };
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(false);
  });

  it("accepts a valid ISO recoveryTargetTime", () => {
    const v = { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1", recoveryTargetTime: "2026-05-17T03:00:00Z" } };
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(true);
  });

  it("rejects an invalid recoveryTargetTime", () => {
    const v = { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1", recoveryTargetTime: "not-a-date" } };
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(false);
  });

  it("still accepts the default new initialization", () => {
    expect(PostgresAddonFormSchema.safeParse(base).success).toBe(true);
  });
});

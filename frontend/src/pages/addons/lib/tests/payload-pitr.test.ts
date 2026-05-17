// @vitest-environment node
import { describe, it, expect } from "vitest";
import { buildCreateInput } from "../payload";
import { defaultFormValues } from "../../schemas/form-schema";

const base = { ...defaultFormValues("cluster-1"), name: "db1" };

describe("payload PITR mapping", () => {
  it("latest restore → no recovery_target_time", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_object_store",
      restore_from_object_store: { object_store_id: "os1", source_postgres_addon_id: "src" },
    });
  });

  it("specific time → recovery_target_time set", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1", recoveryTargetTime: "2026-05-17T03:00:00Z" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_object_store",
      restore_from_object_store: {
        object_store_id: "os1",
        source_postgres_addon_id: "src",
        recovery_target_time: "2026-05-17T03:00:00Z",
      },
    });
  });

  it("new initialization unchanged", () => {
    const input = buildCreateInput({ ...base });
    expect(input.spec.initialization).toEqual({ type: "new" });
  });
});

import { describe, it, expect } from "vitest";
import {
  configurePhaseSchema,
  configSettingsSchema,
  newPreviewEnvSchema,
  syncEnvSchema,
  overridesTextSchema,
} from "../form-schemas";

describe("overridesTextSchema", () => {
  it("accepts an empty string", () => {
    expect(overridesTextSchema.safeParse("").success).toBe(true);
  });

  it("accepts well-formed resource=image lines", () => {
    expect(
      overridesTextSchema.safeParse("web=registry.example.com/web:sha-1\napi=registry.example.com/api:sha-2")
        .success,
    ).toBe(true);
  });

  it("accepts blank lines interspersed with valid lines", () => {
    expect(overridesTextSchema.safeParse("web=image:tag\n\n\n").success).toBe(true);
  });

  it("rejects a line with no '='", () => {
    const result = overridesTextSchema.safeParse("not-a-pair");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toMatch(/resource=image/i);
    }
  });

  it("rejects a line with an empty key", () => {
    expect(overridesTextSchema.safeParse("=image:tag").success).toBe(false);
  });

  it("rejects a valid line mixed with an invalid one", () => {
    expect(overridesTextSchema.safeParse("web=image:tag\nnot-a-pair").success).toBe(false);
  });
});

describe("configurePhaseSchema", () => {
  const valid = { name: "webapp", baseBranch: "main", maxActive: 10 };

  it("accepts a minimal valid form", () => {
    expect(configurePhaseSchema.safeParse(valid).success).toBe(true);
  });

  it("accepts an optional stackfilePath", () => {
    expect(configurePhaseSchema.safeParse({ ...valid, stackfilePath: "deploy/stackfile.yaml" }).success).toBe(true);
  });

  it("rejects an empty name", () => {
    expect(configurePhaseSchema.safeParse({ ...valid, name: "" }).success).toBe(false);
  });

  it("rejects an empty base branch", () => {
    expect(configurePhaseSchema.safeParse({ ...valid, baseBranch: "" }).success).toBe(false);
  });

  it("rejects maxActive below 1", () => {
    expect(configurePhaseSchema.safeParse({ ...valid, maxActive: 0 }).success).toBe(false);
  });

  it("rejects a non-integer maxActive", () => {
    expect(configurePhaseSchema.safeParse({ ...valid, maxActive: 1.5 }).success).toBe(false);
  });
});

describe("configSettingsSchema", () => {
  const valid = { baseBranch: "main", stackfilePath: "stackfile.yaml", maxActive: 10 };

  it("accepts a valid form", () => {
    expect(configSettingsSchema.safeParse(valid).success).toBe(true);
  });

  it("rejects an empty stackfile path", () => {
    const result = configSettingsSchema.safeParse({ ...valid, stackfilePath: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.stackfilePath?.[0]).toMatch(/stackfile path is required/i);
    }
  });

  it("rejects an empty base branch", () => {
    expect(configSettingsSchema.safeParse({ ...valid, baseBranch: "" }).success).toBe(false);
  });

  it("rejects maxActive below 1", () => {
    expect(configSettingsSchema.safeParse({ ...valid, maxActive: 0 }).success).toBe(false);
  });
});

describe("newPreviewEnvSchema", () => {
  const valid = { prNumber: "42", branch: "feat/login", overridesText: "" };

  it("accepts a valid form", () => {
    expect(newPreviewEnvSchema.safeParse(valid).success).toBe(true);
  });

  it("rejects an empty PR number", () => {
    const result = newPreviewEnvSchema.safeParse({ ...valid, prNumber: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.prNumber?.[0]).toMatch(/pr number is required/i);
    }
  });

  it("rejects a non-digits PR number", () => {
    const result = newPreviewEnvSchema.safeParse({ ...valid, prNumber: "42a" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.prNumber?.[0]).toMatch(/digits only/i);
    }
  });

  it("rejects an empty branch", () => {
    expect(newPreviewEnvSchema.safeParse({ ...valid, branch: "" }).success).toBe(false);
  });

  it("rejects invalid override lines", () => {
    expect(newPreviewEnvSchema.safeParse({ ...valid, overridesText: "not-a-pair" }).success).toBe(false);
  });
});

describe("syncEnvSchema", () => {
  it("accepts everything empty", () => {
    expect(syncEnvSchema.safeParse({ commit: "", overridesText: "" }).success).toBe(true);
  });

  it("accepts a valid 7-char short SHA", () => {
    expect(syncEnvSchema.safeParse({ commit: "abc1234", overridesText: "" }).success).toBe(true);
  });

  it("accepts a valid 40-char full SHA", () => {
    expect(syncEnvSchema.safeParse({ commit: "a".repeat(40), overridesText: "" }).success).toBe(true);
  });

  it("accepts uppercase hex", () => {
    expect(syncEnvSchema.safeParse({ commit: "ABC1234", overridesText: "" }).success).toBe(true);
  });

  it("rejects a too-short SHA", () => {
    expect(syncEnvSchema.safeParse({ commit: "abc12", overridesText: "" }).success).toBe(false);
  });

  it("rejects a non-hex SHA", () => {
    expect(syncEnvSchema.safeParse({ commit: "zzzzzzz", overridesText: "" }).success).toBe(false);
  });

  it("rejects invalid override lines", () => {
    expect(syncEnvSchema.safeParse({ commit: "", overridesText: "nope" }).success).toBe(false);
  });
});

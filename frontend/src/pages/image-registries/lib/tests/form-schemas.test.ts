import { describe, it, expect } from "vitest";
import { createRegistrySchema, rotateRegistrySchema, verifyRegistrySchema } from "../form-schemas";

describe("createRegistrySchema", () => {
  it("requires host, username, and password", () => {
    const res = createRegistrySchema.safeParse({ host: "", username: "", password: "" });
    expect(res.success).toBe(false);
    if (!res.success) {
      const errs = res.error.flatten().fieldErrors;
      expect(errs.host?.[0]).toMatch(/host is required/i);
      expect(errs.username?.[0]).toMatch(/username is required/i);
      expect(errs.password?.[0]).toMatch(/password is required/i);
    }
  });

  it("accepts a complete input and trims fields", () => {
    const res = createRegistrySchema.safeParse({ host: " docker.io ", username: " bob ", password: " s3cret " });
    expect(res.success).toBe(true);
    if (res.success) {
      expect(res.data.host).toBe("docker.io");
      expect(res.data.username).toBe("bob");
      expect(res.data.password).toBe("s3cret");
    }
  });
});

describe("rotateRegistrySchema", () => {
  it("requires username and password, has no host field", () => {
    const res = rotateRegistrySchema.safeParse({ username: "", password: "" });
    expect(res.success).toBe(false);
    expect("host" in rotateRegistrySchema.shape).toBe(false);
  });
});

describe("verifyRegistrySchema", () => {
  it("requires repository", () => {
    const res = verifyRegistrySchema.safeParse({ repository: "" });
    expect(res.success).toBe(false);
    if (!res.success) {
      expect(res.error.flatten().fieldErrors.repository?.[0]).toMatch(/repository is required/i);
    }
  });
});

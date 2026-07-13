import { describe, it, expect } from "vitest";
import { gitSourceFormSchema } from "../git-source-form-schema";

const valid = { serviceName: "webapp", branch: "main", port: "3000" };

describe("gitSourceFormSchema", () => {
  it("accepts a minimal valid form", () => {
    expect(gitSourceFormSchema.safeParse(valid).success).toBe(true);
  });

  it("accepts optional dockerfilePath/buildContext when provided", () => {
    const result = gitSourceFormSchema.safeParse({
      ...valid,
      dockerfilePath: "docker/Dockerfile",
      buildContext: "./app",
    });
    expect(result.success).toBe(true);
  });

  it("rejects an empty service name", () => {
    const result = gitSourceFormSchema.safeParse({ ...valid, serviceName: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.serviceName?.[0]).toMatch(/service name is required/i);
    }
  });

  it("rejects an empty branch", () => {
    const result = gitSourceFormSchema.safeParse({ ...valid, branch: "" });
    expect(result.success).toBe(false);
  });

  it("rejects an empty port", () => {
    const result = gitSourceFormSchema.safeParse({ ...valid, port: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.port?.[0]).toMatch(/port is required/i);
    }
  });

  it("rejects a non-numeric port", () => {
    const result = gitSourceFormSchema.safeParse({ ...valid, port: "abc" });
    expect(result.success).toBe(false);
  });

  it("rejects port 0", () => {
    expect(gitSourceFormSchema.safeParse({ ...valid, port: "0" }).success).toBe(false);
  });

  it("rejects port above 65535", () => {
    expect(gitSourceFormSchema.safeParse({ ...valid, port: "65536" }).success).toBe(false);
  });

  it("accepts the port bounds 1 and 65535", () => {
    expect(gitSourceFormSchema.safeParse({ ...valid, port: "1" }).success).toBe(true);
    expect(gitSourceFormSchema.safeParse({ ...valid, port: "65535" }).success).toBe(true);
  });
});

import { describe, it, expect } from "vitest";
import { credentialsFormSchema, verifyIntegrationFormSchema } from "../form-schemas";

describe("credentialsFormSchema", () => {
  it("accepts a host and token with no username", () => {
    const result = credentialsFormSchema.safeParse({ host: "github.com", token: "ghp_abc", username: "" });
    expect(result.success).toBe(true);
  });

  it("accepts a host, username, and token", () => {
    const result = credentialsFormSchema.safeParse({ host: "bitbucket.org", username: "my-user", token: "app-pw" });
    expect(result.success).toBe(true);
  });

  it("rejects an empty host", () => {
    const result = credentialsFormSchema.safeParse({ host: "", token: "tok" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.host?.[0]).toMatch(/host is required/i);
    }
  });

  it("rejects a whitespace-only host", () => {
    const result = credentialsFormSchema.safeParse({ host: "   ", token: "tok" });
    expect(result.success).toBe(false);
  });

  it("rejects an empty token", () => {
    const result = credentialsFormSchema.safeParse({ host: "github.com", token: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.token?.[0]).toMatch(/access token is required/i);
    }
  });
});

describe("verifyIntegrationFormSchema", () => {
  it("accepts a valid https URL", () => {
    const result = verifyIntegrationFormSchema.safeParse({ repoUrl: "https://github.com/acme/webapp" });
    expect(result.success).toBe(true);
  });

  it("accepts a valid http URL", () => {
    const result = verifyIntegrationFormSchema.safeParse({ repoUrl: "http://git.example.com/acme/webapp" });
    expect(result.success).toBe(true);
  });

  it("rejects an empty URL", () => {
    const result = verifyIntegrationFormSchema.safeParse({ repoUrl: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.repoUrl?.[0]).toMatch(/required/i);
    }
  });

  it("rejects a non-URL string", () => {
    const result = verifyIntegrationFormSchema.safeParse({ repoUrl: "not a url" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.repoUrl?.[0]).toMatch(/valid http/i);
    }
  });

  it("rejects a non-http(s) scheme", () => {
    const result = verifyIntegrationFormSchema.safeParse({ repoUrl: "ftp://example.com/acme/webapp" });
    expect(result.success).toBe(false);
  });
});

import { describe, it, expect } from "vitest";
import { inviteSchema, toInviteCreateRequest } from "../invite-schema";

describe("inviteSchema", () => {
  it("rejects an invalid email", () => {
    const r = inviteSchema.safeParse({ email: "nope", project_name: "engineering", role: "Developer" });
    expect(r.success).toBe(false);
  });
  it("rejects missing project", () => {
    const r = inviteSchema.safeParse({ email: "a@b.io", project_name: "", role: "Developer" });
    expect(r.success).toBe(false);
  });
  it("accepts a valid invite and maps to the API request with expires_in_days=1", () => {
    const parsed = inviteSchema.parse({ email: "a@b.io", project_name: "engineering", role: "Viewer" });
    expect(toInviteCreateRequest(parsed)).toEqual({
      email: "a@b.io",
      project_name: "engineering",
      role: "Viewer",
      expires_in_days: 1,
    });
  });
});

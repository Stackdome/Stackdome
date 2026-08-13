import { describe, it, expect } from "vitest";
import { quotaMessage, COMPUTE_QUOTA_EXCEEDED_CODE } from "./quota-error";

describe("COMPUTE_QUOTA_EXCEEDED_CODE", () => {
  it("matches pkg/errors ErrorComputeQuotaExceeded", () => {
    expect(COMPUTE_QUOTA_EXCEEDED_CODE).toBe("30");
  });
});

describe("quotaMessage", () => {
  it("frames the stack-resource limit as organisation-wide", () => {
    const msg = quotaMessage("Stackdome Cloud allows a maximum of 6 stack resources per organisation");
    expect(msg.title).toBe("Stack resource limit reached");
    expect(msg.description).toBe(
      "Stackdome Cloud allows a maximum of 6 stack resources per organisation. " +
        "That's counted across your whole organisation, not per stack. " +
        "Remove a resource here, or delete one from another stack.",
    );
  });

  it("prefers the more specific scope match", () => {
    expect(quotaMessage("Stackdome Cloud allows a maximum of 2 stacks per organisation").title)
      .toBe("Stack limit reached");
    expect(quotaMessage("Stackdome Cloud allows a maximum volume size of 5Gi").title)
      .toBe("Volume size limit reached");
    expect(quotaMessage("Stackdome Cloud allows a maximum of 3 volumes per organisation").title)
      .toBe("Volume limit reached");
  });

  it("drops the organisation-wide note for the per-volume size limit", () => {
    const msg = quotaMessage("Stackdome Cloud allows a maximum volume size of 5Gi");
    expect(msg.description).toBe(
      "Stackdome Cloud allows a maximum volume size of 5Gi. Reduce the volume's size.",
    );
  });

  it("keeps the server reason when the scope is unrecognised", () => {
    const msg = quotaMessage("Stackdome Cloud allows a maximum of 4 widgets per organisation");
    expect(msg.title).toBe("Plan limit reached");
    expect(msg.description).toBe("Stackdome Cloud allows a maximum of 4 widgets per organisation");
  });
});

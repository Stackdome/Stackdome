import { describe, it, expect } from "vitest";
import { AxiosError } from "axios";
import { parseApiError } from "./errors";

function axiosErr(status: number, data: unknown): AxiosError {
  return new AxiosError(
    "Request failed",
    "ERR_BAD_REQUEST",
    { headers: {} } as never,
    {},
    {
      status,
      statusText: "Bad Request",
      data,
      headers: {},
      config: { headers: {} } as never,
    },
  );
}

describe("parseApiError", () => {
  it("extracts rich field errors from details.errors[]", () => {
    const err = axiosErr(400, {
      kind: "Error",
      id: "8",
      code: "8",
      reason: "validation failed: 2 error(s)",
      details: {
        errors: [
          { field: "spec.stack_resources[0].ports[1].protocol", code: "port_protocol_invalid", message: "protocol must be TCP or UDP" },
          { field: "name", code: "stack_name_invalid", message: "stack name is invalid" },
        ],
      },
    });
    const parsed = parseApiError(err);
    expect(parsed.status).toBe(400);
    expect(parsed.fieldErrors).toHaveLength(2);
    expect(parsed.fieldErrors[0]).toEqual({
      field: "spec.stack_resources[0].ports[1].protocol",
      code: "port_protocol_invalid",
      message: "protocol must be TCP or UDP",
    });
    expect(parsed.topLevel).toBe("validation failed: 2 error(s)");
    expect(parsed.credential).toBeUndefined();
    expect(parsed.code).toBe("8");
  });

  it("exposes the envelope code on errors with no details payload", () => {
    const err = axiosErr(400, {
      kind: "Error",
      id: "30",
      code: "30",
      reason: "Stackdome Cloud allows a maximum of 6 stack resources per organisation",
    });
    const parsed = parseApiError(err);
    expect(parsed.code).toBe("30");
    expect(parsed.fieldErrors).toEqual([]);
    expect(parsed.credential).toBeUndefined();
  });

  it("leaves code undefined when the body carries none", () => {
    expect(parseApiError(axiosErr(500, { reason: "boom" })).code).toBeUndefined();
  });

  it("falls back to reason with no field errors on a flat reject", () => {
    const err = axiosErr(400, { kind: "Error", reason: "source.git.repo_url is required" });
    const parsed = parseApiError(err);
    expect(parsed.fieldErrors).toEqual([]);
    expect(parsed.topLevel).toBe("source.git.repo_url is required");
  });

  it("parses the credential error shape", () => {
    const err = axiosErr(400, {
      kind: "Error",
      reason: "credentials required",
      details: {
        code: "credentials_required",
        target: { kind: "git_integration", host: "github.com", ref: "acme/repo" },
      },
    });
    const parsed = parseApiError(err);
    expect(parsed.fieldErrors).toEqual([]);
    expect(parsed.credential).toEqual({
      code: "credentials_required",
      target: { kind: "git_integration", host: "github.com", ref: "acme/repo" },
    });
  });

  it("reads details from the first item of an ErrorList", () => {
    const err = axiosErr(400, {
      kind: "ErrorList",
      page: 0,
      size: 1,
      total: 1,
      items: [
        {
          kind: "Error",
          reason: "validation failed: 1 error(s)",
          details: { errors: [{ field: "replicas", code: "replicas_invalid", message: "replicas cannot be negative" }] },
        },
      ],
    });
    const parsed = parseApiError(err);
    expect(parsed.topLevel).toBe("validation failed: 1 error(s)");
    expect(parsed.fieldErrors).toEqual([
      { field: "replicas", code: "replicas_invalid", message: "replicas cannot be negative" },
    ]);
  });

  it("drops malformed entries missing field or message", () => {
    const err = axiosErr(400, {
      kind: "Error",
      reason: "validation failed",
      details: {
        errors: [
          { field: "name", message: "ok", code: "resource_name_required" },
          { code: "port_protocol_invalid" },
          { field: "ports[0].number" },
          null,
        ],
      },
    });
    const parsed = parseApiError(err);
    expect(parsed.fieldErrors).toEqual([
      { field: "name", code: "resource_name_required", message: "ok" },
    ]);
  });

  it("handles a network error with no response body", () => {
    const err = new AxiosError("Network Error", "ERR_NETWORK");
    const parsed = parseApiError(err);
    expect(parsed.status).toBeUndefined();
    expect(parsed.fieldErrors).toEqual([]);
    expect(parsed.topLevel).toBe("Network Error");
  });

  it("handles a plain Error", () => {
    const parsed = parseApiError(new Error("boom"));
    expect(parsed.fieldErrors).toEqual([]);
    expect(parsed.topLevel).toBe("boom");
  });
});

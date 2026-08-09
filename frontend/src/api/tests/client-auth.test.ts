// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { handleResponseError, isOnAuthPage, type AuthErrorDeps } from "@/api/client";

function mkDeps(over: Partial<AuthErrorDeps> = {}): AuthErrorDeps {
  return {
    refresh: vi.fn(async () => "new-access-token"),
    retry: vi.fn(async () => ({ data: "retried-ok" })),
    onAuthFailure: vi.fn(),
    isAuthPage: () => false,
    ...over,
  };
}

// Backend returns 403 (not 401) for an expired/invalid access token, with the
// reason coming from the authn middleware. RBAC denials are also 403 but carry a
// different reason.
function err(
  status: number | undefined,
  reason = "",
  config: Record<string, unknown> = { headers: {} },
) {
  return {
    response: status === undefined ? undefined : { status, data: { reason } },
    config,
  };
}

beforeEach(() => vi.clearAllMocks());

describe("handleResponseError", () => {
  it("refreshes + retries on a 403 whose reason is an expired token", async () => {
    const deps = mkDeps();
    const config: Record<string, unknown> = { headers: {} };

    const result = await handleResponseError(
      err(403, "token parse error: token is expired by 33h52m26s", config),
      deps,
    );

    expect(deps.refresh).toHaveBeenCalledTimes(1);
    expect((config.headers as Record<string, string>).Authorization).toBe("Bearer new-access-token");
    expect(deps.retry).toHaveBeenCalledWith(config);
    expect(result).toEqual({ data: "retried-ok" });
    expect(deps.onAuthFailure).not.toHaveBeenCalled();
  });

  it("still refreshes + retries on a 401 (defensive — in case the backend ever returns one)", async () => {
    const deps = mkDeps();
    const config: Record<string, unknown> = { headers: {} };
    const result = await handleResponseError(err(401, "", config), deps);
    expect(deps.refresh).toHaveBeenCalledTimes(1);
    expect(result).toEqual({ data: "retried-ok" });
  });

  it("does NOT refresh on an RBAC 403 (valid session, missing permission)", async () => {
    const deps = mkDeps();
    await expect(
      handleResponseError(err(403, "Account is unauthorized to perform this action"), deps),
    ).rejects.toBeDefined();
    expect(deps.refresh).not.toHaveBeenCalled();
    expect(deps.onAuthFailure).not.toHaveBeenCalled();
  });

  it("does NOT log out when refresh succeeds but the retried request fails (e.g. 404)", async () => {
    const retryErr = { response: { status: 404 } };
    const deps = mkDeps({ retry: vi.fn(async () => { throw retryErr; }) });
    await expect(handleResponseError(err(403, "token is expired by 1h"), deps)).rejects.toBe(retryErr);
    expect(deps.refresh).toHaveBeenCalledTimes(1);
    expect(deps.onAuthFailure).not.toHaveBeenCalled();
  });

  it("when refresh fails: calls onAuthFailure and rejects", async () => {
    const deps = mkDeps({ refresh: vi.fn(async () => { throw new Error("refresh failed"); }) });
    await expect(
      handleResponseError(err(403, "token is expired by 1h"), deps),
    ).rejects.toBeDefined();
    expect(deps.onAuthFailure).toHaveBeenCalledTimes(1);
    expect(deps.retry).not.toHaveBeenCalled();
  });

  it("does not refresh on unrelated errors (500)", async () => {
    const deps = mkDeps();
    await expect(handleResponseError(err(500, "boom"), deps)).rejects.toBeDefined();
    expect(deps.refresh).not.toHaveBeenCalled();
  });

  it("does not refresh again on an already-retried request (loop guard)", async () => {
    const deps = mkDeps();
    await expect(
      handleResponseError(err(403, "token is expired by 1h", { headers: {}, _retry: true }), deps),
    ).rejects.toBeDefined();
    expect(deps.refresh).not.toHaveBeenCalled();
  });

  it("does not refresh on an auth page (so a wrong-password error surfaces inline)", async () => {
    const deps = mkDeps({ isAuthPage: () => true });
    await expect(handleResponseError(err(403, "token is expired by 1h"), deps)).rejects.toBeDefined();
    expect(deps.refresh).not.toHaveBeenCalled();
  });
});

describe("isOnAuthPage", () => {
  // The callback page must count as an auth page: with a stale expired session,
  // a failed refresh would otherwise redirect to /sign-in and abort the
  // in-flight OAuth code exchange.
  it.each(["/sign-in", "/sign-up", "/auth/github/callback"])("is true on %s", (path) => {
    window.history.replaceState(null, "", path);
    expect(isOnAuthPage()).toBe(true);
  });

  it("is false on app pages", () => {
    window.history.replaceState(null, "", "/dashboard");
    expect(isOnAuthPage()).toBe(false);
  });
});

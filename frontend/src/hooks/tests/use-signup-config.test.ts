// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useSignupConfig } from "@/hooks/use-signup-config";
import * as configApi from "@/api/config";

vi.mock("@/api/config");

describe("useSignupConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(configApi.getCachedAppConfig).mockReturnValue(null);
  });

  it("returns the public Turnstile settings from app config", async () => {
    vi.mocked(configApi.getAppConfig).mockResolvedValue({
      signup: {
        turnstile: {
          enabled: true,
          site_key: "public-site-key",
          action: "turnstile-spin-v2",
        },
      },
    });

    const { result } = renderHook(() => useSignupConfig());

    await waitFor(() =>
      expect(result.current.turnstile).toEqual({
        enabled: true,
        site_key: "public-site-key",
        action: "turnstile-spin-v2",
      }),
    );
    expect(result.current.error).toBeNull();
  });

  it("fails closed when app config cannot be loaded", async () => {
    vi.mocked(configApi.getAppConfig).mockRejectedValue(new Error("network"));

    const { result } = renderHook(() => useSignupConfig());

    await waitFor(() => expect(result.current.error).toBeDefined());
    expect(result.current.turnstile).toBeUndefined();
    expect(result.current.error).toBeDefined();
  });
});

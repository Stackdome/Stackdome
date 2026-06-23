// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useAppConfig } from "@/hooks/use-app-config";
import * as configApi from "@/api/config";

vi.mock("@/api/config");

describe("useAppConfig", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns githubOAuth true when the server reports it enabled", async () => {
    vi.mocked(configApi.getAppConfig).mockResolvedValue({ github_oauth: true });
    const { result } = renderHook(() => useAppConfig());
    await waitFor(() => expect(result.current.githubOAuth).toBe(true));
  });

  it("fails closed (githubOAuth false) when the config request errors", async () => {
    vi.mocked(configApi.getAppConfig).mockRejectedValue(new Error("boom"));
    const { result } = renderHook(() => useAppConfig());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.githubOAuth).toBe(false);
  });
});

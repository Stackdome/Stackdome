// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

const get = vi.fn();
vi.mock("@/api/client", () => ({ default: { get } }));

describe("getAppConfig", () => {
  beforeEach(() => {
    vi.resetModules(); // fresh module-level cache per test
    get.mockReset();
  });

  it("fetches once and caches the resolved value (single-flight)", async () => {
    get.mockResolvedValue({ data: { github_oauth: true } });
    const { getAppConfig, getCachedAppConfig } = await import("@/api/config");

    const a = await getAppConfig();
    const b = await getAppConfig();

    expect(a).toEqual({ github_oauth: true });
    expect(b).toEqual({ github_oauth: true });
    expect(get).toHaveBeenCalledTimes(1);
    expect(getCachedAppConfig()).toEqual({ github_oauth: true });
  });

  it("does NOT cache a rejection — retries on the next call", async () => {
    get
      .mockRejectedValueOnce(new Error("network"))
      .mockResolvedValueOnce({ data: { github_oauth: true } });
    const { getAppConfig } = await import("@/api/config");

    await expect(getAppConfig()).rejects.toThrow("network");
    const v = await getAppConfig();

    expect(v).toEqual({ github_oauth: true });
    expect(get).toHaveBeenCalledTimes(2);
  });
});

import { describe, expect, it, vi, beforeEach } from "vitest";
import api from "../client";
import { createApiToken, listApiTokens, revokeApiToken, getApiTokenScopes } from "../api-tokens";

vi.mock("../client", () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}));

describe("api-tokens api", () => {
  beforeEach(() => vi.clearAllMocks());

  it("lists tokens", async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { items: [] } });
    await expect(listApiTokens()).resolves.toEqual({ items: [] });
    expect(api.get).toHaveBeenCalledWith("/api-tokens");
  });

  it("creates a token and returns the one-time secret", async () => {
    vi.mocked(api.post).mockResolvedValue({ data: { id: "t1", token: "sd_secret" } });
    const res = await createApiToken({ name: "agent", scopes: ["*"] });
    expect(res.token).toBe("sd_secret");
    expect(api.post).toHaveBeenCalledWith("/api-tokens", { name: "agent", scopes: ["*"] });
  });

  it("revokes a token", async () => {
    vi.mocked(api.delete).mockResolvedValue({});
    await revokeApiToken("t1");
    expect(api.delete).toHaveBeenCalledWith("/api-tokens/t1");
  });

  it("fetches scopes", async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { full_access_scope: "*", items: [] } });
    await expect(getApiTokenScopes()).resolves.toMatchObject({ full_access_scope: "*" });
    expect(api.get).toHaveBeenCalledWith("/api-tokens/scopes");
  });
});

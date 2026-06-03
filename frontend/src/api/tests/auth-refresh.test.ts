// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("axios", () => ({ default: { post: vi.fn() } }));
import axios from "axios";
import { refreshAccessToken } from "@/api/auth-refresh";

const mockedPost = vi.mocked(axios.post);

beforeEach(() => {
  localStorage.clear();
  mockedPost.mockReset();
});

describe("refreshAccessToken", () => {
  it("exchanges the stored refresh token, stores the rotated tokens, and returns the new access token", async () => {
    localStorage.setItem("refreshToken", "r1");
    mockedPost.mockResolvedValueOnce({ data: { token: "a2", refreshToken: "r2" } });

    const token = await refreshAccessToken();

    expect(token).toBe("a2");
    expect(localStorage.getItem("authToken")).toBe("a2");
    expect(localStorage.getItem("refreshToken")).toBe("r2");
    expect(mockedPost).toHaveBeenCalledTimes(1);
    expect(mockedPost).toHaveBeenCalledWith(
      expect.stringContaining("/auth/refresh"),
      { refreshToken: "r1" },
    );
  });

  it("keeps the existing refresh token when the response does not rotate it", async () => {
    localStorage.setItem("refreshToken", "r1");
    mockedPost.mockResolvedValueOnce({ data: { token: "a2" } });

    await refreshAccessToken();

    expect(localStorage.getItem("authToken")).toBe("a2");
    expect(localStorage.getItem("refreshToken")).toBe("r1");
  });

  it("throws without a network call when there is no stored refresh token", async () => {
    await expect(refreshAccessToken()).rejects.toThrow();
    expect(mockedPost).not.toHaveBeenCalled();
  });

  it("throws when the refresh endpoint fails", async () => {
    localStorage.setItem("refreshToken", "r1");
    mockedPost.mockRejectedValueOnce(new Error("401"));

    await expect(refreshAccessToken()).rejects.toThrow();
  });

  it("is single-flight: concurrent callers share one network request", async () => {
    localStorage.setItem("refreshToken", "r1");
    let resolve!: (v: unknown) => void;
    mockedPost.mockReturnValueOnce(new Promise((r) => { resolve = r; }));

    const p1 = refreshAccessToken();
    const p2 = refreshAccessToken();
    resolve({ data: { token: "a2", refreshToken: "r2" } });

    await expect(p1).resolves.toBe("a2");
    await expect(p2).resolves.toBe("a2");
    expect(mockedPost).toHaveBeenCalledTimes(1);
  });
});

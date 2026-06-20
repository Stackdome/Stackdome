// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import axios from "axios";
import { githubOAuthUrl, completeGitHubOAuth } from "@/api/auth-github";

vi.mock("axios");

describe("githubOAuthUrl", () => {
  it("builds the base initiate URL", () => {
    expect(githubOAuthUrl()).toBe("/api/v1/auth/github");
  });
  it("appends invite_token when provided", () => {
    expect(githubOAuthUrl("abc123")).toBe("/api/v1/auth/github?invite_token=abc123");
  });
});

describe("completeGitHubOAuth", () => {
  beforeEach(() => vi.clearAllMocks());
  it("GETs the backend callback with code+state and returns the token data", async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { token: "t", refreshToken: "r" } });
    const result = await completeGitHubOAuth("code1", "state1");
    expect(axios.get).toHaveBeenCalledWith("/api/v1/auth/github/callback", {
      params: { code: "code1", state: "state1" },
    });
    expect(result).toEqual({ token: "t", refreshToken: "r" });
  });
});

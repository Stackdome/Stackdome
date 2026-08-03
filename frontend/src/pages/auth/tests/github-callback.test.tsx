// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import type { User } from "@/api/users";
import GithubCallbackPage from "@/pages/auth/github-callback";
import * as authGithub from "@/api/auth-github";
import * as usersApi from "@/api/users";
import * as common from "@/lib/common";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async (orig) => {
  const actual = await orig<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigateMock };
});
vi.mock("@/api/auth-github");
vi.mock("@/api/users");
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ refresh: vi.fn().mockResolvedValue(undefined) }),
}));

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/auth/github/callback" element={<GithubCallbackPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("GithubCallbackPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it("completes the round trip and navigates to the dashboard", async () => {
    vi.mocked(authGithub.completeGitHubOAuth).mockResolvedValue({ token: "tok", refreshToken: "ref" });
    vi.mocked(usersApi.getCurrentUser).mockResolvedValue({ id: "u1", email: "a@b.c" } as User);
    const setSession = vi.spyOn(common, "setAuthSession").mockImplementation(() => {});

    renderAt("/auth/github/callback?code=c1&state=s1");

    await waitFor(() => expect(navigateMock).toHaveBeenCalledWith("/dashboard", { replace: true }));
    expect(setSession).toHaveBeenCalledWith("tok", { id: "u1", email: "a@b.c" }, "ref");
  });

  it("shows an error and does not call the backend when code/state are missing", async () => {
    renderAt("/auth/github/callback");
    await screen.findByText(/missing authorization code or state/i);
    expect(authGithub.completeGitHubOAuth).not.toHaveBeenCalled();
  });

  it("shows the backend error message when the exchange fails", async () => {
    vi.mocked(authGithub.completeGitHubOAuth).mockRejectedValue(new Error("invalid state parameter"));
    renderAt("/auth/github/callback?code=c1&state=s1");
    await screen.findByText(/invalid state parameter/i);
  });
});

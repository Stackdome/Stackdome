// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
import * as useAppConfigModule from "@/hooks/use-app-config";

vi.mock("@/hooks/use-app-config");

describe("GitHubSignInButton", () => {
  beforeEach(() => vi.clearAllMocks());
  afterEach(() => vi.unstubAllGlobals());

  it("renders nothing when GitHub OAuth is disabled", () => {
    vi.mocked(useAppConfigModule.useAppConfig).mockReturnValue({ githubOAuth: false, loading: false });
    const { container } = render(<GitHubSignInButton />);
    expect(container.firstChild).toBeNull();
  });

  it("renders the button when enabled and navigates with invite_token on click", () => {
    vi.mocked(useAppConfigModule.useAppConfig).mockReturnValue({ githubOAuth: true, loading: false });
    const assign = vi.fn();
    vi.stubGlobal("location", { assign });
    render(<GitHubSignInButton inviteToken="inv1" />);
    screen.getByRole("button", { name: /github/i }).click();
    expect(assign).toHaveBeenCalledWith("/api/v1/auth/github?invite_token=inv1");
  });
});

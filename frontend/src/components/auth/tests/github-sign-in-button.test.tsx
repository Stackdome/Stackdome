// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
import * as useAppConfigModule from "@/hooks/use-app-config";

vi.mock("@/hooks/use-app-config");

describe("GitHubSignInButton", () => {
  beforeEach(() => vi.clearAllMocks());
  afterEach(() => {
    vi.unstubAllGlobals();
    cleanup();
  });

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

  // The email/password submit next to this button is the screen's one filled
  // control (rubric #2) — GitHub is the outline secondary, never `inverse`/`default`.
  it("renders as the outline secondary, not a filled control", () => {
    vi.mocked(useAppConfigModule.useAppConfig).mockReturnValue({ githubOAuth: true, loading: false });
    render(<GitHubSignInButton />);
    const button = screen.getByRole("button", { name: /github/i });
    expect(button.className).toContain("bg-control");
    expect(button.className).not.toContain("bg-foreground");
  });
});

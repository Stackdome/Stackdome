// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

vi.mock("@/api/git-integrations", () => ({
  listGitIntegrations: vi.fn(),
  createGitHubAppManifest: vi.fn(),
  getGitIntegration: vi.fn(),
  listInstallations: vi.fn(),
  searchRepositories: vi.fn().mockResolvedValue({ items: [], page: 1, total_count: 0, has_next: false }),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));

import { listGitIntegrations } from "@/api/git-integrations";
import { EnableRepoWizard } from "../enable-repo-wizard";

function renderWizard() {
  return render(
    <MemoryRouter>
      <EnableRepoWizard open onOpenChange={() => {}} onCreated={() => {}} />
    </MemoryRouter>,
  );
}

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("EnableRepoWizard", () => {
  it("starts at the connect phase when no GitHub App integration exists", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [], total: 0 });
    renderWizard();
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /connect github/i })).toBeTruthy();
    });
  });

  it("skips to the pick phase when an installed integration exists", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "gi1", type: "github_app", status: "installed" }],
      total: 1,
    });
    renderWizard();
    await waitFor(() => {
      expect(screen.getByTestId("pick-phase")).toBeTruthy();
    });
  });

  it("connect phase shows waiting state after clicking connect", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [], total: 0 });
    const { createGitHubAppManifest } = await import("@/api/git-integrations");
    (createGitHubAppManifest as ReturnType<typeof vi.fn>).mockResolvedValue({
      manifest: {}, github_url: "https://github.com/settings/apps/new?state=s", state: "s",
    });
    vi.spyOn(window, "open").mockReturnValue({} as Window);
    renderWizard();
    await waitFor(() => screen.getByRole("button", { name: /connect github/i }));
    await userEvent.click(screen.getByRole("button", { name: /connect github/i }));
    await waitFor(() => {
      expect(screen.getByText(/waiting for installation/i)).toBeTruthy();
    });
  });
});

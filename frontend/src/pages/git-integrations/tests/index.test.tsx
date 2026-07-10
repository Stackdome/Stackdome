// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";

vi.mock("@/api/git-integrations", () => ({
  listGitIntegrations: vi.fn(),
  deleteGitIntegration: vi.fn(),
  listInstallations: vi.fn().mockResolvedValue({ items: [], total: 0 }),
  createGitHubAppManifest: vi.fn(),
  getGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));

import { listGitIntegrations } from "@/api/git-integrations";
import GitIntegrationsPage from "../index";

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("GitIntegrationsPage", () => {
  it("shows the connect CTA when no GitHub App integration exists", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [], total: 0 });
    render(<GitIntegrationsPage />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /connect github/i })).toBeTruthy();
    });
  });

  it("lists integrations with status badges", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "gi1", type: "github_app", host: "github.com", status: "installed", credentials_configured: true }],
      total: 1,
    });
    render(<GitIntegrationsPage />);
    await waitFor(() => {
      expect(screen.getByText("github.com")).toBeTruthy();
      expect(screen.getByText("installed")).toBeTruthy();
      expect(screen.getByText(/credentials set/i)).toBeTruthy();
    });
    expect(screen.queryByRole("button", { name: /connect github/i })).toBeNull();
  });
});

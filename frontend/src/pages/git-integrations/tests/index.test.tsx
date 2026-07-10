// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/git-integrations", () => ({
  listGitIntegrations: vi.fn(),
  deleteGitIntegration: vi.fn(),
  listInstallations: vi.fn().mockResolvedValue({ items: [], total: 0 }),
  createGitHubAppManifest: vi.fn(),
  getGitIntegration: vi.fn(),
  verifyGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { listGitIntegrations, verifyGitIntegration } from "@/api/git-integrations";
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

  it("verifies a repository URL through the verify dialog", async () => {
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "gi1", type: "github_app", host: "github.com", status: "installed", credentials_configured: true }],
      total: 1,
    });
    (verifyGitIntegration as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    render(<GitIntegrationsPage />);
    await waitFor(() => {
      expect(screen.getByText("github.com")).toBeTruthy();
    });

    await userEvent.click(screen.getByRole("button", { name: /verify github\.com integration/i }));
    await userEvent.type(screen.getByLabelText(/repository url/i), "https://github.com/acme/webapp");
    await userEvent.click(screen.getByRole("button", { name: "Verify" }));

    await waitFor(() => {
      expect(verifyGitIntegration).toHaveBeenCalledWith("org1", "gi1", "https://github.com/acme/webapp");
    });
    expect(toastMock).toHaveBeenCalledWith({ title: "Verification succeeded" });
  });
});

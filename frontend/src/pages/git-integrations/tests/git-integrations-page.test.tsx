// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import GitIntegrationsPage from "../index";

afterEach(cleanup);

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listGitIntegrations: vi.fn(),
  listInstallations: vi.fn().mockResolvedValue({ items: [] }),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/pages/previews/hooks/use-github-connect", () => ({
  useGithubConnect: () => ({ state: "idle", error: null, connect: vi.fn() }),
}));

import { listGitIntegrations } from "@/api/git-integrations";

describe("GitIntegrationsPage", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders branded empty state with an Add integration action when list is empty", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({ items: [] });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText(/no git integrations/i)).toBeInTheDocument());
    // Header + empty-state both expose the CTA.
    expect(screen.getAllByRole("button", { name: /add integration/i }).length).toBeGreaterThanOrEqual(1);
  });

  it("opens the wizard from the empty-state action", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({ items: [] });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText(/no git integrations/i)).toBeInTheDocument());
    fireEvent.click(screen.getAllByRole("button", { name: /add integration/i })[0]);
    expect(screen.getByText(/GitLab/)).toBeInTheDocument(); // provider grid visible
  });

  it("lists integrations inside the panel", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({
      items: [
        { id: "g1", host: "github.com", type: "github_app", status: "installed" },
        { id: "g2", host: "gitlab.com", type: "git_credentials", status: "active", credentials_configured: true },
      ],
    });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText("gitlab.com")).toBeInTheDocument());
    expect(screen.getByText("github.com")).toBeInTheDocument();
    expect(screen.getByText(/all integrations/i)).toBeInTheDocument();
  });
});

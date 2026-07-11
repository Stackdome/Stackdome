// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import GitIntegrationsPage from "../index";

afterEach(cleanup);

const toastMock = vi.fn();

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listGitIntegrations: vi.fn(),
  deleteGitIntegration: vi.fn(),
  listInstallations: vi.fn().mockResolvedValue({ items: [] }),
  verifyGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/pages/previews/hooks/use-github-connect", () => ({
  useGithubConnect: () => ({ state: "idle", error: null, connect: vi.fn() }),
}));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { listGitIntegrations, deleteGitIntegration, verifyGitIntegration } from "@/api/git-integrations";

describe("GitIntegrationsPage", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders branded empty state with a Connect provider action when list is empty", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({ items: [] });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText(/no git integrations yet/i)).toBeInTheDocument());
    // Header + empty-state both expose a connect CTA.
    expect(screen.getAllByRole("button", { name: /connect provider/i }).length).toBeGreaterThanOrEqual(1);
  });

  it("opens the wizard from the empty-state action", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({ items: [] });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText(/no git integrations yet/i)).toBeInTheDocument());
    const [, emptyStateButton] = screen.getAllByRole("button", { name: /connect provider/i });
    fireEvent.click(emptyStateButton);
    expect(screen.getByText(/GitLab/)).toBeInTheDocument(); // provider grid visible
  });

  it("lists integrations inside the panel with human copy", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({
      items: [
        { id: "g1", host: "github.com", type: "github_app", status: "installed", credentials_configured: true },
        { id: "g2", host: "gitlab.com", type: "git_credentials", status: "active", credentials_configured: true },
      ],
    });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText("gitlab.com")).toBeInTheDocument());
    expect(screen.getByText("github.com")).toBeInTheDocument();
    expect(screen.getByText(/connected providers/i)).toBeInTheDocument();
    expect(screen.getAllByText("Connected").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("GitHub App").length).toBeGreaterThanOrEqual(1);
  });

  it("renders the summary strip when the list is populated", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({
      items: [
        { id: "g1", host: "github.com", type: "github_app", status: "installed", credentials_configured: true },
        { id: "g2", host: "gitlab.com", type: "git_credentials", status: "pending_install", credentials_configured: true },
      ],
    });
    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText("gitlab.com")).toBeInTheDocument());
    expect(screen.getByText(/connected & ready/i)).toBeInTheDocument();
    expect(screen.getByText(/needs attention/i)).toBeInTheDocument();
  });

  it("shows the error state and retries the fetch on Retry", async () => {
    vi.mocked(listGitIntegrations)
      .mockRejectedValueOnce(new Error("request failed with status 500"))
      .mockResolvedValueOnce({
        items: [{ id: "g1", host: "github.com", type: "github_app", status: "installed", credentials_configured: true }],
      });

    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText(/couldn't load integrations/i)).toBeInTheDocument());

    await userEvent.click(screen.getByRole("button", { name: /retry/i }));

    await waitFor(() => expect(screen.getByText("github.com")).toBeInTheDocument());
    expect(listGitIntegrations).toHaveBeenCalledTimes(2);
  });

  it("verifies a repository URL through the verify dialog", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({
      items: [{ id: "g1", host: "github.com", type: "git_credentials", status: "active", credentials_configured: true }],
    });
    vi.mocked(verifyGitIntegration).mockResolvedValue(undefined);

    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText("github.com")).toBeInTheDocument());

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/verify repository access/i), { pointerEventsCheck: 0 });

    await userEvent.type(screen.getByLabelText(/repository url/i), "https://github.com/acme/webapp");
    await userEvent.click(screen.getByRole("button", { name: "Verify" }));

    await waitFor(() =>
      expect(verifyGitIntegration).toHaveBeenCalledWith("org-1", "g1", "https://github.com/acme/webapp"),
    );
    expect(toastMock).toHaveBeenCalledWith({ title: "Repository access verified" });
  });

  it("removes an integration via the row menu and confirm dialog", async () => {
    vi.mocked(listGitIntegrations)
      .mockResolvedValueOnce({
        items: [{ id: "g1", host: "github.com", type: "github_app", status: "installed", credentials_configured: true }],
      })
      .mockResolvedValueOnce({ items: [] });
    vi.mocked(deleteGitIntegration).mockResolvedValue(undefined);

    render(<GitIntegrationsPage />);
    await waitFor(() => expect(screen.getByText("github.com")).toBeInTheDocument());

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/remove integration/i), { pointerEventsCheck: 0 });

    expect(await screen.findByText(/remove this integration/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Remove" }), { pointerEventsCheck: 0 });

    await waitFor(() => expect(deleteGitIntegration).toHaveBeenCalledWith("org-1", "g1"));
    expect(toastMock).toHaveBeenCalledWith({ title: "Integration removed" });
  });
});

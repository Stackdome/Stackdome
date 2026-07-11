// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import { AddIntegrationWizard } from "../add-integration-wizard";

afterEach(cleanup);

const mockConnect = vi.fn().mockResolvedValue(undefined);
let mockConnectState = "idle";

vi.mock("@/pages/previews/hooks/use-github-connect", () => ({
  useGithubConnect: () => ({ state: mockConnectState, error: null, connect: mockConnect }),
}));
vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  createGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

import { createGitIntegration } from "@/api/git-integrations";

function renderWizard(props: Partial<Parameters<typeof AddIntegrationWizard>[0]> = {}) {
  return render(
    <AddIntegrationWizard
      open
      onOpenChange={vi.fn()}
      hasGithubApp={false}
      onCreated={vi.fn()}
      {...props}
    />,
  );
}

describe("AddIntegrationWizard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockConnect.mockResolvedValue(undefined);
    mockConnectState = "idle";
  });

  it("shows the provider grid first", () => {
    renderWizard();
    for (const label of ["GitHub", "GitLab", "Bitbucket", "Gitea", "Other"]) {
      expect(screen.getByRole("button", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("GitHub tile leads to method choice with App install recommended", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    expect(screen.getByText(/recommended/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /install github app/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /access token/i })).toBeInTheDocument();
  });

  it("disables App install when a GitHub App integration exists", () => {
    renderWizard({ hasGithubApp: true });
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    expect(screen.getByRole("button", { name: /install github app/i })).toBeDisabled();
    expect(screen.getByText(/already connected/i)).toBeInTheDocument();
  });

  it("App install invokes the connect flow", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /install github app/i }));
    expect(mockConnect).toHaveBeenCalledOnce();
  });

  it("GitLab tile opens credentials form with host prefilled", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    expect(screen.getByLabelText(/host/i)).toHaveValue("gitlab.com");
  });

  it("submits credentials integration and fires onCreated", async () => {
    vi.mocked(createGitIntegration).mockResolvedValue({ host: "gitlab.com" });
    const onCreated = vi.fn();
    renderWizard({ onCreated });
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "glpat-abc" } });
    fireEvent.click(screen.getByRole("button", { name: /add integration/i }));
    await waitFor(() => expect(onCreated).toHaveBeenCalledOnce());
    expect(createGitIntegration).toHaveBeenCalledWith("org-1", {
      host: "gitlab.com",
      type: "git_credentials",
      auth: { token: "glpat-abc" },
    });
  });

  it("renders inline error when create fails and keeps the form", async () => {
    vi.mocked(createGitIntegration).mockRejectedValue(new Error("409 conflict"));
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "glpat-abc" } });
    fireEvent.click(screen.getByRole("button", { name: /add integration/i }));
    await waitFor(() => expect(screen.getByText(/conflict/i)).toBeInTheDocument());
    expect(screen.getByLabelText(/host/i)).toHaveValue("gitlab.com");
  });

  it("fires onCreated when GitHub connect completes after the wizard was closed", () => {
    const onCreated = vi.fn();
    const onOpenChange = vi.fn();
    const { rerender } = render(
      <AddIntegrationWizard open onOpenChange={onOpenChange} hasGithubApp={false} onCreated={onCreated} />,
    );

    // User closes the wizard while the GitHub popup is still pending.
    rerender(
      <AddIntegrationWizard open={false} onOpenChange={onOpenChange} hasGithubApp={false} onCreated={onCreated} />,
    );
    expect(onCreated).not.toHaveBeenCalled();

    // The popup finishes in the background after the dialog is already closed.
    mockConnectState = "connected";
    rerender(
      <AddIntegrationWizard open={false} onOpenChange={onOpenChange} hasGithubApp={false} onCreated={onCreated} />,
    );

    expect(onCreated).toHaveBeenCalledOnce();
  });

  it("waiting state keeps Back and Done enabled while Install button shows a spinner", () => {
    mockConnectState = "waiting";
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    expect(screen.getByText(/waiting for the github app installation/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /install github app/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /back/i })).toBeEnabled();
    expect(screen.getByRole("button", { name: /done/i })).toBeEnabled();
  });

  it("submits basic auth when a username is provided (e.g. Bitbucket app passwords)", async () => {
    vi.mocked(createGitIntegration).mockResolvedValue({ host: "bitbucket.org" });
    const onCreated = vi.fn();
    renderWizard({ onCreated });
    fireEvent.click(screen.getByRole("button", { name: /Bitbucket/ }));
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: "my-user" } });
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "app-password" } });
    fireEvent.click(screen.getByRole("button", { name: /add integration/i }));
    await waitFor(() => expect(onCreated).toHaveBeenCalledOnce());
    expect(createGitIntegration).toHaveBeenCalledWith("org-1", {
      host: "bitbucket.org",
      type: "git_credentials",
      auth: { basic: { username: "my-user", password: "app-password" } },
    });
  });

  it("back from GitHub-credentials returns to method choice", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /access token/i }));
    expect(screen.getByLabelText(/host/i)).toHaveValue("github.com");
    fireEvent.click(screen.getByRole("button", { name: /back/i }));
    expect(screen.getByRole("button", { name: /install github app/i })).toBeInTheDocument();
  });
});

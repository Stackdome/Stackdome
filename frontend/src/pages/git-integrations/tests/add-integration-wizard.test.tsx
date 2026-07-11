// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup, within } from "@testing-library/react";
import { AddIntegrationWizard } from "../add-integration-wizard";

afterEach(cleanup);

const mockConnect = vi.fn().mockResolvedValue(undefined);
const mockCheckAgain = vi.fn().mockResolvedValue(undefined);
let mockConnectState = "idle";
let mockConnectError: string | null = null;

vi.mock("@/pages/previews/hooks/use-github-connect", () => ({
  useGithubConnect: () => ({
    state: mockConnectState,
    error: mockConnectError,
    connect: mockConnect,
    checkAgain: mockCheckAgain,
  }),
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
    mockCheckAgain.mockResolvedValue(undefined);
    mockConnectState = "idle";
    mockConnectError = null;
  });

  it("shows the provider grid first", () => {
    renderWizard();
    for (const label of ["GitHub", "GitLab", "Bitbucket", "Gitea", "Other"]) {
      expect(screen.getByRole("button", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("shows the Provider / Connect / Done stepper with the current step highlighted", () => {
    renderWizard();
    const stepper = screen.getByTestId("wizard-stepper");
    const { getByText } = within(stepper);
    expect(getByText("Provider")).toBeInTheDocument();
    expect(getByText("Connect")).toBeInTheDocument();
    expect(getByText("Done")).toBeInTheDocument();
    expect(getByText("Provider")).toHaveAttribute("data-current", "true");

    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    expect(getByText("Connect")).toHaveAttribute("data-current", "true");
    expect(getByText("Provider")).toHaveAttribute("data-current", "false");
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

  it("App install invokes the connect flow and moves to the connecting phase", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /install github app/i }));
    expect(mockConnect).toHaveBeenCalledOnce();
    expect(screen.getByText(/installing the github app/i)).toBeInTheDocument();
  });

  it("connecting phase renders the checklist and wires Check again to checkAgain()", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /install github app/i }));

    expect(screen.getByText(/opening github authorization/i)).toBeInTheDocument();
    expect(screen.getByText(/authorizing the installation/i)).toBeInTheDocument();
    expect(screen.getByText(/fetching accessible repositories/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /check again/i }));
    expect(mockCheckAgain).toHaveBeenCalledOnce();
  });

  it("shows a structured error card on connect failure with a retry that re-calls connect()", () => {
    mockConnectError = "Popup blocked — allow popups for this site and try again.";
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /install github app/i }));

    expect(screen.getByText(/couldn't connect to github/i)).toBeInTheDocument();
    expect(screen.getByText(/popup blocked/i)).toBeInTheDocument();
    mockConnect.mockClear();
    fireEvent.click(screen.getByRole("button", { name: /retry/i }));
    expect(mockConnect).toHaveBeenCalledOnce();
  });

  it("allows leaving the connecting phase via Close without firing onCreated", () => {
    mockConnectState = "waiting";
    const onCreated = vi.fn();
    const onOpenChange = vi.fn();
    renderWizard({ onCreated, onOpenChange });
    fireEvent.click(screen.getByRole("button", { name: /GitHub/ }));
    fireEvent.click(screen.getByRole("button", { name: /install github app/i }));

    expect(screen.getByText(/installing the github app/i)).toBeInTheDocument();

    const closeButton = screen.getByRole("button", { name: /close/i });
    expect(closeButton).toBeEnabled();
    fireEvent.click(closeButton);

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onCreated).not.toHaveBeenCalled();
  });

  it("GitLab tile opens credentials form with host prefilled", () => {
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    expect(screen.getByLabelText(/host/i)).toHaveValue("gitlab.com");
  });

  it("submits credentials integration, shows the done phase, and fires onCreated", async () => {
    vi.mocked(createGitIntegration).mockResolvedValue({ host: "gitlab.com" });
    const onCreated = vi.fn();
    renderWizard({ onCreated });
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "glpat-abc" } });
    fireEvent.click(screen.getByRole("button", { name: /^connect$/i }));
    await waitFor(() => expect(onCreated).toHaveBeenCalledOnce());
    expect(createGitIntegration).toHaveBeenCalledWith("org-1", {
      host: "gitlab.com",
      type: "git_credentials",
      auth: { token: "glpat-abc" },
    });
    expect(screen.getByText(/gitlab connected/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /^done$/i }));
  });

  it("renders inline error when create fails and keeps the form", async () => {
    vi.mocked(createGitIntegration).mockRejectedValue(new Error("409 conflict"));
    renderWizard();
    fireEvent.click(screen.getByRole("button", { name: /GitLab/ }));
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "glpat-abc" } });
    fireEvent.click(screen.getByRole("button", { name: /^connect$/i }));
    await waitFor(() => expect(screen.getByText(/conflict/i)).toBeInTheDocument());
    expect(screen.getByLabelText(/host/i)).toHaveValue("gitlab.com");
  });

  it("shows the done phase and fires onCreated once when GitHub connect completes while open", () => {
    const onCreated = vi.fn();
    const onOpenChange = vi.fn();
    const { rerender } = render(
      <AddIntegrationWizard open onOpenChange={onOpenChange} hasGithubApp={false} onCreated={onCreated} />,
    );

    mockConnectState = "connected";
    rerender(
      <AddIntegrationWizard open onOpenChange={onOpenChange} hasGithubApp={false} onCreated={onCreated} />,
    );

    expect(onCreated).toHaveBeenCalledOnce();
    expect(screen.getByText(/github app installed/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /^done$/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("fires onCreated when GitHub connect completes after the wizard was closed, without rendering the done phase", () => {
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
    expect(screen.queryByText(/github app installed/i)).not.toBeInTheDocument();
  });

  it("submits basic auth when a username is provided (e.g. Bitbucket app passwords)", async () => {
    vi.mocked(createGitIntegration).mockResolvedValue({ host: "bitbucket.org" });
    const onCreated = vi.fn();
    renderWizard({ onCreated });
    fireEvent.click(screen.getByRole("button", { name: /Bitbucket/ }));
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: "my-user" } });
    fireEvent.change(screen.getByLabelText(/token/i), { target: { value: "app-password" } });
    fireEvent.click(screen.getByRole("button", { name: /^connect$/i }));
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

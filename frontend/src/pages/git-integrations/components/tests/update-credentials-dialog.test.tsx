// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  updateGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { updateGitIntegration } from "@/api/git-integrations";
import { UpdateCredentialsDialog } from "../update-credentials-dialog";
import { GIT_INTEGRATION_TYPE_CREDENTIALS } from "../../lib/derive-row";
import type { GitIntegration } from "@/api/git-integrations";

const integration: GitIntegration = {
  id: "g1",
  host: "gitlab.com",
  type: GIT_INTEGRATION_TYPE_CREDENTIALS,
};

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("UpdateCredentialsDialog", () => {
  it("requires a token and does not call the API when empty", async () => {
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog integration={integration} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    expect(await screen.findByText(/access token is required/i)).toBeInTheDocument();
    expect(updateGitIntegration).not.toHaveBeenCalled();
  });

  it("shows the integration host read-only (no host input)", () => {
    render(<UpdateCredentialsDialog integration={integration} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);
    expect(screen.getByText("gitlab.com")).toBeInTheDocument();
    expect(screen.queryByLabelText(/host/i)).not.toBeInTheDocument();
  });

  it("submits token-only auth, toasts, closes, and reports onUpdated", async () => {
    vi.mocked(updateGitIntegration).mockResolvedValue(integration);
    const onOpenChange = vi.fn();
    const onUpdated = vi.fn();
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog integration={integration} onOpenChange={onOpenChange} onUpdated={onUpdated} />);

    await user.type(screen.getByLabelText(/access token/i), "glpat-new");
    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    await waitFor(() => {
      expect(updateGitIntegration).toHaveBeenCalledWith("org-1", "g1", {
        host: "gitlab.com",
        auth: { token: "glpat-new" },
      });
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Credentials updated" }));
      expect(onOpenChange).toHaveBeenCalledWith(false);
      expect(onUpdated).toHaveBeenCalled();
    });
  });

  it("submits basic auth when a username is filled", async () => {
    vi.mocked(updateGitIntegration).mockResolvedValue(integration);
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog integration={integration} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);

    await user.type(screen.getByLabelText(/username/i), "bob");
    await user.type(screen.getByLabelText(/access token/i), "app-pass");
    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    await waitFor(() => {
      expect(updateGitIntegration).toHaveBeenCalledWith("org-1", "g1", {
        host: "gitlab.com",
        auth: { basic: { username: "bob", password: "app-pass" } },
      });
    });
  });

  it("keeps the dialog open and toasts destructively on API failure", async () => {
    vi.mocked(updateGitIntegration).mockRejectedValue(new Error("boom"));
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog integration={integration} onOpenChange={onOpenChange} onUpdated={vi.fn()} />);

    await user.type(screen.getByLabelText(/access token/i), "glpat-new");
    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    await waitFor(() => {
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ variant: "destructive" }));
    });
    expect(onOpenChange).not.toHaveBeenCalled();
  });

  it("resets fields when a different integration is opened", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <UpdateCredentialsDialog integration={integration} onOpenChange={vi.fn()} onUpdated={vi.fn()} />,
    );
    await user.type(screen.getByLabelText(/access token/i), "glpat-new");
    expect(screen.getByLabelText(/access token/i)).toHaveValue("glpat-new");

    const other: GitIntegration = { id: "g2", host: "bitbucket.org", type: GIT_INTEGRATION_TYPE_CREDENTIALS };
    rerender(<UpdateCredentialsDialog integration={other} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);

    expect(screen.getByLabelText(/access token/i)).toHaveValue("");
  });
});

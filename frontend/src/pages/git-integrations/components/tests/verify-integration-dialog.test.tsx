// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  verifyGitIntegration: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { verifyGitIntegration } from "@/api/git-integrations";
import { VerifyIntegrationDialog } from "../verify-integration-dialog";
import type { GitIntegration } from "@/api/git-integrations";

const integration: GitIntegration = { id: "g1", host: "github.com" };

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("VerifyIntegrationDialog", () => {
  it("marks the repository URL as required and does not call the API when empty", async () => {
    const user = userEvent.setup();
    render(<VerifyIntegrationDialog integration={integration} onOpenChange={vi.fn()} />);

    const verifyButton = screen.getByRole("button", { name: /^verify$/i });
    expect(verifyButton).toBeEnabled();
    await user.click(verifyButton);

    expect(await screen.findByText(/repository url is required/i)).toBeInTheDocument();
    expect(verifyGitIntegration).not.toHaveBeenCalled();

    const label = screen.getByText(/^repository url$/i).closest("label");
    expect(label?.querySelector('[aria-hidden]')).toHaveTextContent("*");
  });

  it("rejects a non-URL value", async () => {
    const user = userEvent.setup();
    render(<VerifyIntegrationDialog integration={integration} onOpenChange={vi.fn()} />);

    await user.type(screen.getByLabelText(/repository url/i), "not a url");
    await user.click(screen.getByRole("button", { name: /^verify$/i }));

    expect(await screen.findByText(/valid http/i)).toBeInTheDocument();
    expect(verifyGitIntegration).not.toHaveBeenCalled();
  });

  it("clears the error once the field is edited", async () => {
    const user = userEvent.setup();
    render(<VerifyIntegrationDialog integration={integration} onOpenChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /^verify$/i }));
    expect(await screen.findByText(/repository url is required/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/repository url/i), "h");
    expect(screen.queryByText(/repository url is required/i)).not.toBeInTheDocument();
  });

  it("verifies a valid http(s) URL and closes on success", async () => {
    vi.mocked(verifyGitIntegration).mockResolvedValue(undefined);
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<VerifyIntegrationDialog integration={integration} onOpenChange={onOpenChange} />);

    await user.type(screen.getByLabelText(/repository url/i), "https://github.com/acme/webapp");
    await user.click(screen.getByRole("button", { name: /^verify$/i }));

    await waitFor(() => {
      expect(verifyGitIntegration).toHaveBeenCalledWith("org-1", "g1", "https://github.com/acme/webapp");
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Repository access verified" }));
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it("resets the field when a new integration is opened", () => {
    const { rerender } = render(
      <VerifyIntegrationDialog integration={integration} onOpenChange={vi.fn()} />,
    );
    const input = screen.getByLabelText(/repository url/i) as HTMLInputElement;
    input.value = "https://github.com/acme/webapp";

    const other: GitIntegration = { id: "g2", host: "gitlab.com" };
    rerender(<VerifyIntegrationDialog integration={other} onOpenChange={vi.fn()} />);

    expect(screen.getByLabelText(/repository url/i)).toHaveValue("");
  });
});

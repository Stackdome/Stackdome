// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/registry-credentials", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  updateRegistryCredential: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { updateRegistryCredential } from "@/api/registry-credentials";
import { UpdateCredentialsDialog } from "../update-credentials-dialog";
import { PURPOSE_PULL } from "../../lib/providers";
import type { RegistryCredential } from "@/api/registry-credentials";

const credential: RegistryCredential = {
  id: "r1",
  host: "quay.io",
  username: "old-bot",
  purpose: PURPOSE_PULL,
};

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("UpdateCredentialsDialog", () => {
  it("prefills username, leaves password empty, shows host read-only", () => {
    render(<UpdateCredentialsDialog credential={credential} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);
    expect(screen.getByLabelText(/username/i)).toHaveValue("old-bot");
    expect(screen.getByLabelText(/password/i)).toHaveValue("");
    expect(screen.getByText("quay.io")).toBeInTheDocument();
    expect(screen.queryByLabelText(/host/i)).not.toBeInTheDocument();
  });

  it("submits rotation with host and purpose unchanged", async () => {
    vi.mocked(updateRegistryCredential).mockResolvedValue(credential);
    const onOpenChange = vi.fn();
    const onUpdated = vi.fn();
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog credential={credential} onOpenChange={onOpenChange} onUpdated={onUpdated} />);

    await user.clear(screen.getByLabelText(/username/i));
    await user.type(screen.getByLabelText(/username/i), "new-bot");
    await user.type(screen.getByLabelText(/password/i), "n3w-secret");
    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    await waitFor(() => {
      expect(updateRegistryCredential).toHaveBeenCalledWith("org-1", "r1", {
        host: "quay.io",
        username: "new-bot",
        password: "n3w-secret",
        purpose: PURPOSE_PULL,
      });
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Credentials updated" }));
      expect(onOpenChange).toHaveBeenCalledWith(false);
      expect(onUpdated).toHaveBeenCalled();
    });
  });

  it("requires a password and does not call the API without one", async () => {
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog credential={credential} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    expect(await screen.findByText(/password is required/i)).toBeInTheDocument();
    expect(updateRegistryCredential).not.toHaveBeenCalled();
  });

  it("keeps the dialog open and toasts destructively on failure", async () => {
    vi.mocked(updateRegistryCredential).mockRejectedValue(new Error("boom"));
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<UpdateCredentialsDialog credential={credential} onOpenChange={onOpenChange} onUpdated={vi.fn()} />);

    await user.type(screen.getByLabelText(/password/i), "n3w-secret");
    await user.click(screen.getByRole("button", { name: /update credentials/i }));

    await waitFor(() => {
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ variant: "destructive" }));
    });
    expect(onOpenChange).not.toHaveBeenCalled();
  });

  it("resets fields when a different credential is opened", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <UpdateCredentialsDialog credential={credential} onOpenChange={vi.fn()} onUpdated={vi.fn()} />,
    );
    await user.type(screen.getByLabelText(/password/i), "typed");

    const other: RegistryCredential = { id: "r2", host: "ghcr.io", username: "gh-bot" };
    rerender(<UpdateCredentialsDialog credential={other} onOpenChange={vi.fn()} onUpdated={vi.fn()} />);

    expect(screen.getByLabelText(/username/i)).toHaveValue("gh-bot");
    expect(screen.getByLabelText(/password/i)).toHaveValue("");
  });
});

// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/registry-credentials", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  verifyRegistryCredential: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { verifyRegistryCredential } from "@/api/registry-credentials";
import { VerifyRegistryDialog } from "../verify-registry-dialog";
import type { RegistryCredential } from "@/api/registry-credentials";

const credential: RegistryCredential = { id: "r1", host: "ghcr.io", username: "bot" };

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("VerifyRegistryDialog", () => {
  it("requires a repository and does not call the API when empty", async () => {
    const user = userEvent.setup();
    render(<VerifyRegistryDialog credential={credential} onOpenChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /^verify$/i }));

    expect(await screen.findByText(/repository is required/i)).toBeInTheDocument();
    expect(verifyRegistryCredential).not.toHaveBeenCalled();
  });

  it("verifies a repository and closes on success", async () => {
    vi.mocked(verifyRegistryCredential).mockResolvedValue(undefined);
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<VerifyRegistryDialog credential={credential} onOpenChange={onOpenChange} />);

    await user.type(screen.getByLabelText(/repository/i), "acme/app");
    await user.click(screen.getByRole("button", { name: /^verify$/i }));

    await waitFor(() => {
      expect(verifyRegistryCredential).toHaveBeenCalledWith("org-1", "r1", "acme/app");
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Registry access verified" }));
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it("toasts destructively and stays open on failure", async () => {
    vi.mocked(verifyRegistryCredential).mockRejectedValue(new Error("denied"));
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<VerifyRegistryDialog credential={credential} onOpenChange={onOpenChange} />);

    await user.type(screen.getByLabelText(/repository/i), "acme/app");
    await user.click(screen.getByRole("button", { name: /^verify$/i }));

    await waitFor(() => {
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ variant: "destructive" }));
    });
    expect(onOpenChange).not.toHaveBeenCalled();
  });
});

// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/registry-credentials", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  createRegistryCredential: vi.fn(),
}));
vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { createRegistryCredential } from "@/api/registry-credentials";
import { AddRegistryDialog } from "../add-registry-dialog";
import { PURPOSE_BOTH } from "../../lib/providers";

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("AddRegistryDialog", () => {
  it("prefills the host from the chosen provider tile", async () => {
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={vi.fn()} onCreated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /docker hub/i }));
    expect(screen.getByLabelText(/host/i)).toHaveValue("docker.io");
  });

  it("creates a credential with default purpose and reports success", async () => {
    vi.mocked(createRegistryCredential).mockResolvedValue({ id: "r1", host: "docker.io", username: "bob" });
    const onCreated = vi.fn();
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={onOpenChange} onCreated={onCreated} />);

    await user.click(screen.getByRole("button", { name: /docker hub/i }));
    await user.type(screen.getByLabelText(/username/i), "bob");
    await user.type(screen.getByLabelText(/password/i), "s3cret");
    await user.click(screen.getByRole("button", { name: /add registry/i }));

    await waitFor(() => {
      expect(createRegistryCredential).toHaveBeenCalledWith("org-1", {
        host: "docker.io",
        username: "bob",
        password: "s3cret",
        purpose: PURPOSE_BOTH,
      });
      expect(onCreated).toHaveBeenCalled();
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it("keeps the dialog open and toasts destructively on non-conflict failure", async () => {
    vi.mocked(createRegistryCredential).mockRejectedValue(new Error("boom"));
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={onOpenChange} onCreated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /docker hub/i }));
    await user.type(screen.getByLabelText(/username/i), "bob");
    await user.type(screen.getByLabelText(/password/i), "s3cret");
    await user.click(screen.getByRole("button", { name: /add registry/i }));

    await waitFor(() => {
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ variant: "destructive" }));
    });
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("shows a 409 conflict inline on the host field, not as a toast", async () => {
    const conflict = Object.assign(new Error("conflict"), {
      isAxiosError: true,
      response: { status: 409 },
    });
    vi.mocked(createRegistryCredential).mockRejectedValue(conflict);
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={vi.fn()} onCreated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /docker hub/i }));
    await user.type(screen.getByLabelText(/username/i), "bob");
    await user.type(screen.getByLabelText(/password/i), "s3cret");
    await user.click(screen.getByRole("button", { name: /add registry/i }));

    expect(
      await screen.findByText(/credentials for this registry and purpose already exist/i),
    ).toBeInTheDocument();
    expect(toastMock).not.toHaveBeenCalled();
  });

  it("validates required fields without calling the API", async () => {
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={vi.fn()} onCreated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /^other/i }));
    await user.click(screen.getByRole("button", { name: /add registry/i }));

    expect(await screen.findByText(/host is required/i)).toBeInTheDocument();
    expect(screen.getByText(/username is required/i)).toBeInTheDocument();
    expect(screen.getByText(/password is required/i)).toBeInTheDocument();
    expect(createRegistryCredential).not.toHaveBeenCalled();
  });

  it("clears typed credentials when a different provider is picked", async () => {
    const user = userEvent.setup();
    render(<AddRegistryDialog open onOpenChange={vi.fn()} onCreated={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /docker hub/i }));
    await user.type(screen.getByLabelText(/username/i), "bob");
    await user.type(screen.getByLabelText(/password/i), "hub-secret");
    await user.click(screen.getByRole("button", { name: /back/i }));
    await user.click(screen.getByRole("button", { name: /quay/i }));

    expect(screen.getByLabelText(/host/i)).toHaveValue("quay.io");
    expect(screen.getByLabelText(/username/i)).toHaveValue("");
    expect(screen.getByLabelText(/password/i)).toHaveValue("");
  });
});

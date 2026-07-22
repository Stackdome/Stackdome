// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PendingRowMenu } from "../pending-row-menu";
import { ConfirmProvider } from "@/components/branded/confirm";
import type { PendingRow } from "../../hooks/use-users";

const revoke = vi.fn();
vi.mock("../../hooks/use-invites", () => ({
  useInvites: () => ({ create: vi.fn(), resend: vi.fn(), revoke, reset: vi.fn(), submitting: false, serverError: null, result: "idle" }),
}));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: vi.fn(), dismiss: vi.fn(), toasts: [] }),
}));

afterEach(cleanup);

const row = { kind: "pending", id: "i1", email: "dev@example.com", invite: {} } as PendingRow;

async function openRevoke() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: /invite actions/i }));
  await user.click(await screen.findByText("Revoke"));
  return user;
}

describe("PendingRowMenu revoke", () => {
  beforeEach(() => vi.clearAllMocks());

  it("revokes only after the confirm dialog is accepted", async () => {
    render(
      <ConfirmProvider>
        <PendingRowMenu row={row} onChanged={vi.fn()} />
      </ConfirmProvider>,
    );
    const user = await openRevoke();
    const dialog = await screen.findByRole("alertdialog");
    expect(dialog).toHaveTextContent("Revoke invite for dev@example.com?");
    expect(revoke).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Revoke" }));
    expect(revoke).toHaveBeenCalledWith("i1");
  });

  it("does not revoke when the confirm dialog is cancelled", async () => {
    render(
      <ConfirmProvider>
        <PendingRowMenu row={row} onChanged={vi.fn()} />
      </ConfirmProvider>,
    );
    const user = await openRevoke();
    await screen.findByRole("alertdialog");

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(revoke).not.toHaveBeenCalled();
  });
});

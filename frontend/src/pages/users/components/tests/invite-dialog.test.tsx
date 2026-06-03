// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";

const create = vi.fn();
vi.mock("../../hooks/use-invites", () => ({
  useInvites: () => ({ create, resend: vi.fn(), revoke: vi.fn(), reset: vi.fn(), submitting: false, serverError: null, result: "idle" }),
}));
vi.mock("../../hooks/use-team-options", () => ({
  useTeamOptions: () => ({ teams: [{ name: "engineering", default_team: true }], loading: false }),
}));

import { InviteDialog } from "../invite-dialog";

beforeEach(() => create.mockReset());
afterEach(() => cleanup());

describe("InviteDialog state machine", () => {
  it("shows validation errors and does not call the API for an invalid email", async () => {
    render(<InviteDialog open onOpenChange={() => {}} onCreated={() => {}} />);
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "bad" } });
    fireEvent.click(screen.getByRole("button", { name: /send invitation/i }));
    await waitFor(() => expect(screen.getByText(/valid email/i)).toBeTruthy());
    expect(create).not.toHaveBeenCalled();
  });

  it("on email_sent=true shows the one-time link and the copy-now notice", async () => {
    create.mockResolvedValue({ invite: { email_sent: true }, token: "tok_abc" });
    render(<InviteDialog open onOpenChange={() => {}} onCreated={() => {}} />);
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "a@b.io" } });
    fireEvent.click(screen.getByRole("button", { name: /send invitation/i }));
    await waitFor(() => expect(screen.getAllByText(/tok_abc/).length).toBeGreaterThan(0));
    expect(screen.getAllByText((_, el) => el?.tagName === "P" && !!el?.textContent?.match(/expires in 1 day/i)).length).toBeGreaterThan(0);
  });

  it("on email_sent=false surfaces the link as the fallback", async () => {
    create.mockResolvedValue({ invite: { email_sent: false }, token: "tok_def" });
    render(<InviteDialog open onOpenChange={() => {}} onCreated={() => {}} />);
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "a@b.io" } });
    fireEvent.click(screen.getByRole("button", { name: /send invitation/i }));
    await waitFor(() => expect(screen.getByText(/email delivery failed/i)).toBeTruthy());
    expect(screen.getAllByText(/tok_def/).length).toBeGreaterThan(0);
  });

  it("on API failure shows the server error and keeps the form", async () => {
    create.mockRejectedValueOnce(new Error("boom"));
    render(<InviteDialog open onOpenChange={() => {}} onCreated={() => {}} />);
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "a@b.io" } });
    fireEvent.click(screen.getByRole("button", { name: /send invitation/i }));
    await waitFor(() => expect(screen.getByText(/boom/i)).toBeTruthy());
    expect(screen.getByRole("button", { name: /send invitation/i })).toBeTruthy();
  });
});

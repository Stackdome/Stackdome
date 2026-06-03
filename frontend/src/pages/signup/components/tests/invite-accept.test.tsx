// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const signup = vi.fn();
vi.mock("../../hooks/use-signup", () => ({ useSignup: () => ({ signup, loading: false, error: null }) }));
vi.mock("@/helpers/common", () => ({ isUserLoggedIn: vi.fn(() => false), setAuthSession: vi.fn() }));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, isErrorStatus: vi.fn((_e: unknown, s: number) => s === 409) };
});

import { InviteAcceptForm } from "../invite-accept-form";

const info = { org_name: "Acme", team_name: "engineering", inviter_name: "Jane", expires_at: "2026-05-19T00:00:00Z" };

beforeEach(() => { signup.mockReset(); });
afterEach(() => cleanup());

function renderForm() {
  return render(
    <MemoryRouter>
      <InviteAcceptForm token="tok_1" info={info as never} />
    </MemoryRouter>,
  );
}

describe("InviteAcceptForm", () => {
  it("shows who invited and which org/team", () => {
    renderForm();
    expect(screen.getByText(/Acme/)).toBeTruthy();
    expect(screen.getByText(/engineering/)).toBeTruthy();
    expect(screen.getByText(/Jane/)).toBeTruthy();
  });

  it("creates the account with the invite token and shows the accepted interstitial", async () => {
    signup.mockResolvedValue({ jwt_token: "jwt", user: { id: "u1" } });
    renderForm();
    fireEvent.change(screen.getByLabelText(/full name/i), { target: { value: "Neo" } });
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "neo@acme.io" } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "longenoughpw" } });
    fireEvent.click(screen.getByRole("button", { name: /create account and join/i }));
    await waitFor(() => expect(signup).toHaveBeenCalledWith(
      { name: "Neo", email: "neo@acme.io", password: "longenoughpw" }, "tok_1",
    ));
    await waitFor(() => expect(screen.getByText(/you.?re in/i)).toBeTruthy());
  });

  it("on 409 shows the existing-account → log in message", async () => {
    signup.mockRejectedValue({ response: { status: 409 } });
    renderForm();
    fireEvent.change(screen.getByLabelText(/full name/i), { target: { value: "Neo" } });
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "neo@acme.io" } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "longenoughpw" } });
    fireEvent.click(screen.getByRole("button", { name: /create account and join/i }));
    await waitFor(() => expect(screen.getByText(/already have an account/i)).toBeTruthy());
    expect(screen.getByRole("link", { name: /log in/i })).toBeTruthy();
  });
});

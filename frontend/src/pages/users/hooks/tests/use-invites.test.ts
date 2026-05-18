// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, cleanup, waitFor } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/invites", () => ({
  createInvite: vi.fn(), resendInvite: vi.fn(), revokeInvite: vi.fn(),
}));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "server fail"), isErrorStatus: vi.fn(() => false) };
});

import { createInvite, resendInvite, revokeInvite } from "@/api/invites";
import { useInvites } from "../use-invites";

beforeEach(() => { vi.mocked(createInvite).mockReset(); vi.mocked(resendInvite).mockReset(); vi.mocked(revokeInvite).mockReset(); });
afterEach(() => cleanup());

describe("useInvites", () => {
  it("create() with email_sent=true yields result 'sent' and exposes the one-time token", async () => {
    vi.mocked(createInvite).mockResolvedValue({ id: "i1", email: "a@b.io", email_sent: true, invite_token: "tok_123" } as never);
    const { result } = renderHook(() => useInvites());
    let out: unknown;
    await act(async () => { out = await result.current.create({ email: "a@b.io", team_name: "engineering", role: "Developer" }); });
    expect(result.current.result).toBe("sent");
    expect((out as { token: string }).token).toBe("tok_123");
  });

  it("create() with email_sent=false yields result 'failed'", async () => {
    vi.mocked(createInvite).mockResolvedValue({ id: "i2", email: "a@b.io", email_sent: false, invite_token: "tok_x" } as never);
    const { result } = renderHook(() => useInvites());
    await act(async () => { await result.current.create({ email: "a@b.io", team_name: "engineering", role: "Developer" }); });
    expect(result.current.result).toBe("failed");
  });

  it("create() failure sets serverError and rethrows", async () => {
    vi.mocked(createInvite).mockRejectedValue(new Error("x"));
    const { result } = renderHook(() => useInvites());
    await act(async () => {
      await expect(result.current.create({ email: "a@b.io", team_name: "engineering", role: "Developer" })).rejects.toBeTruthy();
    });
    await waitFor(() => expect(result.current.serverError).toBe("server fail"));
  });

  it("resend and revoke call the API with org id", async () => {
    vi.mocked(resendInvite).mockResolvedValue(undefined as never);
    vi.mocked(revokeInvite).mockResolvedValue(undefined as never);
    const { result } = renderHook(() => useInvites());
    await act(async () => { await result.current.resend("i1"); await result.current.revoke("i1"); });
    expect(resendInvite).toHaveBeenCalledWith("org-1", "i1");
    expect(revokeInvite).toHaveBeenCalledWith("org-1", "i1");
  });
});

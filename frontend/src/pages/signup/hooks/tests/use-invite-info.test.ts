// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup } from "@testing-library/react";

vi.mock("@/api/invites", () => ({ getPublicInviteInfo: vi.fn() }));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return {
    ...actual,
    getErrorStatus: vi.fn(),
    getErrorMessage: vi.fn((e: unknown) => (e as { msg?: string })?.msg ?? ""),
  };
});

import { getPublicInviteInfo } from "@/api/invites";
import { getErrorStatus } from "@/api/client";
import { useInviteInfo } from "../use-invite-info";

beforeEach(() => { vi.mocked(getPublicInviteInfo).mockReset(); vi.mocked(getErrorStatus).mockReset(); });
afterEach(() => cleanup());

describe("useInviteInfo", () => {
  it("resolves to new-user with the info payload on success", async () => {
    vi.mocked(getPublicInviteInfo).mockResolvedValue({ org_name: "Acme", project_name: "engineering", inviter_name: "Jane", expires_at: "2026-05-19T00:00:00Z" } as never);
    const { result } = renderHook(() => useInviteInfo("tok"));
    await waitFor(() => expect(result.current.state).toBe("new-user"));
    expect(result.current.info?.org_name).toBe("Acme");
  });

  it("maps 404 to not-found", async () => {
    vi.mocked(getPublicInviteInfo).mockRejectedValue({});
    vi.mocked(getErrorStatus).mockReturnValue(404);
    const { result } = renderHook(() => useInviteInfo("tok"));
    await waitFor(() => expect(result.current.state).toBe("not-found"));
  });

  it("maps 410 + 'revoked' message to revoked", async () => {
    vi.mocked(getPublicInviteInfo).mockRejectedValue({ msg: "invite has been revoked" });
    vi.mocked(getErrorStatus).mockReturnValue(410);
    const { result } = renderHook(() => useInviteInfo("tok"));
    await waitFor(() => expect(result.current.state).toBe("revoked"));
  });

  it("maps 410 + 'already' message to already-used", async () => {
    vi.mocked(getPublicInviteInfo).mockRejectedValue({ msg: "invite has already been accepted" });
    vi.mocked(getErrorStatus).mockReturnValue(410);
    const { result } = renderHook(() => useInviteInfo("tok"));
    await waitFor(() => expect(result.current.state).toBe("already-used"));
  });

  it("maps 410 + 'expired' message to expired", async () => {
    vi.mocked(getPublicInviteInfo).mockRejectedValue({ msg: "invite has expired" });
    vi.mocked(getErrorStatus).mockReturnValue(410);
    const { result } = renderHook(() => useInviteInfo("tok"));
    await waitFor(() => expect(result.current.state).toBe("expired"));
  });
});

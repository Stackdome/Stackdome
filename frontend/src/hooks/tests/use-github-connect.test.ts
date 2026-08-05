// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";

vi.mock("@/api/git-integrations", () => ({
  createGitHubAppManifest: vi.fn(),
  listInstallations: vi.fn(),
  listGitIntegrations: vi.fn(),
}));
vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));

import { createGitHubAppManifest, listInstallations, listGitIntegrations } from "@/api/git-integrations";
import { GITHUB_APP_INSTALLED_MESSAGE } from "@/hooks/use-github-setup-landing";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  STATUS_INSTALLED,
  STATUS_PENDING_INSTALL,
} from "@/lib/git-integrations";
import { useGithubConnect } from "@/hooks/use-github-connect";

const flow = {
  manifest: { name: "stackdome-org1" },
  github_url: "https://github.com/settings/apps/new?state=s1",
  state: "s1",
};

describe("useGithubConnect", () => {
  let openSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.useFakeTimers();
    openSpy = vi.spyOn(window, "open").mockReturnValue({} as Window);
    (createGitHubAppManifest as ReturnType<typeof vi.fn>).mockResolvedValue(flow);
    (listGitIntegrations as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "gi1", type: GIT_INTEGRATION_TYPE_GITHUB_APP, status: STATUS_PENDING_INSTALL }],
      total: 1,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    openSpy.mockRestore();
    vi.clearAllMocks();
    document.querySelectorAll("form").forEach((f) => f.remove());
  });

  it("moves to waiting and form-posts the manifest into the popup", async () => {
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    expect(openSpy).toHaveBeenCalled();
    expect(result.current.state).toBe("waiting");
    expect(createGitHubAppManifest).toHaveBeenCalledWith("org1");
  });

  it("navigates the popup to the install page when there is no manifest", async () => {
    const popup = { location: { href: "" } } as unknown as Window;
    openSpy.mockReturnValue(popup);
    (createGitHubAppManifest as ReturnType<typeof vi.fn>).mockResolvedValue({
      github_url: "https://github.com/apps/stackdome-cloud/installations/new?state=s2",
      state: "s2",
    });

    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });

    expect(popup.location.href).toBe(
      "https://github.com/apps/stackdome-cloud/installations/new?state=s2",
    );
    expect(document.querySelectorAll("form")).toHaveLength(0);
    expect(result.current.state).toBe("waiting");
  });

  it("errors when the popup is blocked", async () => {
    openSpy.mockReturnValue(null);
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    expect(result.current.state).toBe("error");
    expect(result.current.error).toMatch(/popup blocked/i);
  });

  it("connects on the popup's postMessage", async () => {
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    act(() => {
      window.dispatchEvent(new MessageEvent("message", {
        data: { type: GITHUB_APP_INSTALLED_MESSAGE },
        origin: window.location.origin,
      }));
    });
    expect(result.current.state).toBe("connected");
  });

  it("connects when polling sees an installation appear (refresh self-heals missed webhooks)", async () => {
    (listInstallations as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({ items: [], total: 0 })
      .mockResolvedValueOnce({ items: [{ id: "in1" }], total: 1 });
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });
    expect(result.current.state).toBe("waiting");
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });
    expect(listInstallations).toHaveBeenCalledWith("org1", "gi1", true);
    expect(result.current.state).toBe("connected");
  });

  it("polls to connected even when the integration record only appears after connect", async () => {
    // connect() sees no integration yet — the record is created by GitHub's
    // manifest callback after the user confirms in the popup.
    (listGitIntegrations as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({ items: [], total: 0 }) // during connect()
      .mockResolvedValueOnce({ items: [], total: 0 }) // poll tick 1: still absent
      .mockResolvedValue({
        items: [{ id: "gi1", type: GIT_INTEGRATION_TYPE_GITHUB_APP, status: STATUS_INSTALLED }],
        total: 1,
      });
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });
    expect(result.current.state).toBe("waiting");
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });
    // The integration id only appears once the poll re-resolves it; success
    // here (rather than the earlier absent-record ticks) proves that happened.
    expect(result.current.state).toBe("connected");
  });

  it("checkAgain refreshes installations and connects when one exists", async () => {
    (listInstallations as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [{ id: "in1" }], total: 1 });
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    await act(async () => {
      await result.current.checkAgain();
    });
    expect(listInstallations).toHaveBeenCalledWith("org1", "gi1", true);
    expect(result.current.state).toBe("connected");
  });

  it("closes the popup and errors when manifest creation fails", async () => {
    const close = vi.fn();
    openSpy.mockReturnValue({ close } as unknown as Window);
    (createGitHubAppManifest as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("boom"));
    const { result } = renderHook(() => useGithubConnect());
    await act(async () => {
      await result.current.connect();
    });
    expect(close).toHaveBeenCalled();
    expect(result.current.state).toBe("error");
    expect(result.current.error).not.toBeNull();
  });
});

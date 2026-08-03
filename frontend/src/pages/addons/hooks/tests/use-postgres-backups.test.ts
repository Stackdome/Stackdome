// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup, act } from "@testing-library/react";
import type { PostgresBackup, PostgresBackupList } from "@/api/postgres-backups";

vi.mock("@/api/postgres-backups", async () => {
  const actual = await vi.importActual<typeof import("@/api/postgres-backups")>(
    "@/api/postgres-backups",
  );
  return {
    ...actual,
    listPostgresBackups: vi.fn(),
    // isTerminalPhase: keep the real implementation so polling logic behaves naturally
  };
});

vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: vi.fn(() => "org-1"),
}));

vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return {
    ...actual,
    isNotFoundError: vi.fn((e: unknown) => {
      // Treat any error with a `__notFound` marker as a 404 for the test seam.
      return !!(e as { __notFound?: boolean })?.__notFound;
    }),
    getErrorMessage: vi.fn(
      (e: unknown) => (e as { message?: string })?.message ?? "error",
    ),
  };
});

import { listPostgresBackups } from "@/api/postgres-backups";
import { usePostgresBackups } from "../use-postgres-backups";

const mockedList = vi.mocked(listPostgresBackups);

function mkBackup(
  phase: NonNullable<PostgresBackup["phase"]>,
  id = `b-${phase}`,
): PostgresBackup {
  return { id, phase } as PostgresBackup;
}

function mkList(items: PostgresBackup[]): PostgresBackupList {
  return { items } as PostgresBackupList;
}

beforeEach(() => {
  mockedList.mockReset();
  // shouldAdvanceTime lets waitFor/timeouts inside testing-library still run
  // while we retain manual control via advanceTimersByTime for polling.
  vi.useFakeTimers({ shouldAdvanceTime: true });
});

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.clearAllMocks();
});

// Helper: flush pending microtasks (resolved promises) while fake timers are active.
async function flushPromises() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe("usePostgresBackups", () => {
  it("fetches once on mount and exposes { backups, loading, error, refetch }", async () => {
    const items = [mkBackup("completed")];
    mockedList.mockResolvedValue(mkList(items));

    const { result } = renderHook(() => usePostgresBackups("addon-1"));

    expect(typeof result.current.refetch).toBe("function");

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(mockedList).toHaveBeenCalledTimes(1);
    expect(mockedList).toHaveBeenCalledWith("org-1", "addon-1");
    expect(result.current.backups).toEqual(items);
    expect(result.current.error).toBeNull();
  });

  it("polls every 5s while any backup is pending/running", async () => {
    // Always return at least one non-terminal backup so polling keeps going.
    mockedList.mockResolvedValue(mkList([mkBackup("running")]));

    renderHook(() => usePostgresBackups("addon-1"));

    await waitFor(() => expect(mockedList).toHaveBeenCalledTimes(1));

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });
    await flushPromises();
    expect(mockedList).toHaveBeenCalledTimes(2);

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });
    await flushPromises();
    expect(mockedList).toHaveBeenCalledTimes(3);
  });

  it("stops polling once all backups reach terminal phase", async () => {
    mockedList
      // first fetch: still running -> should start polling
      .mockResolvedValueOnce(mkList([mkBackup("running")]))
      // second fetch (after first 5s tick): all completed -> should stop polling
      .mockResolvedValueOnce(mkList([mkBackup("completed"), mkBackup("failed")]))
      // safety net: any further calls would also resolve cleanly (but should not happen)
      .mockResolvedValue(mkList([mkBackup("completed")]));

    renderHook(() => usePostgresBackups("addon-1"));

    await waitFor(() => expect(mockedList).toHaveBeenCalledTimes(1));

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });
    await flushPromises();
    expect(mockedList).toHaveBeenCalledTimes(2);

    // Now all terminal — advancing another 5s must NOT trigger another fetch.
    await act(async () => {
      vi.advanceTimersByTime(5000);
    });
    await flushPromises();
    expect(mockedList).toHaveBeenCalledTimes(2);
  });

  it("does not fetch again after unmount", async () => {
    mockedList.mockResolvedValue(mkList([mkBackup("running")]));

    const { unmount } = renderHook(() => usePostgresBackups("addon-1"));
    await waitFor(() => expect(mockedList).toHaveBeenCalledTimes(1));

    unmount();
    await act(async () => {
      vi.advanceTimersByTime(5000);
    });
    await flushPromises();

    expect(mockedList).toHaveBeenCalledTimes(1);
  });

  it("fetches against the new id when addonId changes", async () => {
    mockedList.mockResolvedValue(mkList([mkBackup("completed")]));

    const { rerender } = renderHook(({ id }) => usePostgresBackups(id), {
      initialProps: { id: "addon-1" },
    });

    await waitFor(() => expect(mockedList).toHaveBeenCalledTimes(1));
    expect(mockedList).toHaveBeenLastCalledWith("org-1", "addon-1");

    rerender({ id: "addon-2" });

    await waitFor(() => expect(mockedList).toHaveBeenCalledTimes(2));
    expect(mockedList).toHaveBeenLastCalledWith("org-1", "addon-2");
  });

  it("treats 404 as empty list with no error", async () => {
    const notFound = Object.assign(new Error("not found"), { __notFound: true });
    mockedList.mockRejectedValueOnce(notFound);

    const { result } = renderHook(() => usePostgresBackups("addon-1"));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.backups).toEqual([]);
    expect(result.current.error).toBeNull();
  });
});

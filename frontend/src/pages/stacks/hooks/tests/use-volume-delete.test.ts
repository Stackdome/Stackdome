// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { AxiosError, AxiosHeaders } from "axios";
import type { Stack, Volume } from "@/api/stacks";

vi.mock("@/api/volumes", () => ({
  deleteVolume: vi.fn(),
}));
vi.mock("@/api/stacks", () => ({
  getStackById: vi.fn(),
}));

import { deleteVolume } from "@/api/volumes";
import { getStackById } from "@/api/stacks";
import { useVolumeDelete } from "../use-volume-delete";

const mockedDelete = vi.mocked(deleteVolume);
const mockedGetStack = vi.mocked(getStackById);

const ids = { orgId: "org-1", teamName: "alpha", stackId: "stack-1" };

function mkStack(volumes: Partial<Volume>[]): Stack {
  return {
    id: "stack-1",
    spec: { volumes, stack_resources: [], connections: [] },
  } as unknown as Stack;
}

function mkAxiosError(status: number): AxiosError {
  const err = new AxiosError(`Request failed with status code ${status}`);
  err.response = {
    status,
    statusText: "",
    headers: {},
    config: { headers: new AxiosHeaders() } as never,
    data: {},
  };
  return err;
}

function mkArgs(overrides: Partial<Parameters<typeof useVolumeDelete>[0]> = {}) {
  const flush = vi.fn().mockResolvedValue(true);
  const notifyExternalUpdate = vi.fn();
  const onServerRefresh = vi.fn();
  const onRestoreVolume = vi.fn();
  const toast = vi.fn();
  return {
    args: {
      ids,
      draftSync: { flush, notifyExternalUpdate },
      onServerRefresh,
      onRestoreVolume,
      toast,
      ...overrides,
    },
    flush,
    notifyExternalUpdate,
    onServerRefresh,
    onRestoreVolume,
    toast,
  };
}

beforeEach(() => {
  mockedDelete.mockReset();
  mockedGetStack.mockReset();
});

describe("useVolumeDelete", () => {
  it("happy path: flush -> refetch (id lookup) -> deleteVolume(id) -> refetch -> notify+refresh, true", async () => {
    const { args, flush, notifyExternalUpdate, onServerRefresh, toast } = mkArgs();
    mockedGetStack
      .mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]))
      .mockResolvedValueOnce(mkStack([]));
    mockedDelete.mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });

    expect(ok).toBe(true);
    expect(flush).toHaveBeenCalledTimes(1);
    expect(mockedGetStack).toHaveBeenCalledTimes(2);
    expect(mockedDelete).toHaveBeenCalledWith("org-1", "alpha", "vol-1");
    expect(notifyExternalUpdate).toHaveBeenCalledTimes(1);
    expect(onServerRefresh).toHaveBeenCalledTimes(1);
    expect(toast).toHaveBeenCalledWith({
      title: "Volume deleted",
      description: '"data" and its data were deleted.',
    });
    expect(result.current.deleting).toBe(false);
  });

  it("ids null: returns true immediately, makes zero network calls", async () => {
    const { args, flush } = mkArgs({ ids: null });
    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });
    expect(ok).toBe(true);
    expect(flush).not.toHaveBeenCalled();
    expect(mockedGetStack).not.toHaveBeenCalled();
    expect(mockedDelete).not.toHaveBeenCalled();
  });

  it("flush fails: no delete call, restores the volume, returns false", async () => {
    const { args, onRestoreVolume, toast } = mkArgs({
      draftSync: { flush: vi.fn().mockResolvedValue(false), notifyExternalUpdate: vi.fn() },
    });
    mockedGetStack.mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]));

    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });

    expect(ok).toBe(false);
    expect(mockedDelete).not.toHaveBeenCalled();
    expect(mockedGetStack).toHaveBeenCalledTimes(1);
    expect(onRestoreVolume).toHaveBeenCalledWith(expect.objectContaining({ id: "vol-1", name: "data" }));
    expect(toast).toHaveBeenCalled();
  });

  it("volume absent post-flush (never persisted): no delete call, true", async () => {
    const { args, onServerRefresh, onRestoreVolume } = mkArgs();
    mockedGetStack.mockResolvedValueOnce(mkStack([]));

    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });

    expect(ok).toBe(true);
    expect(mockedDelete).not.toHaveBeenCalled();
    expect(onServerRefresh).not.toHaveBeenCalled();
    expect(onRestoreVolume).not.toHaveBeenCalled();
  });

  it("409 conflict: restores the volume, toasts in-use, returns false", async () => {
    const { args, onRestoreVolume, toast } = mkArgs();
    mockedGetStack
      .mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]))
      .mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]));
    mockedDelete.mockRejectedValueOnce(mkAxiosError(409));

    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });

    expect(ok).toBe(false);
    expect(onRestoreVolume).toHaveBeenCalledWith(expect.objectContaining({ id: "vol-1" }));
    expect(toast).toHaveBeenCalledWith(
      expect.objectContaining({
        title: expect.stringMatching(/in use/i),
        description: expect.stringMatching(/data.*mounted by a running deployment/i),
      }),
    );
  });

  it("network reject (non-409): restores the volume, toasts generic failure, returns false", async () => {
    const { args, onRestoreVolume, toast } = mkArgs();
    mockedGetStack
      .mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]))
      .mockResolvedValueOnce(mkStack([{ id: "vol-1", name: "data" }]));
    mockedDelete.mockRejectedValueOnce(new Error("network down"));

    const { result } = renderHook(() => useVolumeDelete(args));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.deleteVolume("data");
    });

    expect(ok).toBe(false);
    expect(onRestoreVolume).toHaveBeenCalledWith(expect.objectContaining({ id: "vol-1" }));
    expect(toast).toHaveBeenCalledWith(
      expect.objectContaining({
        title: expect.stringMatching(/couldn.t delete volume/i),
        description: expect.stringMatching(/data.*was not deleted.*check your connection/i),
      }),
    );
  });
});

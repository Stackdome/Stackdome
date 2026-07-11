// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/projects", () => ({
  listProjects: vi.fn(), createProject: vi.fn(), renameProject: vi.fn(), deleteProject: vi.fn(),
}));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "nope") };
});

import { listProjects, createProject, deleteProject } from "@/api/projects";
import { useProjects } from "../use-projects";

beforeEach(() => { vi.mocked(listProjects).mockReset(); vi.mocked(createProject).mockReset(); vi.mocked(deleteProject).mockReset(); });
afterEach(() => cleanup());

describe("useProjects", () => {
  it("loads projects and computes onlyDefault", async () => {
    vi.mocked(listProjects).mockResolvedValue({ items: [{ id: "t1", name: "engineering", default_project: true }] } as never);
    const { result } = renderHook(() => useProjects());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.projects).toHaveLength(1);
    expect(result.current.onlyDefault).toBe(true);
  });

  it("create returns ok and refetches", async () => {
    vi.mocked(listProjects).mockResolvedValue({ items: [] } as never);
    vi.mocked(createProject).mockResolvedValue({ id: "t2", name: "data" } as never);
    const { result } = renderHook(() => useProjects());
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.create("data"); });
    expect(createProject).toHaveBeenCalledWith("org-1", { name: "data" });
    expect(r).toEqual({ ok: true });
  });

  it("remove maps failure", async () => {
    vi.mocked(listProjects).mockResolvedValue({ items: [] } as never);
    vi.mocked(deleteProject).mockRejectedValue(new Error("x"));
    const { result } = renderHook(() => useProjects());
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.remove("data"); });
    expect(r).toEqual({ ok: false, error: "nope" });
  });
});

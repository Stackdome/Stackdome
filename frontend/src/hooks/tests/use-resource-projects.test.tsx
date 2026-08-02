// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";

vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/projects", () => ({ listProjects: vi.fn() }));

import { listProjects } from "@/api/projects";
import { useResourceProjects } from "@/hooks/use-resource-projects";

const PROJECTS = {
  items: [
    { id: "td", name: "default", default_project: true },
    { id: "tq", name: "quality-eng", default_project: false },
  ],
};

beforeEach(() => {
  vi.mocked(listProjects).mockReset();
  vi.mocked(listProjects).mockResolvedValue(PROJECTS as never);
});

describe("useResourceProjects", () => {
  it("resolves a project_id to its name from the full org project list", async () => {
    const { result } = renderHook(() => useResourceProjects());
    await waitFor(() => expect(result.current.projectNameById("td")).toBe("default"));
    expect(result.current.projectNameById("tq")).toBe("quality-eng");
  });

  it("returns undefined for an unknown or missing project_id", async () => {
    const { result } = renderHook(() => useResourceProjects());
    await waitFor(() => expect(result.current.projectNameById("td")).toBe("default"));
    expect(result.current.projectNameById("nope")).toBeUndefined();
    expect(result.current.projectNameById(undefined)).toBeUndefined();
  });

  it("exposes the org default project name (works even when the user is not a member of it)", async () => {
    const { result } = renderHook(() => useResourceProjects());
    await waitFor(() => expect(result.current.defaultProjectName).toBe("default"));
  });

  it("defaultProjectName is undefined when no project is marked default", async () => {
    vi.mocked(listProjects).mockResolvedValue({ items: [{ id: "tq", name: "quality-eng", default_project: false }] } as never);
    const { result } = renderHook(() => useResourceProjects());
    await waitFor(() => expect(result.current.projectNameById("tq")).toBe("quality-eng"));
    expect(result.current.defaultProjectName).toBeUndefined();
  });
});

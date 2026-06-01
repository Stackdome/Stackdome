// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/teams", () => ({ listTeams: vi.fn() }));

import { listTeams } from "@/api/teams";
import { useResourceTeams } from "@/hooks/use-resource-teams";

const TEAMS = {
  items: [
    { id: "td", name: "default", default_team: true },
    { id: "tq", name: "quality-eng", default_team: false },
  ],
};

beforeEach(() => {
  vi.mocked(listTeams).mockReset();
  vi.mocked(listTeams).mockResolvedValue(TEAMS as never);
});

describe("useResourceTeams", () => {
  it("resolves a team_id to its name from the full org team list", async () => {
    const { result } = renderHook(() => useResourceTeams());
    await waitFor(() => expect(result.current.teamNameById("td")).toBe("default"));
    expect(result.current.teamNameById("tq")).toBe("quality-eng");
  });

  it("returns undefined for an unknown or missing team_id", async () => {
    const { result } = renderHook(() => useResourceTeams());
    await waitFor(() => expect(result.current.teamNameById("td")).toBe("default"));
    expect(result.current.teamNameById("nope")).toBeUndefined();
    expect(result.current.teamNameById(undefined)).toBeUndefined();
  });

  it("exposes the org default team name (works even when the user is not a member of it)", async () => {
    const { result } = renderHook(() => useResourceTeams());
    await waitFor(() => expect(result.current.defaultTeamName).toBe("default"));
  });

  it("defaultTeamName is undefined when no team is marked default", async () => {
    vi.mocked(listTeams).mockResolvedValue({ items: [{ id: "tq", name: "quality-eng", default_team: false }] } as never);
    const { result } = renderHook(() => useResourceTeams());
    await waitFor(() => expect(result.current.teamNameById("tq")).toBe("quality-eng"));
    expect(result.current.defaultTeamName).toBeUndefined();
  });
});

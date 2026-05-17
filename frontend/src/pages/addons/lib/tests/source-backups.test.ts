// @vitest-environment node
import { describe, it, expect } from "vitest";
import { completedNewestFirst } from "../source-backups";
import type { PostgresBackup } from "@/api/postgres-backups";

const b = (p: Partial<PostgresBackup>): PostgresBackup => p as PostgresBackup;

describe("completedNewestFirst", () => {
  it("keeps only completed backups", () => {
    const list = [
      b({ id: "1", phase: "completed", completed_at: "2026-05-01T00:00:00Z" }),
      b({ id: "2", phase: "running" }),
      b({ id: "3", phase: "failed" }),
      b({ id: "4", phase: "pending" }),
    ];
    expect(completedNewestFirst(list).map((x) => x.id)).toEqual(["1"]);
  });

  it("sorts completed newest-first by completed_at", () => {
    const list = [
      b({ id: "old", phase: "completed", completed_at: "2026-05-01T00:00:00Z" }),
      b({ id: "new", phase: "completed", completed_at: "2026-05-10T00:00:00Z" }),
      b({ id: "mid", phase: "completed", completed_at: "2026-05-05T00:00:00Z" }),
    ];
    expect(completedNewestFirst(list).map((x) => x.id)).toEqual(["new", "mid", "old"]);
  });

  it("handles missing completed_at (sorts last) and empty input", () => {
    expect(completedNewestFirst([])).toEqual([]);
    const list = [
      b({ id: "withts", phase: "completed", completed_at: "2026-05-05T00:00:00Z" }),
      b({ id: "nots", phase: "completed" }),
    ];
    expect(completedNewestFirst(list).map((x) => x.id)).toEqual(["withts", "nots"]);
  });
});

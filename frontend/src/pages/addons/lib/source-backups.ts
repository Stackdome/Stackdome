import type { PostgresBackup } from "@/api/postgres-backups";

/**
 * Restorable backups for the WAL-off path: only `completed` runs, newest
 * first by `completed_at`. Entries with no `completed_at` sort last.
 */
export function completedNewestFirst(
  backups: PostgresBackup[],
): PostgresBackup[] {
  return backups
    .filter((x) => x.phase === "completed")
    .sort((a, z) => {
      const ta = a.completed_at ? Date.parse(a.completed_at) : -Infinity;
      const tz = z.completed_at ? Date.parse(z.completed_at) : -Infinity;
      return tz - ta;
    });
}

import type { PostgresAddon } from "@/api/addons";

/**
 * Addons that can be a PITR source: backups land in an object store and WAL
 * archiving was on (continuous WAL is required for point-in-time recovery).
 * Optionally exclude one id (e.g. the addon being created, if relevant).
 */
export function eligibleRestoreSources(
  addons: PostgresAddon[],
  excludeId?: string,
): PostgresAddon[] {
  return addons.filter((a) => {
    if (excludeId && a.id === excludeId) return false;
    const b = a.spec?.backup;
    return Boolean(b?.object_store_id) && b?.wal_archiving === true;
  });
}

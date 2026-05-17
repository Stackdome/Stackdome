import type { PostgresAddon } from "@/api/addons";

/**
 * Addons that can be a restore source: backups land in an object store.
 * WAL archiving is NOT required here — without it, only discrete completed
 * backups are restorable (restore_from_backup); with it, point-in-time is
 * additionally available (restore_from_object_store). The caller decides
 * which mode based on the chosen source's wal_archiving.
 * Optionally exclude one id (e.g. the addon being created, if relevant).
 */
export function eligibleRestoreSources(
  addons: PostgresAddon[],
  excludeId?: string,
): PostgresAddon[] {
  return addons.filter((a) => {
    if (excludeId && a.id === excludeId) return false;
    return Boolean(a.spec?.backup?.object_store_id);
  });
}

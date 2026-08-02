import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import { listPostgresBackups, type PostgresBackup } from "@/api/postgres-backups";
import { completedNewestFirst } from "../lib/source-backups";

/**
 * Completed backups (newest-first) of a source addon, for the WAL-off
 * restore path. Fetches only when sourceAddonId is set.
 */
export function useSourceBackups(sourceAddonId: string | undefined) {
  const [raw, setRaw] = useState<PostgresBackup[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const orgId = getCurrentOrganizationId();
  const cancelRef = useRef<{ cancelled: boolean }>({ cancelled: false });

  const fetchBackups = useCallback(async () => {
    if (!orgId || !sourceAddonId) {
      setRaw([]);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const data = await listPostgresBackups(orgId, sourceAddonId);
      if (cancelRef.current.cancelled) return;
      setRaw(data.items || []);
    } catch (e: unknown) {
      if (cancelRef.current.cancelled) return;
      if (isNotFoundError(e)) setRaw([]);
      else setError(getErrorMessage(e));
    } finally {
      if (!cancelRef.current.cancelled) setLoading(false);
    }
  }, [orgId, sourceAddonId]);

  useEffect(() => {
    cancelRef.current = { cancelled: false };
    void fetchBackups();
    return () => {
      cancelRef.current.cancelled = true;
    };
  }, [fetchBackups]);

  const backups = useMemo(() => completedNewestFirst(raw), [raw]);
  return { backups, loading, error };
}

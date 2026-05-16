import { useCallback, useEffect, useRef, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import {
  listPostgresBackups,
  isTerminalPhase,
  type PostgresBackup,
} from "@/api/postgres-backups";

const POLL_INTERVAL_MS = 5000;
const DEFAULT_PAGE_SIZE = 10;

export function usePostgresBackups(
  addonId: string | undefined,
  pageSize = DEFAULT_PAGE_SIZE,
) {
  const [backups, setBackups] = useState<PostgresBackup[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const orgId = getCurrentOrganizationId();
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const cancelRef = useRef<{ cancelled: boolean }>({ cancelled: false });

  const fetchBackups = useCallback(async () => {
    if (!orgId || !addonId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await listPostgresBackups(orgId, addonId, {
        limit: pageSize,
        offset: page * pageSize,
      });
      if (cancelRef.current.cancelled) return;
      setBackups(data.items || []);
      setTotal(data.total ?? 0);
    } catch (e: unknown) {
      if (cancelRef.current.cancelled) return;
      if (isNotFoundError(e)) {
        setBackups([]);
        setTotal(0);
      } else {
        setError(getErrorMessage(e));
      }
    } finally {
      if (!cancelRef.current.cancelled) setLoading(false);
    }
  }, [orgId, addonId, page, pageSize]);

  // Reset to the first page whenever the addon changes.
  useEffect(() => {
    setPage(0);
  }, [addonId]);

  // Fetch on page/addon change + reset cancellation.
  useEffect(() => {
    cancelRef.current = { cancelled: false };
    void fetchBackups();
    return () => {
      cancelRef.current.cancelled = true;
    };
  }, [fetchBackups]);

  // Poll while any backup on the current page is non-terminal. Newest runs are
  // on page 1, so this still catches in-progress backups in practice.
  useEffect(() => {
    const hasNonTerminal = backups.some((b) => !isTerminalPhase(b.phase));
    if (!hasNonTerminal) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }
    if (intervalRef.current) return; // already polling
    intervalRef.current = setInterval(() => {
      void fetchBackups();
    }, POLL_INTERVAL_MS);
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [backups, fetchBackups]);

  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  return {
    backups,
    total,
    page,
    pageSize,
    pageCount,
    setPage,
    loading,
    error,
    refetch: fetchBackups,
  };
}

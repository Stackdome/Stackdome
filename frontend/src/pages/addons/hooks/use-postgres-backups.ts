import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import {
  listPostgresBackups,
  isTerminalPhase,
  type PostgresBackup,
} from "@/api/postgres-backups";

const POLL_INTERVAL_MS = 5000;
const DEFAULT_PAGE_SIZE = 10;

// The backups list endpoint ignores limit/offset and omits `total`: it always
// returns every run in `items`. So we fetch the full list and paginate on the
// client. Switch back to server paging if/when the API honours limit/offset.
export function usePostgresBackups(
  addonId: string | undefined,
  pageSize = DEFAULT_PAGE_SIZE,
) {
  const [allBackups, setAllBackups] = useState<PostgresBackup[]>([]);
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
      const data = await listPostgresBackups(orgId, addonId);
      if (cancelRef.current.cancelled) return;
      setAllBackups(data.items || []);
    } catch (e: unknown) {
      if (cancelRef.current.cancelled) return;
      if (isNotFoundError(e)) {
        setAllBackups([]);
      } else {
        setError(getErrorMessage(e));
      }
    } finally {
      if (!cancelRef.current.cancelled) setLoading(false);
    }
  }, [orgId, addonId]);

  // Reset to the first page whenever the addon changes.
  useEffect(() => {
    setPage(0);
  }, [addonId]);

  useEffect(() => {
    cancelRef.current = { cancelled: false };
    void fetchBackups();
    return () => {
      cancelRef.current.cancelled = true;
    };
  }, [fetchBackups]);

  // Poll while any run is non-terminal (in-progress backup somewhere in the list).
  useEffect(() => {
    const hasNonTerminal = allBackups.some((b) => !isTerminalPhase(b.phase));
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
  }, [allBackups, fetchBackups]);

  const total = allBackups.length;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  // Clamp the page if the list shrank (e.g. runs aged out between polls).
  const safePage = Math.min(page, pageCount - 1);
  useEffect(() => {
    if (page !== safePage) setPage(safePage);
  }, [page, safePage]);

  const backups = useMemo(
    () => allBackups.slice(safePage * pageSize, safePage * pageSize + pageSize),
    [allBackups, safePage, pageSize],
  );

  return {
    backups,
    total,
    page: safePage,
    pageSize,
    pageCount,
    setPage,
    loading,
    error,
    refetch: fetchBackups,
  };
}

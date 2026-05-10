import { useCallback, useEffect, useRef, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import {
  listPostgresBackups,
  isTerminalPhase,
  type PostgresBackup,
} from "@/api/postgres-backups";

const POLL_INTERVAL_MS = 5000;

export function usePostgresBackups(addonId: string | undefined) {
  const [backups, setBackups] = useState<PostgresBackup[]>([]);
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
      setBackups(data.items || []);
    } catch (e: unknown) {
      if (cancelRef.current.cancelled) return;
      if (isNotFoundError(e)) {
        setBackups([]);
      } else {
        setError(getErrorMessage(e));
      }
    } finally {
      if (!cancelRef.current.cancelled) setLoading(false);
    }
  }, [orgId, addonId]);

  // Initial fetch + reset cancellation on addon change
  useEffect(() => {
    cancelRef.current = { cancelled: false };
    void fetchBackups();
    return () => {
      cancelRef.current.cancelled = true;
    };
  }, [fetchBackups]);

  // Poll while any backup is non-terminal
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

  return { backups, loading, error, refetch: fetchBackups };
}

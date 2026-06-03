import { useCallback, useEffect, useRef, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import * as addonsApi from "@/api/addons";
import type { PostgresAddon } from "@/api/addons";

const POLL_INTERVAL_MS = 5000;

const TERMINAL_STATES: ReadonlySet<string> = new Set([
  "Ready",
  "Error",
  "Hibernated",
  "Fenced",
]);

function hasPendingItem(items: PostgresAddon[]): boolean {
  return items.some((item) => {
    const state = item.status?.state;
    return !state || !TERMINAL_STATES.has(state);
  });
}

export function usePostgresAddons() {
  const [addons, setAddons] = useState<PostgresAddon[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const orgId = getCurrentOrganizationId();
  const { defaultTeamName } = useResourceTeams();

  const fetchAddons = useCallback(async () => {
    if (!orgId || !defaultTeamName) return;
    try {
      const data = await addonsApi.listPostgresAddons(orgId, defaultTeamName);
      setAddons(data.items || []);
      setError(null);
    } catch (e: unknown) {
      if (isNotFoundError(e)) {
        setAddons([]);
      } else {
        console.error("Failed to fetch addons:", e);
        setError(getErrorMessage(e));
      }
    }
  }, [orgId, defaultTeamName]);

  const initialFetch = useCallback(async () => {
    setLoading(true);
    await fetchAddons();
    setLoading(false);
  }, [fetchAddons]);

  const refetch = useCallback(() => {
    fetchAddons();
  }, [fetchAddons]);

  useEffect(() => {
    initialFetch();
  }, [initialFetch]);

  useEffect(() => {
    if (pollRef.current) {
      clearTimeout(pollRef.current);
      pollRef.current = null;
    }
    if (!hasPendingItem(addons)) return;

    pollRef.current = setTimeout(() => {
      fetchAddons();
    }, POLL_INTERVAL_MS);

    return () => {
      if (pollRef.current) {
        clearTimeout(pollRef.current);
        pollRef.current = null;
      }
    };
  }, [addons, fetchAddons]);

  return { addons, loading, error, refetch };
}

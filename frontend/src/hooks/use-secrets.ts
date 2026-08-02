import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import type { Secret } from "@/api/secrets";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import * as secretsApi from "@/api/secrets";

export function useSecrets() {
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const orgId = getCurrentOrganizationId();

  const fetchSecrets = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await secretsApi.getSecrets(orgId);
      setSecrets(data.items || []);
    } catch (e: unknown) {
      if (isNotFoundError(e)) {
        // Treat "not found" as empty state, not an error
        setSecrets([]);
      } else {
        console.error('Failed to fetch secrets:', e);
        setError(getErrorMessage(e));
      }
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  const refetch = useCallback(() => {
    fetchSecrets();
  }, [fetchSecrets]);

  useEffect(() => {
    fetchSecrets();
  }, [fetchSecrets]);

  return {
    secrets,
    loading,
    error,
    refetch,
  };
}

import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import type { ObjectStore } from "@/pages/object-stores/types";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import * as objectStoresApi from "@/api/object-stores";

export function useObjectStores() {
  const [objectStores, setObjectStores] = useState<ObjectStore[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const orgId = getCurrentOrganizationId();

  const fetchObjectStores = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await objectStoresApi.getObjectStores(orgId);
      setObjectStores(data.items || []);
    } catch (e: unknown) {
      if (isNotFoundError(e)) {
        setObjectStores([]);
      } else {
        console.error("Failed to fetch object stores:", e);
        setError(getErrorMessage(e));
      }
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  const refetch = useCallback(() => {
    fetchObjectStores();
  }, [fetchObjectStores]);

  useEffect(() => {
    fetchObjectStores();
  }, [fetchObjectStores]);

  return { objectStores, loading, error, refetch };
}

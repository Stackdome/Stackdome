import { useCallback, useEffect, useState } from "react";
import {
  listAllPreviewEnvs,
  TERMINAL_PHASES,
  type PreviewStack,
} from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";

const POLL_MS = 7_000;

export interface UsePreviewEnvs {
  envs: PreviewStack[];
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

function hasNonTerminal(envs: PreviewStack[]): boolean {
  return envs.some((e) => {
    const phase = e.status?.phase;
    return phase == null || !TERMINAL_PHASES.includes(phase);
  });
}

export function usePreviewEnvs(configId?: string): UsePreviewEnvs {
  const { defaultProjectName } = useResourceProjects();
  const [envs, setEnvs] = useState<PreviewStack[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultProjectName) return;
    try {
      const list = await listAllPreviewEnvs(orgId, defaultProjectName, configId);
      setEnvs(list);
      setError(null);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [configId, defaultProjectName]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!hasNonTerminal(envs)) return;
    const t = setInterval(() => void refresh(), POLL_MS);
    return () => clearInterval(t);
  }, [envs, refresh]);

  return { envs, loading, error, refresh };
}

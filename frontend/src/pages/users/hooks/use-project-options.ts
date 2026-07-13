import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { listProjects, type Project } from "@/api/projects";

export function useProjectOptions() {
  const orgId = getCurrentOrganizationId();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const fetchProjects = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    try {
      const res = await listProjects(orgId);
      setProjects(res.items ?? []);
    } catch {
      setProjects([]);
    } finally {
      setLoading(false);
    }
  }, [orgId]);
  useEffect(() => { void fetchProjects(); }, [fetchProjects]);
  return { projects, loading };
}

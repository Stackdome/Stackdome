import { useEffect, useState } from "react";
import { getAppConfig, getCachedAppConfig } from "@/api/config";

export function useAppConfig() {
  // Seed from the session cache so a warm config resolves on first paint
  // (no loading flash) — the fetch is warmed at app entry (main.tsx).
  const cached = getCachedAppConfig();
  const [githubOAuth, setGithubOAuth] = useState(Boolean(cached?.github_oauth));
  const [loading, setLoading] = useState(cached === null);

  useEffect(() => {
    let active = true;
    getAppConfig()
      .then((cfg) => {
        if (active) setGithubOAuth(Boolean(cfg.github_oauth));
      })
      .catch(() => {
        if (active) setGithubOAuth(false); // fail-closed: hide the button on error
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  return { githubOAuth, loading };
}

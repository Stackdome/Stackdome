import { useEffect, useState } from "react";
import { getAppConfig } from "@/api/config";

export function useAppConfig() {
  const [githubOAuth, setGithubOAuth] = useState(false);
  const [loading, setLoading] = useState(true);

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

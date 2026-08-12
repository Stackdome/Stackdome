import { useEffect, useState } from "react";
import { getAppConfig, getCachedAppConfig } from "@/api/config";
import type { components } from "@/api/types/openapi";

type TurnstileConfig = components["schemas"]["TurnstileConfigResponse"];

export function useSignupConfig() {
  const cached = getCachedAppConfig();
  const [turnstile, setTurnstile] = useState<TurnstileConfig | undefined>(
    cached?.signup?.turnstile,
  );
  const [loading, setLoading] = useState(cached === null);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    getAppConfig()
      .then((config) => {
        if (!active) return;
        setTurnstile(config.signup?.turnstile);
        setError(null);
      })
      .catch((cause: unknown) => {
        if (!active) return;
        setTurnstile(undefined);
        setError(cause instanceof Error ? cause : new Error("failed to load app config"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  return { turnstile, loading, error };
}

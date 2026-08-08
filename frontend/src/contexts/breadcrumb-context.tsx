import { useState, useCallback } from "react";
import type { ReactNode } from "react";
import { createContext } from "react";

export type BreadcrumbContextType = {
  customLabels: Record<string, string>;
  setCustomLabel: (path: string, label: string) => void;
  loadingLabels: Record<string, boolean>;
  setPathLoading: (path: string, isLoading: boolean) => void;
  nonClickablePaths: Record<string, boolean>;
  registerNonClickablePath: (path: string) => () => void;
  /**
   * The origin of the journey currently on screen, or `null` on an ordinary
   * page. A journey is a task launched from a main screen — `New stack` — as
   * opposed to a place you navigated down into (§12a).
   *
   * It lives here rather than as a `SheetHeader` prop because the header is
   * rendered once by `AppLayout`, and because registering it changes the trail
   * as well as adding the back arrow: a journey shows **its title alone**.
   */
  journeyOrigin: string | null;
  registerJourney: (origin: string) => () => void;
};

export const BreadcrumbContext = createContext<BreadcrumbContextType>({
  customLabels: {},
  setCustomLabel: () => {},
  loadingLabels: {},
  setPathLoading: () => {},
  nonClickablePaths: {},
  registerNonClickablePath: () => () => {},
  journeyOrigin: null,
  registerJourney: () => () => {},
});

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const [customLabels, setCustomLabels] = useState<Record<string, string>>({});
  const [loadingLabels, setLoadingLabels] = useState<Record<string, boolean>>({});
  const [nonClickablePaths, setNonClickablePaths] = useState<Record<string, boolean>>({});
  const [journeyOrigin, setJourneyOrigin] = useState<string | null>(null);

  const setCustomLabel = useCallback((path: string, label: string) => {
    setCustomLabels((prev) => ({
      ...prev,
      [path]: label,
    }));
  }, []);

  const setPathLoading = useCallback((path: string, isLoading: boolean) => {
    setLoadingLabels((prev) => ({
      ...prev,
      [path]: isLoading,
    }));
  }, []);

  // Only one journey can be on screen at a time, so this is a single value
  // rather than a map. Unregistering clears it only if it is still ours — on a
  // journey-to-journey move the next page mounts before this one unmounts.
  const registerJourney = useCallback((origin: string) => {
    setJourneyOrigin(origin);
    return () => setJourneyOrigin((prev) => (prev === origin ? null : prev));
  }, []);

  const registerNonClickablePath = useCallback((path: string) => {
    setNonClickablePaths((prev) =>
      prev[path] ? prev : { ...prev, [path]: true },
    );
    return () => {
      setNonClickablePaths((prev) => {
        if (!(path in prev)) return prev;
        const next = { ...prev };
        delete next[path];
        return next;
      });
    };
  }, []);

  return (
    <BreadcrumbContext.Provider
      value={{
        customLabels,
        setCustomLabel,
        loadingLabels,
        setPathLoading,
        nonClickablePaths,
        registerNonClickablePath,
        journeyOrigin,
        registerJourney,
      }}
    >
      {children}
    </BreadcrumbContext.Provider>
  );
}

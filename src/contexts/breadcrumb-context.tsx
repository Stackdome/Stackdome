import { useState, useCallback } from "react";
import type { ReactNode } from "react";
import { createContext } from "react";

export type BreadcrumbContextType = {
  customLabels: Record<string, string>;
  setCustomLabel: (path: string, label: string) => void;
  loadingLabels: Record<string, boolean>;
  setPathLoading: (path: string, isLoading: boolean) => void;
};

export const BreadcrumbContext = createContext<BreadcrumbContextType>({
  customLabels: {},
  setCustomLabel: () => {},
  loadingLabels: {},
  setPathLoading: () => {},
});

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const [customLabels, setCustomLabels] = useState<Record<string, string>>({});
  const [loadingLabels, setLoadingLabels] = useState<Record<string, boolean>>({});

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

  return (
    <BreadcrumbContext.Provider value={{ customLabels, setCustomLabel, loadingLabels, setPathLoading }}>
      {children}
    </BreadcrumbContext.Provider>
  );
}

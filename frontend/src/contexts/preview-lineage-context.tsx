import { createContext, useCallback, useState } from "react";
import type { ReactNode } from "react";

/** Preview-stack labels the API server writes (pkg/models/preview_stack.go). */
export const PREVIEW_STACK_LABEL = "preview.stackdome.io/preview-stack";
export const PREVIEW_CONFIG_ID_LABEL = "preview.stackdome.io/config-id";

export interface PreviewLineage {
  configId: string;
  /** Config name, or undefined until the config request resolves. */
  configName?: string;
}

export type PreviewLineageContextType = {
  /** Set while the open stack is a preview environment; null otherwise. */
  lineage: PreviewLineage | null;
  setLineage: (lineage: PreviewLineage | null) => void;
};

export const PreviewLineageContext = createContext<PreviewLineageContextType>({
  lineage: null,
  setLineage: () => {},
});

/**
 * A preview stack lives at /stacks/<id> like any other, so its owning config is
 * invisible to the chrome. The stack page reads it off the stack's labels and
 * publishes it here; the breadcrumb and sidebar read it back.
 */
export function PreviewLineageProvider({ children }: { children: ReactNode }) {
  const [lineage, setLineageState] = useState<PreviewLineage | null>(null);

  const setLineage = useCallback((next: PreviewLineage | null) => {
    setLineageState((prev) => {
      if (prev === next) return prev;
      if (prev && next && prev.configId === next.configId && prev.configName === next.configName) return prev;
      return next;
    });
  }, []);

  return (
    <PreviewLineageContext.Provider value={{ lineage, setLineage }}>
      {children}
    </PreviewLineageContext.Provider>
  );
}

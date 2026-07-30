import { createContext } from "react";

/** Lets canvas chrome (the zen toggle) drive the editor header collapse
 *  without prop drilling. Default no-op so canvas pieces render standalone
 *  in tests. */
export const HeaderCollapseContext = createContext<{
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
    }>({ collapsed: false, setCollapsed: () => {} });

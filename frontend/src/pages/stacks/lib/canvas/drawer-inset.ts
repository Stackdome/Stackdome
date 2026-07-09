import { createContext, useContext } from "react";

/**
 * Lets the floating drawer stack report the horizontal space it occupies
 * (viewport px from the right edge) so ancestor chrome can push content left
 * instead of being overlaid. Defaults to a no-op so DrawerStack renders fine
 * without a provider (tests, storybook).
 */
export interface DrawerInset {
  setInset: (px: number) => void;
}

export const DrawerInsetContext = createContext<DrawerInset>({ setInset: () => {} });

export function useDrawerInset(): DrawerInset {
  return useContext(DrawerInsetContext);
}

export const DRAWER_BASE_INSET_PX = 12;
export const DRAWER_STAGGER_X_PX = 16;
export const DRAWER_PANEL_WIDTH_PX = 680;

/**
 * Horizontal space (px from the viewport's right edge) the drawer stack
 * occupies for a given panel count. The panel width clamp mirrors the
 * panels' max-w-[calc(100vw-24px)].
 */
export function computeDrawerInset(panelCount: number, viewportWidth: number): number {
  if (panelCount <= 0) return 0;
  const panelWidth = Math.min(DRAWER_PANEL_WIDTH_PX, viewportWidth - 2 * DRAWER_BASE_INSET_PX);
  return DRAWER_BASE_INSET_PX + panelWidth + DRAWER_STAGGER_X_PX * (panelCount - 1);
}

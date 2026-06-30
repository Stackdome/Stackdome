const CANVAS_LS_KEY = "stackCanvas";

/**
 * The stack canvas editor is opt-in while it is being built: a build-time env
 * flag turns it on for everyone, and a per-browser localStorage override lets
 * us flip it on for a single session without a rebuild.
 *
 * Any storage failure (private mode, SSR) resolves to "off" rather than
 * throwing — callers never have to guard this.
 */
export function isCanvasEnabled(): boolean {
  if (import.meta.env.VITE_STACK_CANVAS === "true") return true;
  try {
    return localStorage.getItem(CANVAS_LS_KEY) === "1";
  } catch {
    return false;
  }
}

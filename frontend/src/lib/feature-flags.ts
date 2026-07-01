const CANVAS_LS_KEY = "stackCanvas";

/**
 * The stack canvas editor is now the default stack experience. The legacy
 * accordion form is kept behind an explicit opt-out (to be removed later):
 * set `VITE_STACK_CANVAS=false` at build time, or `localStorage.stackCanvas`
 * to `"0"` per-browser, to fall back to the old form.
 *
 * Any storage failure (private mode, SSR) resolves to the canvas default
 * rather than throwing — callers never have to guard this.
 */
export function isCanvasEnabled(): boolean {
  if (import.meta.env.VITE_STACK_CANVAS === "false") return false;
  try {
    return localStorage.getItem(CANVAS_LS_KEY) !== "0";
  } catch {
    return true;
  }
}

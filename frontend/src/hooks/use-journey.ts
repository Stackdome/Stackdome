import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

/**
 * Declares this page a **journey** — a task launched from a main screen, which
 * you finish or abandon, as opposed to a place you navigated down into (§12a).
 *
 * The header responds by showing a back arrow and **the page title alone**: the
 * trail is dropped, because the sidebar already says which section you are in
 * and `Stacks /` would be the third way back off one screen.
 *
 * `origin` is the fallback destination, used only when there is no history to
 * go back through — a deep link, or a fresh tab.
 *
 * `title` names the page. Without it the header falls back to capitalising the
 * last URL segment, which gives `New` rather than `New stack`.
 *
 * ```tsx
 * useJourney("/stacks", "New stack")   // in the New stack page
 * ```
 */
export function useJourney(origin: string, title?: string) {
  const { registerJourney, setCustomLabel } = useBreadcrumb();
  const { pathname } = useLocation();

  useEffect(() => registerJourney(origin), [registerJourney, origin]);

  useEffect(() => {
    if (title) setCustomLabel(pathname, title);
  }, [setCustomLabel, pathname, title]);
}

import { useEffect } from "react";

/** Message a GitHub-App-install popup posts to its opener before closing. */
export const GITHUB_APP_INSTALLED_MESSAGE = "stackdome:github-app-installed";
export const GITHUB_APP_SETUP_ERROR_MESSAGE = "stackdome:github-app-setup-error";

/**
 * GitHub's App setup URL is the hub root, so after an install the popup lands
 * on the SPA with ?setup_action=install (or ?setup_error=<reason> when the
 * callback failed). When that happens inside a popup (window.opener set),
 * notify the opener and close; the opener's connect flow (use-github-connect)
 * is listening. Outside a popup the page's own setup_error handling toasts.
 */
export function useGithubSetupLanding(): void {
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const setupError = params.get("setup_error");
    if (!params.get("setup_action") && !setupError) return;
    if (!window.opener) return;
    (window.opener as Window).postMessage(
      setupError
        ? { type: GITHUB_APP_SETUP_ERROR_MESSAGE, reason: setupError }
        : { type: GITHUB_APP_INSTALLED_MESSAGE },
      window.location.origin,
    );
    window.close();
  }, []);
}

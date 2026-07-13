import { useEffect } from "react";

/** Message a GitHub-App-install popup posts to its opener before closing. */
export const GITHUB_APP_INSTALLED_MESSAGE = "stackdome:github-app-installed";

/**
 * GitHub's App setup URL is the hub root, so after an install the popup lands
 * on the SPA with ?setup_action=install. When that happens inside a popup
 * (window.opener set), notify the opener and close; the opener's connect flow
 * (use-github-connect) is listening.
 */
export function useGithubSetupLanding(): void {
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (!params.get("setup_action")) return;
    if (!window.opener) return;
    (window.opener as Window).postMessage(
      { type: GITHUB_APP_INSTALLED_MESSAGE },
      window.location.origin,
    );
    window.close();
  }, []);
}

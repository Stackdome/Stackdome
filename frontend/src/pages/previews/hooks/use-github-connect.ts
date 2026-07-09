import { useCallback, useEffect, useRef, useState } from "react";
import {
  createGitHubAppManifest,
  getGitIntegration,
  listGitIntegrations,
  listInstallations,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { GITHUB_APP_INSTALLED_MESSAGE } from "@/hooks/use-github-setup-landing";

export type GithubConnectState = "idle" | "waiting" | "connected" | "error";

const POPUP_NAME = "stackdome-github-connect";
const POPUP_FEATURES = "width=1020,height=800";
const POLL_MS = 5_000;
const MAX_POLLS = 40;

const CONNECTED_STATUSES = new Set(["installed", "active"]);

export interface GithubConnect {
  state: GithubConnectState;
  error: string | null;
  connect: () => Promise<void>;
  checkAgain: () => Promise<void>;
  integrationId: string | null;
}

/** Form-POSTs the manifest JSON to GitHub inside the named popup window. */
function postManifestToPopup(githubUrl: string, manifest: unknown): void {
  const form = document.createElement("form");
  form.method = "POST";
  form.action = githubUrl;
  form.target = POPUP_NAME;
  const input = document.createElement("input");
  input.type = "hidden";
  input.name = "manifest";
  input.value = JSON.stringify(manifest);
  form.appendChild(input);
  document.body.appendChild(form);
  form.submit();
  form.remove();
}

export function useGithubConnect(): GithubConnect {
  const [state, setState] = useState<GithubConnectState>("idle");
  const [error, setError] = useState<string | null>(null);
  const [integrationId, setIntegrationId] = useState<string | null>(null);
  const pollCount = useRef(0);

  const connect = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setError("No organization selected.");
      setState("error");
      return;
    }
    // Open synchronously inside the user gesture; navigate it via form POST after.
    const popup = window.open("", POPUP_NAME, POPUP_FEATURES);
    if (!popup) {
      setError("Popup blocked — allow popups for this site and try again.");
      setState("error");
      return;
    }
    try {
      const flow = await createGitHubAppManifest(orgId);
      postManifestToPopup(flow.github_url ?? "", flow.manifest);
      const list = await listGitIntegrations(orgId);
      const pending = (list.items ?? []).find(
        (i) => i.type === "github_app" && !CONNECTED_STATUSES.has(i.status ?? ""),
      );
      setIntegrationId(pending?.id ?? null);
      pollCount.current = 0;
      setError(null);
      setState("waiting");
    } catch (e) {
      popup.close();
      setError(getErrorMessage(e));
      setState("error");
    }
  }, []);

  const checkAgain = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integrationId) return;
    try {
      const installs = await listInstallations(orgId, integrationId, true);
      if ((installs.items ?? []).length > 0) setState("connected");
    } catch (e) {
      setError(getErrorMessage(e));
    }
  }, [integrationId]);

  // The popup posts a message right before closing itself.
  useEffect(() => {
    if (state !== "waiting") return;
    const onMessage = (ev: MessageEvent) => {
      if (ev.origin !== window.location.origin) return;
      if ((ev.data as { type?: string })?.type === GITHUB_APP_INSTALLED_MESSAGE) {
        setState("connected");
      }
    };
    window.addEventListener("message", onMessage);
    return () => window.removeEventListener("message", onMessage);
  }, [state]);

  // Fallback: poll the integration status in case the message never arrives
  // (popup closed early, blocked message, webhook race).
  useEffect(() => {
    if (state !== "waiting" || !integrationId) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    const timer = setInterval(async () => {
      pollCount.current += 1;
      if (pollCount.current > MAX_POLLS) {
        clearInterval(timer);
        setError("Timed out waiting for the GitHub App installation.");
        setState("error");
        return;
      }
      try {
        const integration = await getGitIntegration(orgId, integrationId);
        if (CONNECTED_STATUSES.has(integration.status ?? "")) setState("connected");
      } catch {
        // transient poll errors are ignored; checkAgain surfaces real ones
      }
    }, POLL_MS);
    return () => clearInterval(timer);
  }, [state, integrationId]);

  return { state, error, connect, checkAgain, integrationId };
}

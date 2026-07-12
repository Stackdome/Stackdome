import { useCallback, useEffect, useRef, useState } from "react";
import {
  createGitHubAppManifest,
  listGitIntegrations,
  listInstallations,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { GITHUB_APP_INSTALLED_MESSAGE } from "@/hooks/use-github-setup-landing";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
} from "@/pages/git-integrations/lib/derive-row";

export type GithubConnectState = "idle" | "waiting" | "connected" | "error";

const POPUP_NAME = "stackdome-github-connect";
const POPUP_FEATURES = "width=1020,height=800";
const POLL_MS = 5_000;
const MAX_POLLS = 40;

export const CONNECTED_STATUSES = new Set<string>([STATUS_INSTALLED, STATUS_ACTIVE]);

export interface GithubConnect {
  state: GithubConnectState;
  error: string | null;
  connect: () => Promise<void>;
  checkAgain: () => Promise<void>;
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
  const pollCount = useRef(0);
  // Holds the integration id across the poll interval, whose closure would
  // otherwise see the value captured when the interval was created. Internal
  // only — no consumer needs the id, so it isn't exposed via GithubConnect.
  const integrationIdRef = useRef<string | null>(null);

  const rememberIntegrationId = useCallback((id: string | null) => {
    integrationIdRef.current = id;
  }, []);

  /**
   * One connection probe, safe to run before the integration record exists:
   * the record is only created by GitHub's manifest callback (after the user
   * clicks "Create app" in the popup), so the id must be re-resolved lazily.
   * Once known, installations are re-listed from GitHub (refresh=true) — the
   * install webhook never arrives in local dev without a public URL, so a
   * status-only check would wait forever.
   */
  const probeConnected = useCallback(async (): Promise<boolean> => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) return false;
    let id = integrationIdRef.current;
    if (!id) {
      const list = await listGitIntegrations(orgId);
      const app = (list.items ?? []).find((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);
      if (!app?.id) return false;
      id = app.id;
      rememberIntegrationId(id);
      if (CONNECTED_STATUSES.has(app.status ?? "")) return true;
    }
    const installs = await listInstallations(orgId, id, true);
    return (installs.items ?? []).length > 0;
  }, [rememberIntegrationId]);

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
      // Usually null here — the integration record appears only after the
      // user confirms app creation in the popup; the poll re-resolves it.
      const list = await listGitIntegrations(orgId);
      const pending = (list.items ?? []).find(
        (i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP && !CONNECTED_STATUSES.has(i.status ?? ""),
      );
      rememberIntegrationId(pending?.id ?? null);
      pollCount.current = 0;
      setError(null);
      setState("waiting");
    } catch (e) {
      popup.close();
      setError(getErrorMessage(e));
      setState("error");
    }
  }, [rememberIntegrationId]);

  const checkAgain = useCallback(async () => {
    try {
      if (await probeConnected()) setState("connected");
    } catch (e) {
      setError(getErrorMessage(e));
    }
  }, [probeConnected]);

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

  // Fallback: poll in case the message never arrives (popup closed early,
  // blocked message, cross-origin landing, missed webhook). Deliberately not
  // gated on integrationId — the record doesn't exist until the user confirms
  // in the popup; probeConnected re-resolves it on every tick.
  useEffect(() => {
    if (state !== "waiting") return;
    const timer = setInterval(async () => {
      pollCount.current += 1;
      if (pollCount.current > MAX_POLLS) {
        clearInterval(timer);
        setError("Timed out waiting for the GitHub App installation.");
        setState("error");
        return;
      }
      try {
        if (await probeConnected()) {
          clearInterval(timer);
          setState("connected");
        }
      } catch {
        // transient poll errors are ignored; checkAgain surfaces real ones
      }
    }, POLL_MS);
    return () => clearInterval(timer);
  }, [state, probeConnected]);

  return { state, error, connect, checkAgain };
}

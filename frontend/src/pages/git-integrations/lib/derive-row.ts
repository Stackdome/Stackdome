import type { GitIntegration, GitInstallation } from "@/api/git-integrations";

export const GIT_INTEGRATION_TYPE_GITHUB_APP = "github_app" as const;
export const GIT_INTEGRATION_TYPE_CREDENTIALS = "git_credentials" as const;
export const STATUS_PENDING_INSTALL = "pending_install" as const;
export const STATUS_INSTALLED = "installed" as const;
export const STATUS_ACTIVE = "active" as const;
export const REPOSITORY_SELECTION_ALL = "all" as const;

export type ProviderId = "github" | "gitlab" | "bitbucket" | "gitea" | "other";

/** Detects the git host provider from a hostname (or URL host substring), for logo selection. */
export function providerIdForHost(host?: string): ProviderId {
  const h = (host ?? "").toLowerCase();
  if (h.includes("github")) return "github";
  if (h.includes("gitlab")) return "gitlab";
  if (h.includes("bitbucket")) return "bitbucket";
  if (h.includes("gitea")) return "gitea";
  return "other";
}

/** Detects the git host provider from the integration type/host, for logo selection. */
export function providerIdFor(integration: GitIntegration): ProviderId {
  if (integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP) return "github";
  return providerIdForHost(integration.host);
}

/** Row-title display name per provider (wizard tiles use their own copy). */
export const PROVIDER_DISPLAY_NAMES: Record<ProviderId, string> = {
  github: "GitHub",
  gitlab: "GitLab",
  bitbucket: "Bitbucket",
  gitea: "Gitea",
  other: "Git host",
};

export type RowTone = "ok" | "attention";

/** One-line summary of what the integration can reach, shown in the row. */
export interface RowAccess {
  label: string;
  /** Optional mono hint rendered right-aligned (e.g. repository scope). */
  hint?: string;
}

export interface RowViewModel {
  host: string;
  authLabel: string;
  statusKey: "connected" | "needs_setup" | "action_needed";
  statusLabel: string;
  tone: RowTone;
  banner?: { message: string; ctaLabel: string; ctaHref?: string };
  access: RowAccess;
}

// Invariant: status must stay derivable from the integration alone — installations are optional row context.
function statusFor(integration: GitIntegration): { key: RowViewModel["statusKey"]; label: string; tone: RowTone } {
  if (integration.credentials_configured === false) {
    return { key: "action_needed", label: "Needs attention", tone: "attention" };
  }
  if (integration.status === STATUS_PENDING_INSTALL) {
    return { key: "needs_setup", label: "Needs setup", tone: "attention" };
  }
  return { key: "connected", label: "Connected", tone: "ok" };
}

function bannerFor(integration: GitIntegration, statusKey: RowViewModel["statusKey"]): RowViewModel["banner"] {
  if (statusKey === "needs_setup") {
    return {
      message: "The app is created but not installed on any account yet, so Stackdome can't see your repositories.",
      ctaLabel: "Finish install →",
      ctaHref: integration.install_url,
    };
  }
  if (statusKey === "action_needed") {
    return {
      message: "No credentials are stored for this integration, so clones will fail.",
      ctaLabel: "Update credentials →",
    };
  }
  return undefined;
}

function accessFor(
  integration: GitIntegration,
  statusKey: RowViewModel["statusKey"],
  installations?: GitInstallation[],
): RowAccess {
  if (statusKey === "action_needed") {
    return { label: "Access blocked" };
  }

  if (integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP) {
    if (statusKey === "needs_setup") {
      return { label: "No repositories yet", hint: "finish install" };
    }
    const count = installations?.length ?? 0;
    const installationWord = count === 1 ? "installation" : "installations";
    const hasAll = installations?.some((installation) => installation.repository_selection === REPOSITORY_SELECTION_ALL) ?? false;
    const scope = hasAll ? "all repositories" : "selected repositories";
    return { label: `${count} ${installationWord}`, hint: scope };
  }

  return { label: `Token-scoped access to ${integration.host}` };
}

export function deriveRow(integration: GitIntegration, installations?: GitInstallation[]): RowViewModel {
  const { key: statusKey, label: statusLabel, tone } = statusFor(integration);

  return {
    host: integration.host,
    authLabel: integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP ? "GitHub App" : "Access token",
    statusKey,
    statusLabel,
    tone,
    banner: bannerFor(integration, statusKey),
    access: accessFor(integration, statusKey, installations),
  };
}

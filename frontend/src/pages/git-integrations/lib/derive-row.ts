import type { GitIntegration, GitInstallation } from "@/api/git-integrations";

export const GIT_INTEGRATION_TYPE_GITHUB_APP = "github_app" as const;
export const GIT_INTEGRATION_TYPE_CREDENTIALS = "git_credentials" as const;
export const STATUS_PENDING_INSTALL = "pending_install" as const;
export const STATUS_INSTALLED = "installed" as const;
export const STATUS_ACTIVE = "active" as const;

export type ProviderId = "github" | "gitlab" | "bitbucket" | "gitea" | "other";

/** Detects the git host provider from the integration type/host, for logo selection. */
export function providerIdFor(integration: GitIntegration): ProviderId {
  if (integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP) return "github";
  const host = (integration.host ?? "").toLowerCase();
  if (host.includes("github")) return "github";
  if (host.includes("gitlab")) return "gitlab";
  if (host.includes("bitbucket")) return "bitbucket";
  if (host.includes("gitea")) return "gitea";
  return "other";
}

export type RowTone = "ok" | "attention";

export interface RowMeter {
  left: string;
  right: string;
  fill: "full" | "partial" | "none";
}

export interface RowViewModel {
  host: string;
  authLabel: string;
  statusKey: "connected" | "needs_setup" | "action_needed";
  statusLabel: string;
  tone: RowTone;
  banner?: { message: string; ctaLabel: string; ctaHref?: string };
  meter: RowMeter;
}

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

function meterFor(
  integration: GitIntegration,
  statusKey: RowViewModel["statusKey"],
  installations?: GitInstallation[],
): RowMeter {
  if (statusKey === "action_needed") {
    return { left: "Access blocked", right: "this host", fill: "none" };
  }

  if (integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP) {
    if (statusKey === "needs_setup") {
      return { left: "No repositories yet", right: "finish install", fill: "none" };
    }
    const count = installations?.length ?? 0;
    const installationWord = count === 1 ? "installation" : "installations";
    const hasAll = installations?.some((installation) => installation.repository_selection === "all") ?? false;
    const scope = hasAll ? "all repositories" : "selected repositories";
    return { left: `${count} ${installationWord}`, right: scope, fill: "full" };
  }

  return { left: "Token-scoped access", right: "this host", fill: "partial" };
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
    meter: meterFor(integration, statusKey, installations),
  };
}

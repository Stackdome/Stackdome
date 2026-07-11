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

export interface RowViewModel {
  host: string;
  authLabel: string;
  statusKey: "connected" | "needs_setup" | "action_needed";
  statusLabel: string;
  tone: RowTone;
  banner?: { message: string; ctaLabel: string; ctaHref?: string };
  accessLine?: string;
}

function statusFor(integration: GitIntegration): { key: RowViewModel["statusKey"]; label: string; tone: RowTone } {
  if (integration.credentials_configured === false) {
    return { key: "action_needed", label: "Action needed", tone: "attention" };
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
      message: "Credentials for this integration need attention before Stackdome can clone repositories.",
      ctaLabel: "Update credentials →",
    };
  }
  return undefined;
}

function accessLineFor(integration: GitIntegration, installations?: GitInstallation[]): string | undefined {
  if (integration.type !== GIT_INTEGRATION_TYPE_GITHUB_APP) return undefined;
  if (!installations || installations.length === 0) return undefined;

  const count = installations.length;
  const installationWord = count === 1 ? "installation" : "installations";
  const hasAll = installations.some((installation) => installation.repository_selection === "all");
  const scope = hasAll ? "all repositories" : "selected repositories";

  return `${count} ${installationWord} · ${scope}`;
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
    accessLine: accessLineFor(integration, installations),
  };
}

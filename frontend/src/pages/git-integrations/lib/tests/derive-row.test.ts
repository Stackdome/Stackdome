import { describe, it, expect } from "vitest";
import type { GitIntegration, GitInstallation } from "@/api/git-integrations";
import {
  deriveRow,
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_PENDING_INSTALL,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
} from "../derive-row";

function integration(overrides: Partial<GitIntegration> = {}): GitIntegration {
  return {
    id: "int-1",
    host: "github.com",
    type: GIT_INTEGRATION_TYPE_GITHUB_APP,
    status: STATUS_INSTALLED,
    credentials_configured: true,
    ...overrides,
  };
}

function installation(overrides: Partial<GitInstallation> = {}): GitInstallation {
  return {
    id: "inst-1",
    repository_selection: "all",
    ...overrides,
  };
}

describe("deriveRow", () => {
  it("marks an installed github_app integration as connected", () => {
    const row = deriveRow(integration({ status: STATUS_INSTALLED }));
    expect(row.statusKey).toBe("connected");
    expect(row.statusLabel).toBe("Connected");
    expect(row.tone).toBe("ok");
    expect(row.banner).toBeUndefined();
  });

  it("marks an active integration as connected", () => {
    const row = deriveRow(integration({ status: STATUS_ACTIVE }));
    expect(row.statusKey).toBe("connected");
    expect(row.statusLabel).toBe("Connected");
    expect(row.tone).toBe("ok");
  });

  it("marks a pending_install integration as needs_setup with install_url CTA", () => {
    const row = deriveRow(
      integration({ status: STATUS_PENDING_INSTALL, install_url: "https://github.com/apps/x/installations/new" }),
    );
    expect(row.statusKey).toBe("needs_setup");
    expect(row.statusLabel).toBe("Needs setup");
    expect(row.tone).toBe("attention");
    expect(row.banner).toEqual({
      message:
        "The app is created but not installed on any account yet, so Stackdome can't see your repositories.",
      ctaLabel: "Finish install →",
      ctaHref: "https://github.com/apps/x/installations/new",
    });
  });

  it("action_needed takes precedence over pending_install when credentials are missing", () => {
    const row = deriveRow(
      integration({
        status: STATUS_PENDING_INSTALL,
        credentials_configured: false,
        install_url: "https://github.com/apps/x/installations/new",
      }),
    );
    expect(row.statusKey).toBe("action_needed");
    expect(row.statusLabel).toBe("Action needed");
    expect(row.tone).toBe("attention");
  });

  it("marks credentials-missing integration as action_needed even when active", () => {
    const row = deriveRow(integration({ status: STATUS_ACTIVE, credentials_configured: false }));
    expect(row.statusKey).toBe("action_needed");
    expect(row.statusLabel).toBe("Action needed");
    expect(row.tone).toBe("attention");
  });

  it("labels github_app auth as GitHub App", () => {
    const row = deriveRow(integration({ type: GIT_INTEGRATION_TYPE_GITHUB_APP }));
    expect(row.authLabel).toBe("GitHub App");
  });

  it("labels git_credentials auth as Access token", () => {
    const row = deriveRow(integration({ type: GIT_INTEGRATION_TYPE_CREDENTIALS, credentials_configured: true }));
    expect(row.authLabel).toBe("Access token");
  });

  it("uses the integration host verbatim", () => {
    const row = deriveRow(integration({ host: "github.acme-corp.com" }));
    expect(row.host).toBe("github.acme-corp.com");
  });

  it("builds a singular access line for one all-repositories installation", () => {
    const row = deriveRow(integration(), [installation({ repository_selection: "all" })]);
    expect(row.accessLine).toBe("1 installation · all repositories");
  });

  it("builds a plural access line for multiple installations with selected repositories", () => {
    const row = deriveRow(integration(), [
      installation({ repository_selection: "selected" }),
      installation({ repository_selection: "selected" }),
    ]);
    expect(row.accessLine).toBe("2 installations · selected repositories");
  });

  it("prefers all repositories when installations are mixed", () => {
    const row = deriveRow(integration(), [
      installation({ repository_selection: "selected" }),
      installation({ repository_selection: "all" }),
    ]);
    expect(row.accessLine).toBe("2 installations · all repositories");
  });

  it("omits the access line for credentials-type integrations", () => {
    const row = deriveRow(
      integration({ type: GIT_INTEGRATION_TYPE_CREDENTIALS, credentials_configured: true }),
      [installation()],
    );
    expect(row.accessLine).toBeUndefined();
  });

  it("omits the access line when there are zero installations", () => {
    const row = deriveRow(integration(), []);
    expect(row.accessLine).toBeUndefined();
  });

  it("omits the access line when installations are undefined", () => {
    const row = deriveRow(integration());
    expect(row.accessLine).toBeUndefined();
  });
});

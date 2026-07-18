// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { IntegrationRow } from "../integration-row";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
  STATUS_PENDING_INSTALL,
  REPOSITORY_SELECTION_ALL,
} from "../../lib/derive-row";
import type { GitIntegration } from "@/api/git-integrations";

afterEach(cleanup);

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listInstallations: vi.fn().mockResolvedValue({ items: [] }),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

import { listInstallations } from "@/api/git-integrations";

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

function renderRow(props: Partial<Parameters<typeof IntegrationRow>[0]> = {}) {
  return render(
    <IntegrationRow
      integration={integration()}
      onVerify={vi.fn()}
      onRemove={vi.fn()}
      {...props}
    />,
  );
}

describe("IntegrationRow", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listInstallations).mockResolvedValue({ items: [] });
  });

  it("renders a quiet row for an installed, credentials-configured integration", async () => {
    renderRow();
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    expect(screen.getByText("Connected")).toBeInTheDocument();
    expect(screen.queryByText(/finish install/i)).not.toBeInTheDocument();
  });

  it("renders a loud row with a banner CTA anchored to install_url for pending_install", async () => {
    renderRow({
      integration: integration({
        status: STATUS_PENDING_INSTALL,
        install_url: "https://github.com/apps/x/installations/new",
      }),
    });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    expect(screen.getByText(/app is created but not installed/i)).toBeInTheDocument();
    const cta = screen.getByRole("link", { name: /finish install/i });
    expect(cta).toHaveAttribute("href", "https://github.com/apps/x/installations/new");
    expect(cta).toHaveAttribute("target", "_blank");
  });

  it("renders the banner CTA as disabled when install_url is missing", async () => {
    renderRow({ integration: integration({ status: STATUS_PENDING_INSTALL, install_url: undefined }) });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    expect(screen.queryByRole("link", { name: /finish install/i })).not.toBeInTheDocument();
    const cta = screen.getByRole("button", { name: /finish install/i });
    expect(cta).toBeDisabled();
  });

  it("routes the action_needed banner CTA to onUpdateCredentials, not onVerify", async () => {
    const user = userEvent.setup();
    const onVerify = vi.fn();
    const onUpdateCredentials = vi.fn();
    renderRow({
      onVerify,
      onUpdateCredentials,
      integration: integration({
        type: GIT_INTEGRATION_TYPE_CREDENTIALS,
        status: STATUS_ACTIVE,
        host: "gitlab.com",
        credentials_configured: false,
      }),
    });
    const cta = screen.getByRole("button", { name: /update credentials/i });
    await user.click(cta, { pointerEventsCheck: 0 });
    expect(onUpdateCredentials).toHaveBeenCalledOnce();
    expect(onVerify).not.toHaveBeenCalled();
  });

  it("renders the access summary derived from installations", async () => {
    vi.mocked(listInstallations).mockResolvedValue({
      items: [{ id: "i1", repository_selection: REPOSITORY_SELECTION_ALL }],
    });
    renderRow();
    await waitFor(() => expect(screen.getByText("1 installation")).toBeInTheDocument());
    expect(screen.getByText("all repositories")).toBeInTheDocument();
  });

  it("opens the kebab menu and dispatches Verify repository access on a credentials row", async () => {
    const user = userEvent.setup();
    const onVerify = vi.fn();
    const row = integration({ type: GIT_INTEGRATION_TYPE_CREDENTIALS, status: STATUS_ACTIVE, host: "gitlab.com" });
    renderRow({ onVerify, integration: row });
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/verify repository access/i), { pointerEventsCheck: 0 });
    await waitFor(() => expect(onVerify).toHaveBeenCalledWith(row));
  });

  it("hides Verify on GitHub App rows (backend only verifies credentials-type directly)", async () => {
    const user = userEvent.setup();
    renderRow();
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    expect(await screen.findByText(/remove integration/i)).toBeInTheDocument();
    expect(screen.queryByText(/verify repository access/i)).not.toBeInTheDocument();
  });

  it("dispatching Verify or Remove defers the callback until after the menu has closed, avoiding the Radix pointer-events lock", async () => {
    // Regression test for a Radix DropdownMenu -> Dialog composition bug: if the
    // dialog-opening callback fires synchronously from onSelect, the menu's
    // close and the dialog's mount race and can leave
    // document.body.style.pointerEvents stuck at "none" forever. Deferring the
    // callback via setTimeout(0) lets the menu finish closing (and reset
    // pointer-events) before the dialog mounts.
    const user = userEvent.setup();
    const onRemove = vi.fn();
    const row = integration();
    renderRow({ onRemove });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    const removeItem = await screen.findByText(/remove integration/i);
    await user.click(removeItem, { pointerEventsCheck: 0 });

    await waitFor(() => expect(onRemove).toHaveBeenCalledWith(row));
    // Once the deferred callback has run, the dropdown menu's own close
    // cleanup must have already reset pointer-events — it is not left
    // stuck at "none" by the callback firing before the menu unmounts.
    expect(document.body.style.pointerEvents).not.toBe("none");
  });

  it("loads installations with refresh=true so missed-webhook state self-heals on every visit", async () => {
    renderRow();
    await waitFor(() => expect(listInstallations).toHaveBeenCalledTimes(1));
    expect(listInstallations).toHaveBeenCalledWith("org-1", "int-1", true);
  });

  it("never fetches installations for a credentials row (no installations exist to refresh)", async () => {
    renderRow({
      integration: integration({ type: GIT_INTEGRATION_TYPE_CREDENTIALS, status: STATUS_ACTIVE, host: "gitlab.com" }),
    });
    expect(screen.getByText("gitlab.com")).toBeInTheDocument();
    await waitFor(() => expect(listInstallations).not.toHaveBeenCalled());
  });

  it("offers Update credentials in the menu for credentials rows and passes the integration", async () => {
    const onUpdateCredentials = vi.fn();
    const creds = integration({ type: GIT_INTEGRATION_TYPE_CREDENTIALS, host: "gitlab.com" });
    const user = userEvent.setup();
    renderRow({ integration: creds, onUpdateCredentials });

    await user.click(screen.getByRole("button", { name: /open row menu/i }));
    await user.click(await screen.findByRole("menuitem", { name: /update credentials/i }));

    await waitFor(() => expect(onUpdateCredentials).toHaveBeenCalledWith(creds));
    expect(screen.queryByRole("menuitem", { name: /manage on github/i })).not.toBeInTheDocument();
  });

  it("offers Manage on GitHub for app rows with an install_url, not Update credentials", async () => {
    const user = userEvent.setup();
    renderRow({
      integration: integration({ install_url: "https://github.com/apps/x/installations/new" }),
      onUpdateCredentials: vi.fn(),
    });

    await user.click(screen.getByRole("button", { name: /open row menu/i }));

    const manage = await screen.findByRole("menuitem", { name: /manage on github/i });
    expect(manage).toHaveAttribute("href", "https://github.com/apps/x/installations/new");
    expect(manage).toHaveAttribute("target", "_blank");
    expect(screen.queryByRole("menuitem", { name: /update credentials/i })).not.toBeInTheDocument();
  });

  it("hides Manage on GitHub when an app row has no install_url", async () => {
    const user = userEvent.setup();
    renderRow({ integration: integration({ install_url: undefined }) });

    await user.click(screen.getByRole("button", { name: /open row menu/i }));
    await screen.findByRole("menuitem", { name: /remove integration/i });
    expect(screen.queryByRole("menuitem", { name: /manage on github/i })).not.toBeInTheDocument();
  });

  it("routes the action_needed banner CTA through onUpdateCredentials with the integration", async () => {
    const onUpdateCredentials = vi.fn();
    const creds = integration({
      type: GIT_INTEGRATION_TYPE_CREDENTIALS,
      credentials_configured: false,
    });
    const user = userEvent.setup();
    renderRow({ integration: creds, onUpdateCredentials });

    await user.click(screen.getByRole("button", { name: /update credentials/i }));
    expect(onUpdateCredentials).toHaveBeenCalledWith(creds);
  });
});

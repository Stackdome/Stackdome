// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { IntegrationRow } from "../integration-row";
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
    type: "github_app",
    status: "installed",
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
      onChanged={vi.fn()}
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
        status: "pending_install",
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
    renderRow({ integration: integration({ status: "pending_install", install_url: undefined }) });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    expect(screen.queryByRole("link", { name: /finish install/i })).not.toBeInTheDocument();
    const cta = screen.getByRole("button", { name: /finish install/i });
    expect(cta).toBeDisabled();
  });

  it("renders the access meter derived from installations", async () => {
    vi.mocked(listInstallations).mockResolvedValue({
      items: [{ id: "i1", repository_selection: "all" }],
    });
    renderRow();
    await waitFor(() => expect(screen.getByText("1 installation")).toBeInTheDocument());
    expect(screen.getByText("all repositories")).toBeInTheDocument();
  });

  it("opens the kebab menu and dispatches Verify repository access on a credentials row", async () => {
    const user = userEvent.setup();
    const onVerify = vi.fn();
    const row = integration({ type: "git_credentials", status: "active", host: "gitlab.com" });
    renderRow({ onVerify, integration: row });
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/verify repository access/i), { pointerEventsCheck: 0 });
    await waitFor(() => expect(onVerify).toHaveBeenCalledWith(row));
    // Sync only applies to GitHub App integrations.
    expect(screen.queryByText(/sync from github/i)).not.toBeInTheDocument();
  });

  it("hides Verify on GitHub App rows (backend only verifies credentials-type directly) but shows Sync", async () => {
    const user = userEvent.setup();
    renderRow();
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    expect(await screen.findByText(/sync from github/i)).toBeInTheDocument();
    expect(screen.queryByText(/verify repository access/i)).not.toBeInTheDocument();
  });

  it("dispatches Remove integration from the kebab menu", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    const row = integration();
    renderRow({ onRemove });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/remove integration/i), { pointerEventsCheck: 0 });
    await waitFor(() => expect(onRemove).toHaveBeenCalledWith(row));
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
    renderRow({ onRemove });
    await waitFor(() => expect(listInstallations).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    const removeItem = await screen.findByText(/remove integration/i);
    await user.click(removeItem, { pointerEventsCheck: 0 });

    await waitFor(() => expect(onRemove).toHaveBeenCalled());
    // Once the deferred callback has run, the dropdown menu's own close
    // cleanup must have already reset pointer-events — it is not left
    // stuck at "none" by the callback firing before the menu unmounts.
    expect(document.body.style.pointerEvents).not.toBe("none");
  });

  it("Sync from GitHub refetches installations with refresh=true then calls onChanged", async () => {
    const user = userEvent.setup();
    const onChanged = vi.fn();
    renderRow({ onChanged });
    await waitFor(() => expect(listInstallations).toHaveBeenCalledTimes(1));
    expect(listInstallations).toHaveBeenNthCalledWith(1, "org-1", "int-1", false);

    await user.click(screen.getByRole("button", { name: /open row menu/i }), { pointerEventsCheck: 0 });
    expect(await screen.findByText(/re-check access now/i)).toBeInTheDocument();
    await user.click(screen.getByText(/sync from github/i), { pointerEventsCheck: 0 });

    await waitFor(() => expect(listInstallations).toHaveBeenNthCalledWith(2, "org-1", "int-1", true));
    await waitFor(() => expect(onChanged).toHaveBeenCalled());
  });
});

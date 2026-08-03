// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CredentialsDropdown, usableIntegrations } from "../credentials-dropdown";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
  STATUS_PENDING_INSTALL,
} from "@/lib/git-integrations";
import type { GitIntegration } from "@/api/git-integrations";

afterEach(cleanup);

const app: GitIntegration = {
  id: "int-app",
  host: "github.com",
  type: GIT_INTEGRATION_TYPE_GITHUB_APP,
  status: STATUS_INSTALLED,
  credentials_configured: true,
};
const creds: GitIntegration = {
  id: "int-creds",
  host: "gitlab.example.com",
  type: GIT_INTEGRATION_TYPE_CREDENTIALS,
  status: STATUS_ACTIVE,
  credentials_configured: true,
};
const pending: GitIntegration = {
  id: "int-pending",
  host: "github.com",
  type: GIT_INTEGRATION_TYPE_GITHUB_APP,
  status: STATUS_PENDING_INSTALL,
  credentials_configured: true,
};

describe("usableIntegrations", () => {
  it("keeps installed apps and active credentials, drops pending installs and missing credentials", () => {
    const noCreds: GitIntegration = { ...creds, id: "int-nocreds", credentials_configured: false };
    const result = usableIntegrations([app, creds, pending, noCreds]);
    expect(result.map((i) => i.id)).toEqual(["int-app", "int-creds"]);
  });
});

describe("CredentialsDropdown", () => {
  it("lists usable integrations by display name + host and dispatches onSelect", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <CredentialsDropdown
        integrations={[app, creds]}
        selectedId="int-app"
        onSelect={onSelect}
        onConnectNew={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /credentials/i }), { pointerEventsCheck: 0 });
    expect(await screen.findByText("gitlab.example.com")).toBeInTheDocument();
    await user.click(screen.getByText("gitlab.example.com"), { pointerEventsCheck: 0 });
    expect(onSelect).toHaveBeenCalledWith(creds);
  });

  it("offers Connect provider…, deferring the callback past menu close (Radix #1836)", async () => {
    const user = userEvent.setup();
    const onConnectNew = vi.fn();
    render(
      <CredentialsDropdown integrations={[app]} selectedId={null} onSelect={vi.fn()} onConnectNew={onConnectNew} />,
    );
    await user.click(screen.getByRole("button", { name: /credentials/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/connect provider/i), { pointerEventsCheck: 0 });
    await vi.waitFor(() => expect(onConnectNew).toHaveBeenCalledOnce());
    expect(document.body.style.pointerEvents).not.toBe("none");
  });
});

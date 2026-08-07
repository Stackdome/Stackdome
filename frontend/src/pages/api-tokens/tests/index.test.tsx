// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import ApiTokensPage from "../index";
import * as tokensApi from "@/api/api-tokens";
import { ConfirmProvider } from "@/components/branded/confirm";

const toastMock = vi.fn();

afterEach(cleanup);

vi.mock("@/api/api-tokens");
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

const token = {
  id: "t1",
  name: "agent",
  token_prefix: "sd_abc1",
  scopes: ["*"],
  created_at: "2026-08-01T00:00:00Z",
};

function renderPage() {
  return render(
    <MemoryRouter>
      <ConfirmProvider>
        <ApiTokensPage />
      </ConfirmProvider>
    </MemoryRouter>,
  );
}

describe("ApiTokensPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(tokensApi.listApiTokens).mockResolvedValue({ items: [token] });
    vi.mocked(tokensApi.getApiTokenScopes).mockResolvedValue({
      full_access_scope: "*",
      items: [{ resource: "stacks", actions: ["read", "write"] }],
    });
  });

  it("lists tokens with their prefix", async () => {
    renderPage();
    expect(await screen.findByText("agent")).toBeInTheDocument();
    expect(screen.getByText(/sd_abc1/)).toBeInTheDocument();
  });

  it("shows the raw token exactly once after create, and never again once dismissed", async () => {
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));
    expect(await screen.findByText(/sd_raw_secret/)).toBeInTheDocument();
    expect(screen.getByText(/won't see this again/i)).toBeInTheDocument();

    // Dismissing the show-once view must drop the secret from state entirely.
    await userEvent.click(screen.getByRole("button", { name: /^done$/i }));
    expect(screen.queryByText(/sd_raw_secret/)).not.toBeInTheDocument();

    // Reopening the create dialog must not resurface the previous secret.
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    expect(await screen.findByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.queryByText(/sd_raw_secret/)).not.toBeInTheDocument();
  });

  it("copies the token via the Clipboard API when available", async () => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));
    await screen.findByText(/sd_raw_secret/);

    await userEvent.click(screen.getByRole("button", { name: "Copy token" }));
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith("sd_raw_secret");
    expect(await screen.findByRole("button", { name: "Copied" })).toBeInTheDocument();
    expect(toastMock).not.toHaveBeenCalled();
  });

  it("falls back to a textarea copy when the Clipboard API is unavailable (insecure context)", async () => {
    Object.assign(navigator, { clipboard: undefined });
    document.execCommand = vi.fn().mockReturnValue(true);
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));
    await screen.findByText(/sd_raw_secret/);

    // No throw / unhandled rejection from the missing Clipboard API.
    await userEvent.click(screen.getByRole("button", { name: "Copy token" }));
    expect(document.execCommand).toHaveBeenCalledWith("copy");
    expect(await screen.findByRole("button", { name: "Copied" })).toBeInTheDocument();
    expect(toastMock).not.toHaveBeenCalled();
  });

  it("surfaces a destructive toast when copying fails entirely", async () => {
    Object.assign(navigator, { clipboard: undefined });
    document.execCommand = vi.fn(() => {
      throw new Error("copy blocked");
    });
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));
    await screen.findByText(/sd_raw_secret/);

    await userEvent.click(screen.getByRole("button", { name: "Copy token" }));
    await waitFor(() =>
      expect(toastMock).toHaveBeenCalledWith(
        expect.objectContaining({ title: "Copy failed", variant: "destructive" }),
      ),
    );
    expect(screen.queryByRole("button", { name: "Copied" })).not.toBeInTheDocument();
  });

  it("disables Create and explains why while scopes haven't loaded", async () => {
    let resolveScopes: (value: tokensApi.ScopeList) => void = () => {};
    vi.mocked(tokensApi.getApiTokenScopes).mockReturnValue(
      new Promise((resolve) => {
        resolveScopes = resolve;
      }),
    );
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");

    expect(screen.getByText(/loading available scopes/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^create$/i })).toBeDisabled();

    resolveScopes({ full_access_scope: "*", items: [] });
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
  });

  it("sends an expires_at that is end-of-day local time for the chosen date", async () => {
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    fireEvent.change(screen.getByLabelText(/expires/i), { target: { value: "2026-12-31" } });
    await waitFor(() => expect(screen.getByRole("button", { name: /^create$/i })).toBeEnabled());
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));

    await waitFor(() => expect(tokensApi.createApiToken).toHaveBeenCalled());
    const { expires_at } = vi.mocked(tokensApi.createApiToken).mock.calls[0][0];
    expect(new Date(expires_at as string)).toEqual(new Date(2026, 11, 31, 23, 59, 59));
  });

  it("revokes a token via the confirm dialog", async () => {
    vi.mocked(tokensApi.revokeApiToken).mockResolvedValue(undefined);
    renderPage();
    await screen.findByText("agent");

    await userEvent.click(screen.getByRole("button", { name: /revoke/i }));
    const dialog = await screen.findByRole("alertdialog");
    expect(dialog).toHaveTextContent(/revoke token\?/i);
    await userEvent.click(screen.getByRole("button", { name: "Revoke" }));

    await waitFor(() => expect(tokensApi.revokeApiToken).toHaveBeenCalledWith("t1"));
  });
});

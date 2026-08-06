// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import ApiTokensPage from "../index";
import * as tokensApi from "@/api/api-tokens";
import { ConfirmProvider } from "@/components/branded/confirm";

afterEach(cleanup);

vi.mock("@/api/api-tokens");

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

  it("shows the raw token exactly once after create", async () => {
    vi.mocked(tokensApi.createApiToken).mockResolvedValue({ id: "t2", name: "ci", token: "sd_raw_secret" });
    renderPage();
    await userEvent.click(await screen.findByRole("button", { name: /create token/i }));
    await userEvent.type(screen.getByLabelText(/name/i), "ci");
    await userEvent.click(screen.getByRole("button", { name: /^create$/i }));
    expect(await screen.findByText(/sd_raw_secret/)).toBeInTheDocument();
    expect(screen.getByText(/won't see this again/i)).toBeInTheDocument();
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

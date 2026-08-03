// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { GitSourcePicker, repoTail } from "../git-source-picker";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
} from "@/lib/git-integrations";

afterEach(cleanup);

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listGitIntegrations: vi.fn(),
  listRepositories: vi.fn(),
  getRepository: vi.fn(),
}));
vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
// The nested connect wizard drags in the whole github-connect flow; the picker
// only needs its open/close contract.
vi.mock("@/components/git-source-picker/add-integration-wizard", () => ({
  AddIntegrationWizard: ({ open }: { open: boolean }) =>
    open ? <div data-testid="add-integration-wizard" /> : null,
}));

import { listGitIntegrations, listRepositories, getRepository } from "@/api/git-integrations";

const app = {
  id: "int-app",
  host: "github.com",
  type: GIT_INTEGRATION_TYPE_GITHUB_APP,
  status: STATUS_INSTALLED,
  credentials_configured: true,
};
const creds = {
  id: "int-creds",
  host: "gitlab.example.com",
  type: GIT_INTEGRATION_TYPE_CREDENTIALS,
  status: STATUS_ACTIVE,
  credentials_configured: true,
};

function renderPicker(props: Partial<Parameters<typeof GitSourcePicker>[0]> = {}) {
  return render(<GitSourcePicker value={null} onChange={vi.fn()} {...props} />);
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listGitIntegrations).mockResolvedValue({ items: [app, creds] });
  vi.mocked(listRepositories).mockResolvedValue({ items: [] });
});

describe("repoTail", () => {
  it("extracts owner/name from clone URLs", () => {
    expect(repoTail("https://github.com/acme/api.git")).toBe("acme/api");
    expect(repoTail("https://gitlab.example.com/group/app/")).toBe("group/app");
  });
});

describe("GitSourcePicker", () => {
  it("searches repos for the auto-selected GitHub App and emits the picked repo with fetched detail", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    vi.mocked(listRepositories).mockResolvedValue({
      items: [{ full_name: "acme/webapp", clone_url: "https://github.com/acme/webapp.git", private: false }],
    });
    vi.mocked(getRepository).mockResolvedValue({
      full_name: "acme/webapp",
      clone_url: "https://github.com/acme/webapp.git",
      default_branch: "main",
    });
    renderPicker({ onChange });
    await user.click(await screen.findByText("acme/webapp"), { pointerEventsCheck: 0 });
    await waitFor(() =>
      expect(onChange).toHaveBeenCalledWith({
        fullName: "acme/webapp",
        cloneUrl: "https://github.com/acme/webapp.git",
        defaultBranch: "main",
        integrationId: "int-app",
      }),
    );
  });

  it("switches to host-scoped URL entry for a credentials integration and validates the host", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderPicker({ onChange });
    await user.click(await screen.findByRole("button", { name: /credentials/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText("gitlab.example.com"), { pointerEventsCheck: 0 });

    const input = await screen.findByPlaceholderText(/gitlab\.example\.com/);
    await user.clear(input);
    await user.type(input, "https://gitlab.example.com/group/app");
    await waitFor(() =>
      expect(onChange).toHaveBeenLastCalledWith({
        fullName: "group/app",
        cloneUrl: "https://gitlab.example.com/group/app",
        defaultBranch: "",
        integrationId: "int-creds",
      }),
    );

    await user.clear(input);
    await user.type(input, "https://github.com/elsewhere/app");
    // "gitlab.example.com" also appears in the dropdown trigger and the helper
    // copy above the input, so match the exact error sentence rather than a
    // host-only regex to avoid an ambiguous "multiple elements" match.
    expect(
      await screen.findByText("URL must be on gitlab.example.com to use this connection."),
    ).toBeInTheDocument();
    await waitFor(() => expect(onChange).toHaveBeenLastCalledWith(null));
  });

  it("emits a public repo (integrationId null) from the Public URL tab and shows the hint", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderPicker({ onChange, publicUrlHint: "PR automation requires a connected provider." });
    await user.click(await screen.findByRole("tab", { name: /public url/i }), { pointerEventsCheck: 0 });
    expect(screen.getByText(/PR automation requires a connected provider/)).toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(/https:\/\//), "https://github.com/acme/site");
    await waitFor(() =>
      expect(onChange).toHaveBeenLastCalledWith({
        fullName: "acme/site",
        cloneUrl: "https://github.com/acme/site",
        defaultBranch: "",
        integrationId: null,
      }),
    );
  });

  it("clears the Public URL field and re-emits null when switching away and back", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderPicker({ onChange });
    await user.click(await screen.findByRole("tab", { name: /public url/i }), { pointerEventsCheck: 0 });
    const urlInput = screen.getByPlaceholderText(/https:\/\//);
    await user.type(urlInput, "https://github.com/acme/site");
    await waitFor(() =>
      expect(onChange).toHaveBeenLastCalledWith({
        fullName: "acme/site",
        cloneUrl: "https://github.com/acme/site",
        defaultBranch: "",
        integrationId: null,
      }),
    );

    await user.click(await screen.findByRole("tab", { name: /connected provider/i }), { pointerEventsCheck: 0 });
    expect(onChange).toHaveBeenLastCalledWith(null);

    await user.click(await screen.findByRole("tab", { name: /public url/i }), { pointerEventsCheck: 0 });
    expect(screen.getByPlaceholderText(/https:\/\//)).toHaveValue("");
    expect(onChange).toHaveBeenLastCalledWith(null);
  });

  it("shows the configure escape hatch from the selected integration's install URL when the search returns nothing", async () => {
    vi.mocked(listGitIntegrations).mockResolvedValue({
      items: [{ ...app, install_url: "https://github.com/apps/x/installations/new" }, creds],
    });
    renderPicker();
    const link = await screen.findByRole("link", { name: /configure in github/i });
    expect(link).toHaveAttribute("href", "https://github.com/apps/x/installations/new");
  });

  it("omits the configure escape hatch when the selected integration has no install URL", async () => {
    renderPicker();
    await screen.findByRole("tab", { name: /connected provider/i });
    expect(await screen.findByText(/no repositories found/i)).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /configure in github/i })).not.toBeInTheDocument();
  });

  it("offers Connect provider when no integrations exist", async () => {
    const user = userEvent.setup();
    vi.mocked(listGitIntegrations).mockResolvedValue({ items: [] });
    renderPicker();
    const btn = await screen.findByRole("button", { name: /connect provider/i });
    await user.click(btn, { pointerEventsCheck: 0 });
    expect(await screen.findByTestId("add-integration-wizard")).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RepoCombobox } from "../repo-combobox";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
} from "@/pages/git-integrations/lib/derive-row";

// cmdk (used by the Command primitive) reads ResizeObserver on mount, which
// jsdom doesn't implement.
beforeAll(() => {
  global.ResizeObserver =
    global.ResizeObserver ||
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  // Radix Select uses pointer-capture APIs that jsdom doesn't implement.
  // Stub them so userEvent.click on the Integration SelectTrigger opens it.
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false,
    releasePointerCapture: () => undefined,
    setPointerCapture: () => undefined,
    scrollIntoView: () => undefined,
  };
  for (const [name, impl] of Object.entries(stubs)) {
    const proto = Element.prototype as unknown as Record<string, unknown>;
    if (!proto[name]) proto[name] = vi.fn(impl);
  }
});

afterEach(cleanup);

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listGitIntegrations: vi.fn(),
  listRepositories: vi.fn(),
  getRepository: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

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
  host: "git.internal.example.com",
  type: GIT_INTEGRATION_TYPE_CREDENTIALS,
  status: STATUS_ACTIVE,
  credentials_configured: true,
};

beforeEach(() => {
  vi.mocked(listGitIntegrations).mockResolvedValue({ items: [app, creds] });
  vi.mocked(listRepositories).mockResolvedValue({
    items: [{ full_name: "acme/api", private: true, default_branch: "main" }],
  });
  vi.mocked(getRepository).mockResolvedValue({
    full_name: "acme/api",
    clone_url: "https://github.com/acme/api.git",
    default_branch: "main",
  });
});

describe("RepoCombobox", () => {
  it("shows the repo tail when the value came from an integration", () => {
    render(
      <RepoCombobox
        id="repo"
        value="https://github.com/acme/api.git"
        integrationId="int-app"
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("combobox")).toHaveTextContent("acme/api");
  });

  it("shows the raw URL when no integration is attached", () => {
    render(<RepoCombobox id="repo" value="https://example.com/x.git" onChange={vi.fn()} />);
    expect(screen.getByRole("combobox")).toHaveTextContent("https://example.com/x.git");
  });

  it("lists searched repositories and picks one with its integration id", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await waitFor(() => expect(listGitIntegrations).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText("acme/api")).toBeInTheDocument());
    await user.click(screen.getByText("acme/api"));
    await waitFor(() =>
      expect(onChange).toHaveBeenCalledWith({
        repo_url: "https://github.com/acme/api.git",
        integration_id: "int-app",
      }),
    );
  });

  it("offers typed text as a repository URL and clears the integration", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "https://example.com/repo.git");
    await user.click(await screen.findByText(/use "https:\/\/example\.com\/repo\.git"/i));
    expect(onChange).toHaveBeenCalledWith({
      repo_url: "https://example.com/repo.git",
      integration_id: undefined,
    });
  });

  it("offers an ssh/scp-style URL as a repository URL", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "git@github.com:acme/api.git");
    await user.click(await screen.findByText(/use "git@github\.com:acme\/api\.git"/i));
    expect(onChange).toHaveBeenCalledWith({
      repo_url: "git@github.com:acme/api.git",
      integration_id: undefined,
    });
  });

  it("falls back to URL entry when integrations fail to load", async () => {
    vi.mocked(listGitIntegrations).mockRejectedValue(new Error("boom"));
    const user = userEvent.setup();
    render(<RepoCombobox id="repo" value="" onChange={vi.fn()} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "https://x.io/y.git");
    expect(await screen.findByText(/use "https:\/\/x\.io\/y\.git"/i)).toBeInTheDocument();
  });

  it("trims surrounding whitespace off a pasted URL before emitting it", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "  https://x.io/y.git  ");
    await user.click(await screen.findByText(/use ".*https:\/\/x\.io\/y\.git.*"/i));
    expect(onChange).toHaveBeenCalledWith({
      repo_url: "https://x.io/y.git",
      integration_id: undefined,
    });
  });

  it("recognizes a non-'git' scp-style user as a URL", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "deploy@git.corp.io:acme/api.git");
    await user.click(await screen.findByText(/use "deploy@git\.corp\.io:acme\/api\.git"/i));
    expect(onChange).toHaveBeenCalledWith({
      repo_url: "deploy@git.corp.io:acme/api.git",
      integration_id: undefined,
    });
  });

  it("collapses a typed trailing .git so useOnHost never doubles it", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RepoCombobox id="repo" value="" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await waitFor(() => expect(listGitIntegrations).toHaveBeenCalled());
    await user.click(await screen.findByRole("combobox", { name: "Integration" }));
    await user.click(await screen.findByRole("option", { name: "git.internal.example.com" }));
    await user.type(screen.getByPlaceholderText(/search repositories/i), "acme/api.git");
    await user.click(await screen.findByText(/use git\.internal\.example\.com\/acme\/api\.git/i));
    expect(onChange).toHaveBeenCalledWith({
      repo_url: "https://git.internal.example.com/acme/api.git",
      integration_id: "int-creds",
    });
  });
});

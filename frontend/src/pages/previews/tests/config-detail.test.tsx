// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";

vi.mock("@/api/preview-configs", () => ({
  getPreviewConfig: vi.fn(),
  updatePreviewConfig: vi.fn(),
  deletePreviewConfig: vi.fn(),
}));
vi.mock("@/pages/previews/hooks/use-preview-envs", () => ({
  usePreviewEnvs: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWriteAnyProject: true }),
}));

import { getPreviewConfig } from "@/api/preview-configs";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import type { PreviewStack } from "@/api/preview-envs";
import PreviewConfigDetailPage from "../config-detail";

const config = {
  id: "c1",
  name: "webapp",
  git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
  stackfile_path: "stackfile.yaml",
  max_active_previews: 10,
};

const mixedEnvs: PreviewStack[] = [
  {
    id: "e1",
    pr_number: "101",
    branch: "feat/ready-one",
    commit: "aaa1111bbbb",
    config_id: "c1",
    status: { phase: "Ready" },
  },
  {
    id: "e2",
    pr_number: "202",
    branch: "feat/pending-two",
    commit: "ccc2222dddd",
    config_id: "c1",
    status: { phase: "Provisioning" },
  },
  {
    id: "e3",
    pr_number: "303",
    branch: "feat/failed-three",
    commit: "eee3333ffff",
    config_id: "c1",
    status: { phase: "Failed", reason: "BuildFailed", message: "image build failed" },
  },
];

function mockEnvs(envs: PreviewStack[], overrides: Partial<ReturnType<typeof usePreviewEnvs>> = {}) {
  vi.mocked(usePreviewEnvs).mockReturnValue({
    envs,
    loading: false,
    error: null,
    refresh: vi.fn(),
    ...overrides,
  });
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/previews/c1"]}>
      <Routes>
        <Route path="/previews/:configId" element={<PreviewConfigDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getPreviewConfig).mockResolvedValue(config);
  mockEnvs([]);
});

afterEach(() => cleanup());

describe("PreviewConfigDetailPage", () => {
  it("loads and shows the configuration", async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("webapp")).toBeInTheDocument();
      expect(screen.getByText("https://github.com/acme/webapp.git")).toBeInTheDocument();
    });
  });

  it("opens the settings modal with seeded fields", async () => {
    renderPage();
    await screen.findByText("webapp");
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /^settings$/i }));

    expect(await screen.findByText(/repository settings/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/base branch/i)).toHaveValue("main");
    expect(screen.getByLabelText(/stackfile path/i)).toHaveValue("stackfile.yaml");
    expect(screen.getByLabelText(/max active previews/i)).toHaveValue(10);
  });

  it("lists every card status word with counts in the status dropdown", async () => {
    mockEnvs(mixedEnvs);
    renderPage();
    await screen.findByText("PR #101");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /status:/i }));

    expect(screen.getByRole("menuitem", { name: /^all/i })).toHaveTextContent(/all\s*3/i);
    expect(screen.getByRole("menuitem", { name: /ready/i })).toHaveTextContent(/ready\s*1/i);
    expect(screen.getByRole("menuitem", { name: /pending/i })).toHaveTextContent(/pending\s*1/i);
    expect(screen.getByRole("menuitem", { name: /failed/i })).toHaveTextContent(/failed\s*1/i);
  });

  it("filters the grid to failed environments via the status dropdown", async () => {
    mockEnvs(mixedEnvs);
    renderPage();
    await screen.findByText("PR #101");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /status:/i }));
    await user.click(screen.getByRole("menuitem", { name: /failed/i }));

    expect(screen.getByText("PR #303")).toBeInTheDocument();
    expect(screen.queryByText("PR #101")).not.toBeInTheDocument();
    expect(screen.queryByText("PR #202")).not.toBeInTheDocument();
  });

  it("narrows the grid by searching for a PR number", async () => {
    mockEnvs(mixedEnvs);
    renderPage();
    await screen.findByText("PR #101");

    const user = userEvent.setup();
    await user.type(screen.getByPlaceholderText(/search pr/i), "202");

    expect(screen.getByText("PR #202")).toBeInTheDocument();
    expect(screen.queryByText("PR #101")).not.toBeInTheDocument();
    expect(screen.queryByText("PR #303")).not.toBeInTheDocument();
  });

  it("shows the empty state with a New preview environment CTA when there are no environments", async () => {
    mockEnvs([]);
    renderPage();
    await screen.findByText("webapp");

    expect(screen.getByText(/no preview environments yet/i)).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /new preview environment/i })).toHaveLength(2);
  });

  it("opens the create modal from the New preview environment button", async () => {
    mockEnvs(mixedEnvs);
    renderPage();
    await screen.findByText("PR #101");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /new preview environment/i }));

    expect(
      await screen.findByText(/deploys the stackfile from a pull request branch of webapp/i),
    ).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import PreviewsPage from "../index";

afterEach(cleanup);

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useNavigate: () => mockNavigate,
}));
vi.mock("@/api/preview-configs", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listAllPreviewConfigs: vi.fn(),
}));
vi.mock("@/hooks/use-preview-envs", () => ({
  usePreviewEnvs: vi.fn(),
}));
vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWriteAnyProject: true }),
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ defaultProjectName: "default" }),
}));
vi.mock("@/pages/previews/components/enable-repo-wizard/enable-repo-wizard", () => ({
  EnableRepoWizard: ({ open }: { open: boolean }) =>
    open ? <div data-testid="enable-repo-wizard" /> : null,
}));

import { listAllPreviewConfigs } from "@/api/preview-configs";
import { usePreviewEnvs } from "@/hooks/use-preview-envs";

const config = {
  id: "cfg-1",
  name: "webapp",
  git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
  stackfile_path: "stackfile.yaml",
  max_active_previews: 10,
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listAllPreviewConfigs).mockResolvedValue([config]);
  vi.mocked(usePreviewEnvs).mockReturnValue({
    envs: [
      { id: "env-1", config_id: "cfg-1", status: { phase: "Ready" } },
      { id: "env-2", config_id: "cfg-1", status: { phase: "Deploying" } },
      { id: "env-3", config_id: "other", status: { phase: "Ready" } },
    ],
    loading: false,
    error: null,
    refresh: vi.fn(),
  });
});

function renderPage() {
  return render(
    <MemoryRouter>
      <PreviewsPage />
    </MemoryRouter>,
  );
}

describe("PreviewsPage", () => {
  it("renders a config row with repo, base branch and its active env count", async () => {
    renderPage();
    expect(await screen.findByText("webapp")).toBeInTheDocument();
    expect(screen.getByText("acme/webapp")).toBeInTheDocument();
    expect(screen.getByText("main")).toBeInTheDocument();
    expect(screen.getByText(/2 environments/)).toBeInTheDocument();
  });

  it("navigates to the config detail on row click", async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(await screen.findByText("webapp"));
    expect(mockNavigate).toHaveBeenCalledWith("/previews/cfg-1");
  });

  it("shows the empty state with an Enable repository CTA when no configs exist", async () => {
    vi.mocked(listAllPreviewConfigs).mockResolvedValue([]);
    const user = userEvent.setup();
    renderPage();
    expect(await screen.findByText(/preview every pull request/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /enable repository/i }));
    expect(await screen.findByTestId("enable-repo-wizard")).toBeInTheDocument();
  });
});

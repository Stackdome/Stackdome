// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

vi.mock("@/api/stacks", () => ({
  getStacksByOrg: vi.fn(),
}));
vi.mock("@/api/preview-configs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-configs")>();
  return { ...orig, listAllPreviewConfigs: vi.fn() };
});
vi.mock("@/pages/previews/hooks/use-preview-envs", () => ({
  usePreviewEnvs: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWriteAnyTeam: true, isOrgAdmin: true }),
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => undefined, defaultTeamName: "default" }),
}));
vi.mock("@/pages/stacks/components/wizard/stack-create-wizard", () => ({
  StackCreateWizard: ({ open }: { open: boolean }) => (open ? <div data-testid="stack-wizard" /> : null),
}));
vi.mock("@/pages/previews/components/enable-repo-wizard/enable-repo-wizard", () => ({
  EnableRepoWizard: ({ open }: { open: boolean }) => (open ? <div data-testid="enable-wizard" /> : null),
}));

import { getStacksByOrg } from "@/api/stacks";
import { listAllPreviewConfigs } from "@/api/preview-configs";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { StackProvider } from "@/pages/stacks/contexts/stack-context";
import StacksPage from "../index";

const stacks = [
  { id: "s-app", name: "tooljet", spec: { stack_resources: [] }, status: { state: "Ready" } },
  { id: "s-preview", name: "pr-1-demo", spec: { stack_resources: [] }, status: { state: "Ready" } },
];

const envs = [
  {
    id: "e1",
    stack_id: "s-preview",
    pr_number: "1",
    branch: "feat/x",
    commit: "abcdef1234",
    config_id: "c1",
    status: { phase: "Ready" as const },
  },
  {
    id: "e2",
    stack_id: "s-other",
    pr_number: "2",
    branch: "feat/y",
    commit: "1234567890",
    config_id: "c2",
    status: { phase: "Deploying" as const },
  },
];

const configs = [
  { id: "c1", name: "demo-repo" },
  { id: "c2", name: "other-repo" },
];

function renderPage(path = "/stacks") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <StackProvider>
        <StacksPage />
      </StackProvider>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  (getStacksByOrg as ReturnType<typeof vi.fn>).mockResolvedValue({ items: stacks });
  (listAllPreviewConfigs as ReturnType<typeof vi.fn>).mockResolvedValue(configs);
  (usePreviewEnvs as ReturnType<typeof vi.fn>).mockReturnValue({
    envs,
    loading: false,
    error: null,
    refresh: vi.fn(),
  });
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("StacksPage views", () => {
  it("deployed view excludes preview-created stacks and shows New Stack", async () => {
    renderPage();
    await waitFor(() => expect(screen.getByText("tooljet")).toBeTruthy());
    expect(screen.queryByText("pr-1-demo")).toBeNull();
    expect(screen.getByRole("button", { name: /new stack/i })).toBeTruthy();
  });

  it("previews view renders env cards with repo names and no create buttons", async () => {
    renderPage("/stacks?view=previews");
    await waitFor(() => expect(screen.getByText("PR #1")).toBeTruthy());
    expect(screen.getByText("PR #2")).toBeTruthy();
    expect(screen.getByText(/demo-repo · feat\/x/)).toBeTruthy();
    expect(screen.queryByRole("button", { name: /new stack/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /enable repository/i })).toBeNull();
  });

  it("repo query param filters previews", async () => {
    renderPage("/stacks?view=previews&repo=c1");
    await waitFor(() => expect(screen.getByText("PR #1")).toBeTruthy());
    expect(screen.queryByText("PR #2")).toBeNull();
  });

  it("toggle switches from deployed to previews", async () => {
    renderPage();
    await waitFor(() => expect(screen.getByText("tooljet")).toBeTruthy());
    await userEvent.click(screen.getByRole("button", { name: /^previews ·/i }));
    await waitFor(() => expect(screen.getByText("PR #1")).toBeTruthy());
    expect(screen.queryByText("tooljet")).toBeNull();
  });

  it("previews empty state offers Enable repository when no configs exist", async () => {
    (listAllPreviewConfigs as ReturnType<typeof vi.fn>).mockResolvedValue([]);
    (usePreviewEnvs as ReturnType<typeof vi.fn>).mockReturnValue({
      envs: [],
      loading: false,
      error: null,
      refresh: vi.fn(),
    });
    renderPage("/stacks?view=previews");
    const cta = await screen.findByRole("button", { name: /enable repository/i });
    await userEvent.click(cta);
    expect(screen.getByTestId("enable-wizard")).toBeTruthy();
  });
});

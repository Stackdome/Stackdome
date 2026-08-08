// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { SheetHost } from "@/test-support/sheet-host";

vi.mock("@/api/stacks", () => ({
  getStacksByOrg: vi.fn(),
}));
vi.mock("@/hooks/use-preview-envs", () => ({
  usePreviewEnvs: vi.fn(),
}));
vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWriteAnyProject: true, canWrite: () => true, isOrgAdmin: true }),
}));

import { getStacksByOrg } from "@/api/stacks";
import { usePreviewEnvs } from "@/hooks/use-preview-envs";
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

function renderPage(path = "/stacks") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <StackProvider>
        <SheetHost>
          <Routes>
            <Route path="/stacks" element={<StacksPage />} />
            <Route path="/previews" element={<div data-testid="previews-page" />} />
          </Routes>
        </SheetHost>
      </StackProvider>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  (getStacksByOrg as ReturnType<typeof vi.fn>).mockResolvedValue({ items: stacks });
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

  it("redirects the old previews tab query param to the dedicated /previews page", async () => {
    renderPage("/stacks?view=previews");
    expect(await screen.findByTestId("previews-page")).toBeTruthy();
    expect(screen.queryByText("tooljet")).toBeNull();
  });
});

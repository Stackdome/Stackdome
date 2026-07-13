// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { PickedRepo } from "@/components/git-source-picker/types";

vi.mock("@/api/git-integrations", () => ({
  listRepositoryBranches: vi.fn().mockResolvedValue({ items: [], total: 0 }),
}));
vi.mock("@/api/preview-configs", () => ({
  createPreviewConfig: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));

// The picker's own network/connect concerns are covered by its dedicated
// tests; here it is a controlled stub that lets the test choose a repo and
// inspect what the wizard passes down.
vi.mock("@/components/git-source-picker/git-source-picker", () => ({
  GitSourcePicker: ({
    onChange,
    publicUrlHint,
  }: {
    onChange: (r: PickedRepo | null) => void;
    publicUrlHint?: string;
  }) => (
    <div>
      <span data-testid="hint">{publicUrlHint}</span>
      <button
        type="button"
        onClick={() =>
          onChange({
            fullName: "acme/webapp",
            cloneUrl: "https://github.com/acme/webapp.git",
            defaultBranch: "main",
            integrationId: "int-app",
          })
        }
      >
        stub-pick-repo
      </button>
    </div>
  ),
}));

import { EnableRepoWizard } from "../enable-repo-wizard";

function renderWizard({ open = true } = {}) {
  return render(
    <MemoryRouter>
      <EnableRepoWizard open={open} onOpenChange={() => {}} onCreated={() => {}} />
    </MemoryRouter>,
  );
}

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("EnableRepoWizard", () => {
  it("opens directly on the pick phase and passes the PR-automation hint", async () => {
    renderWizard({ open: true });
    expect(await screen.findByText("stub-pick-repo")).toBeInTheDocument();
    expect(screen.getByTestId("hint")).toHaveTextContent(
      "PR automation requires a connected provider. Public URLs support manually created previews.",
    );
  });

  it("advances to configure after a repo is picked and Continue is pressed", async () => {
    const user = userEvent.setup();
    renderWizard({ open: true });
    const continueBtn = await screen.findByRole("button", { name: /continue/i });
    expect(continueBtn).toBeDisabled();
    await user.click(screen.getByText("stub-pick-repo"));
    await user.click(screen.getByRole("button", { name: /continue/i }));
    expect(await screen.findByText(/base branch/i)).toBeInTheDocument();
  });
});

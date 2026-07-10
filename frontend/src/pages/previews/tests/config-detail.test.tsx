// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";

vi.mock("@/api/preview-configs", () => ({
  getPreviewConfig: vi.fn(),
  updatePreviewConfig: vi.fn(),
  deletePreviewConfig: vi.fn(),
}));
vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, listPreviewEnvs: vi.fn().mockResolvedValue({ items: [], total: 0 }) };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => undefined, defaultTeamName: "default" }),
}));

import { getPreviewConfig, updatePreviewConfig } from "@/api/preview-configs";
import PreviewConfigDetailPage from "../config-detail";

const config = {
  id: "c1",
  name: "webapp",
  git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
  stackfile_path: "stackfile.yaml",
  max_active_previews: 10,
};

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
  (getPreviewConfig as ReturnType<typeof vi.fn>).mockResolvedValue(config);
});

afterEach(() => cleanup());

describe("PreviewConfigDetailPage", () => {
  it("loads and shows the configuration", async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("webapp")).toBeTruthy();
      expect((screen.getByLabelText(/stackfile path/i) as HTMLInputElement).value).toBe("stackfile.yaml");
    });
  });

  it("saves edits with the repo_url preserved", async () => {
    (updatePreviewConfig as ReturnType<typeof vi.fn>).mockResolvedValue(config);
    renderPage();
    await waitFor(() => screen.getByLabelText(/stackfile path/i));
    const input = screen.getByLabelText(/stackfile path/i);
    await userEvent.clear(input);
    await userEvent.type(input, "deploy/stackfile.yaml");
    await userEvent.click(screen.getByRole("button", { name: /save/i }));
    await waitFor(() => {
      expect(updatePreviewConfig).toHaveBeenCalledWith("org1", "default", "c1", {
        git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
        stackfile_path: "deploy/stackfile.yaml",
        max_active_previews: 10,
      });
    });
  });
});

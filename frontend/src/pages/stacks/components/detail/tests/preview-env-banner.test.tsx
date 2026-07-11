// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, getPreviewEnv: vi.fn(), deletePreviewEnv: vi.fn() };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => "default", defaultTeamName: "default" }),
}));
vi.mock("@/pages/previews/components/sync-env-dialog", () => ({
  SyncEnvDialog: ({ env }: { env: unknown }) => (env ? <div data-testid="sync-dialog" /> : null),
}));

import { getPreviewEnv, PREVIEW_STACK_LABEL, PREVIEW_ID_LABEL } from "@/api/preview-envs";
import { PreviewEnvBanner } from "../preview-env-banner";

const previewLabels = [
  { key: PREVIEW_STACK_LABEL, value: "true" },
  { key: PREVIEW_ID_LABEL, value: "e1" },
];

beforeEach(() => {
  (getPreviewEnv as ReturnType<typeof vi.fn>).mockResolvedValue({
    id: "e1",
    pr_number: "1",
    branch: "feat/x",
    commit: "abcdef1234",
    config_id: "c1",
    status: { phase: "Ready" },
  });
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("PreviewEnvBanner", () => {
  it("renders env details and actions for preview-labeled stacks", async () => {
    render(
      <MemoryRouter>
        <PreviewEnvBanner labels={previewLabels} teamId="t1" />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/preview environment · pr #1/i)).toBeTruthy());
    expect(getPreviewEnv).toHaveBeenCalledWith("org1", "default", "e1");
    expect(screen.getByText("feat/x")).toBeTruthy();
    expect(screen.getByText("abcdef1")).toBeTruthy();
    expect(screen.getByRole("button", { name: /sync/i })).toBeTruthy();
    expect(screen.getByRole("button", { name: /delete environment/i })).toBeTruthy();
    const cfg = screen.getByRole("link", { name: /configuration/i });
    expect(cfg.getAttribute("href")).toBe("/previews/c1");
  });

  it("renders nothing for stacks without the preview label", () => {
    render(
      <MemoryRouter>
        <PreviewEnvBanner labels={[{ key: "app", value: "x" }]} teamId="t1" />
      </MemoryRouter>,
    );
    expect(screen.queryByText(/preview environment/i)).toBeNull();
    expect(getPreviewEnv).not.toHaveBeenCalled();
  });
});

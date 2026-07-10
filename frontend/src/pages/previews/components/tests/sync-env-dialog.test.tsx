// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, syncPreviewEnv: vi.fn() };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => undefined, defaultTeamName: "default" }),
}));

import { SyncEnvDialog } from "../sync-env-dialog";
import type { PreviewStack } from "@/api/preview-envs";

const envA: PreviewStack = { id: "p1", pr_number: "7", branch: "b" };
const envB: PreviewStack = { id: "p2", pr_number: "8", branch: "c" };

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("SyncEnvDialog", () => {
  it("clears stale form state when reopened for a different env after a cancel", async () => {
    const { rerender } = render(
      <SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />,
    );

    await userEvent.type(screen.getByLabelText(/pin to commit/i), "deadbeef");
    expect(screen.getByLabelText(/pin to commit/i)).toHaveValue("deadbeef");

    // Close without submitting (e.g. Cancel/Escape/outside-click).
    rerender(<SyncEnvDialog env={null} onOpenChange={() => {}} onSynced={() => {}} />);

    // Reopen for a different env.
    rerender(<SyncEnvDialog env={envB} onOpenChange={() => {}} onSynced={() => {}} />);

    expect(screen.getByLabelText(/pin to commit/i)).toHaveValue("");
  });
});

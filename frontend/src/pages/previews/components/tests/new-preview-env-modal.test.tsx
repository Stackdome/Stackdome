// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, createPreviewEnv: vi.fn() };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => undefined, defaultTeamName: "default" }),
}));

import { createPreviewEnv } from "@/api/preview-envs";
import { AxiosError } from "axios";
import { NewPreviewEnvModal } from "../new-preview-env-modal";
import type { StackPreviewConfig } from "@/api/preview-configs";

const config: StackPreviewConfig = { id: "c1", name: "webapp" };

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("NewPreviewEnvModal", () => {
  it("creates an environment from PR number and branch", async () => {
    (createPreviewEnv as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "p1" });
    const onCreated = vi.fn();
    render(<NewPreviewEnvModal open onOpenChange={() => {}} config={config} onCreated={onCreated} />);

    await userEvent.type(screen.getByLabelText(/pr number/i), "42");
    await userEvent.type(screen.getByLabelText(/branch/i), "feat/login");
    await userEvent.click(screen.getByRole("button", { name: /create environment/i }));

    await waitFor(() => {
      expect(createPreviewEnv).toHaveBeenCalledWith("org1", "default", {
        config_id: "c1",
        pr_number: "42",
        branch: "feat/login",
      });
      expect(onCreated).toHaveBeenCalled();
    });
  });

  it("sends pasted stackfile content when provided", async () => {
    (createPreviewEnv as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "p1" });
    render(<NewPreviewEnvModal open onOpenChange={() => {}} config={config} onCreated={() => {}} />);

    await userEvent.type(screen.getByLabelText(/pr number/i), "7");
    await userEvent.type(screen.getByLabelText(/branch/i), "fix/nav");
    await userEvent.click(screen.getByRole("button", { name: /advanced/i }));
    await userEvent.type(screen.getByLabelText(/stackfile content/i), "name: test");
    await userEvent.click(screen.getByRole("button", { name: /create environment/i }));

    await waitFor(() => {
      expect(createPreviewEnv).toHaveBeenCalledWith("org1", "default", {
        config_id: "c1",
        pr_number: "7",
        branch: "fix/nav",
        stackfile_content: "name: test",
      });
    });
  });

  it("clears stale form state after closing without submitting and reopening", async () => {
    const { rerender } = render(
      <NewPreviewEnvModal open onOpenChange={() => {}} config={config} onCreated={() => {}} />,
    );

    await userEvent.type(screen.getByLabelText(/pr number/i), "42");
    await userEvent.type(screen.getByLabelText(/branch/i), "feat/login");

    // Close without submitting (e.g. Cancel/Escape/outside-click).
    rerender(
      <NewPreviewEnvModal open={false} onOpenChange={() => {}} config={config} onCreated={() => {}} />,
    );

    // Reopen.
    rerender(
      <NewPreviewEnvModal open onOpenChange={() => {}} config={config} onCreated={() => {}} />,
    );

    expect(screen.getByLabelText(/pr number/i)).toHaveValue(null);
    expect(screen.getByLabelText(/branch/i)).toHaveValue("");
  });

  it("shows inline conflict message on 409", async () => {
    const err = new AxiosError("conflict");
    Object.defineProperty(err, "response", { value: { status: 409, data: { reason: "exists" } } });
    (createPreviewEnv as ReturnType<typeof vi.fn>).mockRejectedValue(err);
    render(<NewPreviewEnvModal open onOpenChange={() => {}} config={config} onCreated={() => {}} />);

    await userEvent.type(screen.getByLabelText(/pr number/i), "42");
    await userEvent.type(screen.getByLabelText(/branch/i), "feat/login");
    await userEvent.click(screen.getByRole("button", { name: /create environment/i }));

    await waitFor(() => {
      expect(screen.getByText(/pr #42 already has an environment/i)).toBeTruthy();
    });
  });
});

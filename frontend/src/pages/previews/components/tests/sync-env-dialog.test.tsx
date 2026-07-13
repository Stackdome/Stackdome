// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, syncPreviewEnv: vi.fn() };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));

import { syncPreviewEnv } from "@/api/preview-envs";
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

    await userEvent.type(screen.getByLabelText(/pin to a specific commit/i), "deadbeef");
    expect(screen.getByLabelText(/pin to a specific commit/i)).toHaveValue("deadbeef");

    // Close without submitting (e.g. Cancel/Escape/outside-click).
    rerender(<SyncEnvDialog env={null} onOpenChange={() => {}} onSynced={() => {}} />);

    // Reopen for a different env.
    rerender(<SyncEnvDialog env={envB} onOpenChange={() => {}} onSynced={() => {}} />);

    expect(screen.getByLabelText(/pin to a specific commit/i)).toHaveValue("");
  });

  it("hides stackfile content and image overrides until Advanced is expanded", () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);

    expect(screen.queryByLabelText(/stackfile content/i)).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/image overrides/i)).not.toBeInTheDocument();
  });

  it("reveals stackfile content and image overrides under Advanced", async () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);

    await userEvent.click(screen.getByRole("button", { name: /advanced/i }));

    expect(screen.getByLabelText(/stackfile content/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/image overrides/i)).toBeInTheDocument();
  });

  it("resets the Advanced collapse when reopened for a different env", async () => {
    const { rerender } = render(
      <SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />,
    );

    await userEvent.click(screen.getByRole("button", { name: /advanced/i }));
    expect(screen.getByLabelText(/stackfile content/i)).toBeInTheDocument();

    rerender(<SyncEnvDialog env={null} onOpenChange={() => {}} onSynced={() => {}} />);
    rerender(<SyncEnvDialog env={envB} onOpenChange={() => {}} onSynced={() => {}} />);

    expect(screen.queryByLabelText(/stackfile content/i)).not.toBeInTheDocument();
  });

  it("sends parsed image_overrides in the sync payload when provided", async () => {
    (syncPreviewEnv as ReturnType<typeof vi.fn>).mockResolvedValue({});
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);

    await userEvent.click(screen.getByRole("button", { name: /advanced/i }));
    await userEvent.type(
      screen.getByLabelText(/image overrides/i),
      "web=registry.example.com/web:sha-1",
    );
    await userEvent.click(screen.getByRole("button", { name: /^sync$/i }));

    await waitFor(() => {
      expect(syncPreviewEnv).toHaveBeenCalledWith("org1", "default", "p1", {
        image_overrides: { web: "registry.example.com/web:sha-1" },
      });
    });
  });

  it("has no required-field asterisks — every field is optional", () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);
    const commitLabel = screen.getByText(/pin to a specific commit/i).closest("label");
    expect(commitLabel?.querySelector('[aria-hidden]')).toBeNull();
  });

  it("rejects a malformed commit SHA and does not sync", async () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);
    await userEvent.type(screen.getByLabelText(/pin to a specific commit/i), "not-hex!");
    await userEvent.click(screen.getByRole("button", { name: /^sync$/i }));

    expect(await screen.findByText(/valid commit sha/i)).toBeInTheDocument();
    expect(syncPreviewEnv).not.toHaveBeenCalled();
  });

  it("clears the commit error once the field is edited", async () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);
    await userEvent.type(screen.getByLabelText(/pin to a specific commit/i), "zzz");
    await userEvent.click(screen.getByRole("button", { name: /^sync$/i }));
    expect(await screen.findByText(/valid commit sha/i)).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText(/pin to a specific commit/i), "1");
    expect(screen.queryByText(/valid commit sha/i)).not.toBeInTheDocument();
  });

  it("rejects an invalid image override line and expands Advanced to show it", async () => {
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);
    await userEvent.click(screen.getByRole("button", { name: /advanced/i }));
    await userEvent.type(screen.getByLabelText(/image overrides/i), "not-a-pair");
    await userEvent.click(screen.getByRole("button", { name: /^sync$/i }));

    expect(await screen.findByText(/resource=image/i)).toBeInTheDocument();
    expect(syncPreviewEnv).not.toHaveBeenCalled();
  });

  it("omits image_overrides from the payload when the field is empty", async () => {
    (syncPreviewEnv as ReturnType<typeof vi.fn>).mockResolvedValue({});
    render(<SyncEnvDialog env={envA} onOpenChange={() => {}} onSynced={() => {}} />);

    await userEvent.type(screen.getByLabelText(/pin to a specific commit/i), "deadbeef");
    await userEvent.click(screen.getByRole("button", { name: /^sync$/i }));

    await waitFor(() => {
      expect(syncPreviewEnv).toHaveBeenCalledWith("org1", "default", "p1", {
        commit: "deadbeef",
      });
    });
  });
});

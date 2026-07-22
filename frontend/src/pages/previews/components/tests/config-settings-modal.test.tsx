// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const toastMock = vi.fn();

vi.mock("@/api/preview-configs", () => ({
  updatePreviewConfig: vi.fn(),
  deletePreviewConfig: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

import { updatePreviewConfig, deletePreviewConfig } from "@/api/preview-configs";
import { ConfigSettingsModal } from "../config-settings-modal";
import { ConfirmProvider } from "@/components/branded/confirm";
import type { StackPreviewConfig } from "@/api/preview-configs";

const config: StackPreviewConfig = {
  id: "c1",
  name: "webapp",
  git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
  stackfile_path: "stackfile.yaml",
  max_active_previews: 10,
};

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("ConfigSettingsModal", () => {
  it("seeds fields from config when opened", () => {
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    expect(screen.getByLabelText(/base branch/i)).toHaveValue("main");
    expect(screen.getByLabelText(/stackfile path/i)).toHaveValue("stackfile.yaml");
    expect(screen.getByLabelText(/max active previews/i)).toHaveValue(10);
  });

  it("re-seeds fields from the latest config on reopen, discarding unsaved edits", async () => {
    const { rerender } = render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );

    const input = screen.getByLabelText(/stackfile path/i);
    await userEvent.clear(input);
    await userEvent.type(input, "deploy/stackfile.yaml");
    expect(input).toHaveValue("deploy/stackfile.yaml");

    // Close without saving.
    rerender(
      <ConfigSettingsModal open={false} config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );
    // Reopen.
    rerender(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );

    expect(screen.getByLabelText(/stackfile path/i)).toHaveValue("stackfile.yaml");
  });

  it("re-seeds when a newer config prop arrives while still open", () => {
    const { rerender } = render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    const updated: StackPreviewConfig = { ...config, stackfile_path: "new/path.yaml" };
    rerender(
      <ConfigSettingsModal open config={updated} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );
    expect(screen.getByLabelText(/stackfile path/i)).toHaveValue("new/path.yaml");
  });

  it("clamps max active previews to a positive integer", async () => {
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    const input = screen.getByLabelText(/max active previews/i);
    fireEvent.change(input, { target: { value: "-5" } });
    expect(input).toHaveValue(1);
  });

  it("requires a non-empty stackfile path and blocks the save", async () => {
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    const input = screen.getByLabelText(/stackfile path/i);
    await userEvent.clear(input);
    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));

    expect(await screen.findByText(/stackfile path is required/i)).toBeInTheDocument();
    expect(updatePreviewConfig).not.toHaveBeenCalled();

    const label = screen.getByText(/^stackfile path$/i).closest("label");
    expect(label?.querySelector('[aria-hidden]')).toHaveTextContent("*");
  });

  it("clears the stackfile path error once the field is edited", async () => {
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    const input = screen.getByLabelText(/stackfile path/i);
    await userEvent.clear(input);
    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));
    expect(await screen.findByText(/stackfile path is required/i)).toBeInTheDocument();

    await userEvent.type(input, "stackfile.yaml");
    expect(screen.queryByText(/stackfile path is required/i)).not.toBeInTheDocument();
  });

  it("saves with the full-replace PUT echo of repo_url/description/labels/annotations", async () => {
    const withExtras: StackPreviewConfig = {
      ...config,
      description: "webapp previews",
      labels: [{ key: "project", value: "platform" }],
      annotations: [{ key: "note", value: "internal" }],
    };
    vi.mocked(updatePreviewConfig).mockResolvedValue(withExtras);
    const onSaved = vi.fn();
    const onOpenChange = vi.fn();

    render(
      <ConfigSettingsModal open config={withExtras} onOpenChange={onOpenChange} onSaved={onSaved} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );

    const input = screen.getByLabelText(/stackfile path/i);
    await userEvent.clear(input);
    await userEvent.type(input, "deploy/stackfile.yaml");
    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(updatePreviewConfig).toHaveBeenCalledWith("org1", "default", "c1", {
        git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
        stackfile_path: "deploy/stackfile.yaml",
        max_active_previews: 10,
        env: [],
        description: "webapp previews",
        labels: [{ key: "project", value: "platform" }],
        annotations: [{ key: "note", value: "internal" }],
      });
    });
    expect(onSaved).toHaveBeenCalledWith(withExtras);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("pre-fills existing env vars from the loaded config", () => {
    const withEnv: StackPreviewConfig = {
      ...config,
      env: [{ name: "FOO", value: "bar" }, { name: "BAZ", value: "" }],
    };
    render(
      <ConfigSettingsModal open config={withEnv} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );
    const names = screen.getAllByLabelText(/^variable name$/i);
    const values = screen.getAllByLabelText(/^variable value$/i);
    expect(names.map((n) => (n as HTMLInputElement).value)).toEqual(["FOO", "BAZ"]);
    expect(values.map((v) => (v as HTMLInputElement).value)).toEqual(["bar", ""]);
  });

  it("saves edited env vars and strips empty-named rows", async () => {
    vi.mocked(updatePreviewConfig).mockResolvedValue(config);
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );

    await userEvent.click(screen.getByRole("button", { name: /add variable/i }));
    await userEvent.click(screen.getByRole("button", { name: /add variable/i }));
    const names = screen.getAllByLabelText(/^variable name$/i);
    const values = screen.getAllByLabelText(/^variable value$/i);
    await userEvent.type(names[0], "FOO");
    await userEvent.type(values[0], "bar");

    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(updatePreviewConfig).toHaveBeenCalledWith(
        "org1",
        "default",
        "c1",
        expect.objectContaining({ env: [{ name: "FOO", value: "bar" }] }),
      );
    });
  });

  it("shows a validation error and blocks save on a duplicate env var name", async () => {
    render(
      <ConfigSettingsModal open config={config} onOpenChange={() => {}} onSaved={() => {}} onDeleted={() => {}} />,
    );

    await userEvent.click(screen.getByRole("button", { name: /add variable/i }));
    await userEvent.click(screen.getByRole("button", { name: /add variable/i }));
    const names = screen.getAllByLabelText(/^variable name$/i);
    await userEvent.type(names[0], "FOO");
    await userEvent.type(names[1], "FOO");

    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));

    expect(await screen.findByText(/duplicate variable name/i)).toBeInTheDocument();
    expect(updatePreviewConfig).not.toHaveBeenCalled();
  });

  it("keeps the modal open on save failure", async () => {
    vi.mocked(updatePreviewConfig).mockRejectedValue(new Error("boom"));
    const onSaved = vi.fn();
    const onOpenChange = vi.fn();

    render(
      <ConfigSettingsModal open config={config} onOpenChange={onOpenChange} onSaved={onSaved} onDeleted={() => {}} />,
      { wrapper: ConfirmProvider },
    );
    await userEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Save failed", variant: "destructive" }),
    ));
    expect(onSaved).not.toHaveBeenCalled();
    expect(onOpenChange).not.toHaveBeenCalled();
  });

  it("deletes the configuration through the confirm dialog", async () => {
    vi.mocked(deletePreviewConfig).mockResolvedValue(undefined);
    const onDeleted = vi.fn();
    const onOpenChange = vi.fn();

    render(
      <ConfigSettingsModal open config={config} onOpenChange={onOpenChange} onSaved={() => {}} onDeleted={onDeleted} />,
      { wrapper: ConfirmProvider },
    );

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /delete configuration/i }));
    expect(await screen.findByText(/delete webapp\?/i)).toBeInTheDocument();
    // fireEvent (not userEvent) for this click: while the confirm AlertDialog
    // is open, the settings Dialog's own focus trap is still mounted too.
    // userEvent's pointer/focus simulation recurses into a stack overflow
    // when jsdom's synchronous focus events bounce between two simultaneously
    // -mounted Radix focus traps. fireEvent dispatches a plain click without
    // that focus machinery, sidestepping the jsdom-only pathological path.
    fireEvent.click(screen.getByRole("button", { name: /^delete$/i }));

    await waitFor(() => expect(deletePreviewConfig).toHaveBeenCalledWith("org1", "default", "c1"));
    expect(onDeleted).toHaveBeenCalledOnce();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("keeps settings open on delete failure (backend rejection)", async () => {
    vi.mocked(deletePreviewConfig).mockRejectedValue(new Error("environments still exist"));
    const onDeleted = vi.fn();
    const onOpenChange = vi.fn();

    render(
      <ConfigSettingsModal open config={config} onOpenChange={onOpenChange} onSaved={() => {}} onDeleted={onDeleted} />,
      { wrapper: ConfirmProvider },
    );

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /delete configuration/i }));
    expect(await screen.findByText(/delete webapp\?/i)).toBeInTheDocument();
    // fireEvent (not userEvent) for this click: on rejection both the settings
    // Dialog and the confirm AlertDialog stay open (by design), and userEvent's
    // pointer/focus simulation recurses into a stack overflow when jsdom's
    // synchronous focus events bounce between two simultaneously-open Radix
    // focus traps. fireEvent dispatches a plain click without that focus
    // machinery, sidestepping the jsdom-only pathological path while still
    // exercising the real click handler and assertions below.
    fireEvent.click(screen.getByRole("button", { name: /^delete$/i }));

    await waitFor(() => expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Delete failed", variant: "destructive" }),
    ));
    expect(onDeleted).not.toHaveBeenCalled();
    expect(onOpenChange).not.toHaveBeenCalled();
    expect(screen.getByLabelText(/stackfile path/i)).toBeInTheDocument();
  });
});

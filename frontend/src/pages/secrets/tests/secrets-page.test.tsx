// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import SecretsPage from "../index";
import { ConfirmProvider } from "@/components/branded/confirm";

afterEach(cleanup);

const toastMock = vi.fn();

vi.mock("@/api/secrets", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  getSecrets: vi.fn().mockResolvedValue({ items: [] }),
  deleteSecret: vi.fn(),
  createSecret: vi.fn(),
  updateSecret: vi.fn(),
}));
vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: () => "org-1" }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));
vi.mock("@/hooks/use-breadcrumb", () => ({
  useBreadcrumb: () => ({ setCustomLabel: vi.fn(), setPathLoading: vi.fn() }),
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({
    projectNameById: () => "default",
    defaultProjectName: "default",
  }),
}));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWrite: () => true, canWriteAnyProject: true }),
}));

import { getSecrets, deleteSecret } from "@/api/secrets";

const secret = {
  id: "s1",
  name: "api-key",
  type: "Generic",
  project_id: "p1",
  data: [],
};

describe("SecretsPage", () => {
  beforeEach(() => vi.clearAllMocks());

  it("deletes a secret via the row menu and confirm dialog", async () => {
    vi.mocked(getSecrets)
      .mockResolvedValueOnce({ items: [secret] })
      .mockResolvedValueOnce({ items: [] });
    vi.mocked(deleteSecret).mockResolvedValue(undefined);

    render(<ConfirmProvider><SecretsPage /></ConfirmProvider>);
    await waitFor(() => expect(screen.getByText("api-key")).toBeInTheDocument());

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Secret actions" }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText("Delete"), { pointerEventsCheck: 0 });

    const dialog = await screen.findByRole("alertdialog");
    expect(dialog).toHaveTextContent(/delete secret\?/i);
    await user.click(screen.getByRole("button", { name: "Delete" }), { pointerEventsCheck: 0 });

    await waitFor(() => expect(deleteSecret).toHaveBeenCalledWith("org-1", "default", "s1"));
    expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Secret deleted", variant: "success" }));
  });

  it("does not delete when the confirm dialog is cancelled", async () => {
    vi.mocked(getSecrets).mockResolvedValue({ items: [secret] });

    render(<ConfirmProvider><SecretsPage /></ConfirmProvider>);
    await waitFor(() => expect(screen.getByText("api-key")).toBeInTheDocument());

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Secret actions" }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText("Delete"), { pointerEventsCheck: 0 });

    await screen.findByRole("alertdialog");
    await user.click(screen.getByRole("button", { name: /cancel/i }), { pointerEventsCheck: 0 });

    await waitFor(() => expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument());
    expect(deleteSecret).not.toHaveBeenCalled();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError, AxiosHeaders } from "axios";
import type { ObjectStore } from "@/pages/object-stores/types";

const toastMock = vi.fn();

vi.mock("@/api/object-stores", () => ({
  deleteObjectStore: vi.fn(),
}));

vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: vi.fn(() => "org-1"),
}));

vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({
    teamNameById: (id: string | undefined) => (id === "t1" ? "alpha" : undefined),
    defaultTeamName: "alpha",
    teams: [],
  }),
}));

import { deleteObjectStore } from "@/api/object-stores";
import { ObjectStoreDeleteDialog } from "../object-store-delete-dialog";

const mockedDelete = vi.mocked(deleteObjectStore);

function mkStore(over: Partial<ObjectStore> = {}): ObjectStore {
  return {
    id: "os-1",
    name: "my-store",
    team_id: "t1",
    ...over,
  } as ObjectStore;
}

// Build a minimal AxiosError that matches what `getErrorStatus` / `getErrorMessage`
// expect: an AxiosError instance with `response.status` and `response.data.reason`.
function mkAxiosError(status: number, reason: string): AxiosError {
  const err = new AxiosError(`Request failed with status code ${status}`);
  err.response = {
    status,
    statusText: "",
    headers: {},
    config: { headers: new AxiosHeaders() } as never,
    data: { reason },
  };
  return err;
}

beforeEach(() => {
  mockedDelete.mockReset();
  toastMock.mockReset();
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

async function clickDelete() {
  const user = userEvent.setup();
  const deleteBtn = screen.getByRole("button", { name: /^delete$/i });
  await user.click(deleteBtn);
}

describe("ObjectStoreDeleteDialog", () => {
  it("happy path: confirms delete, toasts success (destructive), calls onDeleted + onOpenChange(false)", async () => {
    mockedDelete.mockResolvedValueOnce(undefined as never);
    const onOpenChange = vi.fn();
    const onDeleted = vi.fn();

    render(
      <ObjectStoreDeleteDialog
        store={mkStore()}
        onOpenChange={onOpenChange}
        onDeleted={onDeleted}
      />,
    );

    await clickDelete();

    await waitFor(() => expect(mockedDelete).toHaveBeenCalledTimes(1));
    expect(mockedDelete).toHaveBeenCalledWith("org-1", "alpha", "os-1");
    await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
    expect(onOpenChange).toHaveBeenCalledWith(false);
    // success toast fired exactly once, with destructive variant
    const successCalls = toastMock.mock.calls.filter(
      ([arg]) => arg?.title === "Object Store deleted",
    );
    expect(successCalls).toHaveLength(1);
    expect(successCalls[0][0].variant).toBe("destructive");
  });

  it("409 conflict: shows in-dialog banner, disables Delete, does NOT call onDeleted/onOpenChange(false)", async () => {
    mockedDelete.mockRejectedValueOnce(
      mkAxiosError(409, "object store is in use by addon foo"),
    );
    const onOpenChange = vi.fn();
    const onDeleted = vi.fn();

    render(
      <ObjectStoreDeleteDialog
        store={mkStore()}
        onOpenChange={onOpenChange}
        onDeleted={onDeleted}
      />,
    );

    await clickDelete();

    // Conflict banner surfaces the backend reason.
    const banner = await screen.findByText(/in use by addon foo/i);
    expect(banner).toBeInTheDocument();
    // The banner wrapper uses the text-warn class for styling.
    const warnEl = document.querySelector(".text-warn");
    expect(warnEl).not.toBeNull();

    // Delete button is now disabled (conflictMessage truthy).
    expect(screen.getByRole("button", { name: /^delete$/i })).toBeDisabled();

    expect(onDeleted).not.toHaveBeenCalled();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
    // No success toast — confirm we did not accidentally toast success.
    expect(
      toastMock.mock.calls.find(([a]) => a?.title === "Object Store deleted"),
    ).toBeUndefined();
  });

  it("400 + 'in use' regex: surfaces conflict banner via dual-check path", async () => {
    mockedDelete.mockRejectedValueOnce(
      mkAxiosError(400, "object store is in use by one or more postgres addons"),
    );
    const onOpenChange = vi.fn();
    const onDeleted = vi.fn();

    render(
      <ObjectStoreDeleteDialog
        store={mkStore()}
        onOpenChange={onOpenChange}
        onDeleted={onDeleted}
      />,
    );

    await clickDelete();

    expect(
      await screen.findByText(/in use by one or more postgres addons/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^delete$/i })).toBeDisabled();
    expect(onDeleted).not.toHaveBeenCalled();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("400 + unrelated message: falls through to generic destructive toast, NOT conflict banner", async () => {
    mockedDelete.mockRejectedValueOnce(mkAxiosError(400, "something else went wrong"));
    const onOpenChange = vi.fn();
    const onDeleted = vi.fn();

    render(
      <ObjectStoreDeleteDialog
        store={mkStore()}
        onOpenChange={onOpenChange}
        onDeleted={onDeleted}
      />,
    );

    await clickDelete();

    await waitFor(() => {
      expect(
        toastMock.mock.calls.find(([a]) => a?.title === "Failed to delete"),
      ).toBeDefined();
    });
    const failureCall = toastMock.mock.calls.find(
      ([a]) => a?.title === "Failed to delete",
    );
    expect(failureCall?.[0].variant).toBe("destructive");
    expect(failureCall?.[0].description).toMatch(/something else went wrong/i);

    // No conflict banner — element with text-warn class should be absent.
    expect(document.querySelector(".text-warn")).toBeNull();
    // Delete button is re-enabled (no conflict).
    expect(screen.getByRole("button", { name: /^delete$/i })).not.toBeDisabled();
    expect(onDeleted).not.toHaveBeenCalled();
  });
});

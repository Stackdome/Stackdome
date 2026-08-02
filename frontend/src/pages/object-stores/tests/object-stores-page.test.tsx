// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError, AxiosHeaders } from "axios";
import { ConfirmProvider } from "@/components/branded/confirm";
import type { ObjectStore } from "@/pages/object-stores/types";

const toastMock = vi.fn();
const refetchMock = vi.fn();

vi.mock("@/api/object-stores", () => ({
  deleteObjectStore: vi.fn(),
}));

vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: vi.fn(() => "org-1"),
}));

vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({
    projectNameById: (id: string | undefined) => (id === "t1" ? "alpha" : undefined),
    defaultProjectName: "alpha",
    projects: [],
  }),
}));

vi.mock("@/hooks/use-breadcrumb", () => ({
  useBreadcrumb: () => ({ setCustomLabel: vi.fn(), setPathLoading: vi.fn() }),
}));

vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWrite: () => true, canWriteAnyProject: true }),
}));

vi.mock("../components/object-store-form-dialog", () => ({
  ObjectStoreFormDialog: () => null,
}));

vi.mock("../hooks/use-object-stores", () => ({
  useObjectStores: () => ({
    objectStores: [mkStore()],
    loading: false,
    error: null,
    refetch: refetchMock,
  }),
}));

import { deleteObjectStore } from "@/api/object-stores";
import ObjectStoresPage from "../index";

const mockedDelete = vi.mocked(deleteObjectStore);

function mkStore(over: Partial<ObjectStore> = {}): ObjectStore {
  return {
    id: "os-1",
    name: "my-store",
    project_id: "t1",
    spec: { configuration: {}, destination_path: "backups", retention_policy: "7d" },
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
  refetchMock.mockReset();
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

async function confirmDelete() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: /delete my-store/i }));
  await screen.findByRole("alertdialog");
  await user.click(screen.getByRole("button", { name: /^delete$/i }));
}

describe("ObjectStoresPage delete", () => {
  it("confirms, deletes, toasts success and refetches", async () => {
    mockedDelete.mockResolvedValueOnce(undefined as never);

    render(
      <ConfirmProvider>
        <ObjectStoresPage />
      </ConfirmProvider>,
    );

    await confirmDelete();

    await waitFor(() => expect(mockedDelete).toHaveBeenCalledTimes(1));
    expect(mockedDelete).toHaveBeenCalledWith("org-1", "alpha", "os-1");
    await waitFor(() => expect(refetchMock).toHaveBeenCalledTimes(1));
    const successCalls = toastMock.mock.calls.filter(
      ([arg]) => arg?.title === "Object store deleted",
    );
    expect(successCalls).toHaveLength(1);
    expect(successCalls[0][0].variant).toBe("success");
  });

  it("cancelling leaves the store untouched", async () => {
    const user = userEvent.setup();

    render(
      <ConfirmProvider>
        <ObjectStoresPage />
      </ConfirmProvider>,
    );

    await user.click(screen.getByRole("button", { name: /delete my-store/i }));
    await screen.findByRole("alertdialog");
    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(mockedDelete).not.toHaveBeenCalled();
    expect(refetchMock).not.toHaveBeenCalled();
  });

  it("409 conflict surfaces the backend reason as a destructive toast", async () => {
    mockedDelete.mockRejectedValueOnce(mkAxiosError(409, "object store is in use by addon foo"));

    render(
      <ConfirmProvider>
        <ObjectStoresPage />
      </ConfirmProvider>,
    );

    await confirmDelete();

    await waitFor(() => {
      const call = toastMock.mock.calls.find(([a]) => a?.title === "Failed to delete");
      expect(call?.[0].variant).toBe("destructive");
      expect(call?.[0].description).toMatch(/in use by addon foo/i);
    });
    expect(refetchMock).not.toHaveBeenCalled();
  });

  it("400 + 'in use' reason surfaces the conflict through the dual-check path", async () => {
    mockedDelete.mockRejectedValueOnce(
      mkAxiosError(400, "object store is in use by one or more postgres addons"),
    );

    render(
      <ConfirmProvider>
        <ObjectStoresPage />
      </ConfirmProvider>,
    );

    await confirmDelete();

    await waitFor(() => {
      const call = toastMock.mock.calls.find(([a]) => a?.title === "Failed to delete");
      expect(call?.[0].description).toMatch(/in use by one or more postgres addons/i);
    });
  });

  it("400 + unrelated message falls through to the generic destructive toast", async () => {
    mockedDelete.mockRejectedValueOnce(mkAxiosError(400, "something else went wrong"));

    render(
      <ConfirmProvider>
        <ObjectStoresPage />
      </ConfirmProvider>,
    );

    await confirmDelete();

    await waitFor(() => {
      const call = toastMock.mock.calls.find(([a]) => a?.title === "Failed to delete");
      expect(call?.[0].variant).toBe("destructive");
      expect(call?.[0].description).toMatch(/something else went wrong/i);
    });
    expect(refetchMock).not.toHaveBeenCalled();
  });
});

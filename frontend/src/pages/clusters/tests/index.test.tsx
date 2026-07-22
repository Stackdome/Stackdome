// @vitest-environment jsdom
import { describe, it, expect, afterEach, beforeEach, beforeAll, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";

// Radix popper content reads ResizeObserver on mount, which jsdom doesn't implement.
beforeAll(() => {
  global.ResizeObserver =
    global.ResizeObserver ||
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
});

import ClustersPage from "../index";
import type { Cluster } from "../types";
import { ConfirmProvider } from "@/components/branded/confirm";

const useClustersMock = vi.fn();

vi.mock("../hooks/use-clusters", () => ({
  useClusters: () => useClustersMock(),
}));
vi.mock("@/api/clusters", () => ({
  deleteCluster: vi.fn(),
  createCluster: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

const toastMock = vi.fn();
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: toastMock, dismiss: vi.fn(), toasts: [] }),
}));

const cluster = { id: "c1", name: "kind-local" } as Cluster;

function renderPage() {
  return render(
    <ConfirmProvider>
      <MemoryRouter initialEntries={["/clusters"]}>
        <Routes>
          <Route path="/clusters" element={<ClustersPage />} />
          <Route path="/clusters/:id" element={<div>cluster detail</div>} />
        </Routes>
      </MemoryRouter>
    </ConfirmProvider>,
  );
}

afterEach(cleanup);

describe("ClustersPage", () => {
  beforeEach(() => vi.clearAllMocks());

  it("stays on the list page and disables Add Cluster when one cluster exists", () => {
    useClustersMock.mockReturnValue({ clusters: [cluster], loading: false, error: null, refetch: vi.fn() });
    renderPage();
    expect(screen.getByText("All Clusters")).toBeInTheDocument();
    expect(screen.getByText("kind-local")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add Cluster/ })).toBeDisabled();
  });

  it("keeps Add Cluster enabled when no clusters exist", () => {
    useClustersMock.mockReturnValue({ clusters: [], loading: false, error: null, refetch: vi.fn() });
    renderPage();
    expect(screen.getByText("No clusters configured")).toBeInTheDocument();
    for (const button of screen.getAllByRole("button", { name: /Add Cluster/ })) {
      expect(button).toBeEnabled();
    }
  });

  it("navigates to the cluster detail when the row is clicked", async () => {
    useClustersMock.mockReturnValue({ clusters: [cluster], loading: false, error: null, refetch: vi.fn() });
    renderPage();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /kind-local/ }));

    expect(await screen.findByText("cluster detail")).toBeInTheDocument();
  });
});

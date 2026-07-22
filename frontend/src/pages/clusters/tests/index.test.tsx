// @vitest-environment jsdom
import { describe, it, expect, afterEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import ClustersPage from "../index";
import type { Cluster } from "../types";

const useClustersMock = vi.fn();

vi.mock("../hooks/use-clusters", () => ({
  useClusters: () => useClustersMock(),
}));

const cluster = { id: "c1", name: "kind-local" } as Cluster;

afterEach(cleanup);

describe("ClustersPage", () => {
  it("stays on the list page and disables Add Cluster when one cluster exists", () => {
    useClustersMock.mockReturnValue({ clusters: [cluster], loading: false, error: null, refetch: vi.fn() });
    render(
      <MemoryRouter>
        <ClustersPage />
      </MemoryRouter>,
    );
    expect(screen.getByText("All Clusters")).toBeInTheDocument();
    expect(screen.getByText("kind-local")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add Cluster/ })).toBeDisabled();
  });

  it("keeps Add Cluster enabled when no clusters exist", () => {
    useClustersMock.mockReturnValue({ clusters: [], loading: false, error: null, refetch: vi.fn() });
    render(
      <MemoryRouter>
        <ClustersPage />
      </MemoryRouter>,
    );
    expect(screen.getByText("No clusters configured")).toBeInTheDocument();
    for (const button of screen.getAllByRole("button", { name: /Add Cluster/ })) {
      expect(button).toBeEnabled();
    }
  });
});

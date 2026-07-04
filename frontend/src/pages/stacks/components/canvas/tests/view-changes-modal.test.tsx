// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup, within } from "@testing-library/react";
import { ViewChangesModal } from "../ViewChangesModal";
import type { SnapshotDiff } from "@/pages/stacks/components/detail/deployments/release-snapshot-diff";

afterEach(cleanup);

const diff: SnapshotDiff = {
  resources: [
    { name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "nginx:1.25", to: "nginx:1.27", kind: "changed" }] }] },
    { name: "cache", change: "removed", sections: [] },
  ],
  volumes: [{ name: "data", change: "added", rows: [{ key: "size", to: "10Gi", kind: "added" }] }],
  connections: [],
};

const base = {
  open: true,
  onOpenChange: () => {},
  diff,
  count: 3,
  stackName: "payments-api",
  onDiscardResource: () => {},
  onDiscardVolume: () => {},
  onDiscardAll: () => {},
  onDeploy: () => {},
  deployBusy: false,
  canWrite: true,
};

describe("ViewChangesModal", () => {
  it("lists each change and the deploy count", () => {
    render(<ViewChangesModal {...base} />);
    expect(screen.getByText("Undeployed changes")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("cache")).toBeInTheDocument();
    expect(screen.getByText("data")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Deploy" })).toBeInTheDocument();
  });

  it("renders group labels for populated groups only", () => {
    render(<ViewChangesModal {...base} />);
    expect(screen.getByText("Resources")).toBeInTheDocument();
    expect(screen.getByText("Volumes")).toBeInTheDocument();
    expect(screen.queryByText("Connections")).toBeNull();
  });

  it("discards a modified resource by name", () => {
    const onDiscardResource = vi.fn();
    render(<ViewChangesModal {...base} onDiscardResource={onDiscardResource} />);
    // The first Discard button belongs to the "web" (modified) card.
    const discards = screen.getAllByRole("button", { name: "Discard" });
    fireEvent.click(discards[0]);
    expect(onDiscardResource).toHaveBeenCalledWith("web");
  });

  it("disables per-card discard for removed entries", () => {
    render(<ViewChangesModal {...base} />);
    const cacheCard = screen.getByText("cache").closest("[data-change-card]") as HTMLElement;
    expect(within(cacheCard).getByRole("button", { name: "Discard" })).toBeDisabled();
  });

  it("deploys and closes", () => {
    const onDeploy = vi.fn();
    const onOpenChange = vi.fn();
    render(<ViewChangesModal {...base} onDeploy={onDeploy} onOpenChange={onOpenChange} />);
    fireEvent.click(screen.getByRole("button", { name: "Deploy" }));
    expect(onDeploy).toHaveBeenCalled();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows an empty state and disables deploy when there are no changes", () => {
    render(<ViewChangesModal {...base} count={0} diff={{ resources: [], volumes: [], connections: [] }} />);
    expect(screen.getByText("No pending changes.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /deploy/i })).toBeDisabled();
  });

  it("shows a saving hint when the diff has not resolved yet", () => {
    render(<ViewChangesModal {...base} diff={undefined} />);
    expect(screen.getByText(/saving changes/i)).toBeInTheDocument();
  });
});

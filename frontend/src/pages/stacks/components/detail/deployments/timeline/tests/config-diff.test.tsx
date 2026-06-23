// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ConfigDiff } from "../config-diff";
import type { SnapshotDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const diff: SnapshotDiff = {
  resources: [
    { name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "web:1", to: "web:2", kind: "changed" }] }] },
    { name: "worker", change: "added", sections: [{ kind: "configuration", rows: [{ key: "image", to: "worker:1", kind: "added" }] }] },
    { name: "mailhog", change: "removed", sections: [], note: "Resource removed from this release — workload and config deleted from the stack." },
  ],
  volumes: [],
  connections: [],
};

const empty: SnapshotDiff = { resources: [], volumes: [], connections: [] };

describe("ConfigDiff", () => {
  it("renders changed, added and removed resources", () => {
    render(<ConfigDiff diff={diff} hasPrev prevSeq={12} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("web:1")).toHaveClass("line-through");
    expect(screen.getByText("web:2")).toHaveClass("text-success");
    expect(screen.getByText("ADDED")).toBeInTheDocument();
    expect(screen.getByText(/removed from this release/i)).toBeInTheDocument();
  });
  it("renders a volumes group with resized/added volumes", () => {
    render(<ConfigDiff diff={{ resources: [], volumes: [{ name: "data", change: "modified", rows: [{ key: "size", from: "1Gi", to: "2Gi", kind: "changed" }] }], connections: [] }} hasPrev prevSeq={12} />);
    expect(screen.getByText("Volumes")).toBeInTheDocument();
    expect(screen.getByText("data")).toBeInTheDocument();
    expect(screen.getByText("2Gi")).toHaveClass("text-success");
  });
  it("shows the initial-release copy when there is no predecessor", () => {
    render(<ConfigDiff diff={empty} hasPrev={false} />);
    expect(screen.getByText(/initial release/i)).toBeInTheDocument();
  });
  it("shows the no-changes copy when a predecessor exists but nothing changed", () => {
    render(<ConfigDiff diff={empty} hasPrev prevSeq={12} />);
    expect(screen.getByText(/no configuration changes since #12/i)).toBeInTheDocument();
  });
});

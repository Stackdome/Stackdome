// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { ConfigChangesToggle } from "../config-changes-toggle";
import type { SnapshotDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const diff: SnapshotDiff = {
  resources: [{ name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "nginx:1.25", to: "nginx:1.27", kind: "changed" }] }] }],
  volumes: [],
  connections: [],
};

describe("ConfigChangesToggle", () => {
  it("renders a brand-amber label against the previous sequence", () => {
    render(<ConfigChangesToggle diff={diff} prevSeq={7} />);
    const label = screen.getByText(/Config changes · vs #7/);
    expect(label).toHaveClass("text-brand");
  });

  it("is collapsed by default and reveals the diff on click", () => {
    render(<ConfigChangesToggle diff={diff} prevSeq={7} />);
    expect(screen.queryByText("nginx:1.27")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button"));
    expect(screen.getByText("nginx:1.27")).toBeInTheDocument();
  });
});

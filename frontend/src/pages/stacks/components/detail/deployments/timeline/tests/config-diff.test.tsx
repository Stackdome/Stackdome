// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ConfigDiff } from "../config-diff";
import type { ResourceDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const diffs: ResourceDiff[] = [
  { name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "web:1", to: "web:2", kind: "changed" }] }] },
  { name: "worker", change: "added", sections: [{ kind: "configuration", rows: [{ key: "image", to: "worker:1", kind: "added" }] }] },
  { name: "mailhog", change: "removed", sections: [], note: "Resource removed from this release — workload and config deleted from the stack." },
];

describe("ConfigDiff", () => {
  it("renders changed, added and removed resources", () => {
    render(<ConfigDiff diffs={diffs} prevSeq={12} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("web:1")).toHaveClass("line-through");
    expect(screen.getByText("web:2")).toHaveClass("text-success");
    expect(screen.getByText("ADDED")).toBeInTheDocument();
    expect(screen.getByText(/removed from this release/i)).toBeInTheDocument();
  });
  it("shows an empty note when there are no diffs", () => {
    render(<ConfigDiff diffs={[]} />);
    expect(screen.getByText(/nothing to compare/i)).toBeInTheDocument();
  });
});

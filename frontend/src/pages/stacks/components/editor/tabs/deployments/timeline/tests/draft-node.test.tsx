// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DraftNode } from "../draft-node";
import type { SnapshotDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const diff: SnapshotDiff = {
  resources: [{ name: "web-server", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "nginx:1.25", to: "nginx:1.27", kind: "changed" }] }] }],
  volumes: [],
  connections: [],
};

describe("DraftNode", () => {
  it("staged: shows the DRAFT chip and starts collapsed", () => {
    render(<DraftNode phase="staged" diff={diff} vsSeq={7} />);
    expect(screen.getByText("Draft")).toBeInTheDocument();
    expect(screen.getByText(/web-server changed/)).toBeInTheDocument();
    // Collapsed by default: the diff card body is not rendered.
    expect(screen.queryByText("web-server")).not.toBeInTheDocument();
  });

  it("editing: shows the UNSAVED chip", () => {
    render(<DraftNode phase="editing" diff={diff} vsSeq={7} />);
    expect(screen.getByText("Unsaved")).toBeInTheDocument();
  });

  it("expands the card when the row is clicked", async () => {
    render(<DraftNode phase="staged" diff={diff} vsSeq={7} />);
    expect(screen.queryByText("web-server")).not.toBeInTheDocument();
    await userEvent.click(screen.getByText("Staged changes"));
    expect(screen.getByText("web-server")).toBeInTheDocument();
  });

  it("falls back to a plain note when there is no diff to show", async () => {
    render(<DraftNode phase="staged" />);
    await userEvent.click(screen.getByText("Staged changes"));
    expect(screen.getByText(/staged for deploy/i)).toBeInTheDocument();
  });
});

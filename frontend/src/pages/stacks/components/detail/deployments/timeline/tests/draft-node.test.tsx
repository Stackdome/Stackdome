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
  it("staged: shows the DRAFT chip, the staged diff, and the deploy note (open by default)", () => {
    render(<DraftNode phase="staged" diff={diff} liveSeq={7} nextSeq={8} />);
    expect(screen.getByText("Draft")).toBeInTheDocument();
    expect(screen.getByText(/web-server changed/)).toBeInTheDocument();
    // The diff card renders the changed resource + values.
    expect(screen.getByText("web-server")).toBeInTheDocument();
    expect(screen.getByText(/ship as release #8/)).toBeInTheDocument();
  });

  it("editing: shows the UNSAVED chip", () => {
    render(<DraftNode phase="editing" diff={diff} liveSeq={7} nextSeq={8} />);
    expect(screen.getByText("Unsaved")).toBeInTheDocument();
  });

  it("collapses the card when the row is clicked", async () => {
    render(<DraftNode phase="staged" diff={diff} liveSeq={7} nextSeq={8} />);
    expect(screen.getByText(/ship as release #8/)).toBeInTheDocument();
    await userEvent.click(screen.getByText("Staged changes"));
    expect(screen.queryByText(/ship as release #8/)).not.toBeInTheDocument();
  });

  it("falls back to a plain note when there is no diff to show", () => {
    render(<DraftNode phase="staged" nextSeq={1} />);
    expect(screen.getByText(/staged for deploy/i)).toBeInTheDocument();
  });
});

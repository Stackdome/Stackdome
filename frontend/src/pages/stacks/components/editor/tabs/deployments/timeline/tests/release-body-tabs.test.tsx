// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReleaseBodyTabs } from "../release-body-tabs";
import type { SnapshotDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const emptyDiff: SnapshotDiff = { resources: [], volumes: [], connections: [] };
const diff: SnapshotDiff = {
  resources: [{ name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", kind: "changed", from: "web:1", to: "web:2" }] }] }],
  volumes: [],
  connections: [],
};

function renderTabs(props: Partial<React.ComponentProps<typeof ReleaseBodyTabs>> = {}) {
  return render(
    <ReleaseBodyTabs diff={diff} hasPrev prevSeq={12} {...props}>
      <div>outcomes-content</div>
    </ReleaseBodyTabs>,
  );
}

describe("ReleaseBodyTabs", () => {
  it("shows Outcomes by default with a change count on the Changes tab", () => {
    renderTabs();
    expect(screen.getByText("outcomes-content")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Changes 1" })).toBeInTheDocument();
    expect(screen.queryByText("Modified")).not.toBeInTheDocument();
  });

  it("switches to the config diff and back", async () => {
    renderTabs();
    await userEvent.click(screen.getByRole("button", { name: /Changes/ }));
    expect(screen.getByText("Modified")).toBeInTheDocument();
    expect(screen.getByText("vs #12")).toBeInTheDocument();
    expect(screen.queryByText("outcomes-content")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Outcomes" }));
    expect(screen.getByText("outcomes-content")).toBeInTheDocument();
  });

  it("shows the loading placeholder while the previous snapshot resolves", async () => {
    renderTabs({ loading: true });
    await userEvent.click(screen.getByRole("button", { name: /Changes/ }));
    expect(screen.getByText("Loading changes…")).toBeInTheDocument();
  });

  it("reads as initial release when there is no predecessor", async () => {
    renderTabs({ diff: emptyDiff, hasPrev: false, prevSeq: undefined });
    await userEvent.click(screen.getByRole("button", { name: "Changes" }));
    expect(screen.getByText("Initial release — nothing to compare.")).toBeInTheDocument();
    expect(screen.queryByText(/vs #/)).not.toBeInTheDocument();
  });
});

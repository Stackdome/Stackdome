// @vitest-environment jsdom
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";
import { AutosaveStatus } from "../autosave-status";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

describe("AutosaveStatus", () => {
  it("renders nothing when idle", () => {
    const { container } = render(<AutosaveStatus status={SYNC_STATUS.idle} />);
    expect(container).toBeEmptyDOMElement();
  });
  it("shows saving", () => {
    render(<AutosaveStatus status={SYNC_STATUS.saving} />);
    expect(screen.getByText("Saving…")).toBeInTheDocument();
  });
  it("shows saved", () => {
    render(<AutosaveStatus status={SYNC_STATUS.saved} />);
    expect(screen.getByText("All changes saved")).toBeInTheDocument();
  });
  it("shows error", () => {
    render(<AutosaveStatus status={SYNC_STATUS.error} />);
    expect(screen.getByText("Save failed, retrying")).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { LiveViewToggle } from "../live-view-toggle";

afterEach(cleanup);

describe("LiveViewToggle", () => {
  it("marks the active segment and reports switches", () => {
    const onModeChange = vi.fn();
    render(<LiveViewToggle mode="draft" onModeChange={onModeChange} draftDirty={false} />);

    expect(screen.getByRole("button", { name: "Draft" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Live" })).toHaveAttribute("aria-pressed", "false");

    fireEvent.click(screen.getByRole("button", { name: "Live" }));
    expect(onModeChange).toHaveBeenCalledWith("live");
  });

  it("dots the Draft segment when undeployed edits exist", () => {
    const { rerender } = render(<LiveViewToggle mode="live" onModeChange={vi.fn()} draftDirty />);
    expect(screen.getByTestId("draft-dirty-dot")).toBeInTheDocument();

    rerender(<LiveViewToggle mode="live" onModeChange={vi.fn()} draftDirty={false} />);
    expect(screen.queryByTestId("draft-dirty-dot")).not.toBeInTheDocument();
  });
});

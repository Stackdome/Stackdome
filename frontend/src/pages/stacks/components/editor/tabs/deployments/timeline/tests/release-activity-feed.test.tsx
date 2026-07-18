// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ReleaseActivityFeed } from "../release-activity-feed";
import type { ReleaseEvent } from "@/api/releases";

afterEach(cleanup);

const ev = (over: Partial<ReleaseEvent> = {}): ReleaseEvent => ({
  id: "e1", sequence: 1, occurred_at: "2026-07-11T10:00:00Z", type: "build_started", level: "info", message: "Building web", ...over,
});

describe("ReleaseActivityFeed", () => {
  it("renders nothing when there are no events and it's not streaming", () => {
    const { container } = render(<ReleaseActivityFeed events={[]} streaming={false} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the header with a live indicator while streaming, even with no events yet", () => {
    render(<ReleaseActivityFeed events={[]} streaming />);
    expect(screen.getByText("Activity")).toBeInTheDocument();
    expect(screen.getByText("live")).toBeInTheDocument();
  });

  it("does not show the live indicator once streaming stops", () => {
    render(<ReleaseActivityFeed events={[ev()]} streaming={false} />);
    expect(screen.queryByText("live")).not.toBeInTheDocument();
  });

  it("renders the type chip, resource tag, message and time for an event", () => {
    render(<ReleaseActivityFeed events={[ev({ type: "build_started", resource_name: "web", message: "Building web" })]} streaming={false} />);
    expect(screen.getByText("build_started")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("Building web")).toBeInTheDocument();
  });

  it("does not render a resource tag when resource_name is absent", () => {
    render(<ReleaseActivityFeed events={[ev({ resource_name: undefined, message: "Release started" })]} streaming={false} />);
    expect(screen.queryByText("web")).not.toBeInTheDocument();
  });

  it.each([
    ["success", "✓"],
    ["error", "✕"],
    ["warning", "!"],
    ["info", "•"],
  ] as const)("renders the %s level glyph", (level, glyph) => {
    render(<ReleaseActivityFeed events={[ev({ level })]} streaming={false} />);
    expect(screen.getByText(glyph)).toBeInTheDocument();
  });

  it("falls back to the info glyph when level is missing", () => {
    render(<ReleaseActivityFeed events={[ev({ level: undefined })]} streaming={false} />);
    expect(screen.getByText("•")).toBeInTheDocument();
  });

  it("renders links as labelled entries", () => {
    render(
      <ReleaseActivityFeed
        events={[ev({ links: [{ kind: "build_logs", label: "View build logs", target: { build_id: "b1", resource_name: "web" } }] })]}
        streaming={false}
      />,
    );
    expect(screen.getByText(/View build logs/)).toBeInTheDocument();
  });

  it("renders multiple events in the given order", () => {
    render(
      <ReleaseActivityFeed
        events={[ev({ sequence: 1, message: "first" }), ev({ id: "e2", sequence: 2, message: "second" })]}
        streaming={false}
      />,
    );
    const messages = screen.getAllByText(/first|second/).map((el) => el.textContent);
    expect(messages).toEqual(["first", "second"]);
  });
});

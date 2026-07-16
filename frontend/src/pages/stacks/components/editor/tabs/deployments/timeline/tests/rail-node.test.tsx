// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { RailNode } from "../rail-node";

afterEach(cleanup);

describe("RailNode", () => {
  it("renders content and a dot", () => {
    render(<RailNode tone="ok"><span>node body</span></RailNode>);
    expect(screen.getByText("node body")).toBeInTheDocument();
    expect(screen.getByTestId("rail-dot")).toHaveClass("bg-success");
  });
  it("hides the connector on the last node", () => {
    render(<RailNode tone="muted" isLast><span>x</span></RailNode>);
    expect(screen.getByTestId("rail-connector")).toHaveClass("invisible");
  });
});

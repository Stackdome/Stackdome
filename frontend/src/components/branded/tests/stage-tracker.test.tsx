// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { StageTracker } from "../stage-tracker";

afterEach(cleanup);

describe("StageTracker", () => {
  it("renders Build, Deploy, Ready with status marks", () => {
    render(<StageTracker stages={{ build: "done", deploy: "active", ready: "todo" }} />);
    expect(screen.getByText("Build")).toBeInTheDocument();
    expect(screen.getByText("Deploy")).toBeInTheDocument();
    expect(screen.getByText("Ready")).toBeInTheDocument();
    expect(screen.getByText("Deploy").closest("[data-status]")).toHaveAttribute("data-status", "active");
  });

  it("marks a failed stage", () => {
    render(<StageTracker stages={{ build: "failed", deploy: "todo", ready: "todo" }} />);
    expect(screen.getByText("Build").closest("[data-status]")).toHaveAttribute("data-status", "failed");
  });
});

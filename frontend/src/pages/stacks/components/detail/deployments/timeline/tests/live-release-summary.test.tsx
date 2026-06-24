// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { LiveReleaseSummary } from "../live-release-summary";
import type { Stack } from "@/api/stacks";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);

const release = { id: "r9", sequence: 9, state: "Released", cause: { kind: "manual" } } as StackRelease;

const healthyStack = {
  status: { resources: [{ name: "web" }, { name: "api" }] },
  spec: { stack_resources: [{ name: "web" }, { name: "api" }] },
} as unknown as Stack;

const unhealthyStack = {
  status: { resources: [{ name: "web" }, { name: "api" }] },
  spec: {
    stack_resources: [
      { name: "web", status: { state: "CrashLoopBackOff", last_failure: { type: "runtime_crash", container: { reason: "OOMKilled" } } } },
      { name: "api", status: { state: "Ready" } },
    ],
  },
} as unknown as Stack;

describe("LiveReleaseSummary", () => {
  it("identifies the live release and a healthy summary", () => {
    render(<LiveReleaseSummary release={release} stack={healthyStack} />);
    expect(screen.getByText("Live")).toBeInTheDocument();
    expect(screen.getByText("#9")).toBeInTheDocument();
    expect(screen.getByText("Manual deploy")).toBeInTheDocument();
    expect(screen.getByText(/2 resources healthy/)).toBeInTheDocument();
  });

  it("flags an unhealthy live release", () => {
    render(<LiveReleaseSummary release={release} stack={unhealthyStack} />);
    expect(screen.getByText(/1 of 2 unhealthy/)).toBeInTheDocument();
  });

  it("fires onJump and shows the jump affordance when provided", () => {
    const onJump = vi.fn();
    render(<LiveReleaseSummary release={release} stack={healthyStack} onJump={onJump} />);
    expect(screen.getByText("Jump")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button"));
    expect(onJump).toHaveBeenCalledTimes(1);
  });

  it("omits the jump affordance without onJump", () => {
    render(<LiveReleaseSummary release={release} stack={healthyStack} />);
    expect(screen.queryByText("Jump")).not.toBeInTheDocument();
  });
});

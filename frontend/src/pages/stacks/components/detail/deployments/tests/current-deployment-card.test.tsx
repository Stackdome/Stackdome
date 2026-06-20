// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { CurrentDeploymentCard } from "../current-deployment-card";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);

const release: StackRelease = { id: "r1", sequence: 14, state: "Released",
  pins: { resources: { tooljet: { git_sha: "9c69af2" } } }, snapshot_revision: "4f9c1a2" } as StackRelease;

const stack = { status: {
  last_converged: { release_id: "r1" },
  resources: [{ name: "tooljet", phase: "Ready", available_replicas: 1, replicas: 1 }],
}, spec: { stack_resources: [
  { name: "tooljet", status: { state: "Ready", last_failure: { type: "runtime_crash", container: { reason: "CrashLoopBackOff", restart_count: 5 } } } },
] } } as unknown as Stack;

describe("CurrentDeploymentCard", () => {
  it("shows the active release sequence and Ready stage tracker", () => {
    render(<CurrentDeploymentCard release={release} stack={stack} />);
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText("Ready").closest("[data-status]")).toHaveAttribute("data-status", "done");
  });

  it("renders the recovered note for a Ready resource with last_failure", () => {
    render(<CurrentDeploymentCard release={release} stack={stack} />);
    expect(screen.getByText(/recovered/i)).toBeInTheDocument();
    expect(screen.getByText(/CrashLoopBackOff/)).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, afterEach, beforeAll, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent, waitFor } from "@testing-library/react";
import { FailingResourcesAccordion } from "../failing-resources-accordion";
import type { FailingResource } from "../derive";

vi.mock("@/api/observability", () => ({
  fetchLogSnapshot: vi.fn().mockResolvedValue(["panic: boom", "exit status 1"]),
}));

// Radix needs these jsdom stubs (same pattern as release-row.test.tsx)
beforeAll(() => {
  Element.prototype.hasPointerCapture = () => false;
  Element.prototype.releasePointerCapture = () => {};
  Element.prototype.setPointerCapture = () => {};
  Element.prototype.scrollIntoView = () => {};
});
afterEach(cleanup);

const failing: FailingResource[] = [
  { name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", message: "exit 1", exitCode: 1, restartCount: 5 },
  { name: "worker", type: "runtime_crash", stage: "runtime", reason: "OOMKilled", restartCount: 2 },
];

describe("FailingResourcesAccordion", () => {
  it("lists each failing resource as a header", () => {
    render(<FailingResourcesAccordion failing={failing} />);
    expect(screen.getByText("tooljet")).toBeInTheDocument();
    expect(screen.getByText("worker")).toBeInTheDocument();
  });

  it("expands one resource to show its failure detail", () => {
    render(<FailingResourcesAccordion failing={failing} />);
    fireEvent.click(screen.getByText("tooljet"));
    expect(screen.getByText("exit 1")).toBeInTheDocument();
  });

  it("keeps only one item open at a time", () => {
    render(<FailingResourcesAccordion failing={failing} />);
    fireEvent.click(screen.getByText("tooljet"));
    expect(screen.getByText("exit 1")).toBeInTheDocument();
    fireEvent.click(screen.getByText("worker"));
    expect(screen.queryByText("exit 1")).toBeNull();
  });

  it("renders a release-level banner when releaseMessage is set and no per-resource failures", () => {
    render(<FailingResourcesAccordion failing={[]} releaseMessage="apply error: forbidden" />);
    expect(screen.getByText(/apply error: forbidden/)).toBeInTheDocument();
  });

  it("shows a log snapshot when the fetch returns lines", async () => {
    render(<FailingResourcesAccordion failing={failing} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} />);
    fireEvent.click(screen.getByText("tooljet"));
    await waitFor(() => expect(screen.getByText(/panic: boom/)).toBeInTheDocument());
  });
});

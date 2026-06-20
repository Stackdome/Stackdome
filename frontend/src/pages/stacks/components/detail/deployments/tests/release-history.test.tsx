// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ReleaseHistory } from "../release-history";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);

// Radix DropdownMenu touches pointer-capture + scroll APIs that jsdom lacks.
beforeAll(() => {
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false,
    releasePointerCapture: () => undefined,
    setPointerCapture: () => undefined,
    scrollIntoView: () => undefined,
  };
  for (const [name, impl] of Object.entries(stubs)) {
    const proto = Element.prototype as unknown as Record<string, unknown>;
    if (!proto[name]) proto[name] = vi.fn(impl);
  }
});

const noop = vi.fn();
const rel = (id: string, seq: number): StackRelease => ({ id, sequence: seq, state: "Released", cause: { kind: "manual" } } as StackRelease);

describe("ReleaseHistory", () => {
  it("renders one row per release", () => {
    render(<ReleaseHistory releases={[rel("r2", 2), rel("r1", 1)]} onViewDetails={noop} onRollback={noop} onCancel={noop} />);
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("renders an empty state with no releases", () => {
    render(<ReleaseHistory releases={[]} onViewDetails={noop} onRollback={noop} onCancel={noop} />);
    expect(screen.getByText(/No deployments yet/i)).toBeInTheDocument();
  });
});

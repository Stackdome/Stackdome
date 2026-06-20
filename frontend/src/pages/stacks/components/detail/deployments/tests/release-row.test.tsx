// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReleaseRow } from "../release-row";
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

function row(partial: Partial<StackRelease>): StackRelease {
  return { id: "r1", sequence: 14, state: "Released", cause: { kind: "manual" },
    rendered_at: "2026-06-21T12:00:00Z", completed_at: "2026-06-21T12:00:32Z",
    pins: { resources: { api: { git_sha: "9c69af2" } } }, ...partial } as StackRelease;
}

describe("ReleaseRow", () => {
  it("renders sequence, cause and a Released pill", () => {
    render(<ReleaseRow release={row({})} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText(/Manual deploy/)).toBeInTheDocument();
    expect(screen.getByText(/9c69af2/)).toBeInTheDocument();
  });

  it("offers Rollback for a Released release", async () => {
    const user = userEvent.setup();
    const onRollback = vi.fn();
    render(<ReleaseRow release={row({})} onViewDetails={vi.fn()} onRollback={onRollback} onCancel={vi.fn()} />);
    await user.click(screen.getByLabelText("Release actions"));
    await user.click(screen.getByText("Rollback to this"));
    expect(onRollback).toHaveBeenCalledWith("r1");
  });

  it("offers Cancel for a Pending release", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(<ReleaseRow release={row({ state: "Pending" })} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={onCancel} />);
    await user.click(screen.getByLabelText("Release actions"));
    await user.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalledWith("r1");
  });

  it("shows the failure message on a Failed row", () => {
    render(<ReleaseRow release={row({ state: "Failed", message: "render error: bad template" })} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText(/render error: bad template/)).toBeInTheDocument();
  });

  it("does not offer Cancel for a Released release", async () => {
    const user = userEvent.setup();
    render(<ReleaseRow release={row({})} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    await user.click(screen.getByLabelText("Release actions"));
    expect(screen.queryByText("Cancel")).toBeNull();
  });

  it("does not offer Rollback for a Pending release", async () => {
    const user = userEvent.setup();
    render(<ReleaseRow release={row({ state: "Pending" })} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    await user.click(screen.getByLabelText("Release actions"));
    expect(screen.queryByText("Rollback to this")).toBeNull();
  });

  it("View details calls onViewDetails with the id", async () => {
    const user = userEvent.setup();
    const onViewDetails = vi.fn();
    render(<ReleaseRow release={row({})} onViewDetails={onViewDetails} onRollback={vi.fn()} onCancel={vi.fn()} />);
    await user.click(screen.getByLabelText("Release actions"));
    await user.click(screen.getByText("View details"));
    expect(onViewDetails).toHaveBeenCalledWith("r1");
  });
});

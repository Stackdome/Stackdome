// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReleaseMenu } from "../release-menu";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined,
  };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const rel = (over: Partial<StackRelease>) => ({ id: "r1", sequence: 5, state: "Released", ...over } as StackRelease);

describe("ReleaseMenu", () => {
  it("shows Rollback for Released and Copy id; hides Cancel", async () => {
    const onRollback = vi.fn();
    render(<ReleaseMenu release={rel({ state: "Released" })} onRollback={onRollback} onCancel={vi.fn()} onCopyId={vi.fn()} />);
    await userEvent.click(screen.getByRole("button", { name: /release actions/i }));
    expect(screen.getByText("Rollback to this")).toBeInTheDocument();
    expect(screen.queryByText("Cancel release")).not.toBeInTheDocument();
    await userEvent.click(screen.getByText("Rollback to this"));
    expect(onRollback).toHaveBeenCalledWith("r1");
  });

  it("shows Cancel for Pending", async () => {
    render(<ReleaseMenu release={rel({ state: "Pending" })} onRollback={vi.fn()} onCancel={vi.fn()} onCopyId={vi.fn()} />);
    await userEvent.click(screen.getByRole("button", { name: /release actions/i }));
    expect(screen.getByText("Cancel release")).toBeInTheDocument();
    expect(screen.queryByText("Rollback to this")).not.toBeInTheDocument();
  });
});

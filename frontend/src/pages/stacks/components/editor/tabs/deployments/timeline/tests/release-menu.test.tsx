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
  it("shows Cancel for Pending", async () => {
    const onCancel = vi.fn();
    render(<ReleaseMenu release={rel({ state: "Pending" })} onCancel={onCancel} />);
    await userEvent.click(screen.getByRole("button", { name: /release actions/i }));
    await userEvent.click(screen.getByText("Cancel release"));
    expect(onCancel).toHaveBeenCalledWith("r1");
  });

  it("renders nothing for Released — rollback is a button on the detail card", () => {
    render(<ReleaseMenu release={rel({ state: "Released" })} onCancel={vi.fn()} />);
    expect(screen.queryByRole("button", { name: /release actions/i })).not.toBeInTheDocument();
  });

  it("renders nothing once InProgress (the backend rejects cancelling a rollout)", () => {
    render(<ReleaseMenu release={rel({ state: "InProgress" })} onCancel={vi.fn()} />);
    expect(screen.queryByRole("button", { name: /release actions/i })).not.toBeInTheDocument();
  });
});

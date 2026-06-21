// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor, fireEvent } from "@testing-library/react";

vi.mock("@/api/releases", () => ({
  listReleases: vi.fn().mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "Released", cause: { kind: "manual" } }], total: 1 }),
  getRelease: vi.fn(), createRelease: vi.fn().mockResolvedValue({}), rollbackRelease: vi.fn(), cancelRelease: vi.fn(),
}));

vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: vi.fn() }),
}));
import { DeploymentsTab } from "../deployments-tab";
import type { Stack } from "@/api/stacks";

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

const stack = { id: "s1", status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;

describe("DeploymentsTab", () => {
  it("loads releases and renders current deployment + history", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s1" stack={stack} canDeploy />);
    await waitFor(() => expect(screen.getAllByText("#1").length).toBeGreaterThan(0));
    expect(screen.getByText("Current deployment")).toBeInTheDocument();
    expect(screen.getByText("History")).toBeInTheDocument();
  });

  it("deploys and refetches when the Deploy button is clicked", async () => {
    const { listReleases, createRelease } = await import("@/api/releases");
    (createRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r2", sequence: 2, state: "Pending" });
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s1" stack={stack} canDeploy />);
    await waitFor(() => expect(screen.getAllByText("#1").length).toBeGreaterThan(0));
    const callsBefore = (listReleases as ReturnType<typeof vi.fn>).mock.calls.length;
    fireEvent.click(screen.getByRole("button", { name: /^Deploy$/i }));
    await waitFor(() => expect(createRelease).toHaveBeenCalledWith("o", "t", "s1"));
    await waitFor(() => expect((listReleases as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(callsBefore));
  });
});

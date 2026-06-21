// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/components/ui/use-toast", () => ({ useToast: () => ({ toast: vi.fn() }) }));
vi.mock("@/api/releases", () => ({
  listReleases: vi.fn().mockResolvedValue({ items: [{ id: "r1", sequence: 14, state: "Released", cause: { kind: "manual" } }], total: 1 }),
  getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, outcome: { resources: {} }, snapshot: { resources: [] } }),
  createRelease: vi.fn().mockResolvedValue({}), rollbackRelease: vi.fn(), cancelRelease: vi.fn(),
}));
import { listReleases, createRelease } from "@/api/releases";
import { DeploymentsTab } from "../deployments-tab";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;

describe("DeploymentsTab", () => {
  it("deploy button creates a release and refetches", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s" stack={stack} canDeploy />);
    await waitFor(() => expect(screen.getByText("#14")).toBeInTheDocument());
    const before = (listReleases as ReturnType<typeof vi.fn>).mock.calls.length;
    await userEvent.click(screen.getByRole("button", { name: /^deploy$/i }));
    expect(createRelease).toHaveBeenCalledWith("o", "t", "s");
    await waitFor(() => expect((listReleases as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(before));
  });

  it("hides the Deploy button when canDeploy is false", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s" stack={stack} canDeploy={false} />);
    await waitFor(() => expect(screen.getByText("#14")).toBeInTheDocument());
    expect(screen.queryByRole("button", { name: /^deploy$/i })).not.toBeInTheDocument();
  });
});

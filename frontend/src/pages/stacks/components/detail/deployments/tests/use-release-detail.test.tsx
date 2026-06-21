// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { useReleaseDetail } from "../use-release-detail";

afterEach(() => { cleanup(); vi.clearAllMocks(); });

function Harness({ ids }: { ids: string[] }) {
  const detail = useReleaseDetail("o", "t", "s");
  ids.forEach((id) => detail.ensure(id));
  return <div>{ids.map((id) => <span key={id} data-id={id}>{detail.peek(id).data?.sequence ?? "—"}</span>)}</div>;
}

describe("useReleaseDetail", () => {
  it("fetches once per id and caches", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 12 });
    render(<Harness ids={["r1"]} />);
    await waitFor(() => expect(screen.getByText("12")).toBeInTheDocument());
    expect(getRelease).toHaveBeenCalledTimes(1);
    expect(getRelease).toHaveBeenCalledWith("o", "t", "s", "r1");
  });

  it("reuses a fetched release as another row's previous (no double fetch)", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation((_o, _t, _s, id) => Promise.resolve({ id, sequence: id === "r1" ? 12 : 11 }));
    const { rerender } = render(<Harness ids={["r1"]} />);
    await waitFor(() => expect(getRelease).toHaveBeenCalledTimes(1));
    rerender(<Harness ids={["r1", "r2"]} />);
    await waitFor(() => expect(getRelease).toHaveBeenCalledTimes(2)); // only r2 added
    rerender(<Harness ids={["r1", "r2"]} />);
    expect(getRelease).toHaveBeenCalledTimes(2); // r1/r2 already cached
  });

  it("captures error", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("boom"));
    function E() { const d = useReleaseDetail("o", "t", "s"); d.ensure("rx"); return <span>{d.peek("rx").error ?? ""}</span>; }
    render(<E />);
    await waitFor(() => expect(screen.getByText("boom")).toBeInTheDocument());
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { useEffect } from "react";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor, fireEvent } from "@testing-library/react";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { useReleaseDetail, ReleaseDetailProvider, useReleaseDetailContext } from "../use-release-detail";

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

  it("refresh refetches a cached id, keeping old data while loading", async () => {
    let seq = 1;
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation(() => Promise.resolve({ id: "r1", sequence: seq++ }));
    function R() {
      const d = useReleaseDetail("o", "t", "s");
      d.ensure("r1");
      const s = d.peek("r1");
      return (
        <button onClick={() => d.refresh("r1")}>
          {s.data?.sequence ?? "—"}{s.loading ? "*" : ""}
        </button>
      );
    }
    render(<R />);
    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("1"));
    fireEvent.click(screen.getByRole("button"));
    // stale data stays visible while the refetch is in flight
    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("1*"));
    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("2"));
    expect(getRelease).toHaveBeenCalledTimes(2);
  });

  it("keeps ensure/refresh identities stable across cache updates (effect-dep safety)", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 1 });
    const seen: Array<{ ensure: unknown; refresh: unknown }> = [];
    function S() {
      const d = useReleaseDetail("o", "t", "s");
      seen.push({ ensure: d.ensure, refresh: d.refresh });
      d.ensure("r1");
      return <span>{d.peek("r1").data?.sequence ?? "—"}</span>;
    }
    render(<S />);
    await waitFor(() => expect(screen.getByText("1")).toBeInTheDocument());
    expect(seen.length).toBeGreaterThan(1); // cache updates re-rendered
    expect(new Set(seen.map((s) => s.ensure)).size).toBe(1);
    expect(new Set(seen.map((s) => s.refresh)).size).toBe(1);
  });

  it("refresh dedupes while a fetch for the same id is in flight", async () => {
    let resolve!: (v: unknown) => void;
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation(() => new Promise((r) => { resolve = r; }));
    function R() { const d = useReleaseDetail("o", "t", "s"); d.ensure("r1"); d.refresh("r1"); return <span>{d.peek("r1").data?.sequence ?? "—"}</span>; }
    render(<R />);
    expect(getRelease).toHaveBeenCalledTimes(1);
    resolve({ id: "r1", sequence: 5 });
    await waitFor(() => expect(screen.getByText("5")).toBeInTheDocument());
  });

  it("no-ops ensure/refresh until teamName resolves, then fetches once teamName lands", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 1 });
    function T({ teamName }: { teamName: string }) {
      const d = useReleaseDetail("o", teamName, "s");
      // Depend on the stable ensure/refresh identities, not the whole `d` object —
      // peek's identity changes with every cache update, which would re-fire this
      // effect (and refetch) on every render.
      useEffect(() => { d.ensure("r1"); d.refresh("r1"); }, [d.ensure, d.refresh]);
      return <span>{d.peek("r1").data?.sequence ?? "—"}</span>;
    }
    const { rerender } = render(<T teamName="" />);
    expect(getRelease).not.toHaveBeenCalled();

    rerender(<T teamName="t" />);
    await waitFor(() => expect(screen.getByText("1")).toBeInTheDocument());
    expect(getRelease).toHaveBeenCalledTimes(1);
  });

  it("shares one cache across consumers via ReleaseDetailProvider", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 7 });

    function Provider({ children }: { children: React.ReactNode }) {
      const detail = useReleaseDetail("o", "t", "s");
      return <ReleaseDetailProvider value={detail}>{children}</ReleaseDetailProvider>;
    }
    function Consumer() {
      const detail = useReleaseDetailContext();
      useEffect(() => detail.ensure("r1"), [detail]);
      return <span>{detail.peek("r1").data?.sequence ?? "—"}</span>;
    }

    render(
      <Provider>
        <Consumer />
        <Consumer />
      </Provider>,
    );
    await waitFor(() => expect(screen.getAllByText("7")).toHaveLength(2));
    expect(getRelease).toHaveBeenCalledTimes(1); // one shared fetch, not one per consumer
  });
});

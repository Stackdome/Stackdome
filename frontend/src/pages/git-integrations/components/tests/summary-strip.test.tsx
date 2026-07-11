// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { SummaryStrip } from "../summary-strip";
import type { RowViewModel } from "../../lib/derive-row";

afterEach(cleanup);

function row(overrides: Partial<RowViewModel> = {}): RowViewModel {
  return {
    host: "github.com",
    authLabel: "GitHub App",
    statusKey: "connected",
    statusLabel: "Connected",
    tone: "ok",
    ...overrides,
  };
}

describe("SummaryStrip", () => {
  it("counts connected rows and attention rows separately", () => {
    const rows: RowViewModel[] = [
      row({ statusKey: "connected", tone: "ok" }),
      row({ statusKey: "connected", tone: "ok" }),
      row({ statusKey: "needs_setup", tone: "attention" }),
    ];
    render(<SummaryStrip rows={rows} />);
    expect(screen.getByText("Connected & ready")).toBeInTheDocument();
    expect(screen.getByText("Needs attention")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });
});

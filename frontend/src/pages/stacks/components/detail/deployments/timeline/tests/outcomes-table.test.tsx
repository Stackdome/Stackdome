// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { OutcomesTable } from "../outcomes-table";

afterEach(cleanup);

describe("OutcomesTable", () => {
  it("renders a row per resource with phase, replicas, message", () => {
    render(<OutcomesTable outcomes={{
      web: { phase: "Ready", ready_replicas: 1, replicas: 1, message: "" },
      worker: { phase: "CrashLoopBackOff", ready_replicas: 0, replicas: 1, message: "exit 1" },
    }} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("1/1")).toBeInTheDocument();
    expect(screen.getByText("CrashLoopBackOff")).toHaveClass("text-danger");
    expect(screen.getByText("exit 1")).toBeInTheDocument();
  });
});

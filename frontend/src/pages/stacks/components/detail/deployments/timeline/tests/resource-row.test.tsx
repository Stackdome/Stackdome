// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
import { ResourceRow } from "../resource-row";

afterEach(cleanup);

describe("ResourceRow", () => {
  it("renders a healthy row without an expander", () => {
    render(<ResourceRow vm={{ name: "redis", phase: "Ready", replicas: "1/1" }} />);
    expect(screen.getByText("redis")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("expands a failing row to show the FailureCard and Open in Logs", async () => {
    const onOpenLogs = vi.fn();
    render(<ResourceRow
      vm={{ name: "web", phase: "CrashLoopBackOff", replicas: "0/1", failure: { name: "web", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", message: "exit 1", restartCount: 7 } }}
      logContext={{ orgId: "o", teamName: "t", stackId: "s" }}
      onOpenLogs={onOpenLogs}
    />);
    await userEvent.click(screen.getByText("web"));
    expect(screen.getByText("exit 1")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /open in logs/i }));
    expect(onOpenLogs).toHaveBeenCalledWith("web");
  });
});

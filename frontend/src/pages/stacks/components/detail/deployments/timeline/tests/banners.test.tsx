// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DriftBanner, ReleaseErrorBanner } from "../banners";

afterEach(cleanup);

describe("banners", () => {
  it("drift banner deploys", async () => {
    const onDeploy = vi.fn();
    render(<DriftBanner onDeploy={onDeploy} busy={false} />);
    await userEvent.click(screen.getByRole("button", { name: /deploy/i }));
    expect(onDeploy).toHaveBeenCalled();
  });
  it("drift Deploy is disabled while busy", () => {
    render(<DriftBanner onDeploy={vi.fn()} busy />);
    expect(screen.getByRole("button", { name: /deploying/i })).toBeDisabled();
  });
  it("release error banner shows text + View details", async () => {
    const onView = vi.fn();
    render(<ReleaseErrorBanner lead="Deploy #16 failed" text="1 of 3 failing" onView={onView} />);
    expect(screen.getByText(/1 of 3 failing/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /view details/i }));
    expect(onView).toHaveBeenCalled();
  });
});

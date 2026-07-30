// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { DeployFailedBanner } from "../deploy-failed-banner";

afterEach(cleanup);

describe("DeployFailedBanner", () => {
  it("renders the failure heading and message", () => {
    render(<DeployFailedBanner message="timed out waiting for convergence after 15m0s" />);
    expect(screen.getByText("Deploy failed")).toBeInTheDocument();
    expect(screen.getByText(/timed out waiting for convergence/)).toBeInTheDocument();
  });
});

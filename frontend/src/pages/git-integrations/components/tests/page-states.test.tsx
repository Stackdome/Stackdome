// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { IntegrationsErrorState, IntegrationsEmptyState } from "../page-states";

afterEach(cleanup);

describe("IntegrationsErrorState", () => {
  it("renders the error message and fires onRetry", () => {
    const onRetry = vi.fn();
    render(<IntegrationsErrorState message="request failed with status 500" onRetry={onRetry} />);

    expect(screen.getByText("Couldn't load integrations")).toBeInTheDocument();
    expect(screen.getByText(/request failed with status 500/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /retry/i }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

describe("IntegrationsEmptyState", () => {
  it("renders the empty state copy and fires onAdd", () => {
    const onAdd = vi.fn();
    render(<IntegrationsEmptyState onAdd={onAdd} />);

    expect(screen.getByText("No git integrations yet")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /connect a provider/i }));
    expect(onAdd).toHaveBeenCalledTimes(1);
  });
});

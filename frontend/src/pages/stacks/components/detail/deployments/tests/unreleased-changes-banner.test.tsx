// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { UnreleasedChangesBanner } from "../unreleased-changes-banner";

afterEach(cleanup);

describe("UnreleasedChangesBanner", () => {
  it("renders nothing when there is no drift", () => {
    const { container } = render(<UnreleasedChangesBanner hasDrift={false} onDeploy={vi.fn()} busy={false} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the deploy affordance when drift is present", () => {
    const onDeploy = vi.fn();
    render(<UnreleasedChangesBanner hasDrift onDeploy={onDeploy} busy={false} />);
    expect(screen.getByText(/Unreleased changes/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Deploy/i }));
    expect(onDeploy).toHaveBeenCalled();
  });
});

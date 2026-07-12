// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/stack-create-wizard.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StackCreateWizard } from "../stack-create-wizard";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: vi.fn(), toasts: [], dismiss: vi.fn() }),
  toast: vi.fn(),
}));

beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn();
});
afterEach(() => {
  cleanup();
  navigate.mockReset();
});

describe("StackCreateWizard", () => {
  it("opens on the chooser and advances to the composer", async () => {
    const user = userEvent.setup();
    render(<StackCreateWizard open onOpenChange={vi.fn()} />);
    expect(screen.getByText(/How do you want to start\?/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Build from blocks/i }));
    expect(screen.getByText(/What's in your stack\?/i)).toBeInTheDocument();
  });

  it("blank slate navigates to the draft canvas", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    render(<StackCreateWizard open onOpenChange={onOpenChange} />);
    await user.click(screen.getByRole("button", { name: /blank slate/i }));
    expect(navigate).toHaveBeenCalledWith("/stacks/new");
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("template phase has back affordance to return to chooser", async () => {
    const user = userEvent.setup();
    render(<StackCreateWizard open onOpenChange={vi.fn()} />);
    expect(screen.getByText(/How do you want to start\?/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /From template/i }));
    expect(screen.getByText(/Self-hosted apps, ready to deploy/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /back/i }));
    expect(screen.getByText(/How do you want to start\?/i)).toBeInTheDocument();
  });
});

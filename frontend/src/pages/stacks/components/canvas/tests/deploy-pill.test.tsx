// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { DeployPill } from "../DeployPill";

afterEach(cleanup);

const base = {
  hasResources: true,
  dirtyTotal: 0,
  isStaged: false,
  isActive: true,
  deployBusy: false,
  canWrite: true,
  onDeploy: () => {},
  onViewChanges: () => {},
  canDiscardDraft: false,
};

describe("DeployPill visibility", () => {
  it("hides on an existing stack with no changes and no deploy in flight", () => {
    render(<DeployPill {...base} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });

  it("shows on an existing stack with changes", () => {
    render(<DeployPill {...base} dirtyTotal={3} />);
    expect(screen.getByTestId("deploy-pill")).toBeInTheDocument();
    expect(screen.getByText("3 changes")).toBeInTheDocument();
  });

  it("stays visible while a deploy runs even when the count drops to zero", () => {
    render(<DeployPill {...base} deployBusy />);
    expect(screen.getByTestId("deploy-pill")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /deploying/i })).toBeDisabled();
  });

  it("hides on an empty draft, shows once a resource exists", () => {
    const { rerender } = render(<DeployPill {...base} isDraft hasResources={false} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
    rerender(<DeployPill {...base} isDraft hasResources />);
    expect(screen.getByTestId("deploy-pill")).toBeInTheDocument();
  });

  it("draft pill has no count label, no Details, no menu", () => {
    render(<DeployPill {...base} isDraft hasResources dirtyTotal={2} canDiscardDraft onDiscardDraft={() => {}} />);
    expect(screen.queryByText(/change/)).toBeNull();
    expect(screen.queryByRole("button", { name: "Details" })).toBeNull();
    expect(screen.queryByLabelText("Change actions")).toBeNull();
  });
});

describe("DeployPill actions", () => {
  it("Details opens the view-changes modal", () => {
    const onViewChanges = vi.fn();
    render(<DeployPill {...base} dirtyTotal={2} onViewChanges={onViewChanges} />);
    fireEvent.click(screen.getByRole("button", { name: "Details" }));
    expect(onViewChanges).toHaveBeenCalled();
  });

  it("Deploy fires onDeploy for existing stacks, onDraftDeploy for drafts", () => {
    const onDeploy = vi.fn();
    const onDraftDeploy = vi.fn();
    const { rerender } = render(<DeployPill {...base} dirtyTotal={1} onDeploy={onDeploy} />);
    fireEvent.click(screen.getByRole("button", { name: /deploy/i }));
    expect(onDeploy).toHaveBeenCalledTimes(1);
    rerender(<DeployPill {...base} isDraft hasResources onDraftDeploy={onDraftDeploy} />);
    fireEvent.click(screen.getByRole("button", { name: /deploy/i }));
    expect(onDraftDeploy).toHaveBeenCalledTimes(1);
  });

  it("Deploy disabled without write access", () => {
    render(<DeployPill {...base} dirtyTotal={1} canWrite={false} />);
    expect(screen.getByRole("button", { name: /deploy/i })).toBeDisabled();
  });

  it("menu trigger renders only when discard applies", () => {
    const { rerender } = render(<DeployPill {...base} dirtyTotal={1} />);
    expect(screen.queryByLabelText("Change actions")).toBeNull();
    rerender(<DeployPill {...base} dirtyTotal={1} canDiscardDraft onDiscardDraft={() => {}} />);
    expect(screen.getByLabelText("Change actions")).toBeInTheDocument();
  });
});

describe("DeployPill keyboard", () => {
  it("Cmd+Enter deploys when enabled", () => {
    const onDeploy = vi.fn();
    render(<DeployPill {...base} dirtyTotal={1} onDeploy={onDeploy} />);
    fireEvent.keyDown(window, { key: "Enter", metaKey: true });
    expect(onDeploy).toHaveBeenCalledTimes(1);
  });

  it("Ctrl+Enter also deploys", () => {
    const onDeploy = vi.fn();
    render(<DeployPill {...base} dirtyTotal={1} onDeploy={onDeploy} />);
    fireEvent.keyDown(window, { key: "Enter", ctrlKey: true });
    expect(onDeploy).toHaveBeenCalledTimes(1);
  });

  it("shortcut inert when disabled, hidden, or already consumed", () => {
    const onDeploy = vi.fn();
    const { rerender } = render(<DeployPill {...base} dirtyTotal={1} canWrite={false} onDeploy={onDeploy} />);
    fireEvent.keyDown(window, { key: "Enter", metaKey: true });
    rerender(<DeployPill {...base} onDeploy={onDeploy} />); // hidden (no changes)
    fireEvent.keyDown(window, { key: "Enter", metaKey: true });
    rerender(<DeployPill {...base} dirtyTotal={1} onDeploy={onDeploy} />);
    const ev = new KeyboardEvent("keydown", { key: "Enter", metaKey: true, cancelable: true });
    ev.preventDefault();
    window.dispatchEvent(ev);
    expect(onDeploy).not.toHaveBeenCalled();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { CanvasEditorShell } from "../CanvasEditorShell";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

afterEach(cleanup);

const base = {
  statusState: null, subtitle: "0 services · 0 volumes",
  activeTab: "configuration", onTabChange: () => {},
  isActive: true, dirtyResourceCount: 0, dirtyTotal: 0, isStaged: false,
  onViewChanges: () => {},
  syncStatus: SYNC_STATUS.idle,
  deployBusy: false, canWrite: true,
  onCreate: () => {}, isCreating: false,
  onDeploy: () => {}, onDelete: () => {},
  canDiscardDraft: false, canDeleteStack: true,
  configuration: <div />, deployments: <div />, logs: <div />, metrics: <div />,
};

describe("CanvasEditorShell header", () => {
  it("renders an editable name input in draft and reports changes", () => {
    const onNameChange = vi.fn();
    render(<CanvasEditorShell {...base} stackName="" isDraft nameEditable onNameChange={onNameChange} />);
    const input = screen.getByPlaceholderText("name-your-stack");
    fireEvent.change(input, { target: { value: "web" } });
    expect(onNameChange).toHaveBeenCalledWith("web");
  });

  it("renders the name as static text when not editable", () => {
    render(<CanvasEditorShell {...base} stackName="tooljet" nameEditable={false} />);
    expect(screen.getByRole("heading", { name: "tooljet" })).toBeInTheDocument();
    expect(screen.queryByPlaceholderText("name-your-stack")).toBeNull();
  });

  it("renders a single status pill and never a DRAFT pill", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" statusState="Ready" isStaged />);
    expect(screen.getByText("Ready")).toBeInTheDocument();
    expect(screen.queryByText("DRAFT")).toBeNull();
  });
});

describe("CanvasEditorShell primary button matrix", () => {
  it("draft mode shows 'Create stack' and no Deploy / View changes / actions menu", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" isStaged dirtyTotal={3} />);
    expect(screen.getByRole("button", { name: /create stack/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Deploy" })).toBeNull();
    expect(screen.queryByRole("button", { name: /view changes/i })).toBeNull();
    expect(screen.queryByLabelText("Stack actions")).toBeNull();
  });

  it("draft mode shows 'Creating' while creating", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" isCreating />);
    expect(screen.getByRole("button", { name: /creating/i })).toBeInTheDocument();
  });

  it("existing stack with isStaged shows enabled Deploy", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged />);
    const btn = screen.getByRole("button", { name: "Deploy" });
    expect(btn).toBeInTheDocument();
    expect(btn).not.toBeDisabled();
  });

  it("existing stack clean (no staged, no unsaved) shows disabled Deploy and no View changes", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged={false} dirtyTotal={0} />);
    expect(screen.getByRole("button", { name: "Deploy" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: /view changes/i })).toBeNull();
  });

  it("existing stack with hasUnsaved shows enabled Deploy", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={2} isStaged={false} />);
    expect(screen.getByRole("button", { name: "Deploy" })).not.toBeDisabled();
  });

  it("syncStatus saving shows 'Saving…' for existing stacks", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" syncStatus={SYNC_STATUS.saving} />);
    expect(screen.getByText("Saving…")).toBeInTheDocument();
  });
});

describe("CanvasEditorShell View changes entry", () => {
  it("shows 'View changes' with the count and calls onViewChanges when there are changes", () => {
    const onViewChanges = vi.fn();
    render(
      <CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={3} onViewChanges={onViewChanges} />,
    );
    const btn = screen.getByRole("button", { name: /view changes/i });
    expect(btn).toHaveTextContent("3");
    fireEvent.click(btn);
    expect(onViewChanges).toHaveBeenCalled();
  });

  it("shows 'View changes' when staged even with no session dirt is driven by dirtyTotal", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged dirtyTotal={1} />);
    expect(screen.getByRole("button", { name: /view changes/i })).toHaveTextContent("1");
  });

  it("hides 'View changes' when staged but the change count is zero", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged dirtyTotal={0} />);
    expect(screen.queryByRole("button", { name: /view changes/i })).toBeNull();
  });
});

describe("CanvasEditorShell actions menu", () => {
  it("exposes the actions trigger for existing stacks with no 'Discard all changes' item", () => {
    // Radix dropdown content mounts on pointer interaction (not in jsdom), so we
    // assert the trigger exists and that the removed item never renders eagerly.
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={2} />);
    expect(screen.getByLabelText("Stack actions")).toBeInTheDocument();
    expect(screen.queryByText("Discard all changes")).toBeNull();
  });
});

describe("CanvasEditorShell collapse", () => {
  afterEach(() => localStorage.clear());

  it("collapses to a compact bar and hides the subtitle", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    expect(screen.getByText("0 services · 0 volumes")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    expect(screen.queryByText("0 services · 0 volumes")).toBeNull();
    expect(screen.getByText("acme")).toBeInTheDocument(); // compact bar name
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("persists collapsed state per stack id", () => {
    localStorage.setItem("stackdome.editor-header-collapsed.s1", "1");
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("toggles via Cmd+.", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    fireEvent.keyDown(window, { key: ".", metaKey: true });
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("keeps tabs clickable while collapsed", () => {
    const onTabChange = vi.fn();
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" onTabChange={onTabChange} />);
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    fireEvent.click(screen.getByRole("button", { name: /Logs/ }));
    expect(onTabChange).toHaveBeenCalledWith("logs");
  });
});

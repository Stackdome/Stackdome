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
  syncStatus: SYNC_STATUS.idle,
  deployBusy: false, canWrite: true,
  onCreate: () => {}, isCreating: false,
  onDeploy: () => {}, onDiscardAll: () => {}, onDelete: () => {},
  canDiscardDraft: false, canDeleteStack: true,
  configuration: <div />, deployments: <div />, logs: <div />, metrics: <div />,
  labels: [], labelsEditable: true,
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

  it("adds and removes labels", () => {
    const onAddLabel = vi.fn();
    const onRemoveLabel = vi.fn();
    render(
      <CanvasEditorShell {...base} stackName="web" nameEditable
        labels={[{ key: "k", value: "prod" }]} onAddLabel={onAddLabel} onRemoveLabel={onRemoveLabel} />,
    );
    fireEvent.click(screen.getByLabelText("Remove label prod"));
    expect(onRemoveLabel).toHaveBeenCalledWith(0);
    const labelInput = screen.getByPlaceholderText("add label…");
    fireEvent.change(labelInput, { target: { value: "dev" } });
    fireEvent.keyDown(labelInput, { key: "Enter" });
    expect(onAddLabel).toHaveBeenCalledWith("dev");
  });
});

describe("CanvasEditorShell primary button matrix", () => {
  it("draft mode shows 'Create stack' button", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" />);
    expect(screen.getByRole("button", { name: /create stack/i })).toBeInTheDocument();
    // Primary "Deploy" button absent in draft mode (Deployments tab still present)
    expect(screen.queryByRole("button", { name: "Deploy" })).toBeNull();
  });

  it("draft mode shows 'Creating' while creating", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" isCreating />);
    expect(screen.getByRole("button", { name: /creating/i })).toBeInTheDocument();
  });

  it("existing stack with isStaged shows enabled Deploy", () => {
    render(
      <CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged />,
    );
    const btn = screen.getByRole("button", { name: "Deploy" });
    expect(btn).toBeInTheDocument();
    expect(btn).not.toBeDisabled();
  });

  it("existing stack clean (no staged, no unsaved) shows disabled Deploy", () => {
    render(
      <CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged={false} dirtyTotal={0} />,
    );
    const btn = screen.getByRole("button", { name: "Deploy" });
    expect(btn).toBeDisabled();
  });

  it("existing stack with hasUnsaved shows enabled Deploy", () => {
    render(
      <CanvasEditorShell
        {...base}
        nameEditable={false}
        stackName="api"
        isActive
        dirtyTotal={2}
        isStaged={false}
      />,
    );
    const btn = screen.getByRole("button", { name: "Deploy" });
    expect(btn).not.toBeDisabled();
  });

  it("syncStatus saving shows 'Saving…' for existing stacks", () => {
    render(
      <CanvasEditorShell
        {...base}
        nameEditable={false}
        stackName="api"
        syncStatus={SYNC_STATUS.saving}
      />,
    );
    expect(screen.getByText("Saving…")).toBeInTheDocument();
  });

  it("DRAFT pill renders when isStaged and not isDraft", () => {
    render(
      <CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged />,
    );
    expect(screen.getByText("DRAFT")).toBeInTheDocument();
  });

  it("DRAFT pill does not render for isDraft mode", () => {
    render(
      <CanvasEditorShell {...base} isDraft nameEditable stackName="new" isStaged />,
    );
    expect(screen.queryByText("DRAFT")).toBeNull();
  });

  it("Edit menu item is absent", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" />);
    // Open the dropdown
    fireEvent.click(screen.getByLabelText("Stack actions"));
    expect(screen.queryByText("Edit")).toBeNull();
    expect(screen.queryByText("Editing")).toBeNull();
  });
});

describe("CanvasEditorShell collapse", () => {
  afterEach(() => localStorage.clear());

  it("collapses to a compact bar and hides labels/subtitle", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1"
      labels={[{ key: "user", value: "prod" }]} />);
    expect(screen.getByText("prod")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    expect(screen.queryByText("prod")).toBeNull();
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

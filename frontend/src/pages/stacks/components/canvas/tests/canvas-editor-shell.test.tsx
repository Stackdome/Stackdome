// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { CanvasEditorShell } from "../CanvasEditorShell";

afterEach(cleanup);

const base = {
  statusState: null, subtitle: "0 services · 0 volumes",
  activeTab: "configuration", onTabChange: () => {},
  isActive: true, dirtyResourceCount: 0, dirtyTotal: 1, isStaged: false,
  isSaving: false, deployBusy: false, canWrite: true,
  onSave: () => {}, onDeploy: () => {}, onDiscardAll: () => {}, onEdit: () => {}, onDelete: () => {},
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

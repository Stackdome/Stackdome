// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, within, fireEvent, cleanup } from "@testing-library/react";
import { CanvasEditorShell } from "../CanvasEditorShell";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

afterEach(cleanup);

const base = {
  statusState: null, subtitle: "0 services · 0 volumes",
  activeTab: "architecture", onTabChange: () => {},
  isActive: true, dirtyResourceCount: 0, dirtyTotal: 0, isStaged: false,
  hasResources: true,
  onViewChanges: () => {},
  syncStatus: SYNC_STATUS.idle,
  deployBusy: false, canWrite: true,
  onDraftDeploy: () => {}, draftDeploying: false,
  onDeploy: () => {}, onDelete: () => {},
  canDiscardDraft: false, canDeleteStack: true,
  architecture: <div />, deployments: <div />, logs: <div />, metrics: <div />,
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

describe("CanvasEditorShell deploy pill", () => {
  it("draft with resources shows the pill Deploy wired to onDraftDeploy, no Details/menu", () => {
    const onDraftDeploy = vi.fn();
    render(
      <CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" onDraftDeploy={onDraftDeploy} />,
    );
    // Scoped to the pill: the tab rail's "Deployments" tab also matches /deploy/i.
    const pill = within(screen.getByTestId("deploy-pill"));
    fireEvent.click(pill.getByRole("button", { name: /deploy/i }));
    expect(onDraftDeploy).toHaveBeenCalled();
    expect(screen.queryByRole("button", { name: "Details" })).toBeNull();
    expect(screen.queryByLabelText("Stack actions")).toBeNull();
  });

  it("empty draft shows no pill at all", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" hasResources={false} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });

  it("draft shows 'Deploying' while the draft deploy runs", () => {
    render(<CanvasEditorShell {...base} isDraft nameEditable stackName="my-stack" draftDeploying />);
    expect(screen.getByRole("button", { name: /deploying/i })).toBeDisabled();
  });

  it("existing clean stack renders no pill and no rail Deploy", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
    // Exact match — the tab rail's "Deployments" tab also matches a /deploy/i regex.
    expect(screen.queryByRole("button", { name: "Deploy" })).toBeNull();
  });

  it("existing dirty stack shows count + Details wired to onViewChanges", () => {
    const onViewChanges = vi.fn();
    render(
      <CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={3} onViewChanges={onViewChanges} />,
    );
    expect(screen.getByText("Apply 3 changes")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Details" }));
    expect(onViewChanges).toHaveBeenCalled();
  });

  it("pill persists with 'Deploying' while deployBusy even at zero dirt", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" deployBusy />);
    expect(screen.getByRole("button", { name: /deploying/i })).toBeInTheDocument();
  });

  it("pill hidden when an ops tab is active", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={2} activeTab="logs" />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });

  it("staged-but-zero-count nets out — no pill", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged dirtyTotal={0} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
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

  it("collapsed mini-row has no deploy actions, only tabs + expand chevron", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" isActive dirtyTotal={2} />);
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    // The pill still serves deploys; the mini-row itself carries no buttons besides tabs + chevron.
    expect(screen.getByTestId("deploy-pill")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, within, fireEvent, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CanvasEditorShell } from "../canvas-editor-shell";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";
import { EDITOR_TABS } from "../editor-tabs";

afterEach(cleanup);

const base = {
  subtitle: "0 services · 0 volumes",
  activeTab: EDITOR_TABS.architecture, onTabChange: () => {},
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
    render(<CanvasEditorShell {...base} stackName="" isNewStack nameEditable onNameChange={onNameChange} />);
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
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" headerHealth="ok" isStaged />);
    expect(screen.getByText("ok")).toBeInTheDocument();
    expect(screen.queryByText("DRAFT")).toBeNull();
  });

  it("shows a neutral 'Not deployed' pill when no health is derivable (never deployed)", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" />);
    expect(screen.getByText("Not deployed")).toBeInTheDocument();
  });

  it("failed first deploy shows an error pill (health 'failed'), not an empty header", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" headerHealth="failed" />);
    expect(screen.getByText("failed")).toBeInTheDocument();
    expect(screen.queryByText("Not deployed")).toBeNull();
  });

  it("shows a pending 'Deleting' pill when the stack lifecycle is deleting, overriding health", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" headerHealth="ok" lifecycle="deleting" />);
    expect(screen.getByText("Deleting")).toBeInTheDocument();
    expect(screen.queryByText("ok")).toBeNull();
  });
});

describe("CanvasEditorShell deploy pill", () => {
  it("draft with resources shows the pill Deploy wired to onDraftDeploy, no Details/menu", () => {
    const onDraftDeploy = vi.fn();
    render(
      <CanvasEditorShell {...base} isNewStack nameEditable stackName="my-stack" onDraftDeploy={onDraftDeploy} />,
    );
    // Scoped to the pill: the tab rail's "Deployments" tab also matches /deploy/i.
    const pill = within(screen.getByTestId("deploy-pill"));
    fireEvent.click(pill.getByRole("button", { name: /deploy/i }));
    expect(onDraftDeploy).toHaveBeenCalled();
    expect(screen.queryByRole("button", { name: "Details" })).toBeNull();
    expect(screen.queryByLabelText("Stack actions")).toBeNull();
  });

  it("empty draft shows no pill at all", () => {
    render(<CanvasEditorShell {...base} isNewStack nameEditable stackName="my-stack" hasResources={false} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });

  it("draft shows 'Deploying' while the draft deploy runs", () => {
    render(<CanvasEditorShell {...base} isNewStack nameEditable stackName="my-stack" draftDeploying />);
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
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={2} activeTab={EDITOR_TABS.logs} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });

  it("staged-but-zero-count nets out — no pill", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isStaged dirtyTotal={0} />);
    expect(screen.queryByTestId("deploy-pill")).toBeNull();
  });
});

describe("CanvasEditorShell deploy-failed chip", () => {
  it("shows a 'Deploy failed' chip wired to onTabChange('deployments') when latestDeployFailed", () => {
    const onTabChange = vi.fn();
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" latestDeployFailed onTabChange={onTabChange} />);
    const chip = screen.getByRole("button", { name: "Latest deploy failed — view deployments" });
    expect(chip).toBeInTheDocument();
    expect(screen.getByText("Deploy failed")).toBeVisible();
    fireEvent.click(chip);
    expect(onTabChange).toHaveBeenCalledWith(EDITOR_TABS.deployments);
  });

  it("renders no chip when latestDeployFailed is unset", () => {
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" />);
    expect(screen.queryByRole("button", { name: "Latest deploy failed — view deployments" })).toBeNull();
  });
});

describe("CanvasEditorShell actions menu", () => {
  it("exposes the actions trigger for existing stacks with no 'Discard all changes' item", () => {
    // Radix dropdown content mounts on pointer interaction (not in jsdom), so we
    // assert the trigger exists and that the removed item never renders eagerly.
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive dirtyTotal={2} />);
    expect(screen.getByRole("button", { name: "Stack actions" })).toBeInTheDocument();
    expect(screen.queryByText("Discard all changes")).toBeNull();
  });

  it("defers onDelete until after the menu has closed, avoiding the Radix pointer-events lock", async () => {
    // Regression test for a Radix DropdownMenu -> AlertDialog composition bug:
    // if the dialog-opening callback fires synchronously from the menu item,
    // the menu's close and the dialog's mount race and can leave
    // document.body.style.pointerEvents stuck at "none" forever (the dialog
    // captures "none" as the value to restore on unmount). Deferring the
    // callback lets the menu finish closing (and reset pointer-events) before
    // the dialog mounts. See https://github.com/radix-ui/primitives/issues/1836
    const user = userEvent.setup();
    const pointerEventsAtCall: string[] = [];
    const onDelete = vi.fn(() => {
      pointerEventsAtCall.push(document.body.style.pointerEvents);
    });
    render(<CanvasEditorShell {...base} nameEditable={false} stackName="api" isActive onDelete={onDelete} />);
    await user.click(screen.getByRole("button", { name: "Stack actions" }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText("Delete stack"), { pointerEventsCheck: 0 });
    await waitFor(() => expect(onDelete).toHaveBeenCalled());
    // At the moment the dialog-opening callback runs, the menu must already
    // have released its body pointer-events lock.
    expect(pointerEventsAtCall[0]).not.toBe("none");
  });
});

describe("CanvasEditorShell collapse", () => {
  afterEach(() => localStorage.clear());

  // Collapse is driven by zen mode (canvas control / ⌘.) through
  // HeaderCollapseContext; the shell only persists and renders the state.
  it("restores the persisted collapsed state as a compact bar", () => {
    localStorage.setItem("stackdome.editor-header-collapsed.s1", "1");
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    // The expanded header stays mounted for the height animation but is
    // inert + aria-hidden; the compact bar becomes the visible surface.
    expect(screen.getByRole("heading", { name: "acme", hidden: true }).closest("[inert]")).not.toBeNull();
    expect(screen.getAllByText("acme").length).toBeGreaterThan(0); // compact bar name
  });

  it("has no collapse chevron of its own", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    expect(screen.queryByRole("button", { name: /Collapse header|Expand header/ })).toBeNull();
  });

  it("keeps tabs clickable while collapsed", () => {
    const onTabChange = vi.fn();
    localStorage.setItem("stackdome.editor-header-collapsed.s1", "1");
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" onTabChange={onTabChange} />);
    fireEvent.click(screen.getByRole("button", { name: /Logs/ }));
    expect(onTabChange).toHaveBeenCalledWith(EDITOR_TABS.logs);
  });

  it("collapsed mini-row still serves deploys via the pill", () => {
    localStorage.setItem("stackdome.editor-header-collapsed.s1", "1");
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" isActive dirtyTotal={2} />);
    expect(screen.getByTestId("deploy-pill")).toBeInTheDocument();
  });
});

// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { ArchitectureTab } from "../architecture-tab";
import { ConfirmProvider } from "@/components/branded/confirm";
import { useStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";

afterEach(cleanup);

// Same rationale as architecture-tab-volume-delete.test.tsx: the real
// <ReactFlow> never stabilizes under jsdom. The stub keeps the node list and
// click wiring, and surfaces the readOnly flag so the live view's canvas
// lockdown is assertable.
vi.mock("../canvas-editor", () => ({
  CanvasEditor: (props: {
    nodes: { id: string; data: { name?: string } }[];
    onNodeClick?: (e: unknown, n: unknown) => void;
    readOnly?: boolean;
  }) => (
    <div data-testid={props.readOnly ? "canvas-readonly" : "canvas-editable"}>
      {props.nodes.map((n) => (
        <button key={n.id} type="button" onClick={(e) => props.onNodeClick?.(e, n)}>
          {n.data?.name ?? n.id}
        </button>
      ))}
    </div>
  ),
}));

const EMPTY_ADDONS: never[] = [];
const EMPTY_SECRETS: never[] = [];
const NO_TOPOLOGY = { topology: null };
vi.mock("@/pages/addons/hooks/use-postgres-addons", () => ({
  usePostgresAddons: () => ({ addons: EMPTY_ADDONS }),
}));
vi.mock("@/pages/stacks/hooks/use-secrets", () => ({
  useSecrets: () => ({ secrets: EMPTY_SECRETS, isLoading: false }),
}));
vi.mock("@/pages/stacks/hooks/use-stack-topology", () => ({
  useStackTopology: () => NO_TOPOLOGY,
}));

// Module-scoped for referential stability (see the volume-delete harness).
// The draft and the live snapshot deliberately hold differently-named
// resources, so which list the canvas is rendering is directly observable.
const DRAFT_RESOURCE = { name: "web-draft" };
const DRAFT_RESOURCES = [DRAFT_RESOURCE];
const LIVE_RESOURCE = { name: "api-live" };
const NO_VOLUMES: never[] = [];
const NO_ADDON_IDS = new Set<string>();
const NO_ADDON_NAMES = new Map<string, string>();
const NO_ERRORS = {};
const LIVE_VIEW = {
  resources: [LIVE_RESOURCE],
  volumes: NO_VOLUMES,
  linkedAddonIds: NO_ADDON_IDS,
};

function Harness(props: { liveView?: typeof LIVE_VIEW }) {
  const session = useStackEditSession();
  if (!session.isActive) {
    session.start({ resources: DRAFT_RESOURCES, volumes: NO_VOLUMES }, { openTab: "configuration" });
  }
  return (
    <ConfirmProvider>
      <ArchitectureTab
        session={session}
        baselineResources={DRAFT_RESOURCES}
        baselineVolumes={NO_VOLUMES}
        draftResources={DRAFT_RESOURCES}
        draftVolumes={NO_VOLUMES}
        connectionAddonIds={NO_ADDON_IDS}
        addonNameById={NO_ADDON_NAMES}
        errors={NO_ERRORS}
        topologyIds={{ orgId: "o", projectName: "t", stackId: "s" }}
        topologyRefreshKey={0}
        liveView={props.liveView}
      />
    </ConfirmProvider>
  );
}

describe("ArchitectureTab live view", () => {
  it("renders no Draft/Live toggle without a live view (never deployed)", () => {
    render(<Harness />);
    expect(screen.queryByRole("group", { name: "Canvas view" })).not.toBeInTheDocument();
    expect(screen.getByTestId("canvas-editable")).toBeInTheDocument();
  });

  it("switches the canvas to the live snapshot's resources, read-only", () => {
    render(<Harness liveView={LIVE_VIEW} />);
    // Draft view first: draft node shown, live node absent.
    expect(screen.getByRole("button", { name: "web-draft" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "api-live" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Live" }));

    expect(screen.getByTestId("canvas-readonly")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "api-live" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "web-draft" })).not.toBeInTheDocument();
    expect(screen.getByText(/read-only/)).toBeInTheDocument();
  });

  it("opens a disabled, mutation-free drawer for a live node", () => {
    render(<Harness liveView={LIVE_VIEW} />);
    fireEvent.click(screen.getByRole("button", { name: "Live" }));
    fireEvent.click(screen.getByRole("button", { name: "api-live" }));

    const drawer = screen.getByTestId("resource-drawer");
    expect(drawer).toHaveTextContent("Live · read-only");
    expect(screen.queryByRole("button", { name: /remove resource/i })).not.toBeInTheDocument();
    // jest-dom's toBeDisabled honours the ancestor <fieldset disabled> — every
    // form field in the drawer body is dead.
    const nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toBeDisabled();
  });

  it("closes the drawer when switching back to Draft (indexes are not stable across views)", () => {
    render(<Harness liveView={LIVE_VIEW} />);
    fireEvent.click(screen.getByRole("button", { name: "Live" }));
    fireEvent.click(screen.getByRole("button", { name: "api-live" }));
    expect(screen.getByTestId("resource-drawer")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Draft" }));
    expect(screen.queryByTestId("resource-drawer")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "web-draft" })).toBeInTheDocument();
  });
});

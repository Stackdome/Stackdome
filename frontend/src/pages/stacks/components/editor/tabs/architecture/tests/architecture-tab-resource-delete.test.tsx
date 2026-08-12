// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent, within, waitFor } from "@testing-library/react";
import { ArchitectureTab } from "../architecture-tab";
import { ConfirmProvider } from "@/components/branded/confirm";
import { useStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";

afterEach(cleanup);

// Same stub as the volume-delete suite: the real <ReactFlow> never stabilizes
// under jsdom's zero-size layout.
vi.mock("../canvas-editor", () => ({
  CanvasEditor: (props: {
    nodes: { id: string; data: { name?: string } }[];
    onNodeClick?: (e: unknown, n: unknown) => void;
  }) => (
    <div>
      {props.nodes.map((n) => (
        <button key={n.id} type="button" onClick={(e) => props.onNodeClick?.(e, n)}>
          {n.id}
        </button>
      ))}
    </div>
  ),
}));

const EMPTY_ADDONS: never[] = [];
const EMPTY_SECRETS: never[] = [];
const NO_TOPOLOGY = { topology: null };
vi.mock("@/hooks/use-postgres-addons", () => ({
  usePostgresAddons: () => ({ addons: EMPTY_ADDONS }),
}));
vi.mock("@/pages/stacks/hooks/use-secrets", () => ({
  useSecrets: () => ({ secrets: EMPTY_SECRETS, isLoading: false }),
}));
vi.mock("@/pages/stacks/hooks/use-stack-topology", () => ({
  useStackTopology: () => NO_TOPOLOGY,
}));

// Module-scoped so the memo chains ArchitectureTab feeds them into stay stable.
const NO_VOLUMES: never[] = [];
const NO_ADDON_IDS = new Set<string>();
const NO_ADDON_NAMES = new Map<string, string>();
const NO_ERRORS = {};

const NAMED = [{ name: "api" }, { name: "web", depends_on: ["api"] }];
// A resource whose name the user has cleared mid-edit.
const UNNAMED = [{ name: "" }];

function Harness({ resources }: { resources: { name: string; depends_on?: string[] }[] }) {
  const session = useStackEditSession();
  if (!session.isActive) {
    session.start({ resources, volumes: NO_VOLUMES }, { openTab: "configuration" });
  }
  return (
    <ConfirmProvider>
      <ArchitectureTab
        session={session}
        baselineResources={resources}
        baselineVolumes={NO_VOLUMES}
        draftResources={resources}
        draftVolumes={NO_VOLUMES}
        connectionAddonIds={NO_ADDON_IDS}
        addonNameById={NO_ADDON_NAMES}
        errors={NO_ERRORS}
        topologyIds={null}
        topologyRefreshKey={0}
      />
    </ConfirmProvider>
  );
}

const openDrawerAndRemove = async (nodeId: string) => {
  fireEvent.click(await screen.findByRole("button", { name: nodeId }));
  fireEvent.click(await screen.findByText("Remove resource"));
};

describe("ArchitectureTab resource delete from the drawer", () => {
  it("confirms, naming the dependents it will repair", async () => {
    render(<Harness resources={NAMED} />);
    await openDrawerAndRemove("resource:api");

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByRole("heading", { name: "Delete service “api”?" })).toBeInTheDocument();
    expect(within(dialog).getByText("web")).toBeInTheDocument();
    expect(within(dialog).getByText("depends_on")).toBeInTheDocument();
  });

  /** Regression: routing this path through the name-keyed handler made an
   *  unnamed resource undeletable, with no dialog and no error. */
  it("deletes an unnamed resource instead of doing nothing", async () => {
    render(<Harness resources={UNNAMED} />);
    expect(await screen.findByRole("button", { name: "resource:" })).toBeInTheDocument();

    await openDrawerAndRemove("resource:");

    await waitFor(() =>
      expect(screen.queryByRole("button", { name: "resource:" })).not.toBeInTheDocument(),
    );
  });
});

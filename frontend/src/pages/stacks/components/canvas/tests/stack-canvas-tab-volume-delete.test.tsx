// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent, within, act } from "@testing-library/react";
import { StackCanvasTab } from "../StackCanvasTab";
import { useStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";

afterEach(cleanup);

// The real <ReactFlow> canvas measures node/viewport dimensions in a way that
// never stabilizes under jsdom's zero-size layout (infinite update loop).
// Stub it with a plain node list — still wired to the same onNodeClick
// handler — so clicking the volume node still opens the volume drawer (the
// route into the delete-confirm dialog under test) without mounting xyflow.
vi.mock("../CanvasEditor", () => ({
  CanvasEditor: (props: { nodes: { id: string; data: { name?: string } }[]; onNodeClick?: (e: unknown, n: unknown) => void }) => (
    <div>
      {props.nodes.map((n) => (
        <button key={n.id} type="button" onClick={(e) => props.onNodeClick?.(e, n)}>
          {n.data?.name ?? n.id}
        </button>
      ))}
    </div>
  ),
}));

// Heavy network-backed hooks the canvas composes with — none of this test's
// assertions touch their data, so stub them to keep the render synchronous
// and silent.
const EMPTY_ADDONS: never[] = [];
const EMPTY_SECRETS: never[] = [];
const NO_TOPOLOGY = { topology: null };
vi.mock("@/pages/addons/hooks/use-postgres-addons", () => ({
  usePostgresAddons: () => ({ addons: EMPTY_ADDONS }),
}));
vi.mock("@/pages/stacks/hooks/use-secrets", () => ({
  useSecrets: () => ({ secrets: EMPTY_SECRETS, isLoading: false, dockerRegistrySecrets: EMPTY_SECRETS, gitSecrets: EMPTY_SECRETS }),
}));
vi.mock("@/pages/stacks/hooks/use-stack-topology", () => ({
  useStackTopology: () => NO_TOPOLOGY,
}));

// Module-scoped (referentially stable across renders) — StackCanvasTab feeds
// these into memo chains that gate render-triggering effects; fresh literals
// on every render would recompute those memos every time and spin forever.
// Unmounted (no volume_mounts referencing it) so it renders as its own
// floating attachment node — mounted volumes dock onto the resource card
// chip instead, which the CanvasEditor stub below doesn't reproduce.
const RESOURCE = { name: "web" };
const VOLUME = { name: "data", spec: { size: "1Gi" } };
const RESOURCES = [RESOURCE];
const VOLUMES = [VOLUME];
const NO_ADDON_IDS = new Set<string>();
const NO_ADDON_NAMES = new Map<string, string>();
const NO_ERRORS = {};

/** Renders the real edit session hook (pure, no network) seeded with one
 *  resource + one mounted volume, and opens the volume drawer up front so the
 *  "Remove volume" button (already wired to onRequestRemove) is reachable
 *  without simulating xyflow node interactions. */
function Harness(props: {
  topologyIds: { orgId: string; teamName: string; stackId: string } | null;
  onDeleteVolume?: (name: string) => Promise<boolean>;
  deletingVolume?: boolean;
}) {
  const session = useStackEditSession();
  if (!session.isActive) {
    session.start({ resources: RESOURCES, volumes: VOLUMES }, { openTab: "configuration" });
  }
  return (
    <StackCanvasTab
      session={session}
      baselineResources={RESOURCES}
      baselineVolumes={VOLUMES}
      draftResources={RESOURCES}
      draftVolumes={VOLUMES}
      connectionAddonIds={NO_ADDON_IDS}
      addonNameById={NO_ADDON_NAMES}
      errors={NO_ERRORS}
      topologyIds={props.topologyIds}
      topologyRefreshKey={0}
      onDeleteVolume={props.onDeleteVolume}
      deletingVolume={props.deletingVolume}
    />
  );
}

/** Open the volume drawer, then click "Remove volume" — the existing route
 *  into the shared delete-confirm dialog under test. */
async function openDeleteConfirm() {
  fireEvent.click(await screen.findByText("data"));
  fireEvent.click(await screen.findByText("Remove volume"));
}

describe("StackCanvasTab volume delete", () => {
  it("saved stack: dialog carries the immediate data-loss copy", async () => {
    render(<Harness topologyIds={{ orgId: "o", teamName: "t", stackId: "s" }} />);
    await openDeleteConfirm();

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/immediately and permanently destroys/i)).toBeInTheDocument();
    expect(within(dialog).getByText(/cannot be undone/i)).toBeInTheDocument();
    expect(within(dialog).getByRole("button", { name: "Delete volume" })).toBeInTheDocument();
  });

  it("draft stack (topologyIds null): dialog carries the draft-removal copy", async () => {
    render(<Harness topologyIds={null} />);
    await openDeleteConfirm();

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/hasn't been created yet/i)).toBeInTheDocument();
    expect(within(dialog).getByRole("button", { name: "Remove volume" })).toBeInTheDocument();
  });

  it("Action is disabled and reads 'Deleting…' while a delete is in flight", async () => {
    render(<Harness topologyIds={{ orgId: "o", teamName: "t", stackId: "s" }} deletingVolume />);
    await openDeleteConfirm();

    const dialog = await screen.findByRole("alertdialog");
    const action = within(dialog).getByRole("button", { name: "Deleting…" });
    expect(action).toBeDisabled();
  });

  it("applies the local draft edit (volume node gone) before onDeleteVolume settles", async () => {
    let resolveDelete: (v: boolean) => void = () => {};
    const onDeleteVolume = vi.fn(() => new Promise<boolean>((resolve) => { resolveDelete = resolve; }));

    render(<Harness topologyIds={{ orgId: "o", teamName: "t", stackId: "s" }} onDeleteVolume={onDeleteVolume} />);
    await openDeleteConfirm();

    const dialog = await screen.findByRole("alertdialog");
    fireEvent.click(within(dialog).getByRole("button", { name: "Delete volume" }));

    // The local draft mutation (dropping the volume from the draft) has
    // already applied — its node is gone from the canvas — even though
    // onDeleteVolume's promise hasn't settled yet.
    expect(onDeleteVolume).toHaveBeenCalledWith("data");
    expect(screen.queryByRole("button", { name: "data" })).not.toBeInTheDocument();

    await act(async () => {
      resolveDelete(true);
      await Promise.resolve();
    });
  });
});

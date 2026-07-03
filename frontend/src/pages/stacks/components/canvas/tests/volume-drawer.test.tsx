// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { VolumeDrawer } from "../VolumeDrawer";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";

afterEach(cleanup);

function makeSession(volumes: { name: string; spec?: { size?: string } }[]) {
  const updateVolumes = vi.fn();
  const updateResources = vi.fn();
  return {
    session: {
      isActive: true,
      draft: { resources: [{ name: "web", volume_mounts: [{ source_volume_name: "data", target_path: "/data" }] }], volumes },
      updateVolumes,
      updateResources,
    } as unknown as UseStackEditSession,
    updateVolumes,
  };
}

describe("VolumeDrawer", () => {
  it("renders the volume's fields and mount details", () => {
    const { session } = makeSession([{ name: "data", spec: { size: "1Gi" } }]);
    render(<VolumeDrawer volumeName="data" session={session} onClose={vi.fn()} />);
    expect(screen.getByDisplayValue("data")).toBeInTheDocument();
    expect(screen.getByDisplayValue("1Gi")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument(); // mounted-by row
  });

  it("edits flow into session.updateVolumes", () => {
    const { session, updateVolumes } = makeSession([{ name: "data", spec: { size: "1Gi" } }]);
    render(<VolumeDrawer volumeName="data" session={session} onClose={vi.fn()} />);
    fireEvent.change(screen.getByDisplayValue("1Gi"), { target: { value: "2Gi" } });
    expect(updateVolumes).toHaveBeenCalled();
  });

  it("name input is read-only in the drawer; size stays editable", () => {
    // Drawer entries are keyed by volume name — renaming in place would
    // orphan the entry and close the drawer after one keystroke.
    const { session } = makeSession([{ name: "data", spec: { size: "1Gi" } }]);
    render(<VolumeDrawer volumeName="data" session={session} onClose={vi.fn()} />);
    expect(screen.getByDisplayValue("data")).toBeDisabled();
    expect(screen.getByDisplayValue("1Gi")).toBeEnabled();
  });

  it("close button calls onClose", () => {
    const { session } = makeSession([{ name: "data" }]);
    const onClose = vi.fn();
    render(<VolumeDrawer volumeName="data" session={session} onClose={onClose} />);
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalled();
  });
});

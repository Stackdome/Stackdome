// @vitest-environment jsdom
import { describe, expect, it, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, cleanup } from "@testing-library/react";
import { AddVolumeDialog } from "../AddVolumeDialog";

afterEach(cleanup);

const resources = [{ name: "web", volume_mounts: [{ source_volume_name: "old", source_sub_path: "", target_path: "/taken" }] }];
const volumes = [{ name: "old" }];

function renderDialog(onCreate = vi.fn()) {
  render(
    <AddVolumeDialog
      open
      onOpenChange={() => {}}
      resources={resources}
      volumes={volumes}
      initialResourceIdx={0}
      onCreate={onCreate}
    />,
  );
  return onCreate;
}

describe("AddVolumeDialog", () => {
  it("suggests a unique name and creates with valid input", () => {
    const onCreate = renderDialog();
    expect(screen.getByLabelText(/name/i)).toHaveValue("volume");
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(onCreate).toHaveBeenCalledWith({ name: "volume", size: "1Gi", resourceIdx: 0, targetPath: "/data" });
  });

  it("blocks duplicate names", () => {
    const onCreate = renderDialog();
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "old" } });
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(onCreate).not.toHaveBeenCalled();
    expect(screen.getByText(/must be unique/i)).toBeInTheDocument();
  });

  it("blocks relative and already-taken mount paths", () => {
    const onCreate = renderDialog();
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(screen.getByText(/absolute path/i)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/taken" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(screen.getByText(/already mounted/i)).toBeInTheDocument();
    expect(onCreate).not.toHaveBeenCalled();
  });
});

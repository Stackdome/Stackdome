// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, within } from "@testing-library/react";
import { Tabs } from "@/components/ui/tabs";
import { StackResourceConfigurationTab, pickConfigurationDraft } from "../configuration-tab";

afterEach(cleanup);

/**
 * Mounts are attached on the canvas, so the drawer lists them read-only. The
 * row still has to show that it differs from the deployed baseline — without
 * offering a reset, which here would silently detach the volume.
 */

const DATA = { source_volume_name: "data", source_sub_path: "", target_path: "/data" };
const LOGS = { source_volume_name: "logs", source_sub_path: "", target_path: "/logs" };

function renderMounts(draftMounts: Array<Record<string, unknown>>, baselineMounts: Array<Record<string, unknown>>) {
  const resource = { name: "api", sourceType: "image" as const, source: { image: { ref: "api:1" } } };
  render(
    <Tabs defaultValue="general">
      <StackResourceConfigurationTab
        draft={pickConfigurationDraft({ ...resource, volume_mounts: draftMounts } as never)}
        baseline={pickConfigurationDraft({ ...resource, volume_mounts: baselineMounts } as never)}
        index={0}
        errors={{}}
        volumes={[]}
        mountsReadOnly
        onPatchResource={vi.fn()}
        onDiscardField={vi.fn()}
      />
    </Tabs>,
  );
  // The tint lives on the DirtyField wrapper, which is the row's outer element.
  return (targetPath: string) =>
    screen.getByText(targetPath).closest("[class*='border-b']") as HTMLElement | null;
}

describe("read-only mount rows", () => {
  it("marks the row whose volume changed", () => {
    const rowFor = renderMounts([{ ...DATA, source_volume_name: "data-v2" }], [DATA]);
    expect(rowFor("/data")?.className).toContain("bg-brand-bg");
  });

  it("leaves an unchanged row unmarked", () => {
    const rowFor = renderMounts([{ ...DATA, source_volume_name: "data-v2" }, LOGS], [DATA, LOGS]);
    expect(rowFor("/logs")?.className).not.toContain("bg-brand-bg");
  });

  /**
   * DirtyField keeps the reset button mounted and flips its visibility, so that
   * going dirty never remounts the field and eats the keystroke that caused it.
   * The assertion is therefore that the row's button is not offered, not that
   * it is absent.
   */
  it("offers no reset on the row — detaching a volume is a canvas action", () => {
    const rowFor = renderMounts([{ ...DATA, source_volume_name: "data-v2" }], [DATA]);
    const reset = within(rowFor("/data")!).getByLabelText("Reset to original value");
    expect(reset).toHaveAttribute("aria-hidden", "true");
    expect(reset).toHaveAttribute("tabindex", "-1");
    expect(reset.className).toContain("invisible");
  });
});

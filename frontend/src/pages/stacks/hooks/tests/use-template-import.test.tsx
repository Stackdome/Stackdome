// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook, act, cleanup } from "@testing-library/react";
import { useTemplateImport } from "../use-template-import";
import { getTemplateById } from "@/pages/stacks/data/templates/registry";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));

afterEach(() => {
  cleanup();
  navigate.mockReset();
});

describe("useTemplateImport", () => {
  it("opens and closes the dialog", () => {
    const { result } = renderHook(() => useTemplateImport());
    expect(result.current.isDialogOpen).toBe(false);
    act(() => result.current.openDialog());
    expect(result.current.isDialogOpen).toBe(true);
    act(() => result.current.closeDialog());
    expect(result.current.isDialogOpen).toBe(false);
  });

  it("navigates to /stacks/new with a seed prefilled from the template", () => {
    const { result } = renderHook(() => useTemplateImport());
    const tooljet = getTemplateById("tooljet")!;

    act(() => {
      result.current.openDialog();
      result.current.useTemplate(tooljet);
    });

    expect(navigate).toHaveBeenCalledTimes(1);
    expect(navigate).toHaveBeenCalledWith(
      "/stacks/new",
      expect.objectContaining({
        state: expect.objectContaining({
          seed: expect.objectContaining({
            name: expect.any(String),
            resources: expect.any(Array),
            volumes: expect.any(Array),
            linkedAddonIds: expect.any(Array),
          }),
        }),
      }),
    );
    // dialog closes on use
    expect(result.current.isDialogOpen).toBe(false);
  });
});

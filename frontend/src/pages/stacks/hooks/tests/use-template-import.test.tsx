// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook, act, cleanup } from "@testing-library/react";
import { useTemplateImport } from "../use-template-import";
import { getTemplateById } from "@/data/templates/registry";
import { ImportSource } from "@/pages/stacks/lib/import-source";

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

  it("navigates to the create form prefilled from the template", () => {
    const { result } = renderHook(() => useTemplateImport());
    const tooljet = getTemplateById("tooljet")!;

    act(() => {
      result.current.openDialog();
      result.current.useTemplate(tooljet);
    });

    expect(navigate).toHaveBeenCalledTimes(1);
    const [path, opts] = navigate.mock.calls[0];
    expect(path).toBe("/stacks/create");
    expect(opts.state.importSource).toBe(ImportSource.Template);
    expect(opts.state.importedData.name).toBe("tooljet");
    // dialog closes on use
    expect(result.current.isDialogOpen).toBe(false);
  });
});

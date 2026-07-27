// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { DeleteProjectDialog } from "../delete-project-dialog";

afterEach(() => cleanup());

describe("DeleteProjectDialog type-to-confirm", () => {
  it("keeps confirm disabled until the exact project name is typed", () => {
    const onConfirm = vi.fn();
    render(<DeleteProjectDialog open projectName="payments" onOpenChange={() => {}} onConfirm={onConfirm} />);
    const confirmBtn = screen.getByRole("button", { name: /^delete$/i });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(true);
    fireEvent.change(screen.getByLabelText(/type the project name/i), { target: { value: "paymentx" } });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(true);
    fireEvent.change(screen.getByLabelText(/type the project name/i), { target: { value: "payments" } });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(confirmBtn);
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });
});

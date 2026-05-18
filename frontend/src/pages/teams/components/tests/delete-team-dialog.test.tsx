// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { DeleteTeamDialog } from "../delete-team-dialog";

afterEach(() => cleanup());

describe("DeleteTeamDialog type-to-confirm", () => {
  it("keeps confirm disabled until the exact team name is typed", () => {
    const onConfirm = vi.fn();
    render(<DeleteTeamDialog open teamName="payments" onOpenChange={() => {}} onConfirm={onConfirm} />);
    const confirmBtn = screen.getByRole("button", { name: /delete team/i });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(true);
    fireEvent.change(screen.getByLabelText(/type the team name/i), { target: { value: "paymentx" } });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(true);
    fireEvent.change(screen.getByLabelText(/type the team name/i), { target: { value: "payments" } });
    expect((confirmBtn as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(confirmBtn);
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });
});

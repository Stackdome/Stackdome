// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AddFromAddonDialog } from "../src/pages/stacks/components/shared/add-from-addon-dialog";
import type { PostgresAddon } from "@/api/addons";

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => ({
  id: "addon-1",
  name: "tooljet-db",
  status: { state: "Ready" },
  spec: {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "app" }],
    configuration: { enable_superuser_access: false },
  } as any,
  ...(over as any),
});

describe("AddFromAddonDialog", () => {
  it("auto-selects the only database when addon has one", async () => {
    const single = mkAddon({ spec: { ...mkAddon().spec, databases: [{ name: "only" }] } as any });
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[single]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    // Use the dedicated readout for the auto-selected database (Radix Select
    // also renders options containing "only", which makes a generic text
    // query ambiguous in jsdom).
    const readout = await screen.findByTestId("selected-database");
    expect(readout).toHaveTextContent(/only/i);
  });

  it("hides Superuser toggle when addon does not enable_superuser_access", () => {
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    expect(screen.queryByLabelText(/superuser/i)).not.toBeInTheDocument();
  });

  it("shows Superuser toggle when addon enables it", () => {
    const su = mkAddon({
      spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
    });
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[su]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    expect(screen.getByLabelText(/superuser/i)).toBeInTheDocument();
  });

  it("Postgres conventions preset ticks 5 fields with default names", async () => {
    const user = userEvent.setup();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={() => {}}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    expect(screen.getByDisplayValue("PG_HOST")).toBeInTheDocument();
    expect(screen.getByDisplayValue("PG_PASS")).toBeInTheDocument();
    expect(screen.getByDisplayValue("PG_DB")).toBeInTheDocument();
  });

  it("blocks Add when an env name collides with existing env", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set(["PG_HOST"])}
        onAdd={onAdd}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    const addBtn = screen.getByRole("button", { name: /^add$/i });
    expect(addBtn).toBeDisabled();
    expect(screen.getByText(/already exists/i)).toBeInTheDocument();
  });

  it("calls onAdd with one row per ticked field on confirm", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <AddFromAddonDialog
        open
        onOpenChange={() => {}}
        addons={[mkAddon()]}
        existingEnvNames={new Set()}
        onAdd={onAdd}
      />,
    );
    await user.click(screen.getByRole("button", { name: /apply preset/i }));
    await user.click(screen.getByRole("menuitem", { name: /postgres conventions/i }));
    await user.click(screen.getByRole("button", { name: /^add$/i }));
    expect(onAdd).toHaveBeenCalledTimes(1);
    const rows = onAdd.mock.calls[0][0];
    expect(rows).toHaveLength(5);
    rows.forEach((r: any) => {
      expect(r.from).toBe("addon");
      expect(r.addonType).toBe("postgres");
      expect(r.addonId).toBe("addon-1");
    });
    expect(rows.map((r: any) => r.credField).sort()).toEqual(
      ["database", "host", "password", "port", "username"].sort(),
    );
  });
});

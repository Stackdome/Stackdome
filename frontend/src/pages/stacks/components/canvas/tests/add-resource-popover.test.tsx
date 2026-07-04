// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { AddResourcePopover } from "../AddResourcePopover";

afterEach(cleanup);

describe("AddResourcePopover managed addons", () => {
  it("lists linked-able addons and links on click", () => {
    const onLinkAddon = vi.fn();
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[{ id: "a1", name: "prod-db" }]}
        linkedAddonIds={new Set()}
        onLinkAddon={onLinkAddon}
        canAddVolume
        onAddVolume={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.click(screen.getByRole("button", { name: /prod-db/i }));
    expect(onLinkAddon).toHaveBeenCalledWith("a1");
  });

  it("shows check indicator when addon is already linked", () => {
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[{ id: "a1", name: "prod-db" }]}
        linkedAddonIds={new Set(["a1"])}
        onLinkAddon={() => {}}
        canAddVolume
        onAddVolume={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    // The Check icon renders as an SVG inside the tile; the Plus icon must be absent
    const tile = screen.getByRole("button", { name: /prod-db/i });
    // lucide Check renders with aria-hidden; query by the SVG class text-success
    expect(tile.querySelector(".text-success")).toBeInTheDocument();
    expect(tile.querySelector(".text-primary")).toBeNull();
  });
});

describe("AddResourcePopover storage tile search", () => {
  it("shows the Storage section when searching 'storage', matching its header", () => {
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[]}
        linkedAddonIds={new Set()}
        onLinkAddon={() => {}}
        canAddVolume
        onAddVolume={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.change(screen.getByPlaceholderText(/Search services/i), { target: { value: "storage" } });
    expect(screen.getByText("Storage")).toBeInTheDocument();
    expect(screen.getByText("Volume")).toBeInTheDocument();
  });

  it("still shows the Storage section when searching 'volume'", () => {
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[]}
        linkedAddonIds={new Set()}
        onLinkAddon={() => {}}
        canAddVolume
        onAddVolume={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.change(screen.getByPlaceholderText(/Search services/i), { target: { value: "volume" } });
    expect(screen.getByText("Storage")).toBeInTheDocument();
  });

  it("hides the Storage section for unrelated queries", () => {
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[]}
        linkedAddonIds={new Set()}
        onLinkAddon={() => {}}
        canAddVolume
        onAddVolume={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.change(screen.getByPlaceholderText(/Search services/i), { target: { value: "postgres" } });
    expect(screen.queryByText("Storage")).toBeNull();
  });
});

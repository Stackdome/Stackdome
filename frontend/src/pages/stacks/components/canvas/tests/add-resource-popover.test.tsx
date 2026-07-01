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
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.click(screen.getByRole("button", { name: /prod-db/i }));
    expect(onLinkAddon).toHaveBeenCalledWith("a1");
  });
});

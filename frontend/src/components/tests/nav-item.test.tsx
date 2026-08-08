// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { SidebarProvider, SidebarMenu } from "@/components/ui/sidebar";
import { NavItem } from "../nav-item";
import { navGroups, isNavItemActive, type NavItem as NavItemData } from "../nav-items";

// SidebarProvider reads window.matchMedia via useIsMobile; jsdom has neither.
vi.mock("@/hooks/use-mobile", () => ({ useIsMobile: () => false }));

afterEach(cleanup);

const allItems = navGroups.flatMap((g) => g.items);
const byLabel = (label: string): NavItemData => {
  const item = allItems.find((i) => i.label === label);
  if (!item) throw new Error(`no nav item labelled ${label}`);
  return item;
};

function renderItem(item: NavItemData, path = "/") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarProvider>
        <SidebarMenu>
          <NavItem item={item} />
        </SidebarMenu>
      </SidebarProvider>
    </MemoryRouter>,
  );
}

describe("NavItem", () => {
  it("renders a link to the destination", () => {
    renderItem(byLabel("Image Registries"));
    expect(screen.getByRole("link", { name: /image registries/i })).toHaveAttribute(
      "href",
      "/image-registries",
    );
  });

  it("marks the entry active on its own route", () => {
    renderItem(byLabel("Image Registries"), "/image-registries");
    expect(screen.getByRole("link", { name: /image registries/i })).toHaveAttribute(
      "data-active",
      "true",
    );
  });

  it("stays inactive on an unrelated route", () => {
    renderItem(byLabel("Image Registries"), "/secrets");
    expect(screen.getByRole("link", { name: /image registries/i })).toHaveAttribute(
      "data-active",
      "false",
    );
  });
});

describe("isNavItemActive", () => {
  const stacks = byLabel("Stacks");

  it("matches the exact path and anything nested under it", () => {
    expect(isNavItemActive("/stacks", stacks)).toBe(true);
    expect(isNavItemActive("/stacks/s1", stacks)).toBe(true);
  });

  it("does not match a sibling that merely shares a prefix", () => {
    expect(isNavItemActive("/stacks-archive", stacks)).toBe(false);
  });

  // The draft editor is a different screen, not the list — it is full-bleed and
  // has no list behind it, so lighting up Stacks would point at the wrong page.
  it("honours notActiveOn", () => {
    expect(isNavItemActive("/stacks/draft", stacks)).toBe(false);
  });

  // The New stack journey is the exception: §12a drops that page's trail
  // BECAUSE the sidebar keeps saying which section you are in.
  it("keeps the New stack journey active", () => {
    expect(isNavItemActive("/stacks/new", stacks)).toBe(true);
  });
});

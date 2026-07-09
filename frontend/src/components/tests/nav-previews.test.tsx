// @vitest-environment jsdom
import { describe, it, expect, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { SidebarProvider } from "@/components/ui/sidebar";
import { NavPreviews } from "../nav-previews";

afterEach(() => cleanup());

// jsdom in this repo has no matchMedia implementation; SidebarProvider's
// useIsMobile effect calls it on mount, so stub it like other Radix-based
// nav components would need in a browser.
beforeAll(() => {
  window.matchMedia = window.matchMedia || ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia;
});

function renderNav(path = "/") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarProvider>
        <NavPreviews />
      </SidebarProvider>
    </MemoryRouter>,
  );
}

describe("NavPreviews", () => {
  it("links to /previews with the Preview Environments label", () => {
    renderNav();
    const link = screen.getByRole("link", { name: /preview environments/i });
    expect(link.getAttribute("href")).toBe("/previews");
  });

  it("is active on /previews subroutes", () => {
    renderNav("/previews/abc");
    const button = screen.getByRole("link", { name: /preview environments/i }).closest("[data-active]");
    expect(button?.getAttribute("data-active")).toBe("true");
  });
});

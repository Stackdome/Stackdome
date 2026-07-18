// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { SidebarProvider } from "@/components/ui/sidebar";
import { NavImageRegistries } from "../nav-image-registries";

// SidebarProvider reads window.matchMedia via useIsMobile; jsdom has neither.
vi.mock("@/hooks/use-mobile", () => ({ useIsMobile: () => false }));

afterEach(cleanup);

function renderNav(path = "/") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarProvider>
        <NavImageRegistries />
      </SidebarProvider>
    </MemoryRouter>,
  );
}

describe("NavImageRegistries", () => {
  it("renders a link to the image registries page", () => {
    renderNav();
    const link = screen.getByRole("link", { name: /image registries/i });
    expect(link).toHaveAttribute("href", "/image-registries");
  });

  it("marks the entry active on /image-registries routes", () => {
    renderNav("/image-registries");
    expect(screen.getByRole("link", { name: /image registries/i }).closest("a")).toHaveAttribute(
      "data-active",
      "true",
    );
  });
});

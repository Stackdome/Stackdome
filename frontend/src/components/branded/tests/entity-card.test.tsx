// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { EndpointPills, relativeAge } from "@/components/branded/entity-card";

afterEach(cleanup);

describe("relativeAge", () => {
  it("renders compact tokens", () => {
    const now = Date.now();
    expect(relativeAge(new Date(now - 10_000).toISOString())).toBe("just now");
    expect(relativeAge(new Date(now - 5 * 60_000).toISOString())).toBe("5m ago");
    expect(relativeAge(new Date(now - 5 * 3_600_000).toISOString())).toBe("5h ago");
    expect(relativeAge(new Date(now - 3 * 86_400_000).toISOString())).toBe("3d ago");
    expect(relativeAge(null)).toBeNull();
  });
});

describe("EndpointPills", () => {
  it("shows two pills and collapses the rest into a +N popover trigger", () => {
    const urls = [
      { resource: "web", url: "https://web.example.com" },
      { resource: "api", url: "https://api.example.com" },
      { resource: "docs", url: "https://docs.example.com" },
      { resource: "admin", url: "https://admin.example.com" },
    ];
    render(<EndpointPills urls={urls} />);
    expect(screen.getByRole("link", { name: /web/ })).toBeTruthy();
    expect(screen.getByRole("link", { name: /api/ })).toBeTruthy();
    expect(screen.queryByRole("link", { name: /docs/ })).toBeNull();
    expect(screen.getByRole("button", { name: "2 more endpoints" })).toBeTruthy();
  });
});

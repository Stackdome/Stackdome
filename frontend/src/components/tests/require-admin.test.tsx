// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import React from "react";

vi.mock("@/hooks/use-current-user", () => ({ useCurrentUser: vi.fn() }));
import { useCurrentUser } from "@/hooks/use-current-user";
import { RequireAdmin } from "@/components/require-admin";

afterEach(() => cleanup());

function renderAt(initial: string) {
  return render(
    <MemoryRouter initialEntries={[initial]}>
      <Routes>
        <Route path="/" element={<div>home</div>} />
        <Route element={<RequireAdmin />}>
          <Route path="/settings/users" element={<div>users-page</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("RequireAdmin", () => {
  it("renders the gated route for an OrgAdmin", () => {
    vi.mocked(useCurrentUser).mockReturnValue({ isOrgAdmin: true, loading: false } as never);
    renderAt("/settings/users");
    expect(screen.getByText("users-page")).toBeTruthy();
  });

  it("redirects a non-admin to home", () => {
    vi.mocked(useCurrentUser).mockReturnValue({ isOrgAdmin: false, loading: false } as never);
    renderAt("/settings/users");
    expect(screen.getByText("home")).toBeTruthy();
    expect(screen.queryByText("users-page")).toBeNull();
  });

  it("renders nothing while loading (no premature redirect)", () => {
    vi.mocked(useCurrentUser).mockReturnValue({ isOrgAdmin: false, loading: true } as never);
    renderAt("/settings/users");
    expect(screen.queryByText("users-page")).toBeNull();
    expect(screen.queryByText("home")).toBeNull();
  });
});

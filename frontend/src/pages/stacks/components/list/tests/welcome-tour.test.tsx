// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";

vi.mock("@/api/stacks", () => ({
  getStacksByOrg: vi.fn(),
  deleteStack: vi.fn(),
}));
vi.mock("@/hooks/use-preview-envs", () => ({
  usePreviewEnvs: () => ({ envs: [], loading: false }),
}));
vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ canWriteAnyProject: true, canWrite: () => true, isOrgAdmin: true }),
}));
vi.mock("@/pages/stacks/components/wizard/stack-create-wizard", () => ({
  StackCreateWizard: () => null,
}));

import { getStacksByOrg } from "@/api/stacks";
import { StackProvider } from "@/pages/stacks/contexts/stack-context";
import StacksPage from "../index";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/stacks"]}>
      <StackProvider>
        <Routes>
          <Route path="/stacks" element={<StacksPage />} />
        </Routes>
      </StackProvider>
    </MemoryRouter>,
  );
}

describe("first-run tour offer", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(getStacksByOrg).mockResolvedValue({ items: [], total: 0 } as never);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  // The org this runs against has no domain, which is every org on Cloud: the
  // platform base domain covers the demo's public port, so the offer stands.
  it("offers the tour to an org with no stacks and no domain", async () => {
    renderPage();
    await waitFor(() =>
      expect(document.querySelectorAll('[role="dialog"]').length).toBe(1),
    );
  });

  it("stays away once the tour has been retired", async () => {
    localStorage.setItem("stackdome.onboarding-tour.done", "1");
    renderPage();
    await waitFor(() => expect(getStacksByOrg).toHaveBeenCalled());
    expect(document.querySelectorAll('[role="dialog"]').length).toBe(0);
  });
});

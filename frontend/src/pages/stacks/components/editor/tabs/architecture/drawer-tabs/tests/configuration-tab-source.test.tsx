// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { Tabs } from "@/components/ui/tabs";
import { StackResourceConfigurationTab, pickConfigurationDraft } from "../configuration-tab";

afterEach(cleanup);

// The combobox's network behavior is covered by its own tests; here it is a
// controlled stub so we can assert the wiring contract.
vi.mock("@/components/git-source-picker/repo-combobox", () => ({
  RepoCombobox: ({ value, integrationId, onChange }: {
    value: string;
    integrationId?: string;
    onChange: (p: { repo_url: string; integration_id: string | undefined }) => void;
  }) => (
    <button
      data-testid="repo-combobox"
      data-value={value}
      data-integration={integrationId ?? ""}
      onClick={() => onChange({ repo_url: "https://github.com/acme/api.git", integration_id: "int-app" })}
    >
      pick
    </button>
  ),
}));

function renderGitTab(overrides: Record<string, unknown> = {}) {
  const onPatchResource = vi.fn();
  const resource = {
    name: "api",
    sourceType: "git" as const,
    source: { git: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." } },
    ...overrides,
  };
  render(
    // StackResourceConfigurationTab renders <TabsContent value="general">,
    // which requires a Radix <Tabs> ancestor (Tabs provides the context
    // TabsContent reads); the brief's starting spec rendered it bare and
    // failed with "TabsContent must be used within Tabs". Sibling test
    // env-addon-group-render.test.tsx wraps StackResourceEnvironmentTab the
    // same way — follow that convention here.
    <Tabs defaultValue="general">
      <StackResourceConfigurationTab
        draft={pickConfigurationDraft(resource)}
        baseline={pickConfigurationDraft(resource)}
        index={0}
        // `errors` and `volumes` are required props on StackResourceConfigurationTabProps
        // (unlike allResources/onDiscardField/onCreateVolume/onOpenVolume, which are
        // optional). getError() indexes into `errors` unconditionally, so an empty
        // object is required here rather than omitting the prop as the brief's
        // starting spec did.
        errors={{}}
        volumes={[]}
        onPatchResource={onPatchResource}
      />
    </Tabs>,
  );
  return { onPatchResource };
}

describe("git repository row", () => {
  it("renders the repo combobox with the current url", () => {
    renderGitTab({
      source: { git: { repo_url: "https://github.com/a/b.git", dockerfile_path: "Dockerfile", build_context: ".", integration_id: "int-1" } },
    });
    const combo = screen.getByTestId("repo-combobox");
    expect(combo).toHaveAttribute("data-value", "https://github.com/a/b.git");
    expect(combo).toHaveAttribute("data-integration", "int-1");
  });

  it("patches repo_url and integration_id together on pick", () => {
    const { onPatchResource } = renderGitTab();
    screen.getByTestId("repo-combobox").click();
    expect(onPatchResource).toHaveBeenCalledWith({
      source: {
        git: expect.objectContaining({
          repo_url: "https://github.com/acme/api.git",
          integration_id: "int-app",
        }),
      },
    });
  });
});

// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
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

function renderGitTab(
  overrides: Record<string, unknown> = {},
  options: { baselineOverrides?: Record<string, unknown>; onDiscardField?: (path: string) => void } = {},
) {
  const onPatchResource = vi.fn();
  const resource = {
    name: "api",
    sourceType: "git" as const,
    source: { git: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." } },
    ...overrides,
  };
  const baselineResource = {
    name: "api",
    sourceType: "git" as const,
    source: { git: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." } },
    ...(options.baselineOverrides ?? overrides),
  };
  const { rerender } = render(
    // StackResourceConfigurationTab renders <TabsContent value="general">,
    // which requires a Radix <Tabs> ancestor (Tabs provides the context
    // TabsContent reads); the brief's starting spec rendered it bare and
    // failed with "TabsContent must be used within Tabs". Sibling test
    // env-addon-group-render.test.tsx wraps StackResourceEnvironmentTab the
    // same way — follow that convention here.
    <Tabs defaultValue="general">
      <StackResourceConfigurationTab
        draft={pickConfigurationDraft(resource)}
        baseline={pickConfigurationDraft(baselineResource)}
        index={0}
        // `errors` and `volumes` are required props on StackResourceConfigurationTabProps
        // (unlike allResources/onDiscardField/onCreateVolume/onOpenVolume, which are
        // optional). getError() indexes into `errors` unconditionally, so an empty
        // object is required here rather than omitting the prop as the brief's
        // starting spec did.
        errors={{}}
        volumes={[]}
        onPatchResource={onPatchResource}
        onDiscardField={options.onDiscardField}
      />
    </Tabs>,
  );
  // Re-renders with a patched `git` source, simulating the parent store
  // applying an onPatchResource call back into draft (as it does in
  // production). Needed for onBlur-restores-default assertions: these are
  // controlled inputs, and without a real re-render carrying the new value,
  // React resets the DOM node's value back to the original `value` prop
  // right after the change event, so by the time blur fires the input no
  // longer reads as empty.
  const rerenderWithGitSource = (git: Record<string, unknown>) => {
    const next = {
      ...resource,
      source: { git: { ...(resource as { source: { git: Record<string, unknown> } }).source.git, ...git } },
    };
    rerender(
      <Tabs defaultValue="general">
        <StackResourceConfigurationTab
          draft={pickConfigurationDraft(next)}
          baseline={pickConfigurationDraft(baselineResource)}
          index={0}
          errors={{}}
          volumes={[]}
          onPatchResource={onPatchResource}
          onDiscardField={options.onDiscardField}
        />
      </Tabs>,
    );
  };
  // Simulates the parent store applying an onPatchResource call back into the
  // draft (production shallow-merges the patch onto the full resource — see
  // onPatchResource in use-resource-tab-props.ts). Needed for round-trip
  // assertions on the "Build from" toggle, which reads back stashed fields.
  const rerenderWithPatch = (patch: Record<string, unknown>) => {
    const next = { ...resource, ...patch };
    rerender(
      <Tabs defaultValue="general">
        <StackResourceConfigurationTab
          draft={pickConfigurationDraft(next)}
          baseline={pickConfigurationDraft(baselineResource)}
          index={0}
          errors={{}}
          volumes={[]}
          onPatchResource={onPatchResource}
          onDiscardField={options.onDiscardField}
        />
      </Tabs>,
    );
  };
  return { onPatchResource, rerenderWithGitSource, rerenderWithPatch };
}

vi.mock("../image-registry-select", () => ({
  ImageRegistrySelect: ({ imageRef, onChange }: {
    imageRef: string;
    onChange: (p: { ref: string; registry_credentials_id: string | undefined }) => void;
  }) => (
    <button
      data-testid="image-registry-select"
      data-ref={imageRef}
      onClick={() => onChange({ ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" })}
    >
      registry
    </button>
  ),
}));

function renderImageTab(
  overrides: Record<string, unknown> = {},
  options: { baselineOverrides?: Record<string, unknown>; onDiscardField?: (path: string) => void } = {},
) {
  const onPatchResource = vi.fn();
  const baseResource = {
    name: "api",
    sourceType: "image" as const,
    source: { image: { ref: "" } },
  };
  const draftResource = { ...baseResource, ...overrides };
  const baselineResource = { ...baseResource, ...(options.baselineOverrides ?? overrides) };
  const { rerender } = render(
    <Tabs defaultValue="general">
      <StackResourceConfigurationTab
        draft={pickConfigurationDraft(draftResource)}
        baseline={pickConfigurationDraft(baselineResource)}
        index={0}
        errors={{}}
        volumes={[]}
        onPatchResource={onPatchResource}
        onDiscardField={options.onDiscardField}
      />
    </Tabs>,
  );
  // Simulates the parent store applying an onPatchResource call back into the
  // draft (production shallow-merges the patch onto the full resource — see
  // onPatchResource in use-resource-tab-props.ts). Needed for round-trip
  // assertions on the "Build from" toggle, which reads back stashed fields.
  const rerenderWithPatch = (patch: Record<string, unknown>) => {
    const next = { ...draftResource, ...patch };
    rerender(
      <Tabs defaultValue="general">
        <StackResourceConfigurationTab
          draft={pickConfigurationDraft(next)}
          baseline={pickConfigurationDraft(baselineResource)}
          index={0}
          errors={{}}
          volumes={[]}
          onPatchResource={onPatchResource}
          onDiscardField={options.onDiscardField}
        />
      </Tabs>,
    );
  };
  return { onPatchResource, rerenderWithPatch };
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

  it("discards both repo_url and integration_id when the Repository row's reset is clicked", () => {
    const onDiscardField = vi.fn();
    renderGitTab(
      {
        source: {
          git: {
            repo_url: "https://github.com/acme/new-repo.git",
            dockerfile_path: "Dockerfile",
            build_context: ".",
            integration_id: "int-new",
          },
        },
      },
      {
        baselineOverrides: {
          source: {
            git: {
              repo_url: "https://github.com/acme/old-repo.git",
              dockerfile_path: "Dockerfile",
              build_context: ".",
              integration_id: "int-old",
            },
          },
        },
        onDiscardField,
      },
    );
    const resetButton = screen.getByRole("button", { name: /reset to original value/i });
    resetButton.click();
    expect(onDiscardField).toHaveBeenCalledWith("source.git.repo_url");
    expect(onDiscardField).toHaveBeenCalledWith("source.git.integration_id");
  });
});

describe("advanced build fields", () => {
  it("patches dockerfile_path", () => {
    const { onPatchResource } = renderGitTab();
    fireEvent.click(screen.getByText("Advanced"));
    const input = screen.getByLabelText(/dockerfile path/i);
    fireEvent.change(input, { target: { value: "docker/Dockerfile.prod" } });
    expect(onPatchResource).toHaveBeenCalledWith({
      source: { git: expect.objectContaining({ dockerfile_path: "docker/Dockerfile.prod" }) },
    });
  });

  it("patches build_context", () => {
    const { onPatchResource } = renderGitTab();
    fireEvent.click(screen.getByText("Advanced"));
    const input = screen.getByLabelText(/build context/i);
    fireEvent.change(input, { target: { value: "services/api" } });
    expect(onPatchResource).toHaveBeenCalledWith({
      source: { git: expect.objectContaining({ build_context: "services/api" }) },
    });
  });

  it("restores the default on blur when cleared", () => {
    const { onPatchResource, rerenderWithGitSource } = renderGitTab();
    fireEvent.click(screen.getByText("Advanced"));
    const input = screen.getByLabelText(/dockerfile path/i);
    fireEvent.change(input, { target: { value: "" } });
    // Carry the cleared value into a re-render (see renderGitTab) so the
    // controlled input actually reads "" when blur fires, matching how the
    // real app re-renders with the patched draft between change and blur.
    rerenderWithGitSource({ dockerfile_path: "" });
    fireEvent.blur(screen.getByLabelText(/dockerfile path/i));
    expect(onPatchResource).toHaveBeenLastCalledWith({
      source: { git: expect.objectContaining({ dockerfile_path: "Dockerfile" }) },
    });
  });
});

describe("image source rows", () => {
  it("renders the registry select with the full ref", () => {
    renderImageTab({ source: { image: { ref: "ghcr.io/acme/api:1" } } });
    expect(screen.getByTestId("image-registry-select")).toHaveAttribute("data-ref", "ghcr.io/acme/api:1");
  });

  it("patches ref and registry_credentials_id together from the select", () => {
    const { onPatchResource } = renderImageTab({ source: { image: { ref: "acme/api:1" } } });
    screen.getByTestId("image-registry-select").click();
    expect(onPatchResource).toHaveBeenCalledWith({
      source: {
        image: expect.objectContaining({ ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" }),
      },
    });
  });

  it("discards the whole image source when the Registry row's reset is clicked", () => {
    const onDiscardField = vi.fn();
    renderImageTab(
      { source: { image: { ref: "ghcr.io/acme/api:2", registry_credentials_id: "cred-new" } } },
      {
        baselineOverrides: { source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-old" } } },
        onDiscardField,
      },
    );
    // Both the Registry row and the Image reference row wrap "source.image"
    // in their own <DirtyField>, so a dirty ref or credential surfaces a
    // reset button on each — in JSX/DOM order, index 0 is the Registry row's,
    // index 1 the Image reference row's. Discarding the whole "source.image"
    // subtree in one call (rather than two leaf calls) means a credential-only
    // change is covered by the same reset, since it dirties the same path.
    const resetButtons = screen.getAllByRole("button", { name: /reset to original value/i });
    expect(resetButtons).toHaveLength(2);

    resetButtons[0].click();
    expect(onDiscardField).toHaveBeenCalledWith("source.image");
    expect(onDiscardField).toHaveBeenCalledTimes(1);
  });

  it("discards the whole image source when the Image reference row's reset is clicked", () => {
    const onDiscardField = vi.fn();
    renderImageTab(
      { source: { image: { ref: "ghcr.io/acme/api:2", registry_credentials_id: "cred-new" } } },
      {
        baselineOverrides: { source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-old" } } },
        onDiscardField,
      },
    );
    const resetButtons = screen.getAllByRole("button", { name: /reset to original value/i });
    expect(resetButtons).toHaveLength(2);

    resetButtons[1].click();
    expect(onDiscardField).toHaveBeenCalledWith("source.image");
    expect(onDiscardField).toHaveBeenCalledTimes(1);
  });

  it("shows both rows as dirty when only the credential changes and the ref stays the same", () => {
    const onDiscardField = vi.fn();
    renderImageTab(
      { source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-new" } } },
      {
        baselineOverrides: { source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-old" } } },
        onDiscardField,
      },
    );
    // Before the fix, both DirtyFields wrapped "source.image.ref" only, so a
    // credential-only change (same ref, different registry_credentials_id)
    // never surfaced a reset button on either row.
    const resetButtons = screen.getAllByRole("button", { name: /reset to original value/i });
    expect(resetButtons).toHaveLength(2);
  });

  it("shows only the remainder in the ref input and recomposes the host on change", () => {
    const { onPatchResource } = renderImageTab({
      source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" } },
    });
    const input = screen.getByLabelText(/image reference/i) as HTMLInputElement;
    expect(input.value).toBe("acme/api:1");
    fireEvent.change(input, { target: { value: "acme/api:2" } });
    expect(onPatchResource).toHaveBeenCalledWith({
      source: { image: expect.objectContaining({ ref: "ghcr.io/acme/api:2" }) },
    });
  });

  it("replaces the whole ref instead of doubling the host when a full ref is pasted", () => {
    const { onPatchResource } = renderImageTab({
      source: { image: { ref: "ghcr.io/acme/api:1" } },
    });
    const input = screen.getByLabelText(/image reference/i) as HTMLInputElement;
    fireEvent.change(input, { target: { value: "quay.io/other/app:2" } });
    expect(onPatchResource).toHaveBeenCalledWith({
      source: { image: expect.objectContaining({ ref: "quay.io/other/app:2" }) },
    });
  });
});

describe("Build from toggle — source preservation", () => {
  it("stashes the image source when toggling to git, and restores it exactly when toggling back", () => {
    const { onPatchResource, rerenderWithPatch } = renderImageTab({
      source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" } },
    });

    // image -> git: the image source is stashed, git gets fresh defaults.
    screen.getByRole("radio", { name: /git repository/i }).click();
    expect(onPatchResource).toHaveBeenLastCalledWith({
      sourceType: "git",
      source: { git: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." } },
      stashedImageSource: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" },
    });

    const lastPatch = onPatchResource.mock.calls[onPatchResource.mock.calls.length - 1][0];
    rerenderWithPatch(lastPatch);

    // git -> image: the exact stashed image source is restored, and the
    // (untouched, default) git source is stashed in turn.
    screen.getByRole("radio", { name: /container image/i }).click();
    expect(onPatchResource).toHaveBeenLastCalledWith({
      sourceType: "image",
      source: { image: { ref: "ghcr.io/acme/api:1", registry_credentials_id: "cred-ghcr" } },
      stashedGitSource: { repo_url: "", dockerfile_path: "Dockerfile", build_context: "." },
    });
  });
});

describe("Build from toggle — git source preservation", () => {
  it("stashes the git source when toggling to image, and restores it exactly when toggling back", () => {
    const { onPatchResource, rerenderWithPatch } = renderGitTab({
      source: {
        git: {
          repo_url: "https://github.com/acme/api.git",
          dockerfile_path: "Dockerfile",
          build_context: ".",
          integration_id: "int-app",
        },
      },
    });

    // git -> image: the git source is stashed, image gets fresh defaults.
    screen.getByRole("radio", { name: /container image/i }).click();
    expect(onPatchResource).toHaveBeenLastCalledWith({
      sourceType: "image",
      source: { image: { ref: "" } },
      stashedGitSource: {
        repo_url: "https://github.com/acme/api.git",
        dockerfile_path: "Dockerfile",
        build_context: ".",
        integration_id: "int-app",
      },
    });

    const lastPatch = onPatchResource.mock.calls[onPatchResource.mock.calls.length - 1][0];
    rerenderWithPatch(lastPatch);

    // image -> git: the exact stashed git source is restored (repo_url +
    // integration_id survive), and the default image source is stashed.
    screen.getByRole("radio", { name: /git repository/i }).click();
    expect(onPatchResource).toHaveBeenLastCalledWith({
      sourceType: "git",
      source: {
        git: {
          repo_url: "https://github.com/acme/api.git",
          dockerfile_path: "Dockerfile",
          build_context: ".",
          integration_id: "int-app",
        },
      },
      stashedImageSource: { ref: "" },
    });
  });
});

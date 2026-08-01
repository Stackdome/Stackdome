import { describe, expect, it } from "vitest";
import type { Stack } from "@/api/stacks";
import { serverStateFromStack } from "../server-state";
import { buildDesiredState } from "../desired-state";
import { computeSyncOps } from "../ops";
import { draftToSnapshot } from "../draft-snapshot";
import { formResourcesFromSpec, mapVolumeToFormData } from "@/pages/stacks/lib/spec-to-form";
import { diffSnapshots } from "@/pages/stacks/components/editor/tabs/deployments/release-snapshot-diff";

/** A git resource as the API returns it when the build paths were never set —
 *  the API applies its own defaults, so an unset path is not a user edit. */
const bareGitResource = {
  id: "r-worker",
  stack_id: "s1",
  name: "worker",
  workload_type: "Service",
  source: { git: { repo_url: "https://github.com/acme/demo.git", branch: "main" } },
  ports: [],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: [], args: [], environment_variables: [] },
};

const stack = {
  id: "s1",
  name: "demo",
  spec: { stack_resources: [bareGitResource], volumes: [], connections: [] },
} as unknown as Stack;

function draftFrom(s: Stack) {
  return {
    resources: formResourcesFromSpec(s.spec?.stack_resources, s.spec?.connections),
    volumes: (s.spec?.volumes ?? []).map(mapVolumeToFormData),
  };
}

describe("git build defaults", () => {
  it("autosaves nothing for a git resource loaded straight from the server", () => {
    const ops = computeSyncOps(
      serverStateFromStack(stack),
      buildDesiredState(draftFrom(stack) as never),
    );
    expect(ops).toEqual([]);
  });

  it("stages no diff against a release snapshot that omits the build paths", () => {
    const live = { resources: [bareGitResource], volumes: [], connections: [] };
    const diff = diffSnapshots(live as never, draftToSnapshot(draftFrom(stack) as never));
    expect(diff).toEqual({ resources: [], volumes: [], connections: [] });
  });
});

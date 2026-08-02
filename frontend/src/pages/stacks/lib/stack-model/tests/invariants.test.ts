import { describe, expect, it } from "vitest";
import type { Stack } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";
import { formResourcesFromSpec, mapVolumeToFormData } from "@/pages/stacks/lib/spec-to-form";
import { canonicalFromSnapshot, canonicalFromStack } from "../from-api";
import { canonicalFromDraft } from "../from-form";
import { diffStacks, isEmptyDiff } from "../diff";
import { draftToSnapshot } from "@/pages/stacks/lib/draft-sync/draft-snapshot";
import { bumpImage, firstImageResourceName, templateDrafts } from "./template-drafts";

const imageResource = {
  id: "r-web",
  stack_id: "s1",
  name: "web",
  workload_type: "Service",
  source: { image: { ref: "quay.io/stackdome/hello-stack-web" } },
  ports: [{ name: "http-3000", number: 3000, protocol: "http", exposed_to_public: true }],
  volume_mounts: [],
  depends_on: ["redis"],
  execution_config: {
    command: ["node", "server.js"],
    args: [],
    environment_variables: [
      { name: "CELEBRATION", value: "confetti" },
      { name: "PUBLIC_URL", self_output: "public_url" },
    ],
  },
};

const gitResource = {
  id: "r-worker",
  stack_id: "s1",
  name: "worker",
  workload_type: "Service",
  source: {
    git: {
      repo_url: "https://github.com/acme/demo.git",
      branch: "main",
      dockerfile_path: "Dockerfile",
      build_context: "worker",
    },
  },
  ports: [],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: [], args: [], environment_variables: [] },
};

/** A git source without the `dockerfile_path`/`build_context` the form fills in. */
const bareGitResource = {
  ...gitResource,
  source: { git: { repo_url: "https://github.com/acme/demo.git", branch: "main" } },
};

/** No `workload_type`; the form side zod-defaults it. */
const untypedResource = (() => {
  const { workload_type, ...rest } = imageResource;
  void workload_type;
  return rest;
})();

/** Carries the server-written fields: `revision`, `status`, `outputs`. */
const resourceWithServerFields = {
  ...imageResource,
  revision: 7,
  status: { state: "Running", replicas: 1 },
  outputs: [
    { name: "url", type: "string", sensitive: false },
    { name: "public_url", type: "string", sensitive: false },
  ],
};

const redisResource = {
  id: "r-redis",
  stack_id: "s1",
  name: "redis",
  workload_type: "Service",
  source: { image: { ref: "redis:7-alpine" } },
  ports: [{ name: "tcp-6379", number: 6379, protocol: "tcp", exposed_to_public: false }],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: ["redis-server", "--appendonly", "yes"], args: [], environment_variables: [] },
};

const volume = {
  id: "v1",
  name: "redis-data",
  spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false },
};

/** Mounts live in connections; the resource itself always reports `[]`. */
const mountConnection = {
  id: "c-mount",
  kind: "volume_mount",
  from: { type: "volume", name: "redis-data" },
  to: { type: "stack_resource", name: "redis" },
  config: { mount_path: "/data" },
};

const resourceRefConnection = {
  id: "c-env",
  kind: "env",
  from: { type: "stack_resource", name: "redis" },
  to: { type: "stack_resource", name: "web" },
  mappings: [{ target: { type: "env", name: "REDIS_URL" }, value: { output: "url" } }],
};

const secretRefConnection = {
  id: "c-secret",
  kind: "env",
  from: { type: "secret", id: "sec-1" },
  to: { type: "stack_resource", name: "web" },
  mappings: [{ target: { type: "env", name: "API_KEY" }, value: { output: "key" } }],
};

const addonRefConnection = {
  id: "c-addon",
  kind: "env",
  from: { type: "addon/postgres", id: "addon-1" },
  to: { type: "stack_resource", name: "web" },
  config: { database: "app", superuser: false },
  mappings: [{ target: { type: "env", name: "DATABASE_URL" }, value: { output: "connectionString" } }],
};

function stackOf(resources: unknown[], volumes: unknown[] = [], connections: unknown[] = []): Stack {
  return {
    id: "s1",
    name: "demo",
    spec: { stack_resources: resources, volumes, connections },
  } as unknown as Stack;
}

const CORPUS: Record<string, Stack> = {
  "image resource": stackOf([imageResource]),
  "git resource": stackOf([gitResource]),
  "git resource with no build paths": stackOf([bareGitResource]),
  "resource with no workload_type": stackOf([untypedResource]),
  "resource carrying server telemetry": stackOf([resourceWithServerFields]),
  "volume mounted through a connection": stackOf([redisResource], [volume], [mountConnection]),
  "resource reference in env": stackOf([imageResource, redisResource], [], [resourceRefConnection]),
  "secret reference in env": stackOf([imageResource], [], [secretRefConnection]),
  "addon reference in env": stackOf([imageResource], [], [addonRefConnection]),
  "the full demo stack": stackOf(
    [imageResource, gitResource, redisResource],
    [volume],
    [mountConnection, resourceRefConnection],
  ),
  "empty stack": stackOf([]),
};

function draftOf(stack: Stack) {
  return {
    resources: formResourcesFromSpec(stack.spec?.stack_resources, stack.spec?.connections),
    volumes: (stack.spec?.volumes ?? []).map(mapVolumeToFormData),
  };
}

describe("canonical model invariants", () => {
  describe("the read path and the write path agree", () => {
    for (const [label, stack] of Object.entries(CORPUS)) {
      it(label, () => {
        expect(canonicalFromDraft(draftOf(stack))).toMatchObject({
          resources: canonicalFromStack(stack).resources,
          volumes: canonicalFromStack(stack).volumes,
        });
      });
    }
  });

  describe("loading a stack and touching nothing changes nothing", () => {
    for (const [label, stack] of Object.entries(CORPUS)) {
      it(label, () => {
        const diff = diffStacks(canonicalFromStack(stack), canonicalFromDraft(draftOf(stack)));
        expect(diff).toEqual({ resources: [], volumes: [] });
      });
    }
  });

  describe("a draft matching its deployed release is not staged", () => {
    for (const [label, stack] of Object.entries(CORPUS)) {
      it(label, () => {
        const snapshot = {
          resources: stack.spec?.stack_resources,
          volumes: stack.spec?.volumes,
          connections: stack.spec?.connections,
        } as unknown as StackReleaseSnapshot;
        const diff = diffStacks(canonicalFromSnapshot(snapshot), canonicalFromDraft(draftOf(stack)), {
          baselineIsRelease: true,
        });
        expect(isEmptyDiff(diff)).toBe(true);
      });
    }
  });

  it("ignores a revision the pin resolver wrote for a key the spec leaves open", () => {
    const stack = stackOf([bareGitResource]);
    const deployed = {
      resources: [
        {
          ...bareGitResource,
          source: {
            git: { ...bareGitResource.source.git, commit: "9f1c2b7ae4d0", dockerfile_path: "Dockerfile", build_context: "." },
          },
        },
      ],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const diff = diffStacks(canonicalFromSnapshot(deployed), canonicalFromDraft(draftOf(stack)), {
      baselineIsRelease: true,
    });
    expect(isEmptyDiff(diff)).toBe(true);
  });

  it("still reports a revision the spec pins itself", () => {
    const pinned = {
      ...bareGitResource,
      source: { git: { ...bareGitResource.source.git, commit: "aaaaaaa" } },
    };
    const deployed = {
      resources: [{ ...pinned, source: { git: { ...pinned.source.git, commit: "bbbbbbb" } } }],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const diff = diffStacks(canonicalFromSnapshot(deployed), canonicalFromDraft(draftOf(stackOf([pinned]))), {
      baselineIsRelease: true,
    });
    expect(diff.resources).toHaveLength(1);
    expect(diff.resources[0].fields.map((f) => f.path)).toContain("source.git.commit");
  });
});

describe("the diff still sees real edits", () => {
  const base = canonicalFromStack(stackOf([imageResource, redisResource], [volume], [resourceRefConnection]));

  it("reports a changed image", () => {
    const edited = canonicalFromStack(
      stackOf([{ ...imageResource, source: { image: { ref: "web:v2" } } }, redisResource], [volume], [resourceRefConnection]),
    );
    const [d] = diffStacks(base, edited).resources;
    expect(d).toMatchObject({ name: "web", change: "modified" });
    expect(d.fields).toEqual([
      expect.objectContaining({ path: "source.image.ref", from: "quay.io/stackdome/hello-stack-web", to: "web:v2" }),
    ]);
  });

  it("reports an env var edited through a connection", () => {
    const edited = canonicalFromStack(
      stackOf([imageResource, redisResource], [volume], [
        { ...resourceRefConnection, mappings: [{ target: { type: "env", name: "REDIS_URL" }, value: { output: "public_url" } }] },
      ]),
    );
    const [d] = diffStacks(base, edited).resources;
    expect(d.fields).toEqual([
      expect.objectContaining({ path: "env.REDIS_URL", section: "environment" }),
    ]);
  });

  it("reports a removed resource, and the reference it took with it", () => {
    const edited = canonicalFromStack(stackOf([imageResource], [volume], []));
    const diff = diffStacks(base, edited);
    expect(diff.resources.map((r) => [r.name, r.change]).sort()).toEqual([
      ["redis", "removed"],
      ["web", "modified"],
    ]);
    const web = diff.resources.find((r) => r.name === "web")!;
    expect(web.fields).toEqual([
      expect.objectContaining({ path: "env.REDIS_URL", kind: "removed" }),
    ]);
  });

  it("collapses a rename into one entry", () => {
    const edited = canonicalFromStack(stackOf([imageResource, { ...redisResource, name: "cache" }], [volume], []));
    const diff = diffStacks(canonicalFromStack(stackOf([imageResource, redisResource], [volume], [])), edited);
    expect(diff.resources).toEqual([
      expect.objectContaining({ name: "cache", fromName: "redis", change: "renamed" }),
    ]);
  });
});

describe("every shipped template", () => {
  const BUMPED = "example.test/bumped:v2";

  for (const [id, draft] of templateDrafts()) {
    describe(id, () => {
      const deployed = draftToSnapshot(draft);

      it("is not staged against the release it produced", () => {
        const diff = diffStacks(canonicalFromSnapshot(deployed), canonicalFromDraft(draft), {
          baselineIsRelease: true,
        });
        expect(diff).toEqual({ resources: [], volumes: [] });
      });

      it("stages one resource, and only that one, when an image is retagged", () => {
        const name = firstImageResourceName(draft);
        const diff = diffStacks(
          canonicalFromSnapshot(deployed),
          canonicalFromDraft(bumpImage(draft, name, BUMPED)),
          { baselineIsRelease: true },
        );
        expect(diff.resources.map((r) => r.name)).toEqual([name]);
        expect(diff.resources[0].fields).toEqual([
          expect.objectContaining({ path: "source.image.ref", to: BUMPED }),
        ]);
      });

      it("is clean again once that retag is deployed", () => {
        const bumped = bumpImage(draft, firstImageResourceName(draft), BUMPED);
        const diff = diffStacks(canonicalFromSnapshot(draftToSnapshot(bumped)), canonicalFromDraft(bumped), {
          baselineIsRelease: true,
        });
        expect(diff).toEqual({ resources: [], volumes: [] });
      });
    });
  }
});

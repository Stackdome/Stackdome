import type { DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

/** The public demo repo the cluster clones and builds at deploy time. */
export const HELLO_STACK_REPO_URL = "https://github.com/Stackdome/stackdome-demo";
/** Published image for web, so the demo shows both ways to source a resource. */
export const HELLO_STACK_WEB_IMAGE = "quay.io/stackdome/hello-stack-web";
const HELLO_STACK_NAME = "hello-stack";
const HELLO_STACK_BRANCH = "main";

const REDIS_VOLUME = "redis-data";
const REDIS_URL = "redis://redis:6379";

function gitService(
  name: string,
  buildContext: string,
  overrides: Partial<FormStackResourceData>,
): FormStackResourceData {
  return {
    name,
    workload_type: "Service",
    sourceType: "git",
    labels: [],
    depends_on: ["redis"],
    ports: [],
    execution_config: { environment_variables: [] },
    gitRevisionType: "branch",
    gitRevisionValue: HELLO_STACK_BRANCH,
    source: {
      git: {
        repo_url: HELLO_STACK_REPO_URL,
        dockerfile_path: "Dockerfile",
        build_context: buildContext,
      },
    },
    ...overrides,
  } as FormStackResourceData;
}

/** Draft seed for the onboarding demo stack — same shape template and git
    imports hand to /stacks/new via navigation state. Pure — no I/O. */
export function buildHelloStackSeed(): DraftSeed {
  return {
    name: HELLO_STACK_NAME,
    labels: [],
    resources: [
      {
        name: "web",
        workload_type: "Service",
        sourceType: "image",
        labels: [],
        depends_on: ["redis"],
        source: { image: { ref: HELLO_STACK_WEB_IMAGE } },
        ports: [{ name: "http-3000", number: 3000, protocol: "http", exposed_to_public: true }],
        execution_config: {
          environment_variables: [
            { from: "stack", name: "CELEBRATION", value: "confetti" },
            { from: "stack", name: "HAT", value: "party" },
            { from: "stack", name: "HEADLINE", value: "Your stack is now live." },
            { from: "stack", name: "REDIS_URL", value: REDIS_URL },
            // Self-reference resolved at deploy: the page shows its real
            // public address without inferring it client-side.
            { from: "resource", name: "PUBLIC_URL", resourceName: "web", output: "public_url" },
          ],
        },
      } as FormStackResourceData,
      // No ports on the worker — it is reachable by nobody, on purpose.
      gitService("worker", "hello-stack/worker", {
        execution_config: {
          environment_variables: [{ from: "stack", name: "REDIS_URL", value: REDIS_URL }],
        },
      }),
      {
        name: "redis",
        workload_type: "Service",
        sourceType: "image",
        labels: [],
        depends_on: [],
        ports: [{ name: "tcp-6379", number: 6379, protocol: "tcp", exposed_to_public: false }],
        execution_config: {
          command: "redis-server --appendonly yes",
          environment_variables: [],
        },
        source: { image: { ref: "redis:7-alpine" } },
        volume_mounts: [{ source_volume_name: REDIS_VOLUME, target_path: "/data" }],
      } as FormStackResourceData,
    ],
    volumes: [
      {
        name: REDIS_VOLUME,
        sourceType: "None",
        labels: [],
        spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false },
      },
    ],
    linkedAddonIds: [],
  };
}

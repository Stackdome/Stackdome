import type { DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

export const HELLO_STACK_WEB_IMAGE = "quay.io/stackdome/hello-stack-web";
export const HELLO_STACK_WORKER_IMAGE = "quay.io/stackdome/hello-stack-worker";
const HELLO_STACK_NAME = "hello-stack";

const REDIS_VOLUME = "redis-data";
const REDIS_ADDRESS = {
  from: "resource",
  name: "REDIS_URL",
  resourceName: "redis",
  output: "url",
} as const;

/** Same shape the template and git imports hand to /stacks/new via navigation state. */
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
          // Same as the image's own default — spelled out so the deployment
          // tab has something real to show.
          command: "node server.js",
          environment_variables: [
            { from: "stack", name: "CELEBRATION", value: "confetti" },
            { from: "stack", name: "HAT", value: "party" },
            { from: "stack", name: "HEADLINE", value: "Your stack is now live." },
            REDIS_ADDRESS,
            { from: "self", name: "PUBLIC_URL", selfOutput: "public_url" },
          ],
        },
      } as FormStackResourceData,
      {
        name: "worker",
        workload_type: "Service",
        sourceType: "image",
        labels: [],
        depends_on: ["redis"],
        // No ports on purpose — the tour teaches private resources.
        ports: [],
        source: { image: { ref: HELLO_STACK_WORKER_IMAGE } },
        execution_config: { environment_variables: [REDIS_ADDRESS] },
      } as FormStackResourceData,
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

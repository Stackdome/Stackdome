import { sanitizeKubernetesName } from "@/lib/docker-compose-converter";
import type { PickedRepo } from "@/components/git-source-picker/types";

export interface GitServiceForm {
  serviceName: string;
  branch: string;
  dockerfilePath: string;
  buildContext: string;
  port: number;
  exposePublic: boolean;
}

/**
 * `sanitizeKubernetesName` targets the RFC 1123 *subdomain* rules (dots
 * allowed) that volume/PVC names use. A Service object name is stricter —
 * RFC 1035 label, no dots — so repo names like "my-web.app" need the dot
 * folded into a hyphen too, or the generated Service CR is rejected by the
 * cluster.
 */
function sanitizeServiceName(name: string): string {
  return sanitizeKubernetesName(name).replace(/\.+/g, "-");
}

/** Default service/stack name for a repo: last path segment, k8s-sanitized. */
export function defaultServiceName(repo: PickedRepo): string {
  return sanitizeServiceName(repo.fullName.split("/").pop() ?? "service");
}

/** Builds the navigation-state seed consumed by /stacks/new (same shape the
    template and compose imports produce). Pure — no I/O. */
export function buildGitSeed(repo: PickedRepo, form: GitServiceForm) {
  const name = sanitizeServiceName(form.serviceName.trim() || defaultServiceName(repo));
  return {
    name,
    labels: [] as { key: string; value: string }[],
    resources: [
      {
        name,
        workload_type: "Service",
        sourceType: "git" as const,
        labels: [],
        depends_on: [],
        ports: [
          {
            name: `http-${form.port}`,
            number: form.port,
            protocol: "http" as const,
            exposed_to_public: form.exposePublic,
          },
        ],
        execution_config: { environment_variables: [] },
        gitRevisionType: "branch" as const,
        gitRevisionValue: form.branch,
        source: {
          git: {
            repo_url: repo.cloneUrl,
            dockerfile_path: form.dockerfilePath.trim() || "Dockerfile",
            build_context: form.buildContext.trim() || ".",
            ...(repo.integrationId ? { integration_id: repo.integrationId } : {}),
          },
        },
      },
    ],
    volumes: [],
    linkedAddonIds: [] as string[],
  };
}

import type { RegistryCredentialPurpose } from "@/api/registry-credentials";

export const PURPOSE_PULL: RegistryCredentialPurpose = "pull";
export const PURPOSE_PUSH: RegistryCredentialPurpose = "push";
export const PURPOSE_BOTH: RegistryCredentialPurpose = "both";

export const PURPOSE_LABELS: Record<RegistryCredentialPurpose, string> = {
  [PURPOSE_PULL]: "Pull only",
  [PURPOSE_PUSH]: "Push only",
  [PURPOSE_BOTH]: "Pull & push",
};

export type RegistryProviderId = "dockerhub" | "ghcr" | "gitlab" | "quay" | "other";

export interface RegistryProvider {
  id: RegistryProviderId;
  label: string;
  hostPrefill: string;
  hostPlaceholder: string;
  hint: string;
}

export const REGISTRY_PROVIDERS: RegistryProvider[] = [
  {
    id: "dockerhub",
    label: "Docker Hub",
    hostPrefill: "docker.io",
    hostPlaceholder: "docker.io",
    hint: "Use a Docker Hub access token instead of your account password.",
  },
  {
    id: "ghcr",
    label: "GHCR",
    hostPrefill: "ghcr.io",
    hostPlaceholder: "ghcr.io",
    hint: "Use a GitHub personal access token with read:packages (and write:packages for pushes).",
  },
  {
    id: "gitlab",
    label: "GitLab Registry",
    hostPrefill: "registry.gitlab.com",
    hostPlaceholder: "registry.gitlab.com",
    hint: "Use a deploy token or personal access token with read_registry / write_registry scope.",
  },
  {
    id: "quay",
    label: "Quay",
    hostPrefill: "quay.io",
    hostPlaceholder: "quay.io",
    hint: "Use a robot account token with repository access.",
  },
  {
    id: "other",
    label: "Other",
    hostPrefill: "",
    hostPlaceholder: "registry.example.com",
    hint: "Any OCI registry reachable over HTTPS with username/password or token auth.",
  },
];

export function providerIdForHost(host?: string): RegistryProviderId {
  const h = (host ?? "").toLowerCase();
  if (h.includes("docker.io")) return "dockerhub";
  if (h.includes("ghcr.io")) return "ghcr";
  if (h.includes("gitlab")) return "gitlab";
  if (h.includes("quay")) return "quay";
  return "other";
}

import { Container } from "lucide-react";
import dockerUrl from "@/assets/brand/docker.svg";
import dockerLightUrl from "@/assets/brand/docker-light.svg";
import githubUrl from "@/assets/brand/github.svg";
import githubLightUrl from "@/assets/brand/github-light.svg";
import gitlabUrl from "@/assets/brand/gitlab.svg";
import type { RegistryProviderId } from "../lib/providers";

// `light` renders in light mode (chip is light), `dark` renders in dark mode
// (chip is dark). Docker and GitHub ship separate marks for each theme; GitLab
// is colorful enough to reuse the same art for both.
const BRAND: Record<Exclude<RegistryProviderId, "quay" | "other">, { light: string; dark: string }> = {
  dockerhub: { light: dockerUrl, dark: dockerLightUrl },
  ghcr: { light: githubUrl, dark: githubLightUrl },
  gitlab: { light: gitlabUrl, dark: gitlabUrl },
};

/** Docker Hub, GHCR, and GitLab Registry use vendored brand marks (selfh.st);
 *  Quay has no icon in that set, so it and unknown registries fall back to a
 *  generic container glyph. */
export function ProviderLogo({ providerId, className }: { providerId: RegistryProviderId; className?: string }) {
  if (providerId === "dockerhub" || providerId === "ghcr" || providerId === "gitlab") {
    const brand = BRAND[providerId];
    return (
      <>
        <img src={brand.light} alt="" aria-hidden className={`object-contain dark:hidden ${className ?? ""}`} />
        <img
          src={brand.dark}
          alt=""
          aria-hidden
          className={`hidden object-contain dark:block ${className ?? ""}`}
        />
      </>
    );
  }
  return <Container className={className} aria-hidden />;
}

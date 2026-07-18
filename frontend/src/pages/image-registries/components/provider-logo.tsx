import { Container } from "lucide-react";
import githubUrl from "@/assets/brand/github.svg";
import githubLightUrl from "@/assets/brand/github-light.svg";
import gitlabUrl from "@/assets/brand/gitlab.svg";
import type { RegistryProviderId } from "../lib/providers";

// `light` renders in light mode (chip is light), `dark` renders in dark mode
// (chip is dark). GitHub ships separate marks for each theme; GitLab is
// colorful enough to reuse the same art for both.
const BRAND: Record<Exclude<RegistryProviderId, "dockerhub" | "quay" | "other">, { light: string; dark: string }> = {
  ghcr: { light: githubUrl, dark: githubLightUrl },
  gitlab: { light: gitlabUrl, dark: gitlabUrl },
};

/** GHCR and GitLab Registry reuse the vendored git-provider marks; Docker Hub,
 *  Quay, and unknown registries fall back to a generic container glyph until
 *  their brand SVGs are vendored. */
export function ProviderLogo({ providerId, className }: { providerId: RegistryProviderId; className?: string }) {
  if (providerId === "ghcr" || providerId === "gitlab") {
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

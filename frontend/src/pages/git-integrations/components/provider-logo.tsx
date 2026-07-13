import { GitBranch } from "lucide-react";
import githubUrl from "@/assets/brand/github.svg";
import githubLightUrl from "@/assets/brand/github-light.svg";
import gitlabUrl from "@/assets/brand/gitlab.svg";
import bitbucketUrl from "@/assets/brand/bitbucket.svg";
import giteaUrl from "@/assets/brand/gitea.svg";
import type { ProviderId } from "../lib/derive-row";

// `light` renders in light mode (chip is light), `dark` renders in dark mode
// (chip is dark). GitHub ships separate marks for each theme; the rest are
// colorful enough to reuse the same art for both.
const BRAND: Record<Exclude<ProviderId, "other">, { light: string; dark: string }> = {
  github: { light: githubUrl, dark: githubLightUrl },
  gitlab: { light: gitlabUrl, dark: gitlabUrl },
  bitbucket: { light: bitbucketUrl, dark: bitbucketUrl },
  gitea: { light: giteaUrl, dark: giteaUrl },
};

export function ProviderLogo({ providerId, className }: { providerId: ProviderId; className?: string }) {
  if (providerId === "other") {
    return <GitBranch className={className} aria-hidden />;
  }
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

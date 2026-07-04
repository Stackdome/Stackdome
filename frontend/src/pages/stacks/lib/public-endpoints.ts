/** Matches EncodeStackResourceSubdomainPrefix output: 16-char lowercase no-pad std base32. */
const GENERATED_LABEL = /^[a-z2-7]{16}$/;

export type UrlClass = "custom" | "prefix" | "generated";

export function classifyIngressUrl(url: string, orgDomains: string[]): UrlClass {
  let host: string;
  try {
    host = new URL(url).hostname;
  } catch {
    return "custom"; // show unparseable urls as-is rather than hide them
  }
  const under = orgDomains.find((d) => host === d || host.endsWith(`.${d}`));
  if (!under) return "custom";
  const firstLabel = host.slice(0, host.length - under.length).replace(/\.$/, "").split(".")[0] ?? "";
  return GENERATED_LABEL.test(firstLabel) ? "generated" : "prefix";
}

const CLASS_RANK: Record<UrlClass, number> = { custom: 0, prefix: 1, generated: 2 };

/** Best URL for the header pill: custom domain > subdomain-prefix > generated hash. */
export function pickBestIngress(
  ingresses: { url?: string; target_port?: number }[],
  orgDomains: string[],
): { url: string; target_port?: number } | null {
  let best: { url: string; target_port?: number } | null = null;
  let bestRank = Number.POSITIVE_INFINITY;
  for (const ing of ingresses) {
    if (!ing.url) continue;
    const rank = CLASS_RANK[classifyIngressUrl(ing.url, orgDomains)];
    if (rank < bestRank) {
      best = { url: ing.url, target_port: ing.target_port };
      bestRank = rank;
    }
  }
  return best;
}

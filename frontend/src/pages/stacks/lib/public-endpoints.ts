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

export interface SortedIngress {
  url: string;
  target_port?: number;
}

/** All public URLs, best-first: custom domain > subdomain-prefix > generated
 *  hash, lowest target_port breaking ties — deterministic so the "shown" URL
 *  never depends on server array order. Url-less entries are dropped. */
export function sortIngresses(
  ingresses: { url?: string; target_port?: number }[],
  orgDomains: string[],
): SortedIngress[] {
  return ingresses
    .filter((i): i is SortedIngress => !!i.url)
    .map((i) => ({ ing: { url: i.url, target_port: i.target_port }, rank: CLASS_RANK[classifyIngressUrl(i.url, orgDomains)] }))
    .sort(
      (a, b) =>
        a.rank - b.rank ||
        (a.ing.target_port ?? Number.POSITIVE_INFINITY) - (b.ing.target_port ?? Number.POSITIVE_INFINITY),
    )
    .map((e) => e.ing);
}

/** Best URL for the header pill — first entry of sortIngresses. */
export function pickBestIngress(
  ingresses: { url?: string; target_port?: number }[],
  orgDomains: string[],
): SortedIngress | null {
  return sortIngresses(ingresses, orgDomains)[0] ?? null;
}

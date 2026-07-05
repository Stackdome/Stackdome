/**
 * Pure presentation mapper: turns a resource's raw config (image, ports, build)
 * into the display fields a canvas node card shows — a role/tech kind label, a
 * glyph, a status-dot colour, and a rich one-line summary.
 *
 * No React imports; unit-testable in isolation. Heuristic, not authoritative —
 * the backend has no "kind" field, so we infer one from the image + ports.
 */

/** Glyph identifiers a node card can render (mapped to icons by `node-glyph`). */
export type GlyphKind = "web" | "postgres" | "redis" | "database" | "object" | "worker" | "service";

/** Status-dot colour buckets (semantic tokens: ok=green, warn=amber, err=red). */
export type DotState = "ok" | "warn" | "err";

export interface PresentationPort {
  number?: number;
  protocol?: string;
  exposedToPublic?: boolean;
}

export interface PresentationInput {
  /** Managed addon (e.g. postgres) rather than a stack resource. */
  isAddon: boolean;
  /** Container image reference, e.g. "redis:6.2" or "org/app:latest". */
  image?: string;
  /** True when the resource builds from git instead of a prebuilt image. */
  hasBuild?: boolean;
  ports?: PresentationPort[];
}

export interface NodePresentation {
  kindLabel: string;
  glyph: GlyphKind;
  dotState: DotState;
  summary: string;
}

/** Internal kind keys → their display metadata. */
type KindKey = "web" | "postgres" | "redis" | "mysql" | "mongo" | "object" | "service";

const KIND_META: Record<KindKey, { label: string; glyph: GlyphKind; dot: DotState }> = {
  web: { label: "Web", glyph: "web", dot: "ok" },
  postgres: { label: "Postgres", glyph: "postgres", dot: "ok" },
  redis: { label: "Redis", glyph: "redis", dot: "ok" },
  mysql: { label: "MySQL", glyph: "database", dot: "ok" },
  mongo: { label: "Mongo", glyph: "database", dot: "ok" },
  object: { label: "Object", glyph: "object", dot: "warn" },
  service: { label: "Service", glyph: "service", dot: "ok" },
};

/** Image-substring → kind. First match wins; order matters little (disjoint). */
const IMAGE_KINDS: { match: RegExp; kind: KindKey }[] = [
  { match: /redis/, kind: "redis" },
  { match: /postgres/, kind: "postgres" },
  { match: /mariadb|mysql/, kind: "mysql" },
  { match: /mongo/, kind: "mongo" },
  { match: /minio/, kind: "object" },
];

/** Split "registry/org/name:tag" into its base name and tag. */
function imageParts(image: string): { base: string; tag?: string } {
  const lastSlash = image.lastIndexOf("/");
  const afterSlash = lastSlash >= 0 ? image.slice(lastSlash + 1) : image;
  const colon = afterSlash.lastIndexOf(":");
  if (colon > 0) return { base: afterSlash.slice(0, colon), tag: afterSlash.slice(colon + 1) };
  return { base: afterSlash };
}

/** The port to surface: a public one wins, else the first declared. */
function primaryPort(ports?: PresentationPort[]): PresentationPort | undefined {
  if (!ports?.length) return undefined;
  return ports.find((p) => p.exposedToPublic) ?? ports[0];
}

function detectKind(image: string, isPublic: boolean): KindKey {
  const s = image.toLowerCase();
  for (const { match, kind } of IMAGE_KINDS) if (match.test(s)) return kind;
  // A generic service that exposes a public port reads as "Web"; otherwise it's
  // an internal "Service".
  return isPublic ? "web" : "service";
}

function buildSummary(
  kind: KindKey,
  image: string,
  hasBuild: boolean | undefined,
  port: PresentationPort | undefined,
): string {
  if (kind === "redis") return `${image || "redis"} · in-memory`;
  if (kind === "object") return `${image || "minio"} · S3-compatible`;
  if (kind === "web") {
    const base = imageParts(image).base || (hasBuild ? "git build" : "service");
    const num = port?.number;
    const access = port?.exposedToPublic ? "public" : "internal";
    return num ? `${base} · :${num} · ${access}` : base;
  }
  // postgres / mysql / mongo / generic service
  if (image) return image;
  if (hasBuild) return "git build";
  return "service";
}

export function nodePresentation(input: PresentationInput): NodePresentation {
  if (input.isAddon) {
    return { kindLabel: "Postgres", glyph: "postgres", dotState: "ok", summary: "managed postgres" };
  }
  const image = (input.image ?? "").trim();
  const port = primaryPort(input.ports);
  const isPublic = !!input.ports?.some((p) => p.exposedToPublic);
  const kind = detectKind(image, isPublic);
  const meta = KIND_META[kind];
  return {
    kindLabel: meta.label,
    glyph: meta.glyph,
    dotState: meta.dot,
    summary: buildSummary(kind, image, input.hasBuild, port),
  };
}

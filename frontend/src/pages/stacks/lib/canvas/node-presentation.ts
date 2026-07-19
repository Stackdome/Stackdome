/**
 * Pure presentation mapper: turns a resource's raw config (image, ports, build)
 * into the display fields a canvas node card shows — a role/tech kind label, a
 * glyph, and a rich one-line summary. The status dot is computed live from
 * runtime state (see `graph-from-connections.ts`'s use of `statusVariant`),
 * not derived here.
 *
 * No React imports; unit-testable in isolation. Heuristic, not authoritative —
 * the backend has no "kind" field, so we infer one from the image + ports.
 */

/** Glyph identifiers a node card can render (mapped to icons by `node-glyph`). */
export type GlyphKind = "web" | "postgres" | "redis" | "database" | "object" | "worker" | "service";

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
  /** Brand-icon registry key when the image maps to known software (see
   *  `components/branded/brand-icons`); absent → generic Lucide glyph. */
  brandSlug?: string;
  /** First card line: `image[:tag]` (registry/org stripped), or "git build"/"service". */
  summary: string;
  /** Second card line when a port is declared: `port N · public|internal`. */
  detail?: string;
}

/** Internal kind keys → their display metadata. */
type KindKey = "web" | "postgres" | "redis" | "mysql" | "mongo" | "object" | "service";

const KIND_META: Record<KindKey, { label: string; glyph: GlyphKind }> = {
  web: { label: "Web", glyph: "web" },
  postgres: { label: "Postgres", glyph: "postgres" },
  redis: { label: "Redis", glyph: "redis" },
  mysql: { label: "MySQL", glyph: "database" },
  mongo: { label: "Mongo", glyph: "database" },
  object: { label: "Object", glyph: "object" },
  service: { label: "Service", glyph: "service" },
};

/** Image-substring → kind. First match wins; order matters little (disjoint).
 *  `postgres(?!t)` keeps PostgREST (and friends) from reading as a database. */
const IMAGE_KINDS: { match: RegExp; kind: KindKey }[] = [
  { match: /redis/, kind: "redis" },
  { match: /postgres(?!t)/, kind: "postgres" },
  { match: /mariadb|mysql/, kind: "mysql" },
  { match: /mongo/, kind: "mongo" },
  { match: /minio/, kind: "object" },
];

/** Image-substring → brand-icon slug. First match wins — keep more specific
 *  patterns first (otel images come from grafana/, so otel precedes grafana).
 *  Expand alongside `BRAND_ICONS` as the icon set grows. */
const BRAND_MATCHERS: { match: RegExp; slug: string }[] = [
  { match: /otel|opentelemetry/, slug: "opentelemetry" },
  { match: /tooljet/, slug: "tooljet" },
  { match: /grafana/, slug: "grafana" },
  { match: /redis/, slug: "redis" },
  { match: /postgrest/, slug: "postgrest" },
  { match: /postgres(?!t)/, slug: "postgres" },
  { match: /mariadb/, slug: "mariadb" },
  { match: /mysql/, slug: "mysql" },
  { match: /mongo/, slug: "mongo" },
  { match: /minio/, slug: "minio" },
  { match: /clickhouse/, slug: "clickhouse" },
  { match: /elasticsearch/, slug: "elasticsearch" },
  { match: /couchdb/, slug: "couchdb" },
  { match: /influxdb/, slug: "influxdb" },
  { match: /mssql/, slug: "mssql" },
];

function detectBrandSlug(image: string): string | undefined {
  const s = image.toLowerCase();
  for (const { match, slug } of BRAND_MATCHERS) if (match.test(s)) return slug;
  return undefined;
}

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

/** First card line: `image[:tag]` (registry/org stripped); no image → "git build"/"service". */
function buildSummary(image: string, hasBuild: boolean | undefined): string {
  const { base, tag } = imageParts(image);
  return base ? (tag ? `${base}:${tag}` : base) : hasBuild ? "git build" : "service";
}

/** Second card line, only when a port is declared: `port N · public|internal`. */
function buildDetail(port: PresentationPort | undefined): string | undefined {
  if (port?.number == null) return undefined;
  return `port ${port.number} · ${port.exposedToPublic ? "public" : "internal"}`;
}

export function nodePresentation(input: PresentationInput): NodePresentation {
  if (input.isAddon) {
    return { kindLabel: "Postgres", glyph: "postgres", brandSlug: "postgres", summary: "managed postgres" };
  }
  const image = (input.image ?? "").trim();
  const port = primaryPort(input.ports);
  const isPublic = !!input.ports?.some((p) => p.exposedToPublic);
  const kind = detectKind(image, isPublic);
  const meta = KIND_META[kind];
  return {
    kindLabel: meta.label,
    glyph: meta.glyph,
    brandSlug: detectBrandSlug(image),
    summary: buildSummary(image, input.hasBuild),
    detail: buildDetail(port),
  };
}

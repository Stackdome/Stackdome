import { Box, Cloud, Cpu, Database, Globe, Zap, type LucideIcon } from "lucide-react";
import type { GlyphKind } from "@/pages/stacks/lib/canvas/node-presentation";
import { BrandIcon } from "@/components/branded/brand-icons";
import { hasBrandIcon } from "@/components/branded/brand-icon-registry";

const GLYPH_BY_KIND: Record<GlyphKind, LucideIcon> = {
  web: Globe,
  postgres: Database,
  redis: Zap,
  database: Database,
  object: Cloud,
  worker: Cpu,
  service: Box,
};

/**
 * Canvas node icon: the software's brand logo when the image resolved to a
 * registered brand slug, otherwise a generic Lucide glyph for the inferred
 * kind (globe/box/database…).
 */
export function NodeGlyph({
  glyph,
  brandSlug,
  className,
  size = 15,
}: {
  glyph: GlyphKind;
  brandSlug?: string;
  className?: string;
  size?: number;
}) {
  if (hasBrandIcon(brandSlug)) return <BrandIcon slug={brandSlug} size={size} className={className} />;
  const Icon = GLYPH_BY_KIND[glyph] ?? Box;
  return <Icon className={className} aria-hidden />;
}

import { Box, Cloud, Cpu, Database, Globe, Zap, type LucideIcon } from "lucide-react";
import type { GlyphKind } from "@/pages/stacks/lib/canvas/node-presentation";

const GLYPH_BY_KIND: Record<GlyphKind, LucideIcon> = {
  web: Globe,
  postgres: Database,
  redis: Zap,
  database: Database,
  object: Cloud,
  worker: Cpu,
  service: Box,
};

/** Lucide glyph for a canvas node's inferred kind. */
export function NodeGlyph({ glyph, className }: { glyph: GlyphKind; className?: string }) {
  const Icon = GLYPH_BY_KIND[glyph] ?? Box;
  return <Icon className={className} aria-hidden />;
}

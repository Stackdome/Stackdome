import { Box, Database } from "lucide-react";
import { NODE_KIND, type NodeKind } from "@/pages/stacks/lib/canvas/derive-graph";

const GLYPH_BY_KIND = {
  [NODE_KIND.service]: Box,
  [NODE_KIND.addon]: Database,
} as const;

/** Lucide glyph for a canvas node kind. */
export function NodeGlyph({ kind, className }: { kind: NodeKind; className?: string }) {
  const Icon = GLYPH_BY_KIND[kind];
  return <Icon className={className} aria-hidden />;
}

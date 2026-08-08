import { Globe, Database, Zap, Box, type LucideIcon } from "lucide-react";
import { BrandIcon } from "@/components/branded/brand-icons";
import { hasBrandIcon } from "@/components/branded/brand-icon-registry";

const LUCIDE: Record<string, LucideIcon> = { globe: Globe, database: Database, zap: Zap, box: Box };

/** Wizard block icon: registered brand logo when available, Lucide otherwise. */
export function BlockGlyph({ icon, size = 18 }: { icon: string; size?: number }) {
  if (hasBrandIcon(icon)) return <BrandIcon slug={icon} size={size} />;
  const Icon = LUCIDE[icon] ?? Box;
  return <Icon style={{ width: size, height: size }} />;
}

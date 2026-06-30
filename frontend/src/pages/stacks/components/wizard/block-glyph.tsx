import { Globe, Database, Zap, Box, type LucideIcon } from "lucide-react";

const ICONS: Record<string, LucideIcon> = { globe: Globe, database: Database, zap: Zap, box: Box };

export function BlockGlyph({ icon, size = 18 }: { icon: string; size?: number }) {
  const Icon = ICONS[icon] ?? Box;
  return <Icon style={{ width: size, height: size }} />;
}

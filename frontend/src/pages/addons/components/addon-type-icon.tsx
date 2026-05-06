import { Puzzle } from "lucide-react";
import postgresqlIconUrl from "@/assets/addons/postgresql.svg";
import redisIconUrl from "@/assets/addons/redis.svg";
import ollamaIconUrl from "@/assets/addons/ollama.svg";

// Add new addon types here as they ship. Keys mirror the API's addon-type
// discriminator strings; values are URLs to brand SVGs in src/assets/addons/.
const BRAND_ICONS: Record<string, string> = {
  postgres: postgresqlIconUrl,
  redis: redisIconUrl,
  ollama: ollamaIconUrl,
};

interface AddonTypeIconProps {
  /** Addon type id (e.g. "postgres", "redis") */
  type: string;
  className?: string;
  /** Pixel size of the icon — defaults to 16. */
  size?: number;
}

/**
 * Brand icon for an addon type. Falls back to the generic addon Puzzle glyph
 * for types we don't have a brand SVG for yet.
 */
export function AddonTypeIcon({ type, className, size = 16 }: AddonTypeIconProps) {
  const url = BRAND_ICONS[type];
  if (url) {
    return (
      <img
        src={url}
        alt={`${type} logo`}
        width={size}
        height={size}
        className={className}
        aria-hidden
      />
    );
  }
  return <Puzzle className={className} style={{ width: size, height: size }} />;
}
